# goo-grpc gRPC 客户端和服务端库

## 需求

1. 开发语言：golang
2. 包名: googrpc
3. 目录: goo-grpc
4. 功能需求:
   * 定义Client对象
   * 定义Server对象
   * 定义Config对象
   * 定义包方法
   * Client、Server 要支持 etcd、consul 注册支持
   * Client、Server 要支持打印日志，记录traceId

## 功能特性

- ✅ 基于 `google.golang.org/grpc` 封装
- ✅ 支持多 Client 和 Server 切换（通过名称管理多个实例）
- ✅ 提供包级别的便捷方法
- ✅ 线程安全的全局客户端和服务端管理
- ✅ 支持 etcd 服务注册和发现
- ✅ 支持 consul 服务注册和发现
- ✅ 自动日志记录，支持 traceId 追踪
- ✅ 支持 TLS 加密连接
- ✅ 支持 Keepalive 配置
- ✅ 支持自定义消息大小限制
- ✅ 提供拦截器支持（日志和追踪）

## 快速开始

### 安装依赖

```bash
go get google.golang.org/grpc
```

### 基本使用 - 服务端

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    "v2.googo.io/goo-grpc"
    pb "your/proto/package" // 你的 protobuf 生成代码
)

func main() {
    // 1. 创建服务端配置
    config := googrpc.DefaultConfig(
        googrpc.WithAddress(":50051"),
        googrpc.WithServiceName("my-service"),
        googrpc.WithEnableLog(true),
        googrpc.WithEnableTrace(true),
    )
    
    // 2. 注册默认服务端
    if err := googrpc.RegisterDefaultServer(config); err != nil {
        log.Fatal(err)
    }
    
    // 3. 获取服务端并注册服务实现
    server, _ := googrpc.DefaultServer()
    pb.RegisterYourServiceServer(server.Server(), &yourService{})
    
    // 4. 启动服务
    if err := server.Serve(); err != nil {
        log.Fatal(err)
    }
}

type yourService struct {
    pb.UnimplementedYourServiceServer
}

func (s *yourService) YourMethod(ctx context.Context, req *pb.YourRequest) (*pb.YourResponse, error) {
    // 你的业务逻辑
    return &pb.YourResponse{}, nil
}
```

### 基本使用 - 客户端

```go
package main

import (
    "context"
    "log"
    
    "v2.googo.io/goo-grpc"
    pb "your/proto/package" // 你的 protobuf 生成代码
)

func main() {
    // 1. 创建客户端配置
    config := googrpc.DefaultConfig(
        googrpc.WithAddress("localhost:50051"),
        googrpc.WithEnableLog(true),
        googrpc.WithEnableTrace(true),
    )
    
    // 2. 注册默认客户端
    if err := googrpc.RegisterDefaultClient(config); err != nil {
        log.Fatal(err)
    }
    
    // 3. 使用客户端
    client, _ := googrpc.DefaultClient()
    conn := client.Conn()
    
    // 4. 创建服务客户端并调用
    serviceClient := pb.NewYourServiceClient(conn)
    ctx := context.Background()
    
    resp, err := serviceClient.YourMethod(ctx, &pb.YourRequest{})
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("Response:", resp)
}
```

### 使用 etcd 服务注册

```go
package main

import (
    "time"
    "v2.googo.io/goo-grpc"
)

func main() {
    // 服务端配置 - 使用 etcd 注册
    serverConfig := googrpc.DefaultConfig(
        googrpc.WithAddress(":50051"),
        googrpc.WithServiceName("my-service"),
        googrpc.WithRegistryType(googrpc.RegistryTypeEtcd),
        googrpc.WithEtcdConfig(&googrpc.EtcdRegistryConfig{
            Endpoints:   []string{"localhost:2379"},
            Username:    "",
            Password:    "",
            DialTimeout: 5 * time.Second,
            LeaseTTL:    10,
        }),
    )
    
    googrpc.RegisterDefaultServer(serverConfig)
    // ... 启动服务
}
```

### 使用 consul 服务注册

```go
package main

import (
    "v2.googo.io/goo-grpc"
)

func main() {
    // 服务端配置 - 使用 consul 注册
    serverConfig := googrpc.DefaultConfig(
        googrpc.WithAddress(":50051"),
        googrpc.WithServiceName("my-service"),
        googrpc.WithRegistryType(googrpc.RegistryTypeConsul),
        googrpc.WithConsulConfig(&googrpc.ConsulRegistryConfig{
            Address:             "localhost:8500",
            HealthCheckInterval:  10,
            HealthCheckTimeout:   5,
            HealthCheckPath:      "/health",
        }),
    )
    
    googrpc.RegisterDefaultServer(serverConfig)
    // ... 启动服务
}
```

### 使用 traceId 追踪

```go
package main

