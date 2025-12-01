package goorequest

import (
	"crypto/tls"
	"time"
)

type FuncOption func(c *Config)

func (o FuncOption) Apply(c *Config) {
	o(c)
}

// WithBaseURL 设置基础URL
func WithBaseURL(baseURL string) FuncOption {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) FuncOption {
	return func(c *Config) {
		c.Timeout = timeout
	}
}

// WithTLS 设置TLS配置
func WithTLS(tls *tls.Config) FuncOption {
	return func(c *Config) {
		c.TLS = tls
	}
}

// WithHeaders 设置默认请求头
func WithHeaders(headers map[string]string) FuncOption {
	return func(c *Config) {
		if c.Headers == nil {
			c.Headers = make(map[string]string)
		}
		for k, v := range headers {
			c.Headers[k] = v
		}
	}
}

// WithHeader 设置单个请求头
func WithHeader(key, value string) FuncOption {
	return func(c *Config) {
		if c.Headers == nil {
			c.Headers = make(map[string]string)
		}
		c.Headers[key] = value
	}
}

// WithInsecureSkipVerify 设置是否跳过TLS证书验证
func WithInsecureSkipVerify(skip bool) FuncOption {
	return func(c *Config) {
		c.InsecureSkipVerify = skip
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) FuncOption {
	return func(c *Config) {
		c.MaxRetries = maxRetries
	}
}

// WithRetryInterval 设置重试间隔
func WithRetryInterval(interval time.Duration) FuncOption {
	return func(c *Config) {
		c.RetryInterval = interval
	}
}

