package middleware

import (
	"sync"
	"time"

	"file_sys/backend/internal/util"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count    int
	windowStart time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		rl.mu.Lock()

		v, exists := rl.visitors[ip]
		now := time.Now()

		if !exists || now.Sub(v.windowStart) > rl.window {
			rl.visitors[ip] = &visitor{count: 1, windowStart: now}
			rl.mu.Unlock()
			c.Next()
			return
		}

		v.count++
		if v.count > rl.limit {
			rl.mu.Unlock()
			util.Error(c, 429, 42900, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		rl.mu.Unlock()
		c.Next()
	}
}

// Cleanup should be called periodically to purge stale entries.
func (rl *RateLimiter) Cleanup(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for ip, v := range rl.visitors {
		if now.Sub(v.windowStart) > maxAge {
			delete(rl.visitors, ip)
		}
	}
}
