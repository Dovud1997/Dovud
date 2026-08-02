package ws_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/platform/ws"
	"github.com/google/uuid"
)

func TestHubPublishDoesNotPanicWithoutClients(t *testing.T) {
	hub := ws.NewHub(nil)
	hub.BroadcastSyncInvalidate(uuid.New(), "order", "customer")
	hub.Publish(uuid.New(), "notification.created", map[string]any{"id": "1"})
}

func TestMemoryBusFansOutAcrossHubs(t *testing.T) {
	bus := ws.NewMemoryEventBus()
	a := ws.NewHub(nil).WithBus(bus)
	b := ws.NewHub(nil).WithBus(bus)
	defer a.Close()
	defer b.Close()

	tenant := uuid.New()
	env, err := json.Marshal(struct {
		Origin   string   `json:"origin"`
		TenantID string   `json:"tenant_id"`
		Event    ws.Event `json:"event"`
	}{
		Origin: a.InstanceID(), TenantID: tenant.String(),
		Event: ws.Event{
			Type: "order.updated", TenantID: tenant.String(),
			Payload: map[string]any{"id": "o1"}, Timestamp: time.Now().UTC(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	b.DeliverFromBus(env)

	self, _ := json.Marshal(struct {
		Origin   string   `json:"origin"`
		TenantID string   `json:"tenant_id"`
		Event    ws.Event `json:"event"`
	}{
		Origin: b.InstanceID(), TenantID: tenant.String(),
		Event: ws.Event{Type: "order.updated", TenantID: tenant.String()},
	})
	b.DeliverFromBus(self)

	done := make(chan struct{}, 1)
	ch, unsub, err := bus.Subscribe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer unsub()
	go func() {
		select {
		case <-ch:
			done <- struct{}{}
		case <-time.After(2 * time.Second):
		}
	}()
	a.Publish(tenant, "visit.updated", map[string]any{"id": "v1"})
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("expected bus publish")
	}
}
