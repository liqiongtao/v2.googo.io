package googrpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	goocontext "v2.googo.io/goo-context"
	goolog "v2.googo.io/goo-log"
)

// Client gRPC 客户端封装
type Client struct {
	name       string
	config     *Config
	conn       *grpc.ClientConn
	mu         sync.RWMutex
	registry   Registry
}

// Registry 注册中心接口
type Registry interface {
	// Resolve 解析服务地址
	Resolve(ctx context.Context, serviceName string) ([]string, error)
	// Close 关闭注册中心连接
	Close() error
}

// NewClient 创建新的 gRPC 客户端
func NewClient(name string, config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证地址或服务名称
	if config.Address == "" && config.ServiceName == "" {
		return nil, ErrEmptyAddress
	}

	// 如果使用注册中心，验证服务名称
	if config.RegistryType != RegistryTypeNone && config.ServiceName == "" {
		return nil, ErrEmptyServiceName
	}

	// 创建注册中心客户端（如果需要）
	var registry Registry
	var err error
	if config.RegistryType != RegistryTypeNone {
		registry, err = createRegistry(config)
		if err != nil {
			return nil, err
		}
	}

	// 创建 gRPC 连接
	var conn *grpc.ClientConn
	if config.RegistryType != RegistryTypeNone && registry != nil {
		// 使用服务发现
		address := fmt.Sprintf("discovery:///%s", config.ServiceName)
		conn, err = grpc.Dial(address, config.toGrpcDialOptions()...)
		if err != nil {
			if registry != nil {
				registry.Close()
			}
			return nil, err
		}
	} else {
		// 直接连接
		conn, err = grpc.Dial(config.Address, config.toGrpcDialOptions()...)
		if err != nil {
			if registry != nil {
				registry.Close()
			}
			return nil, err
		}
	}

	c := &Client{
		name:     name,
		config:   config,
		conn:     conn,
		registry: registry,
	}

	// 记录日志
	if config.EnableLog {
		goolog.InfoF("[goo-grpc] client '%s' connected to %s", name, config.Address)
	}

	return c, nil
}

// Name 获取客户端名称
func (c *Client) Name() string {
	return c.name
}

// Conn 获取 gRPC 连接
func (c *Client) Conn() *grpc.ClientConn {
	return c.conn
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error
	if c.conn != nil {
		err = c.conn.Close()
	}
	if c.registry != nil {
		if regErr := c.registry.Close(); regErr != nil {
			if err == nil {
				err = regErr
			}
		}
	}

	// 记录日志
	if c.config.EnableLog {
		goolog.InfoF("[goo-grpc] client '%s' closed", c.name)
	}

	return err
}

// Invoke 调用 gRPC 方法（带上下文和日志）
func (c *Client) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	// 如果启用追踪，从上下文提取 traceId 并添加到 metadata
	if c.config.EnableTrace {
		ctx = c.addTraceToContext(ctx)
	}

	// 记录日志
	if c.config.EnableLog {
		goocontextCtx := goocontext.Default(ctx)
		traceId := goocontextCtx.TraceId()
		goolog.WithField("method", method).
			WithField("trace-id", traceId).
			InfoF("[goo-grpc] client '%s' invoking method: %s", c.name, method)
	}

	// 调用 gRPC 方法
	err := c.conn.Invoke(ctx, method, args, reply, opts...)

	// 记录错误日志
	if err != nil && c.config.EnableLog {
		goocontextCtx := goocontext.Default(ctx)
		traceId := goocontextCtx.TraceId()
		goolog.WithField("method", method).
			WithField("trace-id", traceId).
			WithField("error", err.Error()).
			ErrorF("[goo-grpc] client '%s' invoke method '%s' failed", c.name, method)
	}

	return err
}

