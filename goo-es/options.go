package gooes

import (
	"time"
)

type FuncOption func(c *Config)

func (o FuncOption) Apply(c *Config) {
	o(c)
}

// WithAddresses 设置地址列表
func WithAddresses(addresses []string) FuncOption {
	return func(c *Config) {
		c.Addresses = addresses
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

// WithCloudID 设置云ID
func WithCloudID(cloudID string) FuncOption {
	return func(c *Config) {
		c.CloudID = cloudID
	}
}

// WithAPIKey 设置 API Key
func WithAPIKey(apiKey string) FuncOption {
	return func(c *Config) {
		c.APIKey = apiKey
	}
}

// WithServiceToken 设置服务令牌
func WithServiceToken(serviceToken string) FuncOption {
	return func(c *Config) {
		c.ServiceToken = serviceToken
	}
}

// WithTimeout 设置连接超时时间
func WithTimeout(timeout time.Duration) FuncOption {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) FuncOption {
	return func(c *Config) {
		c.MaxRetries = maxRetries
	}
}

// WithEnableCompression 设置是否启用压缩
func WithEnableCompression(enableCompression bool) FuncOption {
	return func(c *Config) {
		c.EnableCompression = enableCompression
	}
}

// WithDisableMetaHeader 设置是否禁用元数据
func WithDisableMetaHeader(disableMetaHeader bool) FuncOption {
	return func(c *Config) {
		c.DisableMetaHeader = disableMetaHeader
	}
}

// WithEnableDebugLogger 设置是否启用调试日志
func WithEnableDebugLogger(enableDebugLogger bool) FuncOption {
	return func(c *Config) {
		c.EnableDebugLogger = enableDebugLogger
	}
}
