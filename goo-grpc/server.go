package googrpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	goocontext "v2.googo.io/goo-context"
	goolog "v2.googo.io/goo-log"
)

// Server gRPC 服务端封装
type Server struct {
	name     string
	config   *Config
	server   *grpc.Server
	listener net.Listener
	mu       sync.RWMutex
	registry Registry
}

// NewServer 创建新的 gRPC 服务端
func NewServer(name string, config *Config) (*Server, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证地址
	if config.Address == "" {
		return nil, ErrEmptyAddress
	}

	// 如果使用注册中心，验证服务名称
	if config.RegistryType != RegistryTypeNone && config.ServiceName == "" {
		return nil, ErrEmptyServiceName
	}

	// 创建监听器
	listener, err := net.Listen("tcp", config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	// 创建 gRPC 服务端选项
	opts := config.toGrpcServerOptions()

	// 如果启用日志和追踪，添加拦截器
	if config.EnableLog || config.EnableTrace {
		opts = append(opts, grpc.UnaryInterceptor(unaryServerInterceptor(config)))
		opts = append(opts, grpc.StreamInterceptor(streamServerInterceptor(config)))
	}

	// 创建 gRPC 服务端
	server := grpc.NewServer(opts...)

	// 创建注册中心客户端（如果需要）
	var registry Registry
	if config.RegistryType != RegistryTypeNone {
		registry, err = createRegistry(config)
		if err != nil {
			listener.Close()
			return nil, err
		}

		// 注册服务
		err = registerService(registry, config, listener.Addr().String())
		if err != nil {
			listener.Close()
			if registry != nil {
				registry.Close()
			}
			return nil, fmt.Errorf("failed to register service: %w", err)
		}
	}

	s := &Server{
		name:     name,
		config:   config,
		server:   server,
		listener: listener,
		registry: registry,
	}

	// 记录日志
	if config.EnableLog {
		goolog.InfoF("[goo-grpc] server '%s' listening on %s", name, config.Address)
		if config.RegistryType != RegistryTypeNone {
			goolog.InfoF("[goo-grpc] server '%s' registered to %s with service name: %s", name, config.RegistryType, config.ServiceName)
		}
	}

	return s, nil
}

// Name 获取服务端名称
func (s *Server) Name() string {
	return s.name
}

// Server 获取 gRPC 服务端
func (s *Server) Server() *grpc.Server {
	return s.server
}

// RegisterService 注册服务实现
func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.server.RegisterService(desc, impl)
}

// Serve 启动服务（阻塞调用）
func (s *Server) Serve() error {
	if s.config.EnableLog {
		goolog.InfoF("[goo-grpc] server '%s' starting to serve", s.name)
	}
	return s.server.Serve(s.listener)
}

// GracefulStop 优雅停止服务
func (s *Server) GracefulStop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config.EnableLog {
		goolog.InfoF("[goo-grpc] server '%s' gracefully stopping", s.name)
	}

	// 注销服务
	if s.registry != nil && s.config.RegistryType != RegistryTypeNone {
		unregisterService(s.registry, s.config)
	}

	s.server.GracefulStop()
}

// Stop 立即停止服务
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config.EnableLog {
		goolog.InfoF("[goo-grpc] server '%s' stopping", s.name)
	}

	// 注销服务
	if s.registry != nil && s.config.RegistryType != RegistryTypeNone {
		unregisterService(s.registry, s.config)
	}

	s.server.Stop()
}

// Close 关闭服务端
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	if s.listener != nil {
		err = s.listener.Close()
	}
	if s.registry != nil {
		if regErr := s.registry.Close(); regErr != nil {
			if err == nil {
				err = regErr
			}
		}
	}

	// 记录日志
	if s.config.EnableLog {
		goolog.InfoF("[goo-grpc] server '%s' closed", s.name)
	}

	return err
}

