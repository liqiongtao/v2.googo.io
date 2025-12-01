package goocos

import (
	"time"
)

type FuncOption func(c *Config)

func (o FuncOption) Apply(c *Config) {
	o(c)
}

// WithSecretID 设置 SecretID
func WithSecretID(secretID string) FuncOption {
	return func(c *Config) {
		c.SecretID = secretID
	}
}

// WithSecretKey 设置 SecretKey
func WithSecretKey(secretKey string) FuncOption {
	return func(c *Config) {
		c.SecretKey = secretKey
	}
}

// WithRegion 设置区域
func WithRegion(region string) FuncOption {
	return func(c *Config) {
		c.Region = region
	}
}

// WithBucket 设置存储桶名称
func WithBucket(bucket string) FuncOption {
	return func(c *Config) {
		c.Bucket = bucket
	}
}

// WithScheme 设置协议
func WithScheme(scheme string) FuncOption {
	return func(c *Config) {
		c.Scheme = scheme
	}
}

// WithBaseURL 设置基础 URL
func WithBaseURL(baseURL string) FuncOption {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

// WithTimeout 设置连接超时时间
func WithTimeout(timeout time.Duration) FuncOption {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithDebug 设置是否启用调试日志
func WithDebug(debug bool) FuncOption {
	return func(c *Config) {
		c.Debug = debug
	}
}

// WithSTS 设置 STS 临时密钥配置
func WithSTS(sts *STSConfig) FuncOption {
	return func(c *Config) {
		c.STS = sts
	}
}

