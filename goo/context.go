package goo

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

type Context struct {
	context.Context
}

func NewContext(ctx context.Context) *Context {
	return &Context{Context: ctx}
}

func (c *Context) AppName() string {
	if v, ok := c.Context.Value("AppName").(string); ok {
		return v
	}
	if v, ok := c.Context.Value("app-name").(string); ok {
		return v
	}
	if v, ok := c.Context.Value("app_name").(string); ok {
		return v
	}
	return ""
}

func (c *Context) TraceId() string {
	if v, ok := c.Context.Value("TraceId").(string); ok {
		return v
	}
	if v, ok := c.Context.Value("trace-id").(string); ok {
		return v
	}
	if v, ok := c.Context.Value("trace_id").(string); ok {
		return v
	}
	c.WithTraceId()
	return c.TraceId()
}

func (c *Context) WithAppName(appName string) *Context {
	return c.WithValue("app-name", appName)
}

func (c *Context) WithTraceId() *Context {
	return c.WithValue("trace-id", uuid.New().String())
}

func (c *Context) WithValue(key, val any) *Context {
	c.Context = context.WithValue(c.Context, key, val)
	return c
}

func (c *Context) WithCancel() (*Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c.Context)
	return NewContext(ctx), cancel
}

func (c *Context) WithGinContext(ctx *gin.Context, data ...any) context.Context {
	ctx.Set("app-name", c.AppName())
	ctx.Set("trace-id", c.TraceId())
	for i := 0; i < len(data); i += 2 {
		if i+1 >= len(data) {
			continue
		}
		ctx.Set(data[i], data[i+1])
	}
	return ctx
}

func (c *Context) WithGrpcContext(kvs ...string) context.Context {
	kvs = append(
		kvs, "app-name", c.AppName(), "trace-id", c.TraceId(),
	)
	return metadata.AppendToOutgoingContext(c.Context, kvs...)
}
