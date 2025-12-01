package gooconsul

import (
	"time"

	"github.com/hashicorp/consul/api"
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

// WithDatacenter 设置数据中心
func WithDatacenter(datacenter string) FuncOption {
	return func(c *Config) {
		c.Datacenter = datacenter
	}
}

// WithToken 设置令牌
func WithToken(token string) FuncOption {
	return func(c *Config) {
		c.Token = token
	}
}

// WithNamespace 设置命名空间
func WithNamespace(namespace string) FuncOption {
	return func(c *Config) {
		c.Namespace = namespace
	}
}

// WithTimeout 设置连接超时时间
func WithTimeout(timeout time.Duration) FuncOption {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithTLSConfig 设置 TLS 配置
func WithTLSConfig(tlsConfig *api.TLSConfig) FuncOption {
	return func(c *Config) {
		c.TLSConfig = tlsConfig
	}
}

// WithHttpClient 设置 HTTP 客户端配置
func WithHttpClient(httpClient *api.HttpClientConfig) FuncOption {
	return func(c *Config) {
		c.HttpClient = httpClient
	}
}

