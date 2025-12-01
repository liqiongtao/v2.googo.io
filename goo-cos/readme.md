# goo-cos 腾讯云对象存储

## 需求

1. 开发语言：golang
2. 包名: goocos
3. 目录: goo-cos
4. 功能需求:

   * 定义Client对象，基于 github.com/tencentyun/cos-go-sdk-v5 github.com/tencentyun/qcloud-cos-sts-sdk/go
   * 定义Config对象
   * 定义包方法

## 功能特性

- ✅ 基于 `github.com/tencentyun/cos-go-sdk-v5` 封装
- ✅ 支持多 Client 切换（通过名称管理多个 COS 连接）
- ✅ 支持永久密钥和 STS 临时密钥
- ✅ 提供包级别的便捷方法
- ✅ 线程安全的全局客户端管理
- ✅ 支持对象上传、下载、删除等基本操作
- ✅ 支持分片上传
- ✅ 支持对象 ACL 管理

## 快速开始

### 安装依赖

```bash
go get github.com/tencentyun/cos-go-sdk-v5
```

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "strings"
    
    "v2.googo.io/goo-cos"
)

func main() {
    // 1. 注册默认客户端
    config := goocos.DefaultConfig()
    config.SecretID = "your-secret-id"
    config.SecretKey = "your-secret-key"
    config.Region = "ap-beijing"
    config.Bucket = "your-bucket-name"
    
    if err := goocos.Register("default", config); err != nil {
        panic(err)
    }
    
    // 2. 使用默认客户端上传文件
    client, _ := goocos.Default()
    ctx := context.Background()
    
    // 上传文件
    content := strings.NewReader("Hello, COS!")
    _, err := client.PutObject(ctx, "test.txt", content, nil)
    if err != nil {
        panic(err)
    }
    
    // 下载文件
    resp, err := client.GetObject(ctx, "test.txt", nil)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    // 读取内容
    buf := make([]byte, 1024)
    n, _ := resp.Body.Read(buf)
    fmt.Println(string(buf[:n])) // 输出: Hello, COS!
    
    // 删除文件
    _, err = client.DeleteObject(ctx, "test.txt")
    if err != nil {
        panic(err)
    }
}
```

### 使用配置选项

```go
// 使用配置选项创建客户端
config := goocos.DefaultConfig(
    goocos.WithSecretID("your-secret-id"),
    goocos.WithSecretKey("your-secret-key"),
    goocos.WithRegion("ap-beijing"),
    goocos.WithBucket("your-bucket-name"),
    goocos.WithTimeout(60*time.Second),
)

goocos.Register("default", config)
```

### 多 Client 切换

```go
// 注册多个客户端
goocos.Register("bucket1", &goocos.Config{
    SecretID:  "your-secret-id",
    SecretKey: "your-secret-key",
    Region:    "ap-beijing",
    Bucket:    "bucket1",
})

goocos.Register("bucket2", &goocos.Config{
    SecretID:  "your-secret-id",
    SecretKey: "your-secret-key",
    Region:    "ap-shanghai",
    Bucket:    "bucket2",
})

// 切换使用不同的客户端
bucket1Client, _ := goocos.GetClient("bucket1")
bucket2Client, _ := goocos.GetClient("bucket2")

// 设置默认客户端
goocos.SetDefault("bucket1")
defaultClient, _ := goocos.Default()
```

### 使用 STS 临时密钥

```go
config := goocos.DefaultConfig()
config.SecretID = "your-secret-id"
config.SecretKey = "your-secret-key"
config.Region = "ap-beijing"
config.Bucket = "your-bucket-name"

// 配置 STS 临时密钥
config.STS = &goocos.STSConfig{
    SecretID:     "sts-secret-id",
    SecretKey:    "sts-secret-key",
    SessionToken: "sts-session-token",
    ExpiredTime:  time.Now().Add(1 * time.Hour),
}

