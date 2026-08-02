package worker

import (
	"context"
	"fmt"
	"strings"

	"github.com/Dovud1997/Dovud/backend/internal/modules/identity/domain"
	"github.com/Dovud1997/Dovud/backend/internal/platform/notify"
)

// PushTarget is a usable device push endpoint.
type PushTarget struct {
	DeviceID string
	Platform string
	Token    string
}

// PushResult is the outcome of one fan-out send.
type PushResult struct {
	Target PushTarget
	Err    error
}

// CollectPushTargets returns non-empty, non-stub device push targets (deduped by token).
func CollectPushTargets(devices []domain.UserDevice) []PushTarget {
	out := make([]PushTarget, 0, len(devices))
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
		out = append(out, PushTarget{
			DeviceID: d.DeviceID,
			Platform: d.Platform,
			Token:    tok,
		})
	}
	return out
}

// CollectPushTokens keeps the token-only helper for callers that do not need device metadata.
func CollectPushTokens(devices []domain.UserDevice) []string {
	targets := CollectPushTargets(devices)
	out := make([]string, 0, len(targets))
	for _, t := range targets {
		out = append(out, t.Token)
	}
	return out
}

// TokenSuffix returns a short non-secret suffix for diagnostics.
func TokenSuffix(token string) string {
	token = strings.TrimSpace(token)
	if len(token) <= 8 {
		return token
	}
	return token[len(token)-8:]
}

// PushSender is satisfied by *notify.Router.
type PushSender interface {
	Send(ctx context.Context, channel string, msg notify.Message) error
}

// FanoutPush sends the same push payload to every target.
func FanoutPush(ctx context.Context, sender PushSender, targets []PushTarget, base notify.Message) (results []PushResult, sent, failed int, lastErr error) {
	if len(targets) == 0 {
		return nil, 0, 0, nil
	}
	results = make([]PushResult, 0, len(targets))
	for _, t := range targets {
		msg := base
		msg.To = t.Token
		msg.Channel = "push"
		err := sender.Send(ctx, "push", msg)
		results = append(results, PushResult{Target: t, Err: err})
		if err != nil {
			failed++
			lastErr = err
			continue
		}
		sent++
	}
	if sent == 0 && failed > 0 {
		return results, sent, failed, fmt.Errorf("push fan-out failed for all %d devices: %w", failed, lastErr)
	}
	return results, sent, failed, nil
}
