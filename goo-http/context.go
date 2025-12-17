package goohttp

import (
	"github.com/gin-gonic/gin"
)

const (
	HttpRequestContextName = "http-request-context"
)

type Context struct {
	*gin.Context
	traceId string
}

func (c *Context) TraceId() string {
	return c.traceId
}

func (c *Context) SetTraceId(traceId string) {
	c.traceId = traceId
}

func GetContext(c *gin.Context) (*Context, bool) {
	v, ok := c.Get(HttpRequestContextName)
	if !ok {
		return nil, false
	}

	ctx, ok := v.(*Context)
	if !ok {
		return nil, false
	}

	return ctx, true
}
