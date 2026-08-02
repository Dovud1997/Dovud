package ws

import (
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
)

// Event is a server→client WebSocket payload.
type Event struct {
	Type      string         `json:"type"`
	TenantID  string         `json:"tenant_id,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	Timestamp time.Time      `json:"ts"`
}

type client struct {
	conn     *websocket.Conn
	tenantID uuid.UUID
	userID   uuid.UUID
}

// Hub fans out live events to tenants.
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
	log     *slog.Logger
}

func NewHub(log *slog.Logger) *Hub {
	if log == nil {
		log = slog.Default()
	}
	return &Hub{clients: map[*client]struct{}{}, log: log}
}

func (h *Hub) Register(conn *websocket.Conn, tenantID, userID uuid.UUID) *client {
	c := &client{conn: conn, tenantID: tenantID, userID: userID}
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
	h.log.Info("ws connected", "tenant_id", tenantID, "user_id", userID, "peers", len(h.clients))
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

// Publish sends an event to all connections of a tenant.
func (h *Hub) Publish(tenantID uuid.UUID, eventType string, payload map[string]any) {
	evt := Event{
		Type: eventType, TenantID: tenantID.String(), Payload: payload, Timestamp: time.Now().UTC(),
	}
	raw, err := json.Marshal(evt)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		if c.tenantID != tenantID {
			continue
		}
		_ = c.conn.WriteMessage(websocket.TextMessage, raw)
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
