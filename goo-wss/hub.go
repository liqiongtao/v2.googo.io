package goowss

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub WebSocket 中心管理器
type Hub struct {
	// 房间映射
	rooms map[string]*Room

	// 客户端映射
	clients map[string]*Client

	// 互斥锁
	mu sync.RWMutex

	// 配置
	config *Config

	// 升级器
	upgrader *websocket.Upgrader
}

// NewHub 创建新的 Hub
func NewHub(config *Config) *Hub {
	if config == nil {
		config = DefaultConfig()
	}

	return &Hub{
		rooms:    make(map[string]*Room),
		clients:  make(map[string]*Client),
		config:   config,
		upgrader: config.toUpgrader(),
	}
}

// Upgrade 升级 HTTP 连接为 WebSocket 连接
func (h *Hub) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error) {
	return h.upgrader.Upgrade(w, r, responseHeader)
}

// RegisterClient 注册客户端
func (h *Hub) RegisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.ID] = client
}

// UnregisterClient 注销客户端
func (h *Hub) UnregisterClient(clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client, exists := h.clients[clientID]
	if exists {
		// 如果客户端在房间中，从房间移除
		roomID := client.GetRoomID()
		if roomID != "" {
			if room, ok := h.rooms[roomID]; ok {
				room.RemoveClient(clientID)
			}
		}
		delete(h.clients, clientID)
		client.Close()
	}
}

// GetOrCreateRoom 获取或创建房间
func (h *Hub) GetOrCreateRoom(roomID string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[roomID]
	if !exists {
		room = NewRoom(roomID)
		h.rooms[roomID] = room
	}

	return room
}

// GetRoom 获取房间
func (h *Hub) GetRoom(roomID string) (*Room, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, exists := h.rooms[roomID]
	if !exists {
		return nil, ErrRoomNotFound
	}

	return room, nil
}

// RemoveRoom 移除房间
func (h *Hub) RemoveRoom(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[roomID]
	if exists {
		room.Close()
		delete(h.rooms, roomID)
	}
}

// GetClient 获取客户端
func (h *Hub) GetClient(clientID string) (*Client, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	client, exists := h.clients[clientID]
	if !exists {
		return nil, ErrClientNotFound
	}

	return client, nil
}

// SendToClient 向指定客户端发送消息
func (h *Hub) SendToClient(clientID string, message []byte) error {
	client, err := h.GetClient(clientID)
	if err != nil {
		return err
	}

	return client.Send(message)
}

// BroadcastToRoom 向房间广播消息
func (h *Hub) BroadcastToRoom(roomID string, message []byte) error {
	room, err := h.GetRoom(roomID)
	if err != nil {
		return err
	}

	room.Broadcast(message)
	return nil
}

// BroadcastToRoomExclude 向房间广播消息，排除指定客户端
func (h *Hub) BroadcastToRoomExclude(roomID string, message []byte, excludeClientID string) error {
	room, err := h.GetRoom(roomID)
	if err != nil {
		return err
	}

	room.BroadcastExclude(message, excludeClientID)
	return nil
}

// GetRoomCount 获取房间数量
func (h *Hub) GetRoomCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms)
}

// GetClientCount 获取客户端数量
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

