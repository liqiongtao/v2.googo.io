package googrpc

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type FuncOption func(c *Config)

func (o FuncOption) Apply(c *Config) {
	o(c)
}

// WithAddress 设置地址
func WithAddress(address string) FuncOption {
	return func(c *Config) {
		c.Address = address
	}
}

// WithServiceName 设置服务名称
func WithServiceName(serviceName string) FuncOption {
	return func(c *Config) {
		c.ServiceName = serviceName
	}
}

// WithRegistryType 设置注册中心类型
func WithRegistryType(registryType RegistryType) FuncOption {
	return func(c *Config) {
		c.RegistryType = registryType
	}
}

// WithEtcdConfig 设置 etcd 配置
func WithEtcdConfig(etcdConfig *EtcdRegistryConfig) FuncOption {
	return func(c *Config) {
		c.EtcdConfig = etcdConfig
	}
}

// WithConsulConfig 设置 consul 配置
func WithConsulConfig(consulConfig *ConsulRegistryConfig) FuncOption {
	return func(c *Config) {
		c.ConsulConfig = consulConfig
	}
}

// WithEnableTLS 设置是否启用 TLS
func WithEnableTLS(enableTLS bool) FuncOption {
	return func(c *Config) {
		c.EnableTLS = enableTLS
	}
}

// WithTLSConfig 设置 TLS 配置
func WithTLSConfig(tlsConfig *TLSConfig) FuncOption {
	return func(c *Config) {
		c.TLSConfig = tlsConfig
	}
}

// WithTimeout 设置连接超时时间
func WithTimeout(timeout time.Duration) FuncOption {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithEnableLog 设置是否启用日志
func WithEnableLog(enableLog bool) FuncOption {
	return func(c *Config) {
		c.EnableLog = enableLog
	}
}

// WithEnableTrace 设置是否启用追踪
func WithEnableTrace(enableTrace bool) FuncOption {
	return func(c *Config) {
		c.EnableTrace = enableTrace
	}
}

// WithMaxSendMsgSize 设置最大发送消息大小
func WithMaxSendMsgSize(size int) FuncOption {
	return func(c *Config) {
		c.MaxSendMsgSize = size
	}
}

// WithMaxRecvMsgSize 设置最大接收消息大小
func WithMaxRecvMsgSize(size int) FuncOption {
	return func(c *Config) {
		c.MaxRecvMsgSize = size
	}
}

// WithKeepaliveParams 设置 Keepalive 参数
func WithKeepaliveParams(params keepalive.ClientParameters) FuncOption {
	return func(c *Config) {
		c.KeepaliveParams = &params
	}
}

// WithKeepaliveEnforcementPolicy 设置 Keepalive 强制策略
func WithKeepaliveEnforcementPolicy(policy keepalive.EnforcementPolicy) FuncOption {
	return func(c *Config) {
		c.KeepaliveEnforcementPolicy = &policy
	}
}

// WithServerOptions 设置服务端选项
func WithServerOptions(options ...grpc.ServerOption) FuncOption {
	return func(c *Config) {
		c.Options = append(c.Options, options...)
	}
}

// WithDialOptions 设置客户端拨号选项
func WithDialOptions(options ...grpc.DialOption) FuncOption {
	return func(c *Config) {
		c.DialOptions = append(c.DialOptions, options...)
	}
}

