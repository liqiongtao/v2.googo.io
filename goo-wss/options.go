package goowss

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type FuncOption func(c *Config)

func (o FuncOption) Apply(c *Config) {
	o(c)
}

// WithReadBufferSize 设置读取缓冲区大小
func WithReadBufferSize(size int) FuncOption {
	return func(c *Config) {
		c.ReadBufferSize = size
	}
}

// WithWriteBufferSize 设置写入缓冲区大小
func WithWriteBufferSize(size int) FuncOption {
	return func(c *Config) {
		c.WriteBufferSize = size
	}
}

// WithEnableCompression 设置是否启用压缩
func WithEnableCompression(enable bool) FuncOption {
	return func(c *Config) {
		c.EnableCompression = enable
	}
}

// WithCheckOrigin 设置来源检查函数
func WithCheckOrigin(checkOrigin func(r *http.Request) bool) FuncOption {
	return func(c *Config) {
		c.CheckOrigin = checkOrigin
	}
}

// WithHandshakeTimeout 设置握手超时时间
func WithHandshakeTimeout(timeout time.Duration) FuncOption {
	return func(c *Config) {
		c.HandshakeTimeout = timeout
	}
}

// WithReadTimeout 设置读取超时时间
func WithReadTimeout(timeout time.Duration) FuncOption {
	return func(c *Config) {
		c.ReadTimeout = timeout
	}
}

// WithWriteTimeout 设置写入超时时间
func WithWriteTimeout(timeout time.Duration) FuncOption {
	return func(c *Config) {
		c.WriteTimeout = timeout
	}
}

// WithPongWait 设置 Pong 等待时间
func WithPongWait(duration time.Duration) FuncOption {
	return func(c *Config) {
		c.PongWait = duration
	}
}

// WithPingPeriod 设置 Ping 周期
func WithPingPeriod(duration time.Duration) FuncOption {
	return func(c *Config) {
		c.PingPeriod = duration
	}
}

// WithMaxMessageSize 设置最大消息大小
func WithMaxMessageSize(size int64) FuncOption {
	return func(c *Config) {
		c.MaxMessageSize = size
	}
}

// WithUpgrader 设置自定义升级器
func WithUpgrader(upgrader *websocket.Upgrader) FuncOption {
	return func(c *Config) {
		c.Upgrader = upgrader
	}
}

