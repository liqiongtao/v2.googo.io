# goo-db 基于xorm封装的数据库

## 需求

1. 开发语言：golang
2. 包名: goodb
3. 目录: goo-db
4. 功能需求:
   * 定义Client对象，基于 github.com/go-xorm/xorm
   * Client对象要支持 github.com/go-sql-driver/mysql
   * Client对象要支持 github.com/lib/pg
   * 定义Config对象
   * 定义包方法
   * 支持多db选择
   * 支持多Client切换

## 功能特性

- ✅ 基于 `github.com/go-xorm/xorm` 封装
- ✅ 支持 MySQL 数据库（通过 `github.com/go-sql-driver/mysql`）
- ✅ 支持 PostgreSQL 数据库（通过 `github.com/lib/pq`）
- ✅ 支持多 Client 切换（通过名称管理多个数据库连接）
- ✅ 支持多 DB 选择（同一个 Client 可以访问不同的数据库）
- ✅ 提供包级别的便捷方法
- ✅ 线程安全的全局客户端管理
- ✅ 自动连接池管理
- ✅ 支持连接超时、最大连接数等配置
- ✅ 支持日志记录和慢查询监控

## 快速开始

### 安装依赖

```bash
go get github.com/go-xorm/xorm
go get github.com/go-sql-driver/mysql
go get github.com/lib/pq
```

### 基本使用

#### MySQL 示例

```go
package main

import (
    "fmt"
    
    "v2.googo.io/goo-db"
)

func main() {
    // 1. 注册默认 MySQL 客户端
    config := goodb.DefaultConfig()
    config.Driver = "mysql"
    config.DSN = "root:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
    
    if err := goodb.Register("default", config); err != nil {
        panic(err)
    }
    
    // 2. 使用默认客户端
    client, _ := goodb.Default()
    engine := client.Engine()
    
    // 执行查询
    var result []map[string]interface{}
    engine.SQL("SELECT * FROM users LIMIT 10").Find(&result)
    fmt.Println(result)
    
    // 3. 使用指定数据库
    db1Engine, _ := goodb.DB("db1")
    db1Engine.SQL("SELECT * FROM orders LIMIT 10").Find(&result)
}
```

#### PostgreSQL 示例

```go
package main

import (
    "fmt"
    
    "v2.googo.io/goo-db"
)

func main() {
    // 注册 PostgreSQL 客户端
    config := goodb.DefaultConfig()
    config.Driver = "postgres"
    config.DSN = "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable"
    
    if err := goodb.Register("pg", config); err != nil {
        panic(err)
    }
    
    // 使用 PostgreSQL 客户端
    client, _ := goodb.GetClient("pg")
    engine := client.Engine()
    
    // 执行查询
    var result []map[string]interface{}
    engine.SQL("SELECT * FROM users LIMIT 10").Find(&result)
    fmt.Println(result)
}
```

### 多 Client 切换

```go
// 注册多个客户端
goodb.Register("mysql-db", &goodb.Config{
    Driver: "mysql",
    DSN:    "root:password@tcp(mysql-server:3306)/db1?charset=utf8mb4&parseTime=True&loc=Local",
})

goodb.Register("pg-db", &goodb.Config{
    Driver: "postgres",
    DSN:    "host=pg-server port=5432 user=postgres password=password dbname=db1 sslmode=disable",
})

// 切换使用不同的客户端
mysqlClient, _ := goodb.GetClient("mysql-db")
pgClient, _ := goodb.GetClient("pg-db")

// 设置默认客户端
goodb.SetDefault("mysql-db")
defaultClient, _ := goodb.Default()
```

### 多 DB 选择

```go
// 获取默认客户端的默认数据库
engine, _ := goodb.GetDefaultEngine()

// 获取默认客户端的指定数据库
db1Engine, _ := goodb.DB("db1")
db2Engine, _ := goodb.DB("db2")

// 从指定客户端获取不同数据库
client, _ := goodb.GetClient("mysql-db")
db3Engine := client.DB("db3")
```

### 使用 ORM 功能

```go
// 定义模型
type User struct {
    Id       int64  `xorm:"pk autoincr"`
    Username string `xorm:"varchar(50) notnull unique"`
    Email    string `xorm:"varchar(100) notnull"`
    Created  time.Time `xorm:"created"`
}

// 获取引擎并操作
client, _ := goodb.Default()
engine := client.Engine()

// 同步表结构
engine.Sync2(new(User))

// 插入数据
user := &User{
    Username: "testuser",
    Email:    "test@example.com",
}
engine.Insert(user)

// 查询数据
var users []User
engine.Where("username = ?", "testuser").Find(&users)

// 更新数据
engine.ID(user.Id).Update(&User{Email: "newemail@example.com"})

// 删除数据
engine.ID(user.Id).Delete(&User{})
```

## API 文档

### Config 配置对象

