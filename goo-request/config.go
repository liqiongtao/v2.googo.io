package goorequest

import (
	"crypto/tls"
	"time"
)

// Config HTTP请求配置
type Config struct {
	// 基础URL，所有请求会基于此URL
	BaseURL string

	// 默认超时时间，默认 30 秒
	Timeout time.Duration

	// TLS配置
	TLS *tls.Config

	// 默认请求头
	Headers map[string]string

	// 是否跳过TLS证书验证（仅用于开发环境）
	InsecureSkipVerify bool

	// 最大重试次数，默认 0（不重试）
	MaxRetries int

	// 重试间隔，默认 1 秒
	RetryInterval time.Duration
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config {
	c := &Config{
		Timeout:       30 * time.Second,
		Headers:       make(map[string]string),
		MaxRetries:    0,
		RetryInterval: 1 * time.Second,
	}
	for _, opt := range opts {
		opt.Apply(c)
	}
	return c
}

