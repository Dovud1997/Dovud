package worker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/modules/identity/domain"
	"github.com/Dovud1997/Dovud/backend/internal/platform/notify"
	"github.com/Dovud1997/Dovud/backend/internal/worker"
)

func TestCollectPushTokensSkipsStubsAndDedupes(t *testing.T) {
	a, b, stub, empty := "tok-a", "tok-b", "stub-push-xyz", "  "
	devices := []domain.UserDevice{
		{PushToken: &a},
		{PushToken: &stub},
		{PushToken: nil},
		{PushToken: &empty},
		{PushToken: &a},
		{PushToken: &b},
	}
	got := worker.CollectPushTokens(devices)
	if len(got) != 2 || got[0] != "tok-a" || got[1] != "tok-b" {
		t.Fatalf("unexpected tokens: %#v", got)
	}
}

type fakeSender struct {
	calls []string
	fail  map[string]bool
}

func (f *fakeSender) Send(ctx context.Context, channel string, msg notify.Message) error {
	_ = ctx
	_ = channel
	f.calls = append(f.calls, msg.To)
	if f.fail[msg.To] {
		return errors.New("boom")
	}
	return nil
}

func TestFanoutPushAllSuccess(t *testing.T) {
	s := &fakeSender{}
	sent, failed, err := worker.FanoutPush(context.Background(), s, []string{"t1", "t2"}, notify.Message{
		Subject: "Hi", Body: "Body",
	})
	if err != nil || sent != 2 || failed != 0 {
		t.Fatalf("sent=%d failed=%d err=%v", sent, failed, err)
	}
	if len(s.calls) != 2 {
		t.Fatalf("calls=%v", s.calls)
	}
}

func TestFanoutPushPartialAndAllFail(t *testing.T) {
	s := &fakeSender{fail: map[string]bool{"bad": true}}
	sent, failed, err := worker.FanoutPush(context.Background(), s, []string{"ok", "bad"}, notify.Message{})
	if err != nil || sent != 1 || failed != 1 {
		t.Fatalf("partial: sent=%d failed=%d err=%v", sent, failed, err)
	}

	s2 := &fakeSender{fail: map[string]bool{"a": true, "b": true}}
	sent, failed, err = worker.FanoutPush(context.Background(), s2, []string{"a", "b"}, notify.Message{})
	if err == nil || sent != 0 || failed != 2 {
		t.Fatalf("all fail: sent=%d failed=%d err=%v", sent, failed, err)
	}
}

func TestFanoutPushEmpty(t *testing.T) {
	s := &fakeSender{}
	sent, failed, err := worker.FanoutPush(context.Background(), s, nil, notify.Message{})
	if err != nil || sent != 0 || failed != 0 || len(s.calls) != 0 {
		t.Fatalf("empty fanout should no-op")
	}
}
