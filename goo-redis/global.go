package gooredis

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
)

var (
	clients     = make(map[string]*Client)
	defaultName = "default"
	mu          sync.RWMutex
)

// Register 注册一个 Redis 客户端（支持多 Client 切换）
func Register(name string, config *Config) error {
	mu.Lock()
	defer mu.Unlock()

	client, err := NewClient(name, config)
	if err != nil {
		return err
	}

	clients[name] = client
	return nil
}

func RegisterDefault(config *Config) error {
	return Register("default", config)
}

// Unregister 注销一个 Redis 客户端
func Unregister(name string) error {
	mu.Lock()
	defer mu.Unlock()

	client, ok := clients[name]
	if !ok {
		return nil
	}

	delete(clients, name)
	return client.Close()
}

func UnregisterDefault() error {
	return Unregister("default")
}

// GetClient 获取指定名称的客户端
func GetClient(name string) (*Client, error) {
	mu.RLock()
	defer mu.RUnlock()

	client, ok := clients[name]
	if !ok {
		return nil, ErrClientNotFound
	}

	return client, nil
}

// Default 获取默认客户端
func Default() (*Client, error) {
	return GetClient(defaultName)
}

// DefaultDB 获取默认客户端的指定数据库（支持多 db 选择）
func DefaultDB(db int) (*redis.Client, error) {
	client, err := Default()
	if err != nil {
		return nil, err
	}
	return client.DB(db), nil
}

// GetDefaultClient 获取默认客户端的默认数据库
func GetDefaultClient() (*redis.Client, error) {
	client, err := Default()
	if err != nil {
		return nil, err
	}
	return client.Client(), nil
}

// DB 获取默认客户端的指定数据库
func DB(db int) (*redis.Client, error) {
	return DefaultDB(db)
}

// CloseAll 关闭所有客户端
func CloseAll() error {
	mu.Lock()
	defer mu.Unlock()

	var err error
	for name, client := range clients {
		if closeErr := client.Close(); closeErr != nil {
			err = closeErr
		}
		delete(clients, name)
	}

	return err
}

// Ping 测试默认客户端连接
func Ping(ctx context.Context) error {
	client, err := Default()
	if err != nil {
		return err
	}
	return client.Ping(ctx)
}