goocos.Register("default", config)
```

### 分片上传大文件

```go
client, _ := goocos.Default()
ctx := context.Background()

// 1. 初始化分片上传
key := "large-file.zip"
initResult, _, err := client.InitiateMultipartUpload(ctx, key, nil)
if err != nil {
    panic(err)
}
uploadID := initResult.UploadID

// 2. 上传分片
partNumber := 1
part1 := strings.NewReader("part1 content")
_, err = client.UploadPart(ctx, key, uploadID, partNumber, part1, nil)
if err != nil {
    panic(err)
}

// 3. 完成分片上传
completeOpt := &cos.CompleteMultipartUploadOptions{
    Parts: []cos.Object{
        {PartNumber: partNumber, ETag: "etag"},
    },
}
_, _, err = client.CompleteMultipartUpload(ctx, key, uploadID, completeOpt)
if err != nil {
    panic(err)
}
```

## API 文档

### Config 配置对象

```go
type Config struct {
    SecretID  string        // 腾讯云 SecretID
    SecretKey string        // 腾讯云 SecretKey
    Region    string        // 区域，例如: ap-beijing
    Bucket    string        // 存储桶名称
    Scheme    string        // 协议，http 或 https，默认 https
    BaseURL   string        // 基础 URL，如果设置则使用此 URL
    Timeout   time.Duration // 连接超时时间，默认 30 秒
    Debug     bool          // 是否启用调试日志，默认 false
    STS       *STSConfig    // STS 临时密钥配置（可选）
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config
```

### STSConfig STS 临时密钥配置

```go
type STSConfig struct {
    SecretID     string        // 临时密钥 SecretID
    SecretKey    string        // 临时密钥 SecretKey
    SessionToken string        // 临时密钥 SessionToken
    ExpiredTime  time.Time     // 过期时间
}
```

### Client 客户端对象

```go
type Client struct {
    // ...
}

// NewClient 创建新的 COS 客户端
func NewClient(name string, config *Config) (*Client, error)

// Name 获取客户端名称
func (c *Client) Name() string

// Client 获取 COS 客户端
func (c *Client) Client() *cos.Client

// Close 关闭客户端连接
func (c *Client) Close() error

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error

// PutObject 上传对象
func (c *Client) PutObject(ctx context.Context, key string, r io.Reader, opt *cos.ObjectPutOptions) (*cos.Response, error)

// GetObject 获取对象
func (c *Client) GetObject(ctx context.Context, key string, opt *cos.ObjectGetOptions) (*cos.Response, error)

// DeleteObject 删除对象
func (c *Client) DeleteObject(ctx context.Context, key string) (*cos.Response, error)

// HeadObject 获取对象元信息
func (c *Client) HeadObject(ctx context.Context, key string, opt *cos.ObjectHeadOptions) (*cos.Response, error)

// CopyObject 复制对象
func (c *Client) CopyObject(ctx context.Context, key, sourceURL string, opt *cos.ObjectCopyOptions) (*cos.CopyObjectResult, *cos.Response, error)

// ListObjects 列出对象
func (c *Client) ListObjects(ctx context.Context, prefix string, opt *cos.BucketGetOptions) (*cos.BucketGetResult, *cos.Response, error)

// PutObjectACL 设置对象 ACL
func (c *Client) PutObjectACL(ctx context.Context, key string, opt *cos.ObjectPutACLOptions) (*cos.Response, error)

// GetObjectACL 获取对象 ACL
func (c *Client) GetObjectACL(ctx context.Context, key string) (*cos.AccessControlPolicy, *cos.Response, error)

// InitiateMultipartUpload 初始化分片上传
func (c *Client) InitiateMultipartUpload(ctx context.Context, key string, opt *cos.InitiateMultipartUploadOptions) (*cos.InitiateMultipartUploadResult, *cos.Response, error)

// UploadPart 上传分片
func (c *Client) UploadPart(ctx context.Context, key string, uploadID string, partNumber int, r io.Reader, opt *cos.ObjectUploadPartOptions) (*cos.Response, error)

// CompleteMultipartUpload 完成分片上传
func (c *Client) CompleteMultipartUpload(ctx context.Context, key string, uploadID string, opt *cos.CompleteMultipartUploadOptions) (*cos.CompleteMultipartUploadResult, *cos.Response, error)

// AbortMultipartUpload 取消分片上传
func (c *Client) AbortMultipartUpload(ctx context.Context, key string, uploadID string) (*cos.Response, error)

// ListMultipartUploads 列出进行中的分片上传
func (c *Client) ListMultipartUploads(ctx context.Context, opt *cos.BucketListMultipartUploadsOptions) (*cos.ListMultipartUploadsResult, *cos.Response, error)
```

### 包级别方法

```go
// Register 注册一个 COS 客户端（支持多 Client 切换）
func Register(name string, config *Config) error

// RegisterDefault 注册默认客户端
func RegisterDefault(config *Config) error

// Unregister 注销一个 COS 客户端
func Unregister(name string) error

// UnregisterDefault 注销默认客户端
func UnregisterDefault() error

// GetClient 获取指定名称的客户端
func GetClient(name string) (*Client, error)

// SetDefault 设置默认客户端名称
func SetDefault(name string)

// Default 获取默认客户端
func Default() (*Client, error)

// GetDefaultClient 获取默认客户端的 COS 客户端
func GetDefaultClient() (*cos.Client, error)

// CloseAll 关闭所有客户端
func CloseAll() error

// Ping 测试默认客户端连接
func Ping(ctx context.Context) error
```

### 配置选项函数

```go
// WithSecretID 设置 SecretID
func WithSecretID(secretID string) FuncOption

// WithSecretKey 设置 SecretKey
func WithSecretKey(secretKey string) FuncOption

// WithRegion 设置区域
func WithRegion(region string) FuncOption

// WithBucket 设置存储桶名称
func WithBucket(bucket string) FuncOption

// WithScheme 设置协议
func WithScheme(scheme string) FuncOption

// WithBaseURL 设置基础 URL
func WithBaseURL(baseURL string) FuncOption

// WithTimeout 设置连接超时时间
func WithTimeout(timeout time.Duration) FuncOption

// WithDebug 设置是否启用调试日志
func WithDebug(debug bool) FuncOption

// WithSTS 设置 STS 临时密钥配置
func WithSTS(sts *STSConfig) FuncOption
```

## 使用建议

1. **初始化时机**: 建议在应用启动时注册所有需要的 COS 客户端
2. **连接管理**: COS 客户端使用 HTTP 连接，会自动管理连接池，无需手动关闭
3. **多 Client 场景**: 当需要连接多个不同的存储桶时，使用不同的名称注册多个客户端
4. **STS 临时密钥**: 在生产环境中建议使用 STS 临时密钥，提高安全性
5. **错误处理**: 始终检查 `Register()` 和 `GetClient()` 等方法的返回值
6. **大文件上传**: 对于大文件（>100MB），建议使用分片上传功能
7. **超时设置**: 根据实际网络情况合理设置超时时间

## 注意事项

1. **依赖版本**: 本库基于 `github.com/tencentyun/cos-go-sdk-v5`，确保已正确安装依赖
2. **线程安全**: 全局客户端管理是线程安全的，可以并发使用
3. **密钥安全**: 不要在代码中硬编码 SecretID 和 SecretKey，建议使用环境变量或配置中心
4. **区域选择**: 选择与业务最近的区域，可以提高上传下载速度
5. **存储桶权限**: 确保配置的 SecretID/SecretKey 有对应存储桶的操作权限
6. **STS 过期**: 使用 STS 临时密钥时，注意检查过期时间，及时更新
7. **BaseURL**: 如果使用自定义域名，可以通过 `BaseURL` 配置，否则会自动根据 Region 和 Bucket 生成
