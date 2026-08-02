package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Dovud1997/Dovud/backend/internal/platform/config"
	"github.com/Dovud1997/Dovud/backend/internal/platform/notify"
)

func TestFileEmailProvider(t *testing.T) {
	dir := t.TempDir()
	r := notify.NewRouter(config.NotifyConfig{Email: config.EmailConfig{Driver: "file", FileDir: dir}}, nil)
	if err := r.Send(context.Background(), "email", notify.Message{
		To: "smoke@demo.local", Subject: "P6", Body: "hello",
	}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 mail file, got %d", len(entries))
	}
	raw, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "Subject: P6") {
		t.Fatalf("missing subject in %q", string(raw))
	}
}

func TestHTTPWebhookSMS(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	router := notify.NewRouter(config.NotifyConfig{
		SMS: config.SMSConfig{Driver: "http", WebhookURL: srv.URL},
	}, nil)
	if err := router.Send(context.Background(), "sms", notify.Message{To: "+99890", Body: "hi"}); err != nil {
		t.Fatal(err)
	}
	if got["to"] != "+99890" || got["body"] != "hi" {
		t.Fatalf("unexpected payload %#v", got)
	}
}
