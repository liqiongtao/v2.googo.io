package goohttp

import (
	"time"

	"github.com/gin-gonic/gin"
)

type FuncOption func(c *Config)

func (o FuncOption) Apply(c *Config) {
	o(c)
}

// WithAddress 设置服务器地址
func WithAddress(address string) FuncOption {
	return func(c *Config) {
		c.Address = address
	}
}

// WithMode 设置运行模式
func WithMode(mode string) FuncOption {
	return func(c *Config) {
		c.Mode = mode
	}
}

// WithEnableTrace 设置是否启用追踪
func WithEnableTrace(enable bool) FuncOption {
	return func(c *Config) {
		c.EnableTrace = enable
	}
}

// WithEncryption 设置加密配置
func WithEncryption(enable bool, key []byte) FuncOption {
	return func(c *Config) {
		c.Encryption.Enable = enable
		if len(key) > 0 {
			c.Encryption.Key = key
		}
	}
}

// WithEncryptor 设置自定义加密器
func WithEncryptor(encryptor Encryptor, decryptor Decryptor) FuncOption {
	return func(c *Config) {
		c.Encryption.Encryptor = encryptor
		c.Encryption.Decryptor = decryptor
		if encryptor != nil || decryptor != nil {
			c.Encryption.Enable = true
		}
	}
}

// WithCORS 设置CORS配置
func WithCORS(enable bool, origins []string, methods []string, headers []string) FuncOption {
	return func(c *Config) {
		c.CORS.Enable = enable
		if origins != nil {
			c.CORS.AllowOrigins = origins
		}
		if methods != nil {
			c.CORS.AllowMethods = methods
		}
		if headers != nil {
			c.CORS.AllowHeaders = headers
		}
	}
}

// WithCORSAllowCredentials 设置是否允许携带凭证
func WithCORSAllowCredentials(allow bool) FuncOption {
	return func(c *Config) {
		c.CORS.AllowCredentials = allow
	}
}

// WithCORSMaxAge 设置预检请求缓存时间
func WithCORSMaxAge(maxAge time.Duration) FuncOption {
	return func(c *Config) {
		c.CORS.MaxAge = maxAge
	}
}

// WithRateLimit 设置限流配置
func WithRateLimit(enable bool, rate float64, burst int) FuncOption {
	return func(c *Config) {
		c.RateLimit.Enable = enable
		c.RateLimit.Rate = rate
		c.RateLimit.Burst = burst
	}
}

// WithRateLimitKeyFunc 设置限流键生成函数
func WithRateLimitKeyFunc(keyFunc func(*gin.Context) string) FuncOption {
	return func(c *Config) {
		c.RateLimit.KeyFunc = keyFunc
	}
}

// WithLog 设置日志配置
func WithLog(enable bool, logRequestBody bool, logResponseBody bool) FuncOption {
	return func(c *Config) {
		c.Log.Enable = enable
		c.Log.LogRequestBody = logRequestBody
		c.Log.LogResponseBody = logResponseBody
	}
}

// WithLogMaxBodySize 设置日志记录的最大请求体和响应体大小
func WithLogMaxBodySize(maxRequestBodySize, maxResponseBodySize int) FuncOption {
	return func(c *Config) {
		c.Log.MaxRequestBodySize = maxRequestBodySize
		c.Log.MaxResponseBodySize = maxResponseBodySize
	}
}

// WithResponseHook 添加响应钩子
func WithResponseHook(hook ResponseHook) FuncOption {
	return func(c *Config) {
		c.ResponseHooks = append(c.ResponseHooks, hook)
	}
}

// WithGinOptions 设置Gin引擎选项
func WithGinOptions(opts ...gin.Option) FuncOption {
	return func(c *Config) {
		c.GinOptions = append(c.GinOptions, opts...)
	}
}

