package ws_test

import (
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/platform/ws"
	"github.com/google/uuid"
)

func TestHubPublishDoesNotPanicWithoutClients(t *testing.T) {
	hub := ws.NewHub(nil)
	hub.BroadcastSyncInvalidate(uuid.New(), "order", "customer")
	hub.Publish(uuid.New(), "notification.created", map[string]any{"id": "1"})
}
