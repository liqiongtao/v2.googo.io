package googrpc

import (
	"sync"
)

var (
	clients     = make(map[string]*Client)
	servers     = make(map[string]*Server)
	defaultClientName = "default"
	defaultServerName = "default"
	mu          sync.RWMutex
)

// RegisterClient 注册一个 gRPC 客户端（支持多 Client 切换）
func RegisterClient(name string, config *Config) error {
	mu.Lock()
	defer mu.Unlock()

	client, err := NewClient(name, config)
	if err != nil {
		return err
	}

	clients[name] = client
	return nil
}

// RegisterDefaultClient 注册默认客户端
func RegisterDefaultClient(config *Config) error {
	return RegisterClient("default", config)
}

// UnregisterClient 注销一个 gRPC 客户端
func UnregisterClient(name string) error {
	mu.Lock()
	defer mu.Unlock()

	client, ok := clients[name]
	if !ok {
		return nil
	}

	delete(clients, name)
	return client.Close()
}

// UnregisterDefaultClient 注销默认客户端
func UnregisterDefaultClient() error {
	return UnregisterClient("default")
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

// SetDefaultClient 设置默认客户端名称
func SetDefaultClient(name string) {
	mu.Lock()
	defer mu.Unlock()
	defaultClientName = name
}

// DefaultClient 获取默认客户端
func DefaultClient() (*Client, error) {
	return GetClient(defaultClientName)
}

// GetDefaultClientConn 获取默认客户端的 gRPC 连接
func GetDefaultClientConn() (*Client, error) {
	return DefaultClient()
}

// CloseAllClients 关闭所有客户端
func CloseAllClients() error {
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

// RegisterServer 注册一个 gRPC 服务端（支持多 Server 切换）
func RegisterServer(name string, config *Config) error {
	mu.Lock()
	defer mu.Unlock()

	server, err := NewServer(name, config)
	if err != nil {
		return err
	}

	servers[name] = server
	return nil
}

// RegisterDefaultServer 注册默认服务端
func RegisterDefaultServer(config *Config) error {
	return RegisterServer("default", config)
}

// UnregisterServer 注销一个 gRPC 服务端
func UnregisterServer(name string) error {
	mu.Lock()
	defer mu.Unlock()

	server, ok := servers[name]
	if !ok {
		return nil
	}

	delete(servers, name)
	return server.Close()
}

// UnregisterDefaultServer 注销默认服务端
func UnregisterDefaultServer() error {
	return UnregisterServer("default")
}

// GetServer 获取指定名称的服务端
func GetServer(name string) (*Server, error) {
	mu.RLock()
	defer mu.RUnlock()

	server, ok := servers[name]
	if !ok {
		return nil, ErrServerNotFound
	}

	return server, nil
}

// SetDefaultServer 设置默认服务端名称
func SetDefaultServer(name string) {
	mu.Lock()
	defer mu.Unlock()
	defaultServerName = name
}

// DefaultServer 获取默认服务端
func DefaultServer() (*Server, error) {
	return GetServer(defaultServerName)
}

// GetDefaultServer 获取默认服务端的 gRPC 服务端
func GetDefaultServer() (*Server, error) {
	return DefaultServer()
}

// CloseAllServers 关闭所有服务端
func CloseAllServers() error {
	mu.Lock()
	defer mu.Unlock()

	var err error
	for name, server := range servers {
		if closeErr := server.Close(); closeErr != nil {
			err = closeErr
		}
		delete(servers, name)
	}

	return err
}

