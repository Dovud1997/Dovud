package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const redisWSChannel = "sfa:ws:events"

// Event is a server→client WebSocket payload.
type Event struct {
	Type      string         `json:"type"`
	TenantID  string         `json:"tenant_id,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	Timestamp time.Time      `json:"ts"`
}

type busEnvelope struct {
	Origin   string `json:"origin"`
	TenantID string `json:"tenant_id"`
	Event    Event  `json:"event"`
}

// EventBus fans WS events across API instances (Redis Pub/Sub or in-process).
type EventBus interface {
	Publish(ctx context.Context, payload []byte) error
	Subscribe(ctx context.Context) (<-chan []byte, func(), error)
}

type client struct {
	conn     *websocket.Conn
	tenantID uuid.UUID
	userID   uuid.UUID
}

// Hub fans out live events to tenants (local sockets + optional cross-instance bus).
type Hub struct {
	mu         sync.RWMutex
	clients    map[*client]struct{}
	log        *slog.Logger
	bus        EventBus
	instanceID string
	stopListen context.CancelFunc
}

func NewHub(log *slog.Logger) *Hub {
	if log == nil {
		log = slog.Default()
	}
	return &Hub{
		clients:    map[*client]struct{}{},
		log:        log,
		instanceID: uuid.NewString(),
	}
}

// WithBus attaches a cross-instance event bus (typically Redis Pub/Sub).
func (h *Hub) WithBus(bus EventBus) *Hub {
	if bus == nil {
		return h
	}
	h.Close()
	h.bus = bus
	ctx, cancel := context.WithCancel(context.Background())
	h.stopListen = cancel
	go h.listenBus(ctx)
	h.log.Info("ws event bus attached", "instance_id", h.instanceID)
	return h
}

// WithRedis is a convenience wrapper around RedisEventBus.
func (h *Hub) WithRedis(rdb *redis.Client) *Hub {
	if rdb == nil {
		return h
	}
	return h.WithBus(NewRedisEventBus(rdb))
}

// Close stops the bus subscriber.
func (h *Hub) Close() {
	if h.stopListen != nil {
		h.stopListen()
		h.stopListen = nil
	}
}

func (h *Hub) Register(conn *websocket.Conn, tenantID, userID uuid.UUID) *client {
	c := &client{conn: conn, tenantID: tenantID, userID: userID}
	h.mu.Lock()
	h.clients[c] = struct{}{}
	n := len(h.clients)
	h.mu.Unlock()
	h.log.Info("ws connected", "tenant_id", tenantID, "user_id", userID, "peers", n)
	return c
}

func (h *Hub) Unregister(c *client) {
	if c == nil {
		return
	}
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	_ = c.conn.Close()
}

// Publish sends an event to local tenant sockets and fans out via the bus.
func (h *Hub) Publish(tenantID uuid.UUID, eventType string, payload map[string]any) {
	evt := Event{
		Type: eventType, TenantID: tenantID.String(), Payload: payload, Timestamp: time.Now().UTC(),
	}
	raw, err := json.Marshal(evt)
	if err != nil {
		return
	}
	h.deliverLocal(tenantID, raw)
	if h.bus == nil {
		return
	}
	env, err := json.Marshal(busEnvelope{
		Origin: h.instanceID, TenantID: tenantID.String(), Event: evt,
	})
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := h.bus.Publish(ctx, env); err != nil {
		h.log.Warn("ws bus publish failed", "error", err, "type", eventType)
	}
}

// DeliverFromBus applies a remote bus envelope to local sockets only (no re-publish).
func (h *Hub) DeliverFromBus(raw []byte) {
	var env busEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return
	}
	if env.Origin == "" || env.Origin == h.instanceID {
		return
	}
	tenantID, err := uuid.Parse(env.TenantID)
	if err != nil {
		return
	}
	payload, err := json.Marshal(env.Event)
	if err != nil {
		return
	}
	h.deliverLocal(tenantID, payload)
}

func (h *Hub) deliverLocal(tenantID uuid.UUID, raw []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		if c.tenantID != tenantID {
			continue
		}
		_ = c.conn.WriteMessage(websocket.TextMessage, raw)
	}
}

func (h *Hub) listenBus(ctx context.Context) {
	ch, unsub, err := h.bus.Subscribe(ctx)
	if err != nil {
		h.log.Warn("ws bus subscribe failed", "error", err)
		return
	}
	defer unsub()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			h.DeliverFromBus(msg)
		}
	}
}

// BroadcastSyncInvalidate notifies tenants that entity types changed.
func (h *Hub) BroadcastSyncInvalidate(tenantID uuid.UUID, entityTypes ...string) {
	types := entityTypes
	if len(types) == 0 {
		types = []string{"*"}
	}
	h.Publish(tenantID, "sync.invalidate", map[string]any{"entity_types": types})
}

// InstanceID is the unique id used to ignore self-originated bus messages.
func (h *Hub) InstanceID() string { return h.instanceID }
