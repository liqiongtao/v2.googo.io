package gooconsul

import (
	"context"
	"sync"

	"github.com/hashicorp/consul/api"
)

var (
	clients     = make(map[string]*Client)
	defaultName = "default"
	mu          sync.RWMutex
)

// Register 注册一个 Consul 客户端（支持多 Client 切换）
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

// Unregister 注销一个 Consul 客户端
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

// GetDefaultClient 获取默认客户端的 Consul 客户端
func GetDefaultClient() (*api.Client, error) {
	client, err := Default()
	if err != nil {
		return nil, err
	}
	return client.Client(), nil
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

// RegisterService 使用默认客户端注册服务
func RegisterService(registration *ServiceRegistration) error {
	client, err := Default()
	if err != nil {
		return err
	}
	return client.RegisterService(registration)
}

// DeregisterService 使用默认客户端注销服务
func DeregisterService(serviceID string) error {
	client, err := Default()
	if err != nil {
		return err
	}
	return client.DeregisterService(serviceID)
}

// ServiceHealth 使用默认客户端获取服务健康状态
func ServiceHealth(serviceName string, passingOnly bool, q *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error) {
	client, err := Default()
	if err != nil {
		return nil, nil, err
	}
	return client.ServiceHealth(serviceName, passingOnly, q)
}

// ServiceNodes 使用默认客户端获取服务节点列表
func ServiceNodes(serviceName string, tag string, passingOnly bool, q *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error) {
	client, err := Default()
	if err != nil {
		return nil, nil, err
	}
	return client.ServiceNodes(serviceName, tag, passingOnly, q)
}

// WatchService 使用默认客户端监听服务变化
func WatchService(ctx context.Context, serviceName string, tag string, passingOnly bool, handler func([]*api.ServiceEntry, error)) error {
	client, err := Default()
	if err != nil {
		return err
	}
	return client.WatchService(ctx, serviceName, tag, passingOnly, handler)
}

// WatchKey 使用默认客户端监听键值变化
func WatchKey(ctx context.Context, key string, handler func(*api.KVPair, error)) error {
	client, err := Default()
	if err != nil {
		return err
	}
	return client.WatchKey(ctx, key, handler)
}

// WatchKeyPrefix 使用默认客户端监听键前缀变化
func WatchKeyPrefix(ctx context.Context, prefix string, handler func(api.KVPairs, error)) error {
	client, err := Default()
	if err != nil {
		return err
	}
	return client.WatchKeyPrefix(ctx, prefix, handler)
}

