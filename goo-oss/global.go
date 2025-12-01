package goooss

import (
	"context"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var (
	clients     = make(map[string]*Client)
	defaultName = "default"
	mu          sync.RWMutex
)

// Register 注册一个 OSS 客户端（支持多 Client 切换）
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

// RegisterDefault 注册默认客户端
func RegisterDefault(config *Config) error {
	return Register("default", config)
}

// Unregister 注销一个 OSS 客户端
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

// UnregisterDefault 注销默认客户端
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

// SetDefault 设置默认客户端名称
func SetDefault(name string) {
	mu.Lock()
	defer mu.Unlock()
	defaultName = name
}

// Default 获取默认客户端
func Default() (*Client, error) {
	return GetClient(defaultName)
}

// GetDefaultClient 获取默认客户端的 OSS 客户端
func GetDefaultClient() (*oss.Client, error) {
	client, err := Default()
	if err != nil {
		return nil, err
	}
	return client.Client(), nil
}

// GetDefaultBucket 获取默认客户端的存储桶对象
func GetDefaultBucket() (*oss.Bucket, error) {
	client, err := Default()
	if err != nil {
		return nil, err
	}
	return client.Bucket(), nil
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

