package goocontext

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Default 创建一个默认的上下文，如果parent为nil则使用context.Background()
func Default(parent context.Context) *Context {
	if parent == nil {
		parent = context.Background()
	}
	return &Context{Context: parent}
}

// WithAppName 设置应用名称
func WithAppName(parent context.Context, appName string, args ...any) *Context {
	return Default(context.WithValue(parent, "app-name", fmt.Sprintf(appName, args...)))
}

// WithTraceId 设置或生成TraceId
func WithTraceId(parent context.Context, traceId ...string) *Context {
	var id string
	if len(traceId) > 0 && traceId[0] != "" {
		id = traceId[0]
	} else {
		id = uuid.New().String()
	}
	return Default(context.WithValue(parent, "trace-id", id))
}

// AppName 获取应用名称
func AppName(c *Context) string {
	if c == nil {
		return ""
	}
	if v := ValueString(c, "AppName"); v != "" {
		return v
	}
	if v := ValueString(c, "app-name"); v != "" {
		return v
	}
	if v := ValueString(c, "app_name"); v != "" {
		return v
	}
	return ""
}

// TraceId 获取TraceId
func TraceId(c *Context) string {
	if c == nil {
		return ""
	}
	if v := ValueString(c, "TraceId"); v != "" {
		return v
	}
	if v := ValueString(c, "trace-id"); v != "" {
		return v
	}
	if v := ValueString(c, "trace_id"); v != "" {
		return v
	}
	if v := ValueString(c, "request_id"); v != "" {
		return v
	}
	return ""
}
