package goohttp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config     *Config
	engine     *gin.Engine
	server     *http.Server
	rateLimits []*RateLimiter // 保存限流器引用，用于优雅关闭
}

func New(opts ...ConfigOption) *Server {
	config := DefaultConfig
	for _, opt := range opts {
		opt.Apply(config)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	server := &Server{
		config: config,
		engine: engine,
	}

	server.setupMiddlewares()

	return server
}

func (s *Server) setupMiddlewares() {
	// TraceId
	s.engine.Use(TraceMiddleware(s.config.TraceIdHeader))

	// 日志
	if s.config.EnableLog {
		s.engine.Use(LogMiddleware(s.config.Logger))
	}

	// CORS
	if s.config.EnableCORS {
		s.engine.Use(CORSMiddleware(s.config.CORSConfig))
	}

	// 限流
	if s.config.EnableRateLimit && len(s.config.RateLimiters) > 0 {
		s.rateLimits = s.config.RateLimiters
		s.engine.Use(RateLimitMiddleware(s.config.RateLimiters))
	}

	// 响应钩子（需要在加密之前，以便解析JSON响应）
	if s.config.ResponseHooks != nil && len(s.config.ResponseHooks) > 0 {
		s.engine.Use(ResponseHookMiddleware(s.config.ResponseHooks))
	}

	// 加解密（最后执行，确保响应钩子能先处理原始响应）
	if s.config.EnableEncrypt && s.config.Encryptor != nil {
		s.engine.Use(EncryptMiddleware(s.config.Encryptor))
	}
}

func (s *Server) Run() error {
	s.server = &http.Server{
		Addr:    s.config.Addr,
		Handler: s.engine,
	}

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		s.Shutdown(ctx)
	}()

	fmt.Printf("Listening on %s\n", s.config.Addr)

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	// 停止所有限流器的清理goroutine
	for _, limiter := range s.rateLimits {
		limiter.Stop()
	}
	return s.server.Shutdown(ctx)
}
