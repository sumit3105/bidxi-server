package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"bidxi/internal/utils"
)

// rateLimiterEntry tracks request counts per client.
type rateLimiterEntry struct {
	count    int
	resetAt  time.Time
}

// RateLimiter implements a simple in-memory sliding-window rate limiter.
type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*rateLimiterEntry
	maxRPM   int
}

// NewRateLimiter creates a new RateLimiter with the given max requests per minute.
func NewRateLimiter(maxRPM int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*rateLimiterEntry),
		maxRPM:  maxRPM,
	}

	// Background cleanup goroutine to prevent memory leaks
	go rl.cleanup()

	return rl
}

// Middleware returns a Gin middleware that enforces rate limiting
// based on client IP (or authenticated user ID if available).
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use user ID if authenticated, otherwise fall back to IP
		key := c.ClientIP()
		if userID, exists := c.Get("userID"); exists {
			key = userID.(string)
		}

		rl.mu.Lock()
		entry, exists := rl.clients[key]

		now := time.Now()

		if !exists || now.After(entry.resetAt) {
			// New window
			rl.clients[key] = &rateLimiterEntry{
				count:   1,
				resetAt: now.Add(1 * time.Minute),
			}
			rl.mu.Unlock()
			c.Next()
			return
		}

		if entry.count >= rl.maxRPM {
			rl.mu.Unlock()
			retryAfter := entry.resetAt.Sub(now).Seconds()
			c.Header("Retry-After", time.Duration(retryAfter * float64(time.Second)).String())
			utils.RespondError(c, http.StatusTooManyRequests, "Rate limit exceeded. Try again shortly.")
			c.Abort()
			return
		}

		entry.count++
		rl.mu.Unlock()
		c.Next()
	}
}

// cleanup removes expired entries every 5 minutes to prevent memory leaks.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.clients {
			if now.After(entry.resetAt) {
				delete(rl.clients, key)
			}
		}
		rl.mu.Unlock()
	}
}