// unaryServerInterceptor 一元 RPC 拦截器（用于日志和追踪）
func unaryServerInterceptor(config *Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 从 metadata 提取 traceId
		if config.EnableTrace {
			ctx = extractTraceFromContext(ctx)
		}

		// 记录请求日志
		if config.EnableLog {
			goocontextCtx := goocontext.Default(ctx)
			traceId := goocontextCtx.TraceId()
			goolog.WithField("method", info.FullMethod).
				WithField("trace-id", traceId).
				InfoF("[goo-grpc] server handling unary request: %s", info.FullMethod)
		}

		// 调用处理函数
		resp, err := handler(ctx, req)

		// 记录响应日志
		if config.EnableLog {
			goocontextCtx := goocontext.Default(ctx)
			traceId := goocontextCtx.TraceId()
			if err != nil {
				goolog.WithField("method", info.FullMethod).
					WithField("trace-id", traceId).
					WithField("error", err.Error()).
					ErrorF("[goo-grpc] server unary request '%s' failed", info.FullMethod)
			} else {
				goolog.WithField("method", info.FullMethod).
					WithField("trace-id", traceId).
					InfoF("[goo-grpc] server unary request '%s' completed", info.FullMethod)
			}
		}

		return resp, err
	}
}

// streamServerInterceptor 流式 RPC 拦截器（用于日志和追踪）
func streamServerInterceptor(config *Config) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		// 从 metadata 提取 traceId
		if config.EnableTrace {
			ctx = extractTraceFromContext(ctx)
		}

		// 记录请求日志
		if config.EnableLog {
			goocontextCtx := goocontext.Default(ctx)
			traceId := goocontextCtx.TraceId()
			goolog.WithField("method", info.FullMethod).
				WithField("trace-id", traceId).
				InfoF("[goo-grpc] server handling stream request: %s", info.FullMethod)
		}

		// 调用处理函数
		err := handler(srv, ss)

		// 记录响应日志
		if config.EnableLog {
			goocontextCtx := goocontext.Default(ctx)
			traceId := goocontextCtx.TraceId()
			if err != nil {
				goolog.WithField("method", info.FullMethod).
					WithField("trace-id", traceId).
					WithField("error", err.Error()).
					ErrorF("[goo-grpc] server stream request '%s' failed", info.FullMethod)
			} else {
				goolog.WithField("method", info.FullMethod).
					WithField("trace-id", traceId).
					InfoF("[goo-grpc] server stream request '%s' completed", info.FullMethod)
			}
		}

		return err
	}
}

// extractTraceFromContext 从 gRPC metadata 提取 traceId 并设置到上下文
func extractTraceFromContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	var traceId, appName string
	if values := md.Get("trace-id"); len(values) > 0 {
		traceId = values[0]
	}
	if values := md.Get("app-name"); len(values) > 0 {
		appName = values[0]
	}

	if traceId == "" && appName == "" {
		return ctx
	}

	goocontextCtx := goocontext.Default(ctx)
	if traceId != "" {
		goocontextCtx = goocontextCtx.WithTraceId(traceId)
	}
	if appName != "" {
		goocontextCtx = goocontextCtx.WithAppName(appName)
	}

	return goocontextCtx.Context
}

// registerService 注册服务到注册中心
func registerService(registry Registry, config *Config, address string) error {
	// TODO: 实现服务注册逻辑
	// 这里需要根据注册中心类型调用相应的注册方法
	// 暂时只记录日志
	if config.EnableLog {
		goolog.InfoF("[goo-grpc] registering service '%s' at %s to %s", config.ServiceName, address, config.RegistryType)
	}
	return nil
}

// unregisterService 从注册中心注销服务
func unregisterService(registry Registry, config *Config) {
	// TODO: 实现服务注销逻辑
	// 这里需要根据注册中心类型调用相应的注销方法
	// 暂时只记录日志
	if config.EnableLog {
		goolog.InfoF("[goo-grpc] unregistering service '%s' from %s", config.ServiceName, config.RegistryType)
	}
}

