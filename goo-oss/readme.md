# goo-oss 阿里云对象存储

## 需求

1. 开发语言：golang
2. 包名: goooss
3. 目录: goo-oss
4. 功能需求:

   * 定义Client对象，基于 github.com/aliyun/aliyun-oss-go-sdk/oss
   * 定义Config对象
   * 定义包方法

## 功能特性

- ✅ 基于 `github.com/aliyun/aliyun-oss-go-sdk/oss` 封装
- ✅ 支持多 Client 切换（通过名称管理多个 OSS 连接）
- ✅ 支持永久密钥和 STS 临时密钥
- ✅ 提供包级别的便捷方法
- ✅ 线程安全的全局客户端管理
- ✅ 支持对象上传、下载、删除等基本操作
- ✅ 支持分片上传
- ✅ 支持对象 ACL 管理
- ✅ 支持生成签名 URL

## 快速开始

### 安装依赖

```bash
go get github.com/aliyun/aliyun-oss-go-sdk/oss
```

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "strings"
    
    "v2.googo.io/goo-oss"
)

func main() {
    // 1. 注册默认客户端
    config := goooss.DefaultConfig()
    config.AccessKeyID = "your-access-key-id"
    config.AccessKeySecret = "your-access-key-secret"
    config.Endpoint = "oss-cn-hangzhou.aliyuncs.com"
    config.Bucket = "your-bucket-name"
    
    if err := goooss.Register("default", config); err != nil {
        panic(err)
    }
    
    // 2. 使用默认客户端上传文件
    client, _ := goooss.Default()
    ctx := context.Background()
    
    // 上传文件
    content := strings.NewReader("Hello, OSS!")
    err := client.PutObject(ctx, "test.txt", content)
    if err != nil {
        panic(err)
    }
    
    // 下载文件
    reader, err := client.GetObject(ctx, "test.txt")
    if err != nil {
        panic(err)
    }
    defer reader.Close()
    
    // 读取内容
    buf := make([]byte, 1024)
    n, _ := reader.Read(buf)
    fmt.Println(string(buf[:n])) // 输出: Hello, OSS!
    
    // 删除文件
    err = client.DeleteObject(ctx, "test.txt")
    if err != nil {
        panic(err)
    }
}
```

### 使用配置选项

```go
// 使用配置选项创建客户端
config := goooss.DefaultConfig(
    goooss.WithAccessKeyID("your-access-key-id"),
    goooss.WithAccessKeySecret("your-access-key-secret"),
    goooss.WithEndpoint("oss-cn-hangzhou.aliyuncs.com"),
    goooss.WithBucket("your-bucket-name"),
    goooss.WithTimeout(60*time.Second),
)

goooss.Register("default", config)
```

### 多 Client 切换

```go
// 注册多个客户端
goooss.Register("bucket1", &goooss.Config{
    AccessKeyID:     "your-access-key-id",
    AccessKeySecret: "your-access-key-secret",
    Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
    Bucket:          "bucket1",
})

goooss.Register("bucket2", &goooss.Config{
    AccessKeyID:     "your-access-key-id",
    AccessKeySecret: "your-access-key-secret",
    Endpoint:        "oss-cn-shanghai.aliyuncs.com",
    Bucket:          "bucket2",
})

// 切换使用不同的客户端
bucket1Client, _ := goooss.GetClient("bucket1")
bucket2Client, _ := goooss.GetClient("bucket2")

// 设置默认客户端
goooss.SetDefault("bucket1")
defaultClient, _ := goooss.Default()
```

### 使用 STS 临时密钥

```go
config := goooss.DefaultConfig()
config.AccessKeyID = "your-access-key-id"
config.AccessKeySecret = "your-access-key-secret"
config.Endpoint = "oss-cn-hangzhou.aliyuncs.com"
config.Bucket = "your-bucket-name"

// 配置 STS 临时密钥
config.STS = &goooss.STSConfig{
    AccessKeyID:     "sts-access-key-id",
    AccessKeySecret: "sts-access-key-secret",
    SecurityToken:   "sts-security-token",
    ExpiredTime:     time.Now().Add(1 * time.Hour),
}

goooss.Register("default", config)
```

### 分片上传大文件

```go
client, _ := goooss.Default()
ctx := context.Background()

