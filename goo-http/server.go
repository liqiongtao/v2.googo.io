package goohttp

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	goocontext "v2.googo.io/goo-context"
	goolog "v2.googo.io/goo-log"
)

// Server HTTP服务器对象
type Server struct {
	name    string
	config  *Config
	engine  *gin.Engine
	server  *http.Server
	limiter *rate.Limiter
	mu      sync.RWMutex
}

// NewServer 创建新的HTTP服务器
func NewServer(name string, config *Config) (*Server, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证地址
	if config.Address == "" {
		return nil, ErrEmptyAddress
	}

	// 设置Gin模式
	gin.SetMode(config.Mode)

	// 创建Gin引擎
	engine := gin.New()
	if len(config.GinOptions) > 0 {
		// 应用Gin选项（如果有）
		for _, opt := range config.GinOptions {
			// Gin选项通常是函数，这里需要根据实际情况调整
			_ = opt
		}
	}

	// 创建限流器（如果启用）
	var limiter *rate.Limiter
	if config.RateLimit.Enable {
		limiter = rate.NewLimiter(rate.Limit(config.RateLimit.Rate), config.RateLimit.Burst)
	}

	s := &Server{
		name:    name,
		config:  config,
		engine:  engine,
		limiter: limiter,
		server: &http.Server{
			Addr:    config.Address,
			Handler: engine,
		},
	}

	// 设置中间件
	s.setupMiddleware()

	// 记录日志
	if config.Log != nil && config.Log.Enable {
		goolog.InfoF("[goo-http] server '%s' created on %s", name, config.Address)
	}

	return s, nil
}

// Name 获取服务器名称
func (s *Server) Name() string {
	return s.name
}

// Engine 获取Gin引擎
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// setupMiddleware 设置中间件
func (s *Server) setupMiddleware() {
	// 恢复中间件
	s.engine.Use(gin.Recovery())

	// 追踪中间件
	if s.config.EnableTrace {
		s.engine.Use(s.traceMiddleware())
	}

	// 日志中间件
	if s.config.Log != nil && s.config.Log.Enable {
		s.engine.Use(s.logMiddleware())
	}

	// 限流中间件
	if s.config.RateLimit != nil && s.config.RateLimit.Enable {
		s.engine.Use(s.rateLimitMiddleware())
	}

	// CORS中间件
	if s.config.CORS != nil && s.config.CORS.Enable {
		s.engine.Use(s.corsMiddleware())
	}

	// 响应包装中间件（用于捕获响应体）
	s.engine.Use(s.responseWrapperMiddleware())

	// 设置响应钩子到gin.Context
	if len(s.config.ResponseHooks) > 0 {
		s.engine.Use(s.responseHookMiddleware())
	}
}

// traceMiddleware 追踪中间件
func (s *Server) traceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取traceId
		traceId := c.GetHeader("X-Trace-Id")
		if traceId == "" {
			// 生成新的traceId
			ctx := goocontext.Default(c.Request.Context())
			ctx = ctx.WithTraceId()
			traceId = ctx.TraceId()
		}

		// 设置到gin.Context
		c.Set("trace-id", traceId)
		c.Header("X-Trace-Id", traceId)

		// 更新请求上下文
		ctx := goocontext.Default(c.Request.Context())
		ctx = ctx.WithTraceId(traceId)
		c.Request = c.Request.WithContext(ctx.Context)

		// 设置Server实例到上下文
		c.Set("goo-http-server", s)

		c.Next()
	}
}

// logMiddleware 日志中间件
func (s *Server) logMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// 包装响应写入器
		wrapResponseWriter(c)

		c.Next()

		// 计算耗时
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		traceId := getTraceId(c)

		// 记录访问日志
		goolog.WithField("method", method).
			WithField("path", path).
			WithField("status", statusCode).
			WithField("latency", latency.String()).
			WithField("ip", c.ClientIP()).
			WithField("trace-id", traceId).
			InfoF("[goo-http] %s %s %d %s", method, path, statusCode, latency)
	}
}

