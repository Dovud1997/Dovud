package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Dovud1997/Dovud/backend/internal/platform/redisx"
	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
)

// LoginGuard protects against brute-force password guessing.
type LoginGuard interface {
	Check(ctx context.Context, key string) error
	Fail(ctx context.Context, key string) error
	Success(ctx context.Context, key string) error
}

type LockoutConfig struct {
	MaxAttempts int
	Window      time.Duration
	LockFor     time.Duration
}

func (c LockoutConfig) withDefaults() LockoutConfig {
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 5
	}
	if c.Window <= 0 {
		c.Window = 15 * time.Minute
	}
	if c.LockFor <= 0 {
		c.LockFor = 15 * time.Minute
	}
	return c
}

type MemoryLoginGuard struct {
	cfg LockoutConfig
	mu  sync.Mutex
	m   map[string]*attemptState
}

type attemptState struct {
	fails    int
	windowAt time.Time
	lockedUntil time.Time
}

func NewMemoryLoginGuard(cfg LockoutConfig) *MemoryLoginGuard {
	return &MemoryLoginGuard{cfg: cfg.withDefaults(), m: map[string]*attemptState{}}
}

func (g *MemoryLoginGuard) Check(ctx context.Context, key string) error {
	_ = ctx
	g.mu.Lock()
	defer g.mu.Unlock()
	st := g.m[key]
	if st == nil {
		return nil
	}
	if time.Now().Before(st.lockedUntil) {
		return apperrors.New("AUTH_LOCKED", "Too many failed login attempts; try again later", 429)
	}
	return nil
}

func (g *MemoryLoginGuard) Fail(ctx context.Context, key string) error {
	_ = ctx
	cfg := g.cfg
	g.mu.Lock()
	defer g.mu.Unlock()
	now := time.Now()
	st := g.m[key]
	if st == nil || now.After(st.windowAt) {
		st = &attemptState{windowAt: now.Add(cfg.Window)}
		g.m[key] = st
	}
	st.fails++
	if st.fails >= cfg.MaxAttempts {
		st.lockedUntil = now.Add(cfg.LockFor)
	}
	return nil
}

func (g *MemoryLoginGuard) Success(ctx context.Context, key string) error {
	_ = ctx
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.m, key)
	return nil
}

type RedisLoginGuard struct {
	cfg    LockoutConfig
	client *redisx.Client
}

func NewRedisLoginGuard(client *redisx.Client, cfg LockoutConfig) *RedisLoginGuard {
	return &RedisLoginGuard{cfg: cfg.withDefaults(), client: client}
}

func (g *RedisLoginGuard) keyFails(key string) string  { return "login:fails:" + key }
func (g *RedisLoginGuard) keyLock(key string) string   { return "login:lock:" + key }

func (g *RedisLoginGuard) Check(ctx context.Context, key string) error {
	ok, err := g.client.Exists(ctx, g.keyLock(key))
	if err != nil {
		return nil // fail open if redis blips
	}
	if ok {
		return apperrors.New("AUTH_LOCKED", "Too many failed login attempts; try again later", 429)
	}
	return nil
}

func (g *RedisLoginGuard) Fail(ctx context.Context, key string) error {
	rdb := g.client.Raw()
	failsKey := g.keyFails(key)
	n, err := rdb.Incr(ctx, failsKey).Result()
	if err != nil {
		return nil
	}
	if n == 1 {
		_ = rdb.Expire(ctx, failsKey, g.cfg.Window).Err()
	}
	if int(n) >= g.cfg.MaxAttempts {
		_ = g.client.Set(ctx, g.keyLock(key), fmt.Sprintf("%d", n), g.cfg.LockFor)
	}
	return nil
}

func (g *RedisLoginGuard) Success(ctx context.Context, key string) error {
	_ = g.client.Del(ctx, g.keyFails(key), g.keyLock(key))
	return nil
}
