package gooetcd

import (
	"time"
)

type FuncOption func(c *Config)

func (o FuncOption) Apply(c *Config) {
	o(c)
}

// WithEndpoints 设置端点列表
func WithEndpoints(endpoints []string) FuncOption {
	return func(c *Config) {
		c.Endpoints = endpoints
	}
}

// WithUsername 设置用户名
func WithUsername(username string) FuncOption {
	return func(c *Config) {
		c.Username = username
	}
}

// WithPassword 设置密码
func WithPassword(password string) FuncOption {
	return func(c *Config) {
		c.Password = password
	}
}

// WithDialTimeout 设置连接超时时间
func WithDialTimeout(dialTimeout time.Duration) FuncOption {
	return func(c *Config) {
		c.DialTimeout = dialTimeout
	}
}

// WithAutoSyncInterval 设置自动同步间隔
func WithAutoSyncInterval(autoSyncInterval time.Duration) FuncOption {
	return func(c *Config) {
		c.AutoSyncInterval = autoSyncInterval
		c.AutoSync = autoSyncInterval > 0
	}
}

// WithAutoSync 设置是否启用自动同步
func WithAutoSync(autoSync bool) FuncOption {
	return func(c *Config) {
		c.AutoSync = autoSync
	}
}

// WithMaxCallSendMsgSize 设置最大调用发送大小
func WithMaxCallSendMsgSize(maxCallSendMsgSize int) FuncOption {
	return func(c *Config) {
		c.MaxCallSendMsgSize = maxCallSendMsgSize
	}
}

// WithMaxCallRecvMsgSize 设置最大调用接收大小
func WithMaxCallRecvMsgSize(maxCallRecvMsgSize int) FuncOption {
	return func(c *Config) {
		c.MaxCallRecvMsgSize = maxCallRecvMsgSize
	}
}

// WithEnableCompression 设置是否启用压缩
func WithEnableCompression(enableCompression bool) FuncOption {
	return func(c *Config) {
		c.EnableCompression = enableCompression
	}
}

// WithEnableGRPCDebugLog 设置是否启用 gRPC 调试日志
func WithEnableGRPCDebugLog(enableGRPCDebugLog bool) FuncOption {
	return func(c *Config) {
		c.EnableGRPCDebugLog = enableGRPCDebugLog
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
		c.EnableTLS = tlsConfig != nil
	}
}
