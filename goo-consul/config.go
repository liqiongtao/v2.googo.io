package gooconsul

import (
	"time"

	"github.com/hashicorp/consul/api"
)

// Config Consul 配置
type Config struct {
	// 地址，格式: "localhost:8500"
	Address string

	// 数据中心
	Datacenter string

	// 令牌
	Token string

	// 命名空间（Consul Enterprise 功能）
	Namespace string

	// 连接超时时间，默认 5 秒
	Timeout time.Duration

	// TLS 配置
	TLSConfig *api.TLSConfig

	// HTTP 客户端配置
	HttpClient *api.HttpClientConfig
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config {
	c := &Config{
		Address:  "localhost:8500",
		Timeout:  5 * time.Second,
		TLSConfig: nil,
		HttpClient: nil,
	}
	for _, opt := range opts {
		opt.Apply(c)
	}
	return c
}

// toConsulConfig 转换为 Consul API 的 Config
func (c *Config) toConsulConfig() (*api.Config, error) {
	cfg := api.DefaultConfig()

	if c.Address != "" {
		cfg.Address = c.Address
	}

	if c.Datacenter != "" {
		cfg.Datacenter = c.Datacenter
	}

	if c.Token != "" {
		cfg.Token = c.Token
	}

	if c.Namespace != "" {
		cfg.Namespace = c.Namespace
	}

	if c.Timeout > 0 {
		cfg.HttpClient.Timeout = c.Timeout
	}

	if c.TLSConfig != nil {
		cfg.TLSConfig = *c.TLSConfig
	}

	if c.HttpClient != nil {
		if c.HttpClient.Timeout > 0 {
			cfg.HttpClient.Timeout = c.HttpClient.Timeout
		}
		if c.HttpClient.TLSConfig != nil {
			cfg.HttpClient.TLSConfig = *c.HttpClient.TLSConfig
		}
	}

	return cfg, nil
}

