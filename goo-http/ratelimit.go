package goohttp

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	DefaultRateLimitConfig = &RateLimitConfig{
		Rate:  100,
		Burst: 200,
		KeyFunc: func(c *Context) string {
			return c.ClientIP()
		},
		CleanupInterval: 5 * time.Minute,
		MaxIdleTime:     10 * time.Minute,
	}
)

type RateLimitKeyFunc func(*Context) string

type RateLimitConfig struct {
	Rate            float64          // 每秒允许的请求数
	Burst           int              // 突发请求数
	KeyFunc         RateLimitKeyFunc // 提取限流key的函数
	CleanupInterval time.Duration    // 清理不活跃限流器的间隔（默认 5 分钟）
	MaxIdleTime     time.Duration    // 限流器最大空闲时间（默认 10 分钟）
}

type RateLimiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
	mu         sync.RWMutex
}

type RateLimiter struct {
	config  *RateLimitConfig
	entries map[string]*RateLimiterEntry
	mu      sync.RWMutex
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

func NewRateLimiter(config *RateLimitConfig) *RateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig
	}

	rl := &RateLimiter{
		config:  config,
		entries: make(map[string]*RateLimiterEntry),
		stopCh:  make(chan struct{}),
	}

	rl.wg.Add(1)
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) getLimiterEntry(key string) *RateLimiterEntry {
	rl.mu.RLock()
	entry, exists := rl.entries[key]
	rl.mu.RUnlock()

	if exists {
		return entry
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if entry, exists = rl.entries[key]; exists {
		return entry
	}

	entry = &RateLimiterEntry{
		limiter:    rate.NewLimiter(rate.Limit(rl.config.Rate), rl.config.Burst),
		lastAccess: time.Now(),
	}

	rl.entries[key] = entry

	return entry
}

func (rl *RateLimiter) Allow(key string) bool {
	entry := rl.getLimiterEntry(key)

	entry.mu.RLock()
	allowed := entry.limiter.Allow()
	entry.mu.RUnlock()

	entry.mu.Lock()
	entry.lastAccess = time.Now()
	entry.mu.Unlock()

	return allowed
}

func (rl *RateLimiter) UpdateConfig(config *RateLimitConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for _, entry := range rl.entries {
		entry.mu.Lock()
		entry.limiter = rate.NewLimiter(rate.Limit(config.Rate), config.Burst)
		entry.mu.Unlock()
	}
}

func (rl *RateLimiter) cleanup() {
	defer rl.wg.Done()

	ticker := time.NewTicker(rl.config.CleanupInterval)
	defer ticker.Stop()

	handler := func(now time.Time) {
		rl.mu.Lock()
		defer rl.mu.Unlock()

		for key, entry := range rl.entries {
			entry.mu.RLock()
			idleTime := now.Sub(entry.lastAccess)
			entry.mu.RUnlock()

			if idleTime > rl.config.MaxIdleTime {
				delete(rl.entries, key)
			}
		}
	}

	for {
		select {
		case <-ticker.C:
			handler(time.Now())
		case <-rl.stopCh:
			return
		}
	}
}

func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
	rl.wg.Wait()
}

func RateLimitMiddleware(limiters []*RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := &Context{Context: c}
		for _, limiter := range limiters {
			key := limiter.config.KeyFunc(ctx)
			if !limiter.Allow(key) {
				ErrorWithStatus(ctx, http.StatusTooManyRequests, 4290, "Rate limit exceeded")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
