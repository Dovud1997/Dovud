package httpx

import (
	"sync"
	"time"

	apperrors "github.com/Dovud1997/Dovud/backend/internal/platform/errors"
	"github.com/gofiber/fiber/v2"
)

type SecurityConfig struct {
	CORSOrigins     string
	BodyLimitBytes  int
	RateLimitMax    int
	RateLimitWindow time.Duration
}

func SecurityHeadersMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "no-referrer")
		c.Set("X-XSS-Protection", "0")
		c.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		if c.Protocol() == "https" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		return c.Next()
	}
}

type rateBucket struct {
	count int
	reset time.Time
}

// MemoryRateLimiter is a simple per-IP fixed-window limiter (no Redis required).
func MemoryRateLimiter(max int, window time.Duration) fiber.Handler {
	if max <= 0 {
		max = 120
	}
	if window <= 0 {
		window = time.Minute
	}
	var mu sync.Mutex
	buckets := map[string]*rateBucket{}

	return func(c *fiber.Ctx) error {
		ip := c.IP()
		now := time.Now()
		mu.Lock()
		b, ok := buckets[ip]
		if !ok || now.After(b.reset) {
			b = &rateBucket{count: 0, reset: now.Add(window)}
			buckets[ip] = b
		}
		b.count++
		count := b.count
		reset := b.reset
		// opportunistic cleanup
		if len(buckets) > 10000 {
			for k, v := range buckets {
				if now.After(v.reset) {
					delete(buckets, k)
				}
			}
		}
		mu.Unlock()
		if count > max {
			retry := int(reset.Sub(now).Seconds())
			if retry < 1 {
				retry = 1
			}
			c.Set("Retry-After", itoa(retry))
			return Fail(c, apperrors.New("RATE_LIMITED", "Too many requests", 429))
		}
		return c.Next()
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
