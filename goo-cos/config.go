package goocos

import (
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// Config COS 配置
type Config struct {
	// SecretID 腾讯云 SecretID
	SecretID string

	// SecretKey 腾讯云 SecretKey
	SecretKey string

	// Region 区域，例如: ap-beijing
	Region string

	// Bucket 存储桶名称
	Bucket string

	// Scheme 协议，http 或 https，默认 https
	Scheme string

	// BaseURL 基础 URL，如果设置则使用此 URL，否则根据 Region 和 Bucket 自动生成
	BaseURL string

	// 连接超时时间，默认 30 秒
	Timeout time.Duration

	// 是否启用调试日志，默认 false
	Debug bool

	// STS 临时密钥配置（可选）
	STS *STSConfig
}

// STSConfig STS 临时密钥配置
type STSConfig struct {
	// SecretID 临时密钥 SecretID
	SecretID string

	// SecretKey 临时密钥 SecretKey
	SecretKey string

	// SessionToken 临时密钥 SessionToken
	SessionToken string

	// ExpiredTime 过期时间
	ExpiredTime time.Time
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config {
	c := &Config{
		Scheme:  "https",
		Timeout: 30 * time.Second,
		Debug:   false,
	}
	for _, opt := range opts {
		opt.Apply(c)
	}
	return c
}

// toCOSOptions 转换为 COS SDK 的配置
func (c *Config) toCOSOptions() (*cos.BaseURL, *cos.ClientOptions) {
	// 构建 BaseURL
	var baseURL *cos.BaseURL
	if c.BaseURL != "" {
		baseURL = &cos.BaseURL{
			BucketURL: c.BaseURL,
		}
	} else {
		// 根据 Region 和 Bucket 自动生成
		scheme := c.Scheme
		if scheme == "" {
			scheme = "https"
		}
		// 构建 bucket URL: https://bucket-name.cos.region.myqcloud.com
		bucketURL := scheme + "://" + c.Bucket + ".cos." + c.Region + ".myqcloud.com"
		baseURL = &cos.BaseURL{
			BucketURL: bucketURL,
		}
	}

	// 构建 ClientOptions
	opts := &cos.ClientOptions{
		Timeout: c.Timeout,
	}

	// 设置默认值
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	return baseURL, opts
}

// getCredentials 获取认证信息
func (c *Config) getCredentials() (string, string, string) {
	// 如果使用 STS 临时密钥
	if c.STS != nil && c.STS.SecretID != "" && c.STS.SecretKey != "" {
		// 检查是否过期
		if !c.STS.ExpiredTime.IsZero() && time.Now().After(c.STS.ExpiredTime) {
			// 已过期，使用永久密钥
			return c.SecretID, c.SecretKey, ""
		}
		return c.STS.SecretID, c.STS.SecretKey, c.STS.SessionToken
	}

	// 使用永久密钥
	return c.SecretID, c.SecretKey, ""
}

