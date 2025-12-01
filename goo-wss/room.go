package goowss

import (
	"sync"
)

// Room WebSocket 房间
type Room struct {
	// 房间 ID
	ID string

	// 客户端映射
	clients map[string]*Client

	// 互斥锁
	mu sync.RWMutex
}

// NewRoom 创建新的房间
func NewRoom(id string) *Room {
	return &Room{
		ID:      id,
		clients: make(map[string]*Client),
	}
}

// AddClient 添加客户端到房间
func (r *Room) AddClient(client *Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查客户端是否已在房间中
	if client.GetRoomID() == r.ID {
		return ErrClientAlreadyInRoom
	}

	// 如果客户端在其他房间，先从其他房间移除
	if client.GetRoomID() != "" {
		// 这里假设有全局的房间管理器，实际使用时需要处理
	}

	// 添加到房间
	r.clients[client.ID] = client
	client.SetRoomID(r.ID)

	return nil
}

// RemoveClient 从房间移除客户端
func (r *Room) RemoveClient(clientID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	client, exists := r.clients[clientID]
	if !exists {
		return ErrClientNotFound
	}

	delete(r.clients, clientID)
	client.SetRoomID("")

	return nil
}

// GetClient 获取房间中的客户端
func (r *Room) GetClient(clientID string) (*Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, exists := r.clients[clientID]
	if !exists {
		return nil, ErrClientNotFound
	}

	return client, nil
}

// Broadcast 向房间内所有客户端广播消息
func (r *Room) Broadcast(message []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, client := range r.clients {
		if !client.IsClosed() {
			client.Send(message)
		}
	}
}

// BroadcastExclude 向房间内所有客户端广播消息，排除指定客户端
func (r *Room) BroadcastExclude(message []byte, excludeClientID string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for id, client := range r.clients {
		if id != excludeClientID && !client.IsClosed() {
			client.Send(message)
		}
	}
}

// SendToClient 向房间内指定客户端发送消息
func (r *Room) SendToClient(clientID string, message []byte) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	client, exists := r.clients[clientID]
	if !exists {
		return ErrClientNotFound
	}

	if client.IsClosed() {
		return ErrConnectionClosed
	}

	return client.Send(message)
}

// GetClientCount 获取房间内客户端数量
func (r *Room) GetClientCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

// GetClients 获取房间内所有客户端 ID
func (r *Room) GetClients() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.clients))
	for id := range r.clients {
		ids = append(ids, id)
	}
	return ids
}

// Close 关闭房间（移除所有客户端）
func (r *Room) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, client := range r.clients {
		client.SetRoomID("")
	}
	r.clients = make(map[string]*Client)
}

