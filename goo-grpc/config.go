package googrpc

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// RegistryType 注册中心类型
type RegistryType string

const (
	// RegistryTypeNone 不使用注册中心
	RegistryTypeNone RegistryType = "none"
	// RegistryTypeEtcd etcd 注册中心
	RegistryTypeEtcd RegistryType = "etcd"
	// RegistryTypeConsul consul 注册中心
	RegistryTypeConsul RegistryType = "consul"
)

// Config gRPC 配置
type Config struct {
	// 地址（客户端：目标地址，服务端：监听地址）
	Address string

	// 服务名称（用于服务注册和发现）
	ServiceName string

	// 注册中心类型
	RegistryType RegistryType

	// etcd 配置（当 RegistryType 为 etcd 时使用）
	EtcdConfig *EtcdRegistryConfig

	// consul 配置（当 RegistryType 为 consul 时使用）
	ConsulConfig *ConsulRegistryConfig

	// 是否启用 TLS，默认 false
	EnableTLS bool

	// TLS 配置（当 EnableTLS 为 true 时使用）
	TLSConfig *TLSConfig

	// 连接超时时间，默认 5 秒
	Timeout time.Duration

	// 是否启用日志，默认 true
	EnableLog bool

	// 是否启用追踪（traceId），默认 true
	EnableTrace bool

	// 最大发送消息大小（字节），默认 4MB
	MaxSendMsgSize int

	// 最大接收消息大小（字节），默认 4MB
	MaxRecvMsgSize int

	// Keepalive 参数
	KeepaliveParams *keepalive.ClientParameters

	// 服务端 Keepalive 强制策略
	KeepaliveEnforcementPolicy *keepalive.EnforcementPolicy

	// 其他 gRPC 选项
	Options     []grpc.ServerOption
	DialOptions []grpc.DialOption
}

// EtcdRegistryConfig etcd 注册中心配置
type EtcdRegistryConfig struct {
	// 端点列表，格式: ["localhost:2379"]
	Endpoints []string

	// 用户名
	Username string

	// 密码
	Password string

	// 连接超时时间，默认 5 秒
	DialTimeout time.Duration

	// 租约 TTL（秒），默认 10 秒
	LeaseTTL int64
}

// ConsulRegistryConfig consul 注册中心配置
type ConsulRegistryConfig struct {
	// 地址，格式: "localhost:8500"
	Address string

	// 健康检查间隔（秒），默认 10 秒
	HealthCheckInterval int

	// 健康检查超时（秒），默认 5 秒
	HealthCheckTimeout int

	// 健康检查路径，默认 "/health"
	HealthCheckPath string
}

// TLSConfig TLS 配置
type TLSConfig struct {
	// 证书文件路径
	CertFile string

	// 密钥文件路径
	KeyFile string

	// CA 证书文件路径
	CAFile string

	// 是否跳过证书验证，默认 false
	InsecureSkipVerify bool
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config {
	c := &Config{
		Address:        "localhost:50051",
		ServiceName:    "",
		RegistryType:   RegistryTypeNone,
		EnableTLS:      false,
		Timeout:        5 * time.Second,
		EnableLog:      true,
		EnableTrace:    true,
		MaxSendMsgSize: 4 * 1024 * 1024, // 4MB
		MaxRecvMsgSize: 4 * 1024 * 1024, // 4MB
		Options:        []grpc.ServerOption{},
		DialOptions:    []grpc.DialOption{},
	}
	for _, opt := range opts {
		opt.Apply(c)
	}
	return c
}

// toGrpcServerOptions 转换为 gRPC 服务端选项
func (c *Config) toGrpcServerOptions() []grpc.ServerOption {
	opts := make([]grpc.ServerOption, 0)

	// 设置最大消息大小
	if c.MaxSendMsgSize > 0 {
		opts = append(opts, grpc.MaxSendMsgSize(c.MaxSendMsgSize))
	}
	if c.MaxRecvMsgSize > 0 {
		opts = append(opts, grpc.MaxRecvMsgSize(c.MaxRecvMsgSize))
	}

	// Keepalive 强制策略
	if c.KeepaliveEnforcementPolicy != nil {
		opts = append(opts, grpc.KeepaliveEnforcementPolicy(*c.KeepaliveEnforcementPolicy))
	}

	// TLS 配置
	if c.EnableTLS && c.TLSConfig != nil {
		creds, err := credentials.NewServerTLSFromFile(c.TLSConfig.CertFile, c.TLSConfig.KeyFile)
		if err == nil {
			opts = append(opts, grpc.Creds(creds))
		}
	}

	// 合并其他选项
	opts = append(opts, c.Options...)

	return opts
}

// toGrpcDialOptions 转换为 gRPC 客户端拨号选项
func (c *Config) toGrpcDialOptions() []grpc.DialOption {
	opts := make([]grpc.DialOption, 0)

	// 设置最大消息大小
	if c.MaxSendMsgSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(c.MaxSendMsgSize)))
	}
	if c.MaxRecvMsgSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(c.MaxRecvMsgSize)))
	}

	// Keepalive 参数
	if c.KeepaliveParams != nil {
		opts = append(opts, grpc.WithKeepaliveParams(*c.KeepaliveParams))
	}

	// TLS 配置
	if c.EnableTLS {
		if c.TLSConfig != nil && c.TLSConfig.CAFile != "" {
			creds, err := credentials.NewClientTLSFromFile(c.TLSConfig.CAFile, "")
			if err == nil {
				opts = append(opts, grpc.WithTransportCredentials(creds))
			} else {
				opts = append(opts, grpc.WithInsecure())
			}
		} else if c.TLSConfig != nil && c.TLSConfig.InsecureSkipVerify {
			opts = append(opts, grpc.WithInsecure())
		} else {
			opts = append(opts, grpc.WithInsecure())
		}
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	// 合并其他选项
	opts = append(opts, c.DialOptions...)

	return opts
}