// rateLimitMiddleware 限流中间件
func (s *Server) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if s.limiter == nil {
			c.Next()
			return
		}

		// 获取限流键
		var key string
		if s.config.RateLimit.KeyFunc != nil {
			key = s.config.RateLimit.KeyFunc(c)
		} else {
			key = c.ClientIP()
		}

		// 检查限流（这里简化处理，实际可能需要按key分别限流）
		if !s.limiter.Allow() {
			Error(c, http.StatusTooManyRequests, "rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}

// corsMiddleware CORS中间件
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		corsConfig := s.config.CORS

		// 检查是否允许该源
		allowed := false
		if len(corsConfig.AllowOrigins) == 0 {
			allowed = true
		} else {
			for _, allowedOrigin := range corsConfig.AllowOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
		}

		if allowed {
			if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			} else if len(corsConfig.AllowOrigins) > 0 && corsConfig.AllowOrigins[0] == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			}

			if corsConfig.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			if len(corsConfig.AllowMethods) > 0 {
				methods := ""
				for i, method := range corsConfig.AllowMethods {
					if i > 0 {
						methods += ", "
					}
					methods += method
				}
				c.Header("Access-Control-Allow-Methods", methods)
			}

			if len(corsConfig.AllowHeaders) > 0 {
				headers := ""
				for i, header := range corsConfig.AllowHeaders {
					if i > 0 {
						headers += ", "
					}
					headers += header
				}
				c.Header("Access-Control-Allow-Headers", headers)
			}

			if len(corsConfig.ExposeHeaders) > 0 {
				headers := ""
				for i, header := range corsConfig.ExposeHeaders {
					if i > 0 {
						headers += ", "
					}
					headers += header
				}
				c.Header("Access-Control-Expose-Headers", headers)
			}

			if corsConfig.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", int(corsConfig.MaxAge.Seconds())))
			}
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// responseWrapperMiddleware 响应包装中间件
func (s *Server) responseWrapperMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		wrapResponseWriter(c)
		c.Next()
	}
}

// responseHookMiddleware 响应钩子中间件
func (s *Server) responseHookMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置响应钩子到gin.Context
		c.Set("response-hooks", s.config.ResponseHooks)
		c.Next()
	}
}

// GET 注册GET路由
func (s *Server) GET(relativePath string, handler HandlerFunc) {
	s.engine.GET(relativePath, Handler(handler))
}

// POST 注册POST路由
func (s *Server) POST(relativePath string, handler HandlerFunc) {
	s.engine.POST(relativePath, Handler(handler))
}

// PUT 注册PUT路由
func (s *Server) PUT(relativePath string, handler HandlerFunc) {
	s.engine.PUT(relativePath, Handler(handler))
}

// DELETE 注册DELETE路由
func (s *Server) DELETE(relativePath string, handler HandlerFunc) {
	s.engine.DELETE(relativePath, Handler(handler))
}

// PATCH 注册PATCH路由
func (s *Server) PATCH(relativePath string, handler HandlerFunc) {
	s.engine.PATCH(relativePath, Handler(handler))
}

// OPTIONS 注册OPTIONS路由
func (s *Server) OPTIONS(relativePath string, handler HandlerFunc) {
	s.engine.OPTIONS(relativePath, Handler(handler))
}

// Any 注册任意方法的路由
func (s *Server) Any(relativePath string, handler HandlerFunc) {
	s.engine.Any(relativePath, Handler(handler))
}

// Use 添加中间件
func (s *Server) Use(middleware ...gin.HandlerFunc) {
	s.engine.Use(middleware...)
}

// Group 创建路由组
func (s *Server) Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup {
	return s.engine.Group(relativePath, handlers...)
}

// Serve 启动服务器（阻塞调用）
func (s *Server) Serve() error {
	if s.config.Log != nil && s.config.Log.Enable {
		goolog.InfoF("[goo-http] server '%s' starting on %s", s.name, s.config.Address)
	}
	return s.server.ListenAndServe()
}

// ServeTLS 启动HTTPS服务器（阻塞调用）
func (s *Server) ServeTLS(certFile, keyFile string) error {
	if s.config.Log != nil && s.config.Log.Enable {
		goolog.InfoF("[goo-http] server '%s' starting TLS on %s", s.name, s.config.Address)
	}
	return s.server.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config.Log != nil && s.config.Log.Enable {
		goolog.InfoF("[goo-http] server '%s' shutting down", s.name)
	}

	// 实现优雅关闭逻辑
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// Close 关闭服务器
func (s *Server) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.config.Log != nil && s.config.Log.Enable {
		goolog.InfoF("[goo-http] server '%s' closed", s.name)
	}

	return s.server.Close()
}
