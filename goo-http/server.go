package goohttp

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config *Config
	engine *gin.Engine
	server *http.Server
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
	if s.config.EnableRateLimit {
		s.engine.Use(RateLimitMiddleware(s.config.RateLimitConfig))
	}

	// 加解密
	if s.config.EnableEncrypt {
		s.engine.Use(EncryptMiddleware(s.config.Encryptor))
	}

	// 响应钩子
	if s.config.ResponseHook != nil {
		s.engine.Use(ResponseHookMiddleware(s.config.ResponseHook))
	}
}

func (s *Server) Run() error {
	s.server = &http.Server{
		Addr:    s.config.Addr,
		Handler: s.engine,
	}
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

type HandlerFunc func(*Context)

func (s *Server) Get(path string, handlers ...HandlerFunc) {
	s.engine.GET(path, func(c *gin.Context) {
		ctx := &Context{Context: c}
		for _, handler := range handlers {
			handler(ctx)
		}
	})
}
