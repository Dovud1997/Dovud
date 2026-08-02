package notify_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
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

func TestFCMPushProvider(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})

	var gotAuth string
	var gotBody map[string]any
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("grant_type") == "" || r.Form.Get("assertion") == "" {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-access", "expires_in": 3600, "token_type": "Bearer",
		})
	})
	mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"projects/demo/messages/1"}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	sa, _ := json.Marshal(map[string]string{
		"client_email": "fcm@demo.iam.gserviceaccount.com",
		"private_key":  string(pemKey),
		"project_id":   "demo-project",
		"token_uri":    srv.URL + "/token",
	})
	p, err := notify.NewFCMPush(config.PushConfig{
		Driver: "fcm", ProjectID: "demo-project", CredentialsJSON: string(sa),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	p.WithEndpoints(srv.URL+"/token", srv.URL+"/send").WithHTTPClient(srv.Client())

	if err := p.Send(context.Background(), notify.Message{
		To: "device-token-1", Subject: "Hello", Body: "World",
		Data: map[string]any{"order_id": "42"},
	}); err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer test-access" {
		t.Fatalf("auth header: %q", gotAuth)
	}
	msg, _ := gotBody["message"].(map[string]any)
	if msg["token"] != "device-token-1" {
		t.Fatalf("payload %#v", gotBody)
	}
	notif, _ := msg["notification"].(map[string]any)
	if notif["title"] != "Hello" || notif["body"] != "World" {
		t.Fatalf("notification %#v", notif)
	}
	data, _ := msg["data"].(map[string]any)
	if data["order_id"] != "42" {
		t.Fatalf("data %#v", data)
	}

	// stub tokens are no-ops
	if err := p.Send(context.Background(), notify.Message{To: "stub-push-abc"}); err != nil {
		t.Fatal(err)
	}
}

func TestFCMPushRequiresCredentials(t *testing.T) {
	if _, err := notify.NewFCMPush(config.PushConfig{Driver: "fcm"}, nil); err == nil {
		t.Fatal("expected credentials error")
	}
	router := notify.NewRouter(config.NotifyConfig{Push: config.PushConfig{Driver: "fcm"}}, nil)
	if router.Push.Name() != "log-push" {
		t.Fatalf("expected log fallback, got %s", router.Push.Name())
	}
}
