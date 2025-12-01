package goohttp

import (
	"crypto/aes"
	"crypto/cipher"
	"time"

	"github.com/gin-gonic/gin"
)

// EncryptionConfig 加密配置
type EncryptionConfig struct {
	// 是否启用加密
	Enable bool
	// 加密密钥（AES-256需要32字节）
	Key []byte
	// 加密算法，默认 AES-256-GCM
	Algorithm string
	// 加密器接口
	Encryptor Encryptor
	// 解密器接口
	Decryptor Decryptor
}

// CORSConfig CORS配置
type CORSConfig struct {
	// 是否启用CORS
	Enable bool
	// 允许的源，空表示允许所有
	AllowOrigins []string
	// 允许的方法
	AllowMethods []string
	// 允许的请求头
	AllowHeaders []string
	// 是否允许携带凭证
	AllowCredentials bool
	// 暴露的响应头
	ExposeHeaders []string
	// 预检请求缓存时间
	MaxAge time.Duration
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 是否启用限流
	Enable bool
	// 每秒允许的请求数
	Rate float64
	// 突发请求数
	Burst int
	// 限流键生成函数，nil表示使用IP地址
	KeyFunc func(*gin.Context) string
}

// LogConfig 日志配置
type LogConfig struct {
	// 是否启用日志
	Enable bool
	// 是否记录请求体
	LogRequestBody bool
	// 是否记录响应体
	LogResponseBody bool
	// 请求体最大记录长度（字节），0表示不限制
	MaxRequestBodySize int
	// 响应体最大记录长度（字节），0表示不限制
	MaxResponseBodySize int
}

// Config HTTP服务器配置
type Config struct {
	// 服务器地址，格式: host:port
	Address string

	// 模式：debug, release, test
	Mode string

	// 加密配置
	Encryption *EncryptionConfig

	// CORS配置
	CORS *CORSConfig

	// 限流配置
	RateLimit *RateLimitConfig

	// 日志配置
	Log *LogConfig

	// 是否启用追踪（traceId）
	EnableTrace bool

	// 响应钩子函数列表
	ResponseHooks []ResponseHook

	// Gin引擎选项
	GinOptions []gin.Option
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config {
	c := &Config{
		Address:     ":8080",
		Mode:        gin.ReleaseMode,
		EnableTrace: true,
		Encryption: &EncryptionConfig{
			Enable:    false,
			Algorithm: "AES-256-GCM",
		},
		CORS: &CORSConfig{
			Enable:          true,
			AllowOrigins:   []string{"*"},
			AllowMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowHeaders:   []string{"*"},
			AllowCredentials: false,
			MaxAge:         12 * time.Hour,
		},
		RateLimit: &RateLimitConfig{
			Enable: false,
			Rate:   100,
			Burst:  200,
		},
		Log: &LogConfig{
			Enable:             true,
			LogRequestBody:     false,
			LogResponseBody:    false,
			MaxRequestBodySize: 1024,
			MaxResponseBodySize: 1024,
		},
		ResponseHooks: make([]ResponseHook, 0),
		GinOptions:    make([]gin.Option, 0),
	}

	for _, opt := range opts {
		opt.Apply(c)
	}

	// 如果启用了加密但没有提供加密器，使用默认的AES-256-GCM
	if c.Encryption.Enable && c.Encryption.Encryptor == nil {
		if len(c.Encryption.Key) == 0 {
			// 如果没有提供密钥，生成一个默认密钥（仅用于开发环境）
			c.Encryption.Key = make([]byte, 32) // AES-256需要32字节
		}
		if len(c.Encryption.Key) != 32 {
			// 密钥长度必须是32字节
			c.Encryption.Key = make([]byte, 32)
		}
		block, err := aes.NewCipher(c.Encryption.Key)
		if err == nil {
			c.Encryption.Encryptor = &AESGCMEncryptor{block: block}
			c.Encryption.Decryptor = &AESGCMDecryptor{block: block}
		}
	}

	return c
}

