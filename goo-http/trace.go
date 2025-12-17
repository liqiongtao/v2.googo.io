package goohttp

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	DefaultTraceIdHeader = "X-Trace-Id"
)

func TraceMiddleware(traceIdHeader string) gin.HandlerFunc {
	if traceIdHeader == "" {
		traceIdHeader = DefaultTraceIdHeader
	}

	return func(c *gin.Context) {
		traceId := c.GetHeader(traceIdHeader)
		if traceId == "" {
			traceId = uuid.New().String()
		}

		ctx := &Context{Context: c}
		ctx.SetTraceId(traceId)

		c.Header(traceIdHeader, traceId)
		c.Next()
	}
}
