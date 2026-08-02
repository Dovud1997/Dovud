package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/platform/auth"
)

func TestMemoryLoginGuardLockout(t *testing.T) {
	g := auth.NewMemoryLoginGuard(auth.LockoutConfig{
		MaxAttempts: 3,
		Window:      time.Minute,
		LockFor:     time.Minute,
	})
	ctx := context.Background()
	key := "demo:bad@demo.local"
	for i := 0; i < 3; i++ {
		if err := g.Fail(ctx, key); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.Check(ctx, key); err == nil {
		t.Fatal("expected lockout")
	}
	if err := g.Success(ctx, key); err != nil {
		t.Fatal(err)
	}
	if err := g.Check(ctx, key); err != nil {
		t.Fatalf("expected clear after success: %v", err)
	}
}
