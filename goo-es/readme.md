# goo-es Elasticsearch 客户端库

## 需求

1. 开发语言：golang
2. 包名: gooes
3. 目录: goo-es
4. 功能需求:

   * 定义Client对象，基于 github.com/elastic/go-elasticsearch/v7
   * 定义Config对象
   * 定义包方法

## 功能特性

- ✅ 基于 `github.com/elastic/go-elasticsearch/v7` 封装
- ✅ 支持多 Client 切换（通过名称管理多个 Elasticsearch 连接）
- ✅ 提供包级别的便捷方法
- ✅ 线程安全的全局客户端管理
- ✅ 支持用户名密码认证
- ✅ 支持 API Key 认证
- ✅ 支持 Elastic Cloud 连接
- ✅ 支持服务令牌认证
- ✅ 提供常用的 CRUD 操作方法（Index、Get、Update、Delete）
- ✅ 支持搜索操作
- ✅ 支持批量操作
- ✅ 自动连接测试

## 快速开始

### 安装依赖

```bash
go get github.com/elastic/go-elasticsearch/v7
```

### 基本使用

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"
    
    "v2.googo.io/goo-es"
)

func main() {
    // 1. 注册默认客户端
    config := gooes.DefaultConfig()
    config.Addresses = []string{"http://localhost:9200"}
    config.Username = "elastic"
    config.Password = "changeme"
    
    if err := gooes.Register("default", config); err != nil {
        panic(err)
    }
    
    // 2. 使用默认客户端
    client, _ := gooes.Default()
    
    // 测试连接
    ctx := context.Background()
    if err := client.Ping(ctx); err != nil {
        panic(err)
    }
    
    // 索引文档
    doc := map[string]interface{}{
        "title": "Elasticsearch",
        "content": "A distributed search and analytics engine",
    }
    res, _ := client.Index("test-index", doc)
    defer res.Body.Close()
    
    // 搜索文档
    var buf strings.Builder
    query := map[string]interface{}{
        "query": map[string]interface{}{
            "match": map[string]interface{}{
                "title": "Elasticsearch",
            },
        },
    }
    json.NewEncoder(&buf).Encode(query)
    
    searchRes, _ := client.Search(
        client.Search.WithIndex("test-index"),
        client.Search.WithBody(strings.NewReader(buf.String())),
    )
    defer searchRes.Body.Close()
    
    fmt.Println("Search completed")
}
```

### 使用配置选项

```go
// 使用配置选项函数
config := gooes.DefaultConfig(
    gooes.WithAddresses([]string{"http://localhost:9200"}),
    gooes.WithUsername("elastic"),
    gooes.WithPassword("changeme"),
    gooes.WithMaxRetries(5),
    gooes.WithEnableCompression(true),
)

if err := gooes.Register("default", config); err != nil {
    panic(err)
}
```

### 多 Client 切换

```go
// 注册多个客户端
gooes.Register("local", &gooes.Config{
    Addresses: []string{"http://localhost:9200"},
    Username:  "elastic",
    Password:  "changeme",
})

gooes.Register("remote", &gooes.Config{
    Addresses: []string{"https://remote-elasticsearch:9200"},
    Username:  "admin",
    Password:  "password",
    APIKey:    "your-api-key",
})

// 切换使用不同的客户端
localClient, _ := gooes.GetClient("local")
remoteClient, _ := gooes.GetClient("remote")

// 设置默认客户端
gooes.SetDefault("local")
defaultClient, _ := gooes.Default()
```

### 使用 Elastic Cloud

```go
config := gooes.DefaultConfig()
config.CloudID = "your-cloud-id"
config.APIKey = "your-api-key"

if err := gooes.Register("cloud", config); err != nil {
    panic(err)
}
```

### 执行搜索操作

```go
client, _ := gooes.Default()

var buf strings.Builder
query := map[string]interface{}{
    "query": map[string]interface{}{
        "match_all": map[string]interface{}{},
    },
}
json.NewEncoder(&buf).Encode(query)

res, err := client.Search(
    client.Search.WithIndex("my-index"),
    client.Search.WithBody(strings.NewReader(buf.String())),
    client.Search.WithPretty(),
)
if err != nil {
    panic(err)
}
defer res.Body.Close()
```

### 批量操作

```go
client, _ := gooes.Default()

var buf strings.Builder
buf.WriteString(`{"index":{"_index":"test"}}` + "\n")
buf.WriteString(`{"title":"Document 1"}` + "\n")
buf.WriteString(`{"index":{"_index":"test"}}` + "\n")
buf.WriteString(`{"title":"Document 2"}` + "\n")

