package gooes

import (
	"time"

	"github.com/elastic/go-elasticsearch/v7"
)

// Config Elasticsearch 配置
type Config struct {
	// 地址列表，格式: ["http://localhost:9200"]
	Addresses []string

	// 用户名
	Username string

	// 密码
	Password string

	// 云ID（Elastic Cloud）
	CloudID string

	// API Key
	APIKey string

	// 服务令牌
	ServiceToken string

	// 连接超时时间，默认 5 秒
	Timeout time.Duration

	// 最大重试次数，默认 3
	MaxRetries int

	// 是否启用压缩，默认 false
	EnableCompression bool

	// 是否禁用元数据，默认 false
	DisableMetaHeader bool

	// 是否启用调试日志，默认 false
	EnableDebugLogger bool
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config {
	c := &Config{
		Addresses:         []string{"http://localhost:9200"},
		Timeout:           5 * time.Second,
		MaxRetries:        3,
		EnableCompression: false,
		DisableMetaHeader: false,
		EnableDebugLogger: false,
	}
	for _, opt := range opts {
		opt.Apply(c)
	}
	return c
}

// toElasticsearchConfig 转换为 go-elasticsearch 的 Config
func (c *Config) toElasticsearchConfig() elasticsearch.Config {
	cfg := elasticsearch.Config{
		Addresses:         c.Addresses,
		Username:          c.Username,
		Password:          c.Password,
		CloudID:           c.CloudID,
		APIKey:            c.APIKey,
		ServiceToken:      c.ServiceToken,
		MaxRetries:        c.MaxRetries,
		EnableCompression: c.EnableCompression,
		DisableMetaHeader: c.DisableMetaHeader,
	}

	// 设置默认值
	if len(cfg.Addresses) == 0 {
		cfg.Addresses = []string{"http://localhost:9200"}
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}

	// 注意: Timeout 需要通过 Transport 配置，这里保留在 Config 中供将来使用

	return cfg
}