// 1. 初始化分片上传
key := "large-file.zip"
imur, err := client.InitiateMultipartUpload(ctx, key)
if err != nil {
    panic(err)
}

// 2. 上传分片
partNumber := 1
part1 := strings.NewReader("part1 content")
part, err := client.UploadPart(ctx, imur, partNumber, part1)
if err != nil {
    panic(err)
}

// 3. 完成分片上传
parts := []oss.UploadPart{part}
_, err = client.CompleteMultipartUpload(ctx, imur, parts)
if err != nil {
    panic(err)
}
```

### 生成签名 URL

```go
client, _ := goooss.Default()
ctx := context.Background()

// 生成 GET 请求的签名 URL，有效期 1 小时
url, err := client.SignURL(ctx, "test.txt", oss.HTTPGet, 3600)
if err != nil {
    panic(err)
}
fmt.Println("Signed URL:", url)
```

## API 文档

### Config 配置对象

```go
type Config struct {
    AccessKeyID     string        // 阿里云 AccessKeyID
    AccessKeySecret string        // 阿里云 AccessKeySecret
    Endpoint        string        // 访问域名，例如: oss-cn-hangzhou.aliyuncs.com
    Bucket          string        // 存储桶名称
    Timeout         time.Duration // 连接超时时间，默认 30 秒
    Debug           bool          // 是否启用调试日志，默认 false
    STS             *STSConfig    // STS 临时密钥配置（可选）
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config
```

### STSConfig STS 临时密钥配置

```go
type STSConfig struct {
    AccessKeyID     string        // 临时密钥 AccessKeyID
    AccessKeySecret string        // 临时密钥 AccessKeySecret
    SecurityToken   string        // 临时密钥 SecurityToken
    ExpiredTime     time.Time     // 过期时间
}
```

### Client 客户端对象

```go
type Client struct {
    // ...
}

// NewClient 创建新的 OSS 客户端
func NewClient(name string, config *Config) (*Client, error)

// Name 获取客户端名称
func (c *Client) Name() string

// Client 获取 OSS 客户端
func (c *Client) Client() *oss.Client

// Bucket 获取存储桶对象
func (c *Client) Bucket() *oss.Bucket

// Close 关闭客户端连接
func (c *Client) Close() error

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error

// PutObject 上传对象
func (c *Client) PutObject(ctx context.Context, objectKey string, reader io.Reader, options ...oss.Option) error

// GetObject 获取对象
func (c *Client) GetObject(ctx context.Context, objectKey string, options ...oss.Option) (io.ReadCloser, error)

// DeleteObject 删除对象
func (c *Client) DeleteObject(ctx context.Context, objectKey string, options ...oss.Option) error

// HeadObject 获取对象元信息
func (c *Client) HeadObject(ctx context.Context, objectKey string, options ...oss.Option) (oss.GetObjectMetaResult, error)

// CopyObject 复制对象
func (c *Client) CopyObject(ctx context.Context, destObjectKey, srcObjectKey string, options ...oss.Option) (oss.CopyObjectResult, error)

// ListObjects 列出对象
func (c *Client) ListObjects(ctx context.Context, options ...oss.Option) (oss.ListObjectsResult, error)

// PutObjectACL 设置对象 ACL
func (c *Client) PutObjectACL(ctx context.Context, objectKey string, objectACL oss.ACLType, options ...oss.Option) error

// GetObjectACL 获取对象 ACL
func (c *Client) GetObjectACL(ctx context.Context, objectKey string, options ...oss.Option) (oss.GetObjectACLResult, error)

// InitiateMultipartUpload 初始化分片上传
func (c *Client) InitiateMultipartUpload(ctx context.Context, objectKey string, options ...oss.Option) (oss.InitiateMultipartUploadResult, error)

// UploadPart 上传分片
func (c *Client) UploadPart(ctx context.Context, imur oss.InitiateMultipartUploadResult, partNumber int, reader io.Reader, options ...oss.Option) (oss.UploadPart, error)

// CompleteMultipartUpload 完成分片上传
func (c *Client) CompleteMultipartUpload(ctx context.Context, imur oss.InitiateMultipartUploadResult, parts []oss.UploadPart, options ...oss.Option) (oss.CompleteMultipartUploadResult, error)

// AbortMultipartUpload 取消分片上传
func (c *Client) AbortMultipartUpload(ctx context.Context, imur oss.InitiateMultipartUploadResult, options ...oss.Option) error

// ListMultipartUploads 列出进行中的分片上传
func (c *Client) ListMultipartUploads(ctx context.Context, options ...oss.Option) (oss.ListMultipartUploadResult, error)

// IsObjectExist 检查对象是否存在
func (c *Client) IsObjectExist(ctx context.Context, objectKey string, options ...oss.Option) (bool, error)

// SignURL 生成签名 URL
func (c *Client) SignURL(ctx context.Context, objectKey string, method oss.HTTPMethod, expiredInSec int64, options ...oss.Option) (string, error)
```

### 包级别方法

```go
// Register 注册一个 OSS 客户端（支持多 Client 切换）
func Register(name string, config *Config) error

// RegisterDefault 注册默认客户端
func RegisterDefault(config *Config) error

// Unregister 注销一个 OSS 客户端
func Unregister(name string) error

// UnregisterDefault 注销默认客户端
func UnregisterDefault() error

// GetClient 获取指定名称的客户端
func GetClient(name string) (*Client, error)

// SetDefault 设置默认客户端名称
func SetDefault(name string)

// Default 获取默认客户端
func Default() (*Client, error)

// GetDefaultClient 获取默认客户端的 OSS 客户端
func GetDefaultClient() (*oss.Client, error)

// GetDefaultBucket 获取默认客户端的存储桶对象
func GetDefaultBucket() (*oss.Bucket, error)

// CloseAll 关闭所有客户端
func CloseAll() error

// Ping 测试默认客户端连接
func Ping(ctx context.Context) error
```

### 配置选项函数

```go
// WithAccessKeyID 设置 AccessKeyID
func WithAccessKeyID(accessKeyID string) FuncOption

// WithAccessKeySecret 设置 AccessKeySecret
func WithAccessKeySecret(accessKeySecret string) FuncOption

// WithEndpoint 设置访问域名
func WithEndpoint(endpoint string) FuncOption

// WithBucket 设置存储桶名称
func WithBucket(bucket string) FuncOption

// WithTimeout 设置连接超时时间
func WithTimeout(timeout time.Duration) FuncOption

// WithDebug 设置是否启用调试日志
func WithDebug(debug bool) FuncOption

// WithSTS 设置 STS 临时密钥配置
func WithSTS(sts *STSConfig) FuncOption
```

## 使用建议

1. **初始化时机**: 建议在应用启动时注册所有需要的 OSS 客户端
2. **连接管理**: OSS 客户端使用 HTTP 连接，会自动管理连接池，无需手动关闭
3. **多 Client 场景**: 当需要连接多个不同的存储桶时，使用不同的名称注册多个客户端
4. **STS 临时密钥**: 在生产环境中建议使用 STS 临时密钥，提高安全性
5. **错误处理**: 始终检查 `Register()` 和 `GetClient()` 等方法的返回值
6. **大文件上传**: 对于大文件（>100MB），建议使用分片上传功能
7. **超时设置**: 根据实际网络情况合理设置超时时间
8. **Endpoint 选择**: 选择与业务最近的区域，可以提高上传下载速度

## 注意事项

1. **依赖版本**: 本库基于 `github.com/aliyun/aliyun-oss-go-sdk/oss`，确保已正确安装依赖
2. **线程安全**: 全局客户端管理是线程安全的，可以并发使用
3. **密钥安全**: 不要在代码中硬编码 AccessKeyID 和 AccessKeySecret，建议使用环境变量或配置中心
4. **区域选择**: 选择与业务最近的区域，可以提高上传下载速度
5. **存储桶权限**: 确保配置的 AccessKeyID/AccessKeySecret 有对应存储桶的操作权限
6. **STS 过期**: 使用 STS 临时密钥时，注意检查过期时间，及时更新
7. **Endpoint 格式**: Endpoint 格式为 `oss-{region}.aliyuncs.com`，例如 `oss-cn-hangzhou.aliyuncs.com`
8. **对象键命名**: 对象键（ObjectKey）是对象在存储桶中的唯一标识，建议使用有意义的命名规范
