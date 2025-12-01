package goowss

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// Config WebSocket 配置
type Config struct {
	// 读取缓冲区大小（字节），默认 4096
	ReadBufferSize int

	// 写入缓冲区大小（字节），默认 4096
	WriteBufferSize int

	// 允许跨域请求，默认 false
	EnableCompression bool

	// 检查来源，默认 false
	CheckOrigin func(r *http.Request) bool

	// 握手超时时间，默认 10 秒
	HandshakeTimeout time.Duration

	// 读取超时时间，默认 60 秒
	ReadTimeout time.Duration

	// 写入超时时间，默认 10 秒
	WriteTimeout time.Duration

	// Pong 等待时间，默认 60 秒
	PongWait time.Duration

	// Ping 周期，默认 54 秒（PongWait 的 90%）
	PingPeriod time.Duration

	// 最大消息大小（字节），默认 512KB
	MaxMessageSize int64

	// 升级器选项
	Upgrader *websocket.Upgrader
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config {
	c := &Config{
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		EnableCompression: false,
		CheckOrigin:       nil,
		HandshakeTimeout:  10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      10 * time.Second,
		PongWait:          60 * time.Second,
		PingPeriod:        54 * time.Second,
		MaxMessageSize:    512 * 1024, // 512KB
		Upgrader:          nil,
	}
	for _, opt := range opts {
		opt.Apply(c)
	}
	return c
}

// toUpgrader 转换为 websocket.Upgrader
func (c *Config) toUpgrader() *websocket.Upgrader {
	if c.Upgrader != nil {
		return c.Upgrader
	}

	upgrader := &websocket.Upgrader{
		ReadBufferSize:  c.ReadBufferSize,
		WriteBufferSize: c.WriteBufferSize,
		CheckOrigin:     c.CheckOrigin,
	}

	if c.EnableCompression {
		upgrader.EnableCompression = true
	}

	return upgrader
}