// NewStream 创建新的流（带上下文和日志）
func (c *Client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	// 如果启用追踪，从上下文提取 traceId 并添加到 metadata
	if c.config.EnableTrace {
		ctx = c.addTraceToContext(ctx)
	}

	// 记录日志
	if c.config.EnableLog {
		goocontextCtx := goocontext.Default(ctx)
		traceId := goocontextCtx.TraceId()
		goolog.WithField("method", method).
			WithField("trace-id", traceId).
			InfoF("[goo-grpc] client '%s' creating stream: %s", c.name, method)
	}

	// 创建流
	stream, err := c.conn.NewStream(ctx, desc, method, opts...)

	// 记录错误日志
	if err != nil && c.config.EnableLog {
		goocontextCtx := goocontext.Default(ctx)
		traceId := goocontextCtx.TraceId()
		goolog.WithField("method", method).
			WithField("trace-id", traceId).
			WithField("error", err.Error()).
			ErrorF("[goo-grpc] client '%s' create stream '%s' failed", c.name, method)
	}

	return stream, err
}

// addTraceToContext 将 traceId 添加到 gRPC metadata
func (c *Client) addTraceToContext(ctx context.Context) context.Context {
	goocontextCtx := goocontext.Default(ctx)
	traceId := goocontextCtx.TraceId()
	appName := goocontextCtx.AppName()

	if traceId == "" && appName == "" {
		return ctx
	}

	md := make([]string, 0, 4)
	if traceId != "" {
		md = append(md, "trace-id", traceId)
	}
	if appName != "" {
		md = append(md, "app-name", appName)
	}

	return metadata.AppendToOutgoingContext(ctx, md...)
}

// createRegistry 创建注册中心客户端
func createRegistry(config *Config) (Registry, error) {
	switch config.RegistryType {
	case RegistryTypeEtcd:
		if config.EtcdConfig == nil {
			return nil, fmt.Errorf("etcd config is required when registry type is etcd")
		}
		return NewEtcdRegistry(config.EtcdConfig)
	case RegistryTypeConsul:
		if config.ConsulConfig == nil {
			return nil, fmt.Errorf("consul config is required when registry type is consul")
		}
		return NewConsulRegistry(config.ConsulConfig)
	default:
		return nil, ErrRegistryNotSupported
	}
}

// EtcdRegistry etcd 注册中心实现
type EtcdRegistry struct {
	endpoints  []string
	username   string
	password   string
	dialTimeout time.Duration
}

// NewEtcdRegistry 创建 etcd 注册中心
func NewEtcdRegistry(config *EtcdRegistryConfig) (*EtcdRegistry, error) {
	if len(config.Endpoints) == 0 {
		return nil, fmt.Errorf("etcd endpoints cannot be empty")
	}

	return &EtcdRegistry{
		endpoints:   config.Endpoints,
		username:    config.Username,
		password:    config.Password,
		dialTimeout: config.DialTimeout,
	}, nil
}

// Resolve 解析服务地址
func (r *EtcdRegistry) Resolve(ctx context.Context, serviceName string) ([]string, error) {
	// TODO: 实现 etcd 服务发现
	// 这里需要集成 etcd 客户端来查询服务地址
	// 暂时返回空，实际使用时需要实现
	return []string{}, nil
}

// Close 关闭 etcd 连接
func (r *EtcdRegistry) Close() error {
	// TODO: 关闭 etcd 客户端连接
	return nil
}

// ConsulRegistry consul 注册中心实现
type ConsulRegistry struct {
	address string
}

// NewConsulRegistry 创建 consul 注册中心
func NewConsulRegistry(config *ConsulRegistryConfig) (*ConsulRegistry, error) {
	if config.Address == "" {
		return nil, fmt.Errorf("consul address cannot be empty")
	}

	return &ConsulRegistry{
		address: config.Address,
	}, nil
}

// Resolve 解析服务地址
func (r *ConsulRegistry) Resolve(ctx context.Context, serviceName string) ([]string, error) {
	// TODO: 实现 consul 服务发现
	// 这里需要集成 consul 客户端来查询服务地址
	// 暂时返回空，实际使用时需要实现
	return []string{}, nil
}

// Close 关闭 consul 连接
func (r *ConsulRegistry) Close() error {
	// TODO: 关闭 consul 客户端连接
	return nil
}

