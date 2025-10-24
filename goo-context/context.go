package goocontext

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Context struct {
	context.Context
}

func New(ctx context.Context) *Context {
	c := &Context{Context: ctx}
	c.WithTraceId()
	return c
}

func (c *Context) AppName() string {
	if v := c.ValueString("AppName"); v != "" {
		return v
	}
	if v := c.ValueString("app-name"); v != "" {
		return v
	}
	if v := c.ValueString("app_name"); v != "" {
		return v
	}
	return ""
}

func (c *Context) TraceId() string {
	if v := c.ValueString("TraceId"); v != "" {
		return v
	}
	if v := c.ValueString("trace-id"); v != "" {
		return v
	}
	if v := c.ValueString("trace_id"); v != "" {
		return v
	}
	return ""
}

func (c *Context) WithAppName(appName string, args ...any) *Context {
	return c.WithValue("app-name", fmt.Sprintf(appName, args...))
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
	return New(ctx), cancel
}

func (c *Context) WithTimeout(d time.Duration) (*Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.Context, d)
	return New(ctx), cancel
}

func (c *Context) WithSignalNotify(signals ...os.Signal) *Context {
	if len(signals) == 0 {
		signals = []os.Signal{
			syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGHUP,
			syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL,
		}
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, signals...)

	ctx, cancel := c.WithCancel()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("WithSignal recover:", r)
			}
		}()

		select {
		case <-sig:
			cancel()
		}
	}()

	return ctx
}

func (c *Context) WithGinContext(ctx *gin.Context, data ...any) *Context {
	ctx.Set("app-name", c.AppName())
	ctx.Set("trace-id", c.TraceId())
	for i := 0; i < len(data)-1; i += 2 {
		ctx.Set(data[i], data[i+1])
	}
	return &Context{Context: ctx}
}

func (c *Context) WithGrpcContext(kvs ...string) *Context {
	kvs = append(
		kvs,
		"app-name", c.AppName(),
		"trace-id", c.TraceId(),
	)
	return &Context{Context: metadata.AppendToOutgoingContext(c.Context, kvs...)}
}
