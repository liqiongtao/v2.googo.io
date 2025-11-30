package goocontext

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// FromGrpcContext 从gRPC的metadata中提取上下文信息
func FromGrpcContext(ctx context.Context) *Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return Default(ctx)
	}

	c := Default(ctx)

	// 提取app-name
	if values := md.Get("app-name"); len(values) > 0 {
		c = c.WithValue("app-name", values[0])
	}

	// 提取trace-id
	if values := md.Get("trace-id"); len(values) > 0 {
		c = c.WithValue("trace-id", values[0])
	}

	return c
}
