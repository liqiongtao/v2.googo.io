package goooss

import (
	"time"
)

// Config OSS 配置
type Config struct {
	// AccessKeyID 阿里云 AccessKeyID
	AccessKeyID string

	// AccessKeySecret 阿里云 AccessKeySecret
	AccessKeySecret string

	// Endpoint 访问域名，例如: oss-cn-hangzhou.aliyuncs.com
	Endpoint string

	// Bucket 存储桶名称
	Bucket string

	// 连接超时时间，默认 30 秒
	Timeout time.Duration

	// 是否启用调试日志，默认 false
	Debug bool

	// STS 临时密钥配置（可选）
	STS *STSConfig
}

// STSConfig STS 临时密钥配置
type STSConfig struct {
	// AccessKeyID 临时密钥 AccessKeyID
	AccessKeyID string

	// AccessKeySecret 临时密钥 AccessKeySecret
	AccessKeySecret string

	// SecurityToken 临时密钥 SecurityToken
	SecurityToken string

	// ExpiredTime 过期时间
	ExpiredTime time.Time
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config {
	c := &Config{
		Timeout: 30 * time.Second,
		Debug:   false,
	}
	for _, opt := range opts {
		opt.Apply(c)
	}
	return c
}

// getCredentials 获取认证信息
func (c *Config) getCredentials() (string, string, string) {
	// 如果使用 STS 临时密钥
	if c.STS != nil && c.STS.AccessKeyID != "" && c.STS.AccessKeySecret != "" {
		// 检查是否过期
		if !c.STS.ExpiredTime.IsZero() && time.Now().After(c.STS.ExpiredTime) {
			// 已过期，使用永久密钥
			return c.AccessKeyID, c.AccessKeySecret, ""
		}
		return c.STS.AccessKeyID, c.STS.AccessKeySecret, c.STS.SecurityToken
	}

	// 使用永久密钥
	return c.AccessKeyID, c.AccessKeySecret, ""
}

