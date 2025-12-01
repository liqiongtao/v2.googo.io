# goo-consul

基于 `github.com/hashicorp/consul/api` 的 Consul 客户端封装库，提供服务注册、服务发现、健康检查和配置管理功能。

## 需求

1. 开发语言：golang
2. 包名: gooconsul
3. 目录: goo-consul
4. 功能需求:

   * 定义Client对象，基于 github.com/hashicorp/consul/api
   * 定义Config对象
   * 要支持服务注册、解析
   * 要支持watch
   * 定义包方法

## 功能特性

- ✅ 基于 `github.com/hashicorp/consul/api` 封装
- ✅ 支持多 Client 切换（通过名称管理多个 Consul 连接）
- ✅ 提供包级别的便捷方法
- ✅ 线程安全的全局客户端管理
- ✅ 支持服务注册与注销
- ✅ 支持服务发现与健康检查
- ✅ 支持服务监听（Watch）
- ✅ 支持键值存储（KV）操作
- ✅ 支持键值监听（Watch）
- ✅ 支持连接超时、TLS 等配置

## 快速开始

### 安装依赖

```bash
go get github.com/hashicorp/consul/api
```

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "v2.googo.io/goo-consul"
    "github.com/hashicorp/consul/api"
)

func main() {
    // 1. 注册默认客户端
    config := gooconsul.DefaultConfig(
        gooconsul.WithAddress("localhost:8500"),
        gooconsul.WithTimeout(5*time.Second),
    )
    
    if err := gooconsul.RegisterDefault(config); err != nil {
        panic(err)
    }
    
    // 2. 服务注册
    registration := &gooconsul.ServiceRegistration{
        ID:      "my-service-1",
        Name:    "my-service",
        Tags:    []string{"v1", "web"},
        Address: "127.0.0.1",
        Port:    8080,
        Check: &api.AgentServiceCheck{
            HTTP:     "http://127.0.0.1:8080/health",
            Interval: "10s",
            Timeout:  "3s",
        },
    }
    
    if err := gooconsul.RegisterService(registration); err != nil {
        panic(err)
    }
    
    // 3. 服务发现
    ctx := context.Background()
    entries, _, err := gooconsul.ServiceNodes("my-service", "", true, nil)
    if err != nil {
        panic(err)
    }
    
    for _, entry := range entries {
        fmt.Printf("Service: %s, Address: %s:%d\n", 
            entry.Service.Service, entry.Service.Address, entry.Service.Port)
    }
    
    // 4. 注销服务
    defer gooconsul.DeregisterService("my-service-1")
}
```

### 多 Client 切换

```go
// 注册多个客户端
gooconsul.Register("consul1", &gooconsul.Config{
    Address: "consul1:8500",
    Token:   "token1",
})

gooconsul.Register("consul2", &gooconsul.Config{
    Address: "consul2:8500",
    Token:   "token2",
})

// 切换使用不同的客户端
consul1Client, _ := gooconsul.GetClient("consul1")
consul2Client, _ := gooconsul.GetClient("consul2")

// 设置默认客户端
gooconsul.SetDefault("consul1")
defaultClient, _ := gooconsul.Default()
```

### Watch 监听服务变化

```go
client, _ := gooconsul.Default()
ctx := context.Background()

// 监听服务变化
err := client.WatchService(ctx, "my-service", "", true, func(entries []*api.ServiceEntry, err error) {
    if err != nil {
        fmt.Printf("Watch error: %v\n", err)
        return
    }
    
    fmt.Printf("Service entries updated, count: %d\n", len(entries))
    for _, entry := range entries {
        fmt.Printf("Service: %s, Address: %s:%d, Status: %s\n",
            entry.Service.Service,
            entry.Service.Address,
            entry.Service.Port,
            entry.Checks.AggregatedStatus(),
        )
    }
})

if err != nil {
    panic(err)
}
```

### Watch 监听键值变化

```go
client, _ := gooconsul.Default()
ctx := context.Background()

// 监听单个键
err := client.WatchKey(ctx, "config/app", func(pair *api.KVPair, err error) {
    if err != nil {
        fmt.Printf("Watch error: %v\n", err)
        return
    }
    
    if pair != nil {
        fmt.Printf("Key updated: %s = %s\n", pair.Key, string(pair.Value))
    }
})

