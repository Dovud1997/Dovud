package ws

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
)

// RedisEventBus publishes WS envelopes on a shared Redis channel.
type RedisEventBus struct {
	rdb     *redis.Client
	channel string
}

func NewRedisEventBus(rdb *redis.Client) *RedisEventBus {
	return &RedisEventBus{rdb: rdb, channel: redisWSChannel}
}

func (b *RedisEventBus) Publish(ctx context.Context, payload []byte) error {
	return b.rdb.Publish(ctx, b.channel, payload).Err()
}

func (b *RedisEventBus) Subscribe(ctx context.Context) (<-chan []byte, func(), error) {
	pubsub := b.rdb.Subscribe(ctx, b.channel)
	out := make(chan []byte, 64)
	go func() {
		defer close(out)
		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				select {
				case out <- []byte(msg.Payload):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	unsub := func() { _ = pubsub.Close() }
	return out, unsub, nil
}

// MemoryEventBus is an in-process pub/sub for tests / single-process fallback.
type MemoryEventBus struct {
	mu   sync.RWMutex
	subs map[chan []byte]struct{}
}

func NewMemoryEventBus() *MemoryEventBus {
	return &MemoryEventBus{subs: map[chan []byte]struct{}{}}
}

func (b *MemoryEventBus) Publish(_ context.Context, payload []byte) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs {
		cp := append([]byte(nil), payload...)
		select {
		case ch <- cp:
		default:
		}
	}
	return nil
}

func (b *MemoryEventBus) Subscribe(_ context.Context) (<-chan []byte, func(), error) {
	ch := make(chan []byte, 64)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	var once sync.Once
	unsub := func() {
		once.Do(func() {
			b.mu.Lock()
			delete(b.subs, ch)
			b.mu.Unlock()
			close(ch)
		})
	}
	return ch, unsub, nil
}