```go
type Config struct {
    Driver          string        // 数据库驱动: "mysql" 或 "postgres"
    DSN             string        // 数据源名称（连接字符串）
    MaxIdleConns    int           // 连接池最大空闲连接数，默认 10
    MaxOpenConns    int           // 连接池最大打开连接数，默认 100
    ConnMaxLifetime time.Duration // 连接最大生存时间，默认 1 小时
    ConnMaxIdleTime time.Duration // 连接最大空闲时间，默认 30 分钟
    ShowSQL         bool          // 是否显示 SQL 语句，默认 false
    LogLevel        int           // 日志级别，默认 0（不记录）
    SlowQueryTime   time.Duration // 慢查询阈值，默认 1 秒
}
```

#### MySQL DSN 格式

```
username:password@tcp(host:port)/database?param1=value1&param2=value2
```

常用参数：
- `charset`: 字符集，推荐 `utf8mb4`
- `parseTime`: 是否解析时间，推荐 `True`
- `loc`: 时区，推荐 `Local`
- `timeout`: 连接超时时间

#### PostgreSQL DSN 格式

```
host=host port=port user=user password=password dbname=database sslmode=mode
```

常用参数：
- `sslmode`: SSL 模式，可选 `disable`, `require`, `verify-ca`, `verify-full`
- `connect_timeout`: 连接超时时间（秒）

### Client 客户端对象

```go
type Client struct {
    // ...
}

// NewClient 创建新的数据库客户端
func NewClient(name string, config *Config) (*Client, error)

// Name 获取客户端名称
func (c *Client) Name() string

// Engine 获取默认数据库的引擎
func (c *Client) Engine() *xorm.Engine

// DB 获取指定数据库的引擎（支持多 db 选择）
func (c *Client) DB(dbName string) *xorm.Engine

// Close 关闭客户端连接
func (c *Client) Close() error

// Ping 测试连接
func (c *Client) Ping() error
```

### 包级别方法

```go
// Register 注册一个数据库客户端（支持多 Client 切换）
func Register(name string, config *Config) error

// Unregister 注销一个数据库客户端
func Unregister(name string) error

// GetClient 获取指定名称的客户端
func GetClient(name string) (*Client, error)

// SetDefault 设置默认客户端名称
func SetDefault(name string)

// Default 获取默认客户端
func Default() (*Client, error)

// GetDefaultEngine 获取默认客户端的默认数据库引擎
func GetDefaultEngine() (*xorm.Engine, error)

// DB 获取默认客户端的指定数据库引擎（支持多 db 选择）
func DB(dbName string) (*xorm.Engine, error)

// CloseAll 关闭所有客户端
func CloseAll() error

// Ping 测试默认客户端连接
func Ping() error
```

## 使用建议

1. **初始化时机**: 建议在应用启动时注册所有需要的数据库客户端
2. **连接管理**: 使用 `CloseAll()` 在应用退出时关闭所有连接
3. **多 Client 场景**: 当需要连接多个不同的数据库服务器时，使用不同的名称注册多个客户端
4. **多 DB 场景**: 当需要访问同一个数据库服务器的不同数据库时，使用 `DB()` 方法
5. **错误处理**: 始终检查 `Register()` 和 `GetClient()` 等方法的返回值
6. **连接池配置**: 根据实际并发需求合理配置 `MaxIdleConns` 和 `MaxOpenConns`
7. **慢查询监控**: 设置 `SlowQueryTime` 来监控慢查询，帮助优化数据库性能
8. **SQL 日志**: 开发环境可以开启 `ShowSQL` 来查看执行的 SQL 语句

## 注意事项

1. **依赖版本**: 本库基于 `github.com/go-xorm/xorm`，确保已正确安装依赖
   - MySQL 驱动: `github.com/go-sql-driver/mysql`
   - PostgreSQL 驱动: `github.com/lib/pq`
2. **线程安全**: 全局客户端管理是线程安全的，可以并发使用
3. **连接池**: 每个 Client 和 DB 都会维护自己的连接池，注意合理配置连接池大小
4. **资源释放**: 使用完毕后记得调用 `Close()` 或 `CloseAll()` 释放连接资源
5. **默认客户端**: 如果没有设置默认客户端，使用 `Default()` 等方法会返回错误
6. **DSN 格式**: MySQL 和 PostgreSQL 的 DSN 格式不同，请确保使用正确的格式
7. **时区设置**: MySQL 连接时建议设置 `loc=Local` 参数，确保时区正确
8. **字符集**: MySQL 建议使用 `utf8mb4` 字符集以支持完整的 Unicode 字符（包括 emoji）
9. **事务处理**: 使用 `engine.Begin()` 开始事务，确保在出错时调用 `Rollback()`，成功时调用 `Commit()`
10. **连接字符串安全**: 不要在代码中硬编码包含密码的连接字符串，建议使用环境变量或配置文件
