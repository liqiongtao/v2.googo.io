package goohttp

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type Logger interface {
	Info(ctx context.Context, msg string, fields ...interface{})
	Error(ctx context.Context, msg string, fields ...interface{})
	Debug(ctx context.Context, msg string, fields ...interface{})
	Warn(ctx context.Context, msg string, fields ...interface{})
}

type DefaultLogger struct{}

func (l *DefaultLogger) Info(ctx context.Context, msg string, fields ...interface{}) {
	fmt.Printf("[INFO] %s %v\n", msg, fields)
}

func (l *DefaultLogger) Error(ctx context.Context, msg string, fields ...interface{}) {
	fmt.Printf("[ERROR] %s %v\n", msg, fields)
}

func (l *DefaultLogger) Debug(ctx context.Context, msg string, fields ...interface{}) {
	fmt.Printf("[DEBUG] %s %v\n", msg, fields)
}

func (l *DefaultLogger) Warn(ctx context.Context, msg string, fields ...interface{}) {
	fmt.Printf("[WARN] %s %v\n", msg, fields)
}

func LogMiddleware(logger Logger) gin.HandlerFunc {
	if logger == nil {
		logger = &DefaultLogger{}
	}

	return func(c *gin.Context) {
		startTime := time.Now()

		data := map[string]any{
			"method": c.Request.Method,
			"uri":    c.Request.RequestURI,
			"ip":     c.ClientIP(),
		}

		if ctx, ok := GetContext(c); ok {
			data["trace_id"] = ctx.TraceId()
		}

		logger.Info(c.Request.Context(), "Request", data)

		c.Next()

		status := c.Writer.Status()

		data["status"] = status
		data["time"] = time.Since(startTime)

		if status >= 500 {
			logger.Error(c.Request.Context(), "Response", data)
		} else if status >= 400 {
			logger.Warn(c.Request.Context(), "Response", data)
		} else {
			logger.Info(c.Request.Context(), "Response", data)
		}
	}
}
