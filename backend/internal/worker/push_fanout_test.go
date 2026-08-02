package worker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/modules/identity/domain"
	"github.com/Dovud1997/Dovud/backend/internal/platform/notify"
	"github.com/Dovud1997/Dovud/backend/internal/worker"
)

func TestCollectPushTargetsSkipsStubsAndDedupes(t *testing.T) {
	a, b, stub, empty := "tok-a", "tok-b", "stub-push-xyz", "  "
	devices := []domain.UserDevice{
		{DeviceID: "d1", Platform: "android", PushToken: &a},
		{DeviceID: "d2", Platform: "ios", PushToken: &stub},
		{PushToken: nil},
		{DeviceID: "d3", PushToken: &empty},
		{DeviceID: "d4", Platform: "web", PushToken: &a},
		{DeviceID: "d5", Platform: "android", PushToken: &b},
	}
	got := worker.CollectPushTargets(devices)
	if len(got) != 2 || got[0].Token != "tok-a" || got[0].DeviceID != "d1" || got[1].Token != "tok-b" {
		t.Fatalf("unexpected targets: %#v", got)
	}
	if worker.TokenSuffix("abcdefghij") != "cdefghij" {
		t.Fatalf("suffix=%q", worker.TokenSuffix("abcdefghij"))
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
	results, sent, failed, err := worker.FanoutPush(context.Background(), s, []worker.PushTarget{
		{DeviceID: "1", Token: "t1"},
		{DeviceID: "2", Token: "t2"},
	}, notify.Message{Subject: "Hi", Body: "Body"})
	if err != nil || sent != 2 || failed != 0 || len(results) != 2 {
		t.Fatalf("sent=%d failed=%d err=%v results=%d", sent, failed, err, len(results))
	}
}

func TestFanoutPushPartialAndAllFail(t *testing.T) {
	s := &fakeSender{fail: map[string]bool{"bad": true}}
	results, sent, failed, err := worker.FanoutPush(context.Background(), s, []worker.PushTarget{
		{DeviceID: "ok", Token: "ok"},
		{DeviceID: "bad", Token: "bad"},
	}, notify.Message{})
	if err != nil || sent != 1 || failed != 1 || results[1].Err == nil {
		t.Fatalf("partial: sent=%d failed=%d err=%v", sent, failed, err)
	}

	s2 := &fakeSender{fail: map[string]bool{"a": true, "b": true}}
	_, sent, failed, err = worker.FanoutPush(context.Background(), s2, []worker.PushTarget{
		{Token: "a"}, {Token: "b"},
	}, notify.Message{})
	if err == nil || sent != 0 || failed != 2 {
		t.Fatalf("all fail: sent=%d failed=%d err=%v", sent, failed, err)
	}
}

func TestFanoutPushEmpty(t *testing.T) {
	s := &fakeSender{}
	results, sent, failed, err := worker.FanoutPush(context.Background(), s, nil, notify.Message{})
	if err != nil || sent != 0 || failed != 0 || len(results) != 0 || len(s.calls) != 0 {
		t.Fatalf("empty fanout should no-op")
	}
}
