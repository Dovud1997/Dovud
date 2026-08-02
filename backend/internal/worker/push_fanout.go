package worker

import (
	"context"
	"fmt"
	"strings"

	"github.com/Dovud1997/Dovud/backend/internal/modules/identity/domain"
	"github.com/Dovud1997/Dovud/backend/internal/platform/notify"
)

// CollectPushTokens returns non-empty, non-stub device push tokens.
func CollectPushTokens(devices []domain.UserDevice) []string {
	out := make([]string, 0, len(devices))
	seen := map[string]struct{}{}
	for _, d := range devices {
		if d.PushToken == nil {
			continue
		}
		tok := strings.TrimSpace(*d.PushToken)
		if tok == "" || strings.HasPrefix(tok, "stub-push-") {
			continue
		}
		if _, ok := seen[tok]; ok {
			continue
		}
		seen[tok] = struct{}{}
		out = append(out, tok)
	}
	return out
}

// PushSender is satisfied by *notify.Router.
type PushSender interface {
	Send(ctx context.Context, channel string, msg notify.Message) error
}

// FanoutPush sends the same push payload to every token.
// Returns counts of successful and failed sends. lastErr is the last failure (if any).
func FanoutPush(ctx context.Context, sender PushSender, tokens []string, base notify.Message) (sent, failed int, lastErr error) {
	if len(tokens) == 0 {
		return 0, 0, nil
	}
	for _, tok := range tokens {
		msg := base
		msg.To = tok
		msg.Channel = "push"
		if err := sender.Send(ctx, "push", msg); err != nil {
			failed++
			lastErr = err
			continue
		}
		sent++
	}
	if sent == 0 && failed > 0 {
		return sent, failed, fmt.Errorf("push fan-out failed for all %d devices: %w", failed, lastErr)
	}
	return sent, failed, nil
}
