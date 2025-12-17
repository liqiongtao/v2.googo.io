package goohttp

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	DefaultRateLimitConfig = &RateLimitConfig{
		Rate:  100,
		Burst: 200,
	}
)

type RateLimiter struct {
	limiter *rate.Limiter
	mu      sync.RWMutex
}

type RateLimitConfig struct {
	Rate  float64 // 每秒允许的请求数
	Burst int     // 突发请求数
}

func (rl *RateLimiter) Allow() bool {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.limiter.Allow()
}

func (rl *RateLimiter) UpdateConfig(config *RateLimitConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.limiter = rate.NewLimiter(rate.Limit(config.Rate), config.Burst)
}

func RateLimitMiddleware(config *RateLimitConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultRateLimitConfig
	}

	limiter := &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(config.Rate), config.Burst),
	}

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{})
			c.Abort()
			return
		}

		c.Next()
	}
}
