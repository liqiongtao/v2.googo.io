package goorequest

import (
	"context"
	"io"
	"net/http"
	"sync"
)

var (
	clients     = make(map[string]*Request)
	defaultName = "default"
	mu          sync.RWMutex
)

// Register 注册一个请求客户端（支持多 Client 切换）
func Register(name string, config *Config) error {
	mu.Lock()
	defer mu.Unlock()

	client, err := NewRequest(name, config)
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

// Unregister 注销一个请求客户端
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
func GetClient(name string) (*Request, error) {
	mu.RLock()
	defer mu.RUnlock()

	client, ok := clients[name]
	if !ok {
		return nil, ErrClientNotFound
	}

	return client, nil
}

// Default 获取默认客户端
func Default() (*Request, error) {
	return GetClient(defaultName)
}

// Get 使用默认客户端发送GET请求
func Get(ctx context.Context, path string, headers map[string]string, params map[string]string) (*http.Response, error) {
	client, err := Default()
	if err != nil {
		return nil, err
	}
	return client.Get(ctx, path, headers, params)
}

// Post 使用默认客户端发送POST请求
func Post(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error) {
	client, err := Default()
	if err != nil {
		return nil, err
	}
	return client.Post(ctx, path, headers, body)
}

// Put 使用默认客户端发送PUT请求
func Put(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error) {
	client, err := Default()
	if err != nil {
		return nil, err
	}
	return client.Put(ctx, path, headers, body)
}

// Delete 使用默认客户端发送DELETE请求
func Delete(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error) {
	client, err := Default()
	if err != nil {
		return nil, err
	}
	return client.Delete(ctx, path, headers, body)
}

// Head 使用默认客户端发送HEAD请求
func Head(ctx context.Context, path string, headers map[string]string, params map[string]string) (*http.Response, error) {
	client, err := Default()
	if err != nil {
		return nil, err
	}
	return client.Head(ctx, path, headers, params)
}

// Patch 使用默认客户端发送PATCH请求
func Patch(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error) {
	client, err := Default()
	if err != nil {
		return nil, err
	}
	return client.Patch(ctx, path, headers, body)
}

// UploadFile 使用默认客户端上传文件
func UploadFile(ctx context.Context, path string, headers map[string]string, fieldName, filePath string, extraFields map[string]string) (*http.Response, error) {
	client, err := Default()
	if err != nil {
		return nil, err
	}
	return client.UploadFile(ctx, path, headers, fieldName, filePath, extraFields)
}

// DownloadFile 使用默认客户端下载文件
func DownloadFile(ctx context.Context, path string, headers map[string]string, params map[string]string, savePath string) error {
	client, err := Default()
	if err != nil {
		return err
	}
	return client.DownloadFile(ctx, path, headers, params, savePath)
}

// DownloadFileToWriter 使用默认客户端下载文件到Writer
func DownloadFileToWriter(ctx context.Context, path string, headers map[string]string, params map[string]string, writer io.Writer) error {
	client, err := Default()
	if err != nil {
		return err
	}
	return client.DownloadFileToWriter(ctx, path, headers, params, writer)
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

