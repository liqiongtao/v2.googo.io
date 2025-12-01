package goowss

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client WebSocket 客户端
type Client struct {
	// 客户端唯一标识
	ID string

	// WebSocket 连接
	conn *websocket.Conn

	// 发送消息通道
	send chan []byte

	// 配置
	config *Config

	// 上下文
	ctx    context.Context
	cancel context.CancelFunc

	// 互斥锁
	mu sync.RWMutex

	// 是否已关闭
	closed bool

	// 所属房间 ID
	roomID string
}

// NewClient 创建新的客户端
func NewClient(id string, conn *websocket.Conn, config *Config) *Client {
	if config == nil {
		config = DefaultConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		ID:     id,
		conn:   conn,
		send:   make(chan []byte, 256),
		config: config,
		ctx:    ctx,
		cancel: cancel,
		closed: false,
		roomID: "",
	}

	// 设置读取限制
	conn.SetReadLimit(config.MaxMessageSize)

	// 设置 Pong 处理器
	conn.SetReadDeadline(time.Now().Add(config.PongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(config.PongWait))
		return nil
	})

	return client
}

// ID 获取客户端 ID
func (c *Client) GetID() string {
	return c.ID
}

// RoomID 获取客户端所属房间 ID
func (c *Client) GetRoomID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.roomID
}

// SetRoomID 设置客户端所属房间 ID
func (c *Client) SetRoomID(roomID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.roomID = roomID
}

// Send 发送消息给客户端
func (c *Client) Send(message []byte) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrConnectionClosed
	}
	c.mu.RUnlock()

	select {
	case c.send <- message:
		return nil
	case <-c.ctx.Done():
		return ErrConnectionClosed
	default:
		return ErrConnectionClosed
	}
}

// ReadPump 读取消息循环
func (c *Client) ReadPump(onMessage func(*Client, []byte), onError func(*Client, error)) {
	defer func() {
		c.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			// 设置读取超时
			if c.config.ReadTimeout > 0 {
				c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))
			}

			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					if onError != nil {
						onError(c, err)
					}
				}
				return
			}

			if onMessage != nil {
				onMessage(c, message)
			}
		}
	}
}

// WritePump 写入消息循环
func (c *Client) WritePump() {
	ticker := time.NewTicker(c.config.PingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case message, ok := <-c.send:
			// 设置写入超时
			if c.config.WriteTimeout > 0 {
				c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
			} else {
				c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			}

			if !ok {
				// 通道已关闭
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 批量发送队列中的消息
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			// 发送 Ping
			if c.config.WriteTimeout > 0 {
				c.conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
			} else {
				c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	c.cancel()
	close(c.send)
	return c.conn.Close()
}

// IsClosed 检查客户端是否已关闭
func (c *Client) IsClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

