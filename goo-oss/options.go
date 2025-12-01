package goooss

import (
	"time"
)

type FuncOption func(c *Config)

func (o FuncOption) Apply(c *Config) {
	o(c)
}

// WithAccessKeyID 设置 AccessKeyID
func WithAccessKeyID(accessKeyID string) FuncOption {
	return func(c *Config) {
		c.AccessKeyID = accessKeyID
	}
}

// WithAccessKeySecret 设置 AccessKeySecret
func WithAccessKeySecret(accessKeySecret string) FuncOption {
	return func(c *Config) {
		c.AccessKeySecret = accessKeySecret
	}
}

// WithEndpoint 设置访问域名
func WithEndpoint(endpoint string) FuncOption {
	return func(c *Config) {
		c.Endpoint = endpoint
	}
}

// WithBucket 设置存储桶名称
func WithBucket(bucket string) FuncOption {
	return func(c *Config) {
		c.Bucket = bucket
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

