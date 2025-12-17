package goohttp

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	DefaultCORSConfig = &CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Trace-Id", "User-Agent", "X-Timestamp", "X-AppId", "X-Sign"},
		ExposeHeaders:    []string{"X-Trace-Id"},
		AllowCredentials: false,
		MaxAge:           86400,
	}
)

type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins" json:"allow_origins"`         // 允许的源
	AllowMethods     []string `yaml:"allow_methods" json:"allow_methods"`         // 允许的方法
	AllowHeaders     []string `yaml:"allow_headers" json:"allow_headers"`         // 允许的请求头
	ExposeHeaders    []string `yaml:"expose_headers" json:"expose_headers"`       // 暴露的响应头
	AllowCredentials bool     `yaml:"allow_credentials" json:"allow_credentials"` // 是否允许凭证
	MaxAge           int      `yaml:"max_age" json:"max_age"`                     // 预检请求缓存时间（秒）
}

func CORSMiddleware(config *CORSConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultCORSConfig
	}

	allowMethods := strings.Join(config.AllowMethods, ", ")
	allowHeaders := strings.Join(config.AllowHeaders, ", ")
	exposeHeaders := strings.Join(config.ExposeHeaders, ", ")

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			if isOriginAllowed(origin, config.AllowOrigins) {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Methods", allowMethods)
				c.Header("Access-Control-Allow-Headers", allowHeaders)
				if config.AllowCredentials {
					c.Header("Access-Control-Allow-Credentials", "true")
				}
				if config.MaxAge > 0 {
					c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
				}
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 处理实际请求
		if isOriginAllowed(origin, config.AllowOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Expose-Headers", exposeHeaders)
			if config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		c.Next()
	}
}

// 检查源是否被允许
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
	}

	return false
}
