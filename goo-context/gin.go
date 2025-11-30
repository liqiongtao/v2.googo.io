package goocontext

import (
	"context"

	"github.com/gin-gonic/gin"
)

// FromGinContext 从gin.Context创建新的上下文
func FromGinContext(ginCtx *gin.Context) *Context {
	if ginCtx == nil {
		return Default(context.TODO())
	}

	ctx := Default(ginCtx.Request.Context())

	// 从gin.Context中提取app-name
	if v, exists := ginCtx.Get("app-name"); exists {
		if name, ok := v.(string); ok {
			ctx = ctx.WithValue("app-name", name)
		}
	}

	// 从gin.Context中提取trace-id
	if v, exists := ginCtx.Get("trace-id"); exists {
		if id, ok := v.(string); ok {
			ctx = ctx.WithValue("trace-id", id)
		}
	}

	return ctx
}
