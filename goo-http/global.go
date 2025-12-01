package goohttp

import (
	"sync"
)

var (
	servers     = make(map[string]*Server)
	defaultName = "default"
	mu          sync.RWMutex
)

// Register 注册一个HTTP服务器（支持多 Server 切换）
func Register(name string, config *Config) error {
	mu.Lock()
	defer mu.Unlock()

	server, err := NewServer(name, config)
	if err != nil {
		return err
	}

	servers[name] = server
	return nil
}

// RegisterDefault 注册默认服务器
func RegisterDefault(config *Config) error {
	return Register("default", config)
}

// Unregister 注销一个服务器
func Unregister(name string) error {
	mu.Lock()
	defer mu.Unlock()

	server, ok := servers[name]
	if !ok {
		return nil
	}

	delete(servers, name)
	return server.Close()
}

// UnregisterDefault 注销默认服务器
func UnregisterDefault() error {
	return Unregister("default")
}

// GetServer 获取指定名称的服务器
func GetServer(name string) (*Server, error) {
	mu.RLock()
	defer mu.RUnlock()

	server, ok := servers[name]
	if !ok {
		return nil, ErrServerNotFound
	}

	return server, nil
}

// Default 获取默认服务器
func Default() (*Server, error) {
	return GetServer(defaultName)
}

// CloseAll 关闭所有服务器
func CloseAll() error {
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

