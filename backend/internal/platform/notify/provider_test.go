package notify_test

import (
	"context"
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
