package goocontext

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

// Context 封装了 context.Context，提供增强功能
type Context struct {
	context.Context
}

// WithValue 设置key-value对到上下文中
func (c *Context) WithValue(key string, value any) *Context {
	return Default(context.WithValue(c.Context, key, value))
}

// WithCancel 创建一个可取消的上下文
func (c *Context) WithCancel() (*Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c.Context)
	return Default(ctx), cancel
}

// WithTimeout 创建一个带超时的上下文
func (c *Context) WithTimeout(d time.Duration) (*Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.Context, d)
	return Default(ctx), cancel
}

// WithDeadline 创建一个带截止时间的上下文
func (c *Context) WithDeadline(d time.Time) (*Context, context.CancelFunc) {
	ctx, cancel := context.WithDeadline(c.Context, d)
	return Default(ctx), cancel
}

// WithSignalNotify 创建一个监听系统信号的上下文
// 当接收到指定信号时，上下文会被取消
// 如果不指定signals，默认监听所有常见退出信号
func (c *Context) WithSignalNotify(signals ...os.Signal) *Context {
	if len(signals) == 0 {
		signals = []os.Signal{
			syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGHUP,
			syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT,
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, signals...)

	ctx, cancel := c.WithCancel()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("WithSignalNotify recover:", r)
			}
		}()

		<-sig
		cancel()
	}()

	return ctx
}

// WithGinContext 从gin.Context创建或更新上下文
// 会自动从gin.Context中提取app-name和trace-id，并设置到上下文中
// 如果gin.Context中没有这些值，会从当前上下文继承
func (c *Context) WithGinContext(ginCtx *gin.Context) *Context {
	if ginCtx == nil {
		return c
	}

	// 从gin.Context中获取或设置app-name
	appName := AppName(c)
	if appName == "" {
		if v, exists := ginCtx.Get("app-name"); exists {
			if name, ok := v.(string); ok {
				appName = name
			}
		}
	} else {
		ginCtx.Set("app-name", appName)
	}

	// 从gin.Context中获取或设置trace-id
	traceId := TraceId(c)
	if traceId == "" {
		if v, exists := ginCtx.Get("trace-id"); exists {
			if id, ok := v.(string); ok {
				traceId = id
			}
		}
		if traceId == "" {
			traceId = uuid.New().String()
		}
	}
	ginCtx.Set("trace-id", traceId)

	// 创建新的上下文，包含app-name和trace-id
	ctx := Default(ginCtx.Request.Context())
	if appName != "" {
		ctx = ctx.WithValue("app-name", appName)
	}
	if traceId != "" {
		ctx = ctx.WithValue("trace-id", traceId)
	}

	return ctx
}

// WithGrpcContext 将上下文中的app-name和trace-id添加到gRPC的metadata中
// 可以额外指定其他key-value对
func (c *Context) WithGrpcContext(kvs ...string) *Context {
	appName := AppName(c)
	traceId := TraceId(c)

	// 构建metadata key-value对
	mdKVs := make([]string, 0, len(kvs)+4)
	mdKVs = append(mdKVs, kvs...)

	if appName != "" {
		mdKVs = append(mdKVs, "app-name", appName)
	}
	if traceId != "" {
		mdKVs = append(mdKVs, "trace-id", traceId)
	}

	ctx := metadata.AppendToOutgoingContext(c.Context, mdKVs...)
	return Default(ctx)
}
