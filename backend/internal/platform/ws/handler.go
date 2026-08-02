package ws

import (
	"strings"

	"github.com/Dovud1997/Dovud/backend/internal/platform/auth"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Register mounts `/ws/v1` with JWT query-token auth.
func Register(app *fiber.App, hub *Hub, tokens *auth.TokenService) {
	app.Use("/ws/v1", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/v1", websocket.New(func(conn *websocket.Conn) {
		token := strings.TrimSpace(conn.Query("token"))
		if token == "" {
			authz := conn.Headers("Authorization")
			if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
				token = strings.TrimSpace(authz[7:])
			}
		}
		if token == "" || tokens == nil {
			_ = conn.WriteJSON(map[string]any{"type": "error", "payload": map[string]any{"message": "unauthorized"}})
			_ = conn.Close()
			return
		}
		claims, err := tokens.ParseAccessToken(token)
		if err != nil {
			_ = conn.WriteJSON(map[string]any{"type": "error", "payload": map[string]any{"message": "invalid token"}})
			_ = conn.Close()
			return
		}
		client := hub.Register(conn, claims.TenantID, claims.UserID)
		defer hub.Unregister(client)

		_ = conn.WriteJSON(map[string]any{
			"type": "ready",
			"payload": map[string]any{
				"tenant_id": claims.TenantID.String(),
				"user_id":   claims.UserID.String(),
			},
		})

		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if mt != websocket.TextMessage {
				continue
			}
			text := strings.TrimSpace(string(msg))
			switch {
			case text == "ping" || strings.Contains(text, `"ping"`):
				_ = conn.WriteJSON(map[string]any{"type": "pong", "ts": uuid.New().String()})
			default:
				// Ignore client events for now (server-push channel).
			}
		}
	}))
}
