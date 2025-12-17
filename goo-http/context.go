package goohttp

import (
	"github.com/gin-gonic/gin"
)

type Context struct {
	*gin.Context
}

func (c *Context) TraceId() string {
	return c.Context.GetString("trace-id")
}

func (c *Context) SetTraceId(traceId string) {
	c.Context.Set("trace-id", traceId)
}

func (c *Context) ClientIP() string {
	if v := c.Context.GetHeader("X-Real-Ip"); v != "" {
		return v
	}
	return c.Context.ClientIP()
}