res, err := client.Bulk(strings.NewReader(buf.String()))
if err != nil {
    panic(err)
}
defer res.Body.Close()
```

## API 文档

### Config 配置对象

```go
type Config struct {
    Addresses         []string      // 地址列表，格式: ["http://localhost:9200"]
    Username          string        // 用户名
    Password          string        // 密码
    CloudID           string        // 云ID（Elastic Cloud）
    APIKey            string        // API Key
    ServiceToken      string        // 服务令牌
    Timeout           time.Duration // 连接超时时间，默认 5 秒
    MaxRetries        int           // 最大重试次数，默认 3
    EnableCompression bool          // 是否启用压缩，默认 false
    DisableMetaHeader bool          // 是否禁用元数据，默认 false
    EnableDebugLogger bool          // 是否启用调试日志，默认 false
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config
```

### Client 客户端对象

```go
type Client struct {
    // ...
}

// NewClient 创建新的 Elasticsearch 客户端
func NewClient(name string, config *Config) (*Client, error)

// Name 获取客户端名称
func (c *Client) Name() string

// Client 获取 Elasticsearch 客户端
func (c *Client) Client() *elasticsearch.Client

// Close 关闭客户端连接
func (c *Client) Close() error

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error

// Info 获取集群信息
func (c *Client) Info(ctx context.Context) (*esapi.Response, error)

// Search 执行搜索
func (c *Client) Search(opts ...func(*esapi.SearchRequest)) (*esapi.Response, error)

// Index 索引文档
func (c *Client) Index(index string, body interface{}, opts ...func(*esapi.IndexRequest)) (*esapi.Response, error)

// Get 获取文档
func (c *Client) Get(index, documentID string, opts ...func(*esapi.GetRequest)) (*esapi.Response, error)

// Delete 删除文档
func (c *Client) Delete(index, documentID string, opts ...func(*esapi.DeleteRequest)) (*esapi.Response, error)

// Update 更新文档
func (c *Client) Update(index, documentID string, body interface{}, opts ...func(*esapi.UpdateRequest)) (*esapi.Response, error)

// Bulk 批量操作
func (c *Client) Bulk(body interface{}, opts ...func(*esapi.BulkRequest)) (*esapi.Response, error)
```

### 包级别方法

```go
// Register 注册一个 Elasticsearch 客户端（支持多 Client 切换）
func Register(name string, config *Config) error

// RegisterDefault 注册默认客户端
func RegisterDefault(config *Config) error

// Unregister 注销一个 Elasticsearch 客户端
func Unregister(name string) error

// UnregisterDefault 注销默认客户端
func UnregisterDefault() error

// GetClient 获取指定名称的客户端
func GetClient(name string) (*Client, error)

// SetDefault 设置默认客户端名称
func SetDefault(name string)

// Default 获取默认客户端
func Default() (*Client, error)

// GetDefaultClient 获取默认客户端的 Elasticsearch 客户端
func GetDefaultClient() (*elasticsearch.Client, error)

// CloseAll 关闭所有客户端
func CloseAll() error

// Ping 测试默认客户端连接
func Ping(ctx context.Context) error
```

### 配置选项函数

```go
// WithAddresses 设置地址列表
func WithAddresses(addresses []string) FuncOption

// WithUsername 设置用户名
func WithUsername(username string) FuncOption

// WithPassword 设置密码
func WithPassword(password string) FuncOption

// WithCloudID 设置云ID
func WithCloudID(cloudID string) FuncOption

// WithAPIKey 设置 API Key
func WithAPIKey(apiKey string) FuncOption

// WithServiceToken 设置服务令牌
func WithServiceToken(serviceToken string) FuncOption

// WithTimeout 设置连接超时时间
func WithTimeout(timeout time.Duration) FuncOption

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) FuncOption

// WithEnableCompression 设置是否启用压缩
func WithEnableCompression(enableCompression bool) FuncOption

// WithDisableMetaHeader 设置是否禁用元数据
func WithDisableMetaHeader(disableMetaHeader bool) FuncOption

// WithEnableDebugLogger 设置是否启用调试日志
func WithEnableDebugLogger(enableDebugLogger bool) FuncOption
```

## 使用建议

1. **初始化时机**: 建议在应用启动时注册所有需要的 Elasticsearch 客户端
2. **连接管理**: 使用 `CloseAll()` 在应用退出时关闭所有连接（虽然 Elasticsearch 客户端通常不需要显式关闭）
3. **多 Client 场景**: 当需要连接多个不同的 Elasticsearch 集群时，使用不同的名称注册多个客户端
4. **错误处理**: 始终检查 `Register()` 和 `GetClient()` 等方法的返回值
5. **认证方式**: 根据 Elasticsearch 集群的配置选择合适的认证方式（用户名密码、API Key、服务令牌等）
6. **重试机制**: 根据网络环境合理配置 `MaxRetries`，默认 3 次重试通常足够
7. **压缩传输**: 对于大量数据传输，可以启用 `EnableCompression` 来减少网络带宽
8. **响应处理**: 使用完 `*esapi.Response` 后记得调用 `res.Body.Close()` 释放资源
9. **上下文使用**: 所有操作都支持传入 `context.Context`，建议使用带超时的上下文
10. **批量操作**: 对于大量文档操作，使用 `Bulk()` 方法可以提高性能

## 注意事项

1. **依赖版本**: 本库基于 `github.com/elastic/go-elasticsearch/v7`，确保已正确安装依赖
2. **线程安全**: 全局客户端管理是线程安全的，可以并发使用
3. **资源释放**: 使用完响应对象后记得调用 `res.Body.Close()` 释放资源
4. **默认客户端**: 如果没有设置默认客户端，使用 `Default()` 等方法会返回错误
5. **地址格式**: 地址列表中的 URL 应该包含协议（http:// 或 https://）
6. **认证配置**: 用户名密码、API Key、服务令牌等认证方式可以同时配置，但通常只需要一种
7. **Elastic Cloud**: 使用 Elastic Cloud 时，需要配置 `CloudID` 和 `APIKey`
8. **超时设置**: `Timeout` 字段目前保留在 Config 中，实际超时需要通过 Transport 配置
9. **错误处理**: Elasticsearch 的响应可能包含错误信息，需要检查 `res.IsError()` 来判断请求是否成功
10. **连接测试**: 创建客户端时会自动进行 Ping 测试，如果连接失败会返回错误
11. **索引名称**: 索引名称应该遵循 Elasticsearch 的命名规范（小写字母、数字、连字符、下划线）
12. **文档 ID**: 如果不指定文档 ID，Elasticsearch 会自动生成一个唯一的 ID
