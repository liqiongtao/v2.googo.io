# goo-request HTTP请求库

## 需求

1. 开发语言：golang
2. 包名: goorequest
3. 目录: goo-request
4. 功能需求:

   * 定义Request对象
   * 支持Get Post Put Head等请求
   * 支持自定义header
   * 支持上传文件，考虑大文件处理机制
   * 支持下载文件，考虑大文件处理机制
   * 支持配置tls
   * 支持设置Timeout
   * 定义包方法

## 功能特性

- ✅ 基于标准库 `net/http` 封装
- ✅ 支持 GET、POST、PUT、DELETE、HEAD、PATCH 等HTTP方法
- ✅ 支持自定义请求头
- ✅ 支持文件上传（流式传输，支持大文件）
- ✅ 支持文件下载（流式传输，支持大文件）
- ✅ 支持TLS配置
- ✅ 支持超时设置
- ✅ 支持请求重试机制
- ✅ 支持多 Client 切换（通过名称管理多个请求客户端）
- ✅ 提供包级别的便捷方法
- ✅ 线程安全的全局客户端管理
- ✅ 自动处理JSON编码/解码
- ✅ 支持查询参数和表单数据

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "io"
    
    "v2.googo.io/goo-request"
)

func main() {
    // 1. 注册默认客户端
    config := goorequest.DefaultConfig(
        goorequest.WithBaseURL("https://api.example.com"),
        goorequest.WithTimeout(30*time.Second),
    )
    
    if err := goorequest.RegisterDefault(config); err != nil {
        panic(err)
    }
    
    // 2. 发送GET请求
    ctx := context.Background()
    resp, err := goorequest.Get(ctx, "/users", nil, map[string]string{
        "page": "1",
        "size": "10",
    })
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    fmt.Println(string(body))
    
    // 3. 发送POST请求（JSON）
    data := map[string]interface{}{
        "name": "John",
        "age":  30,
    }
    resp, err = goorequest.Post(ctx, "/users", nil, data)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
}
```

### 自定义请求头

```go
headers := map[string]string{
    "Authorization": "Bearer token123",
    "Content-Type":  "application/json",
}

resp, err := goorequest.Get(ctx, "/api/data", headers, nil)
```

### 文件上传

```go
// 上传单个文件
resp, err := goorequest.UploadFile(
    ctx,
    "/upload",
    nil,
    "file",           // 字段名
    "/path/to/file",  // 文件路径
    map[string]string{
        "description": "My file",
    },
)
if err != nil {
    panic(err)
}
defer resp.Body.Close()
```

### 文件下载

```go
// 下载文件到指定路径
err := goorequest.DownloadFile(
    ctx,
    "/download/file.pdf",
    nil,
    nil,
    "/local/path/file.pdf",
)
if err != nil {
    panic(err)
}

// 下载文件到Writer
var buf bytes.Buffer
err = goorequest.DownloadFileToWriter(
    ctx,
    "/download/file.pdf",
    nil,
    nil,
    &buf,
)
```

### TLS配置

```go
config := goorequest.DefaultConfig(
    goorequest.WithBaseURL("https://api.example.com"),
    goorequest.WithTLS(&tls.Config{
        InsecureSkipVerify: false,
        // 其他TLS配置...
    }),
)
```

### 多 Client 切换

```go
// 注册多个客户端
goorequest.Register("api1", &goorequest.Config{
    BaseURL: "https://api1.example.com",
    Timeout: 30 * time.Second,
})

goorequest.Register("api2", &goorequest.Config{
    BaseURL: "https://api2.example.com",
    Timeout: 60 * time.Second,
})

// 切换使用不同的客户端
client1, _ := goorequest.GetClient("api1")
resp1, _ := client1.Get(ctx, "/endpoint", nil, nil)