import (
    "context"
    "v2.googo.io/goo-context"
    "v2.googo.io/goo-grpc"
    pb "your/proto/package"
)

func main() {
    // 创建带 traceId 的上下文
    ctx := goocontext.WithTraceId(context.Background(), "trace-12345")
    
    // 客户端会自动将 traceId 添加到请求的 metadata 中
    client, _ := googrpc.DefaultClient()
    conn := client.Conn()
    serviceClient := pb.NewYourServiceClient(conn)
    
    // 服务端会自动从 metadata 中提取 traceId 并记录到日志
    resp, err := serviceClient.YourMethod(ctx, &pb.YourRequest{})
    // ...
}
```

## API 文档

### Config 配置

```go
type Config struct {
    Address          string              // 地址
    ServiceName      string              // 服务名称
    RegistryType     RegistryType        // 注册中心类型
    EtcdConfig       *EtcdRegistryConfig // etcd 配置
    ConsulConfig     *ConsulRegistryConfig // consul 配置
    EnableTLS        bool                // 是否启用 TLS
    TLSConfig        *TLSConfig          // TLS 配置
    Timeout          time.Duration       // 连接超时时间
    EnableLog        bool                // 是否启用日志
    EnableTrace      bool                // 是否启用追踪
    MaxSendMsgSize   int                 // 最大发送消息大小
    MaxRecvMsgSize   int                 // 最大接收消息大小
    // ...
}
```

### Client 客户端

```go
// NewClient 创建新的客户端
func NewClient(name string, config *Config) (*Client, error)

// Conn 获取 gRPC 连接
func (c *Client) Conn() *grpc.ClientConn

// Invoke 调用 gRPC 方法（带日志和追踪）
func (c *Client) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error

// NewStream 创建新的流（带日志和追踪）
func (c *Client) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error)

// Close 关闭客户端连接
func (c *Client) Close() error
```

### Server 服务端

```go
// NewServer 创建新的服务端
func NewServer(name string, config *Config) (*Server, error)

// Server 获取 gRPC 服务端
func (s *Server) Server() *grpc.Server

// RegisterService 注册服务实现
func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl interface{})

// Serve 启动服务（阻塞调用）
func (s *Server) Serve() error

// GracefulStop 优雅停止服务
func (s *Server) GracefulStop()

// Stop 立即停止服务
func (s *Server) Stop()

// Close 关闭服务端
func (s *Server) Close() error
```

### 全局方法

```go
// 客户端管理
func RegisterClient(name string, config *Config) error
func RegisterDefaultClient(config *Config) error
func GetClient(name string) (*Client, error)
func DefaultClient() (*Client, error)
func UnregisterClient(name string) error
func CloseAllClients() error

// 服务端管理
func RegisterServer(name string, config *Config) error
func RegisterDefaultServer(config *Config) error
func GetServer(name string) (*Server, error)
func DefaultServer() (*Server, error)
func UnregisterServer(name string) error
func CloseAllServers() error
```

## 使用建议

1. **服务注册**: 在生产环境中，建议使用 etcd 或 consul 进行服务注册和发现，以实现服务的动态发现和负载均衡。

2. **日志记录**: 启用日志和追踪功能可以帮助你调试和监控 gRPC 调用。traceId 会自动在客户端和服务端之间传递。

3. **TLS 配置**: 在生产环境中，建议启用 TLS 加密以确保通信安全。

4. **优雅关闭**: 使用 `GracefulStop()` 方法可以优雅地关闭服务端，等待正在处理的请求完成。

5. **多实例管理**: 通过名称管理多个客户端和服务端实例，可以在不同的场景下使用不同的配置。

## 注意事项

1. **服务注册实现**: 当前 etcd 和 consul 的服务注册和发现功能为框架代码，实际使用时需要根据具体的 etcd/consul 客户端库实现完整的注册和发现逻辑。

2. **依赖管理**: 如果需要使用 etcd 或 consul，需要额外安装相应的客户端库：
   ```bash
   go get go.etcd.io/etcd/client/v3
   go get github.com/hashicorp/consul/api
   ```

3. **日志集成**: 日志功能依赖于 `v2.googo.io/goo-log` 包，确保已正确配置日志适配器。

4. **上下文追踪**: traceId 追踪功能依赖于 `v2.googo.io/goo-context` 包，确保在调用时正确设置上下文。

5. **并发安全**: 所有全局方法都是线程安全的，可以在多个 goroutine 中安全使用。