// 监听键前缀
err = client.WatchKeyPrefix(ctx, "config/", func(pairs api.KVPairs, err error) {
    if err != nil {
        fmt.Printf("Watch error: %v\n", err)
        return
    }
    
    fmt.Printf("Keys updated, count: %d\n", len(pairs))
    for _, pair := range pairs {
        fmt.Printf("Key: %s, Value: %s\n", pair.Key, string(pair.Value))
    }
})
```

### 键值存储操作

```go
client, _ := gooconsul.Default()

// 设置键值
pair := &api.KVPair{
    Key:   "config/app",
    Value: []byte("value"),
}
_, err := client.PutKV(pair, nil)
if err != nil {
    panic(err)
}

// 获取键值
pair, _, err := client.GetKV("config/app", nil)
if err != nil {
    panic(err)
}
fmt.Printf("Value: %s\n", string(pair.Value))

// 列出键值（前缀匹配）
pairs, _, err := client.ListKV("config/", nil)
if err != nil {
    panic(err)
}

for _, p := range pairs {
    fmt.Printf("Key: %s, Value: %s\n", p.Key, string(p.Value))
}

// 删除键值
_, err = client.DeleteKV("config/app", nil)
if err != nil {
    panic(err)
}
```

## API 文档

### Config 配置对象

```go
type Config struct {
    Address    string              // 地址，格式: "localhost:8500"
    Datacenter string              // 数据中心
    Token      string              // 令牌
    Namespace  string              // 命名空间（Consul Enterprise 功能）
    Timeout    time.Duration       // 连接超时时间，默认 5 秒
    TLSConfig  *api.TLSConfig      // TLS 配置
    HttpClient *api.HttpClientConfig // HTTP 客户端配置
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config
```

### ServiceRegistration 服务注册信息

```go
type ServiceRegistration struct {
    ID      string                  // 服务 ID
    Name    string                  // 服务名称
    Tags    []string                // 服务标签
    Address string                  // 服务地址
    Port    int                     // 服务端口
    Meta    map[string]string       // 服务元数据
    Check   *api.AgentServiceCheck  // 健康检查（单个）
    Checks  api.AgentServiceChecks  // 健康检查（多个）
}
```

### Client 客户端对象

```go
type Client struct {
    // ...
}

// NewClient 创建新的 Consul 客户端
func NewClient(name string, config *Config) (*Client, error)

// Name 获取客户端名称
func (c *Client) Name() string

// Client 获取 Consul 客户端
func (c *Client) Client() *api.Client

// Close 关闭客户端连接
func (c *Client) Close() error

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error

// RegisterService 注册服务
func (c *Client) RegisterService(registration *ServiceRegistration) error

// DeregisterService 注销服务
func (c *Client) DeregisterService(serviceID string) error

// Service 获取服务信息
func (c *Client) Service(serviceID string, q *api.QueryOptions) (*api.AgentService, *api.QueryMeta, error)

// Services 获取所有服务
func (c *Client) Services(q *api.QueryOptions) (map[string]*api.AgentService, error)

// ServiceHealth 获取服务健康状态
func (c *Client) ServiceHealth(serviceName string, passingOnly bool, q *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error)

// ServiceNodes 获取服务节点列表
func (c *Client) ServiceNodes(serviceName string, tag string, passingOnly bool, q *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error)

// CatalogServices 获取目录服务列表
func (c *Client) CatalogServices(q *api.QueryOptions) (map[string][]string, *api.QueryMeta, error)

// CatalogService 获取目录服务详情
func (c *Client) CatalogService(serviceName string, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error)

// WatchService 监听服务变化
func (c *Client) WatchService(ctx context.Context, serviceName string, tag string, passingOnly bool, handler func([]*api.ServiceEntry, error)) error

// WatchKey 监听键值变化
func (c *Client) WatchKey(ctx context.Context, key string, handler func(*api.KVPair, error)) error

// WatchKeyPrefix 监听键前缀变化
func (c *Client) WatchKeyPrefix(ctx context.Context, prefix string, handler func(api.KVPairs, error)) error

// GetKV 获取键值
func (c *Client) GetKV(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error)

// PutKV 设置键值
func (c *Client) PutKV(pair *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error)

// DeleteKV 删除键值
func (c *Client) DeleteKV(key string, q *api.WriteOptions) (*api.WriteMeta, error)

// ListKV 列出键值（前缀匹配）
func (c *Client) ListKV(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error)
```

### 包级别方法

```go
// Register 注册一个 Consul 客户端（支持多 Client 切换）
func Register(name string, config *Config) error

// RegisterDefault 注册默认客户端
func RegisterDefault(config *Config) error

// Unregister 注销一个 Consul 客户端
func Unregister(name string) error

// UnregisterDefault 注销默认客户端
func UnregisterDefault() error

// GetClient 获取指定名称的客户端
func GetClient(name string) (*Client, error)

// SetDefault 设置默认客户端名称
func SetDefault(name string)

// Default 获取默认客户端
func Default() (*Client, error)

// GetDefaultClient 获取默认客户端的 Consul 客户端
func GetDefaultClient() (*api.Client, error)

// CloseAll 关闭所有客户端
func CloseAll() error

// Ping 测试默认客户端连接
func Ping(ctx context.Context) error

// RegisterService 使用默认客户端注册服务
func RegisterService(registration *ServiceRegistration) error

// DeregisterService 使用默认客户端注销服务
func DeregisterService(serviceID string) error

// ServiceHealth 使用默认客户端获取服务健康状态
func ServiceHealth(serviceName string, passingOnly bool, q *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error)

// ServiceNodes 使用默认客户端获取服务节点列表
func ServiceNodes(serviceName string, tag string, passingOnly bool, q *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error)

// WatchService 使用默认客户端监听服务变化
func WatchService(ctx context.Context, serviceName string, tag string, passingOnly bool, handler func([]*api.ServiceEntry, error)) error

// WatchKey 使用默认客户端监听键值变化
func WatchKey(ctx context.Context, key string, handler func(*api.KVPair, error)) error

// WatchKeyPrefix 使用默认客户端监听键前缀变化
func WatchKeyPrefix(ctx context.Context, prefix string, handler func(api.KVPairs, error)) error
```

### 配置选项

```go
// WithAddress 设置地址
func WithAddress(address string) FuncOption

// WithDatacenter 设置数据中心
func WithDatacenter(datacenter string) FuncOption

// WithToken 设置令牌
func WithToken(token string) FuncOption

// WithNamespace 设置命名空间
func WithNamespace(namespace string) FuncOption

// WithTimeout 设置连接超时时间
func WithTimeout(timeout time.Duration) FuncOption

// WithTLSConfig 设置 TLS 配置
func WithTLSConfig(tlsConfig *api.TLSConfig) FuncOption

// WithHttpClient 设置 HTTP 客户端配置
func WithHttpClient(httpClient *api.HttpClientConfig) FuncOption
```

## 使用建议

1. **初始化时机**: 建议在应用启动时注册所有需要的 Consul 客户端
2. **连接管理**: 使用 `CloseAll()` 在应用退出时关闭所有连接
3. **多 Client 场景**: 当需要连接多个不同的 Consul 集群时，使用不同的名称注册多个客户端
4. **错误处理**: 始终检查 `Register()` 和 `GetClient()` 等方法的返回值
5. **上下文使用**: 所有操作都使用 context，建议设置合理的超时时间
6. **Watch 使用**: Watch 操作会持续监听，注意合理管理 goroutine 和 context 取消
7. **服务注册**: 服务注册后建议在应用退出时调用 `DeregisterService` 注销服务
8. **健康检查**: 建议为每个服务配置健康检查，确保服务可用性
9. **服务发现**: 使用 `passingOnly=true` 参数只获取健康状态为 passing 的服务节点

## 注意事项

1. **依赖版本**: 本库基于 `github.com/hashicorp/consul/api`，确保已正确安装依赖
2. **线程安全**: 全局客户端管理是线程安全的，可以并发使用
3. **资源释放**: 使用完毕后记得调用 `Close()` 或 `CloseAll()` 释放连接资源
4. **默认客户端**: 如果没有设置默认客户端，使用 `Default()` 等方法会返回错误
5. **服务 ID**: 服务 ID 必须唯一，建议使用 UUID 或包含主机名和端口的组合
6. **Watch 阻塞**: Watch 方法会阻塞当前 goroutine，建议在独立的 goroutine 中运行
7. **健康检查**: 健康检查失败的服务不会被 `ServiceNodes` 返回（当 `passingOnly=true` 时）
8. **键值存储**: Consul 的 KV 存储适合存储配置信息，不适合存储大量数据
9. **连接超时**: 根据网络环境合理设置 `Timeout`，避免连接时间过长
10. **TLS 配置**: 生产环境建议启用 TLS 加密连接