client2, _ := goorequest.GetClient("api2")
resp2, _ := client2.Get(ctx, "/endpoint", nil, nil)
```

## API 文档

### Config 配置对象

```go
type Config struct {
    BaseURL            string        // 基础URL，所有请求会基于此URL
    Timeout            time.Duration // 默认超时时间，默认 30 秒
    TLS                *tls.Config   // TLS配置
    Headers            map[string]string // 默认请求头
    InsecureSkipVerify bool          // 是否跳过TLS证书验证
    MaxRetries         int           // 最大重试次数，默认 0
    RetryInterval      time.Duration // 重试间隔，默认 1 秒
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config
```

### Request 客户端对象

```go
type Request struct {
    // ...
}

// NewRequest 创建新的请求客户端
func NewRequest(name string, config *Config) (*Request, error)

// Name 获取客户端名称
func (r *Request) Name() string

// Client 获取HTTP客户端
func (r *Request) Client() *http.Client

// Get 发送GET请求
func (r *Request) Get(ctx context.Context, path string, headers map[string]string, params map[string]string) (*http.Response, error)

// Post 发送POST请求
func (r *Request) Post(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error)

// Put 发送PUT请求
func (r *Request) Put(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error)

// Delete 发送DELETE请求
func (r *Request) Delete(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error)

// Head 发送HEAD请求
func (r *Request) Head(ctx context.Context, path string, headers map[string]string, params map[string]string) (*http.Response, error)

// Patch 发送PATCH请求
func (r *Request) Patch(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error)

// UploadFile 上传文件（支持大文件）
func (r *Request) UploadFile(ctx context.Context, path string, headers map[string]string, fieldName, filePath string, extraFields map[string]string) (*http.Response, error)

// DownloadFile 下载文件到指定路径（支持大文件）
func (r *Request) DownloadFile(ctx context.Context, path string, headers map[string]string, params map[string]string, savePath string) error

// DownloadFileToWriter 下载文件到Writer（支持大文件）
func (r *Request) DownloadFileToWriter(ctx context.Context, path string, headers map[string]string, params map[string]string, writer io.Writer) error

// Close 关闭客户端
func (r *Request) Close() error
```

### 包级别方法

```go
// Register 注册一个请求客户端（支持多 Client 切换）
func Register(name string, config *Config) error

// RegisterDefault 注册默认客户端
func RegisterDefault(config *Config) error

// Unregister 注销一个请求客户端
func Unregister(name string) error

// UnregisterDefault 注销默认客户端
func UnregisterDefault() error

// GetClient 获取指定名称的客户端
func GetClient(name string) (*Request, error)

// Default 获取默认客户端
func Default() (*Request, error)

// Get 使用默认客户端发送GET请求
func Get(ctx context.Context, path string, headers map[string]string, params map[string]string) (*http.Response, error)

// Post 使用默认客户端发送POST请求
func Post(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error)

// Put 使用默认客户端发送PUT请求
func Put(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error)

// Delete 使用默认客户端发送DELETE请求
func Delete(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error)

// Head 使用默认客户端发送HEAD请求
func Head(ctx context.Context, path string, headers map[string]string, params map[string]string) (*http.Response, error)

// Patch 使用默认客户端发送PATCH请求
func Patch(ctx context.Context, path string, headers map[string]string, body interface{}) (*http.Response, error)

// UploadFile 使用默认客户端上传文件
func UploadFile(ctx context.Context, path string, headers map[string]string, fieldName, filePath string, extraFields map[string]string) (*http.Response, error)

// DownloadFile 使用默认客户端下载文件
func DownloadFile(ctx context.Context, path string, headers map[string]string, params map[string]string, savePath string) error

// DownloadFileToWriter 使用默认客户端下载文件到Writer
func DownloadFileToWriter(ctx context.Context, path string, headers map[string]string, params map[string]string, writer io.Writer) error

// CloseAll 关闭所有客户端
func CloseAll() error
```

### 配置选项函数

```go
// WithBaseURL 设置基础URL
func WithBaseURL(baseURL string) FuncOption

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) FuncOption

// WithTLS 设置TLS配置
func WithTLS(tls *tls.Config) FuncOption

// WithHeaders 设置默认请求头
func WithHeaders(headers map[string]string) FuncOption

// WithHeader 设置单个请求头
func WithHeader(key, value string) FuncOption

// WithInsecureSkipVerify 设置是否跳过TLS证书验证
func WithInsecureSkipVerify(skip bool) FuncOption

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(maxRetries int) FuncOption

// WithRetryInterval 设置重试间隔
func WithRetryInterval(interval time.Duration) FuncOption
```

## 使用建议

1. **初始化时机**: 建议在应用启动时注册所有需要的请求客户端
2. **资源管理**: 使用 `CloseAll()` 在应用退出时关闭所有客户端（虽然HTTP客户端通常不需要显式关闭）
3. **多 Client 场景**: 当需要连接多个不同的API服务器时，使用不同的名称注册多个客户端
4. **超时设置**: 根据实际需求合理设置超时时间，避免请求长时间阻塞
5. **错误处理**: 始终检查请求方法的返回值，并正确处理HTTP响应状态码
6. **大文件处理**: 上传和下载大文件时，库已使用流式传输，无需担心内存占用
7. **Context使用**: 使用context.Context来控制请求的生命周期和取消操作
8. **TLS配置**: 生产环境建议使用正确的TLS配置，避免使用 `InsecureSkipVerify`

## 注意事项

1. **依赖**: 本库基于标准库 `net/http`，无需额外依赖
2. **线程安全**: 全局客户端管理是线程安全的，可以并发使用
3. **响应处理**: 使用完毕后记得调用 `resp.Body.Close()` 关闭响应体
4. **文件上传**: 上传文件时，文件会在请求发送前完全读取到内存（用于multipart编码），对于超大文件请考虑分块上传
5. **文件下载**: 下载文件使用流式传输，支持大文件下载而不会占用过多内存
6. **默认客户端**: 如果没有设置默认客户端，使用包级别方法会返回错误
7. **Body类型**: POST/PUT等方法支持多种body类型：
   - `string`: 作为文本发送
   - `[]byte`: 作为二进制数据发送
   - `io.Reader`: 直接作为请求体
   - 其他类型: 自动JSON编码
8. **重试机制**: 默认不重试，可通过配置启用重试功能
