package goodb

import (
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-xorm/xorm"
)

// Client 数据库客户端封装
type Client struct {
	name    string
	config  *Config
	engine  *xorm.Engine
	dbs     map[string]*xorm.Engine // 多 db 支持
	mu      sync.RWMutex
}

// NewClient 创建新的数据库客户端
func NewClient(name string, config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证驱动
	if config.Driver != "mysql" && config.Driver != "postgres" {
		return nil, ErrInvalidDriver
	}

	// 验证 DSN
	if config.DSN == "" {
		return nil, ErrEmptyDSN
	}

	// 创建引擎
	engine, err := xorm.NewEngine(config.Driver, config.DSN)
	if err != nil {
		return nil, err
	}

	// 配置连接池
	if config.MaxIdleConns > 0 {
		engine.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.MaxOpenConns > 0 {
		engine.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.ConnMaxLifetime > 0 {
		engine.SetConnMaxLifetime(config.ConnMaxLifetime)
	}
	// 注意: SetConnMaxIdleTime 需要 Go 1.15+ 和较新版本的 xorm
	// 如果编译错误，请升级 xorm 版本或移除此行
	if config.ConnMaxIdleTime > 0 {
		if setConnMaxIdleTime, ok := interface{}(engine).(interface{ SetConnMaxIdleTime(time.Duration) }); ok {
			setConnMaxIdleTime.SetConnMaxIdleTime(config.ConnMaxIdleTime)
		}
	}

	// 配置日志
	engine.ShowSQL(config.ShowSQL)
	if config.LogLevel > 0 {
		engine.SetLogLevel(xorm.LogLevel(config.LogLevel))
	}

	// 测试连接
	if err := engine.Ping(); err != nil {
		engine.Close()
		return nil, err
	}

	c := &Client{
		name:   name,
		config: config,
		engine: engine,
		dbs:    make(map[string]*xorm.Engine),
	}

	// 将默认引擎添加到 dbs map（使用默认数据库名）
	defaultDBName := c.extractDBName(config.DSN, config.Driver)
	if defaultDBName != "" {
		c.dbs[defaultDBName] = engine
	}

	return c, nil
}

// Name 获取客户端名称
func (c *Client) Name() string {
	return c.name
}

// Engine 获取默认数据库的引擎
func (c *Client) Engine() *xorm.Engine {
	return c.engine
}

// DB 获取指定数据库的引擎（支持多 db 选择）
func (c *Client) DB(dbName string) *xorm.Engine {
	if dbName == "" {
		return c.engine
	}

	c.mu.RLock()
	if engine, ok := c.dbs[dbName]; ok {
		c.mu.RUnlock()
		return engine
	}
	c.mu.RUnlock()

	// 创建新的引擎连接（使用相同的配置，但不同的数据库）
	newDSN := c.replaceDBName(c.config.DSN, c.config.Driver, dbName)
	
	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查，避免并发创建
	if engine, ok := c.dbs[dbName]; ok {
		return engine
	}

	newEngine, err := xorm.NewEngine(c.config.Driver, newDSN)
	if err != nil {
		// 如果连接失败，返回默认引擎
		return c.engine
	}

	// 配置连接池（使用相同的配置）
	if c.config.MaxIdleConns > 0 {
		newEngine.SetMaxIdleConns(c.config.MaxIdleConns)
	}
	if c.config.MaxOpenConns > 0 {
		newEngine.SetMaxOpenConns(c.config.MaxOpenConns)
	}
	if c.config.ConnMaxLifetime > 0 {
		newEngine.SetConnMaxLifetime(c.config.ConnMaxLifetime)
	}
	// 注意: SetConnMaxIdleTime 需要 Go 1.15+ 和较新版本的 xorm
	// 如果编译错误，请升级 xorm 版本或移除此行
	if c.config.ConnMaxIdleTime > 0 {
		if setConnMaxIdleTime, ok := interface{}(newEngine).(interface{ SetConnMaxIdleTime(time.Duration) }); ok {
			setConnMaxIdleTime.SetConnMaxIdleTime(c.config.ConnMaxIdleTime)
		}
	}

	// 配置日志
	newEngine.ShowSQL(c.config.ShowSQL)
	if c.config.LogLevel > 0 {
		newEngine.SetLogLevel(xorm.LogLevel(c.config.LogLevel))
	}

	// 测试连接
	if err := newEngine.Ping(); err != nil {
		newEngine.Close()
		// 如果连接失败，返回默认引擎
		return c.engine
	}

	// 缓存到 dbs map
	c.dbs[dbName] = newEngine
	return newEngine
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error
	for _, engine := range c.dbs {
		if closeErr := engine.Close(); closeErr != nil {
			err = closeErr
		}
	}
	c.dbs = make(map[string]*xorm.Engine)
	return err
}

// Ping 测试连接
func (c *Client) Ping() error {
	return c.engine.Ping()
}

// extractDBName 从 DSN 中提取数据库名
func (c *Client) extractDBName(dsn, driver string) string {
	if driver == "mysql" {
		// MySQL DSN 格式: username:password@tcp(host:port)/database?params
		re := regexp.MustCompile(`/([^/?]+)`)
		matches := re.FindStringSubmatch(dsn)
		if len(matches) > 1 {
			return matches[1]
		}
	} else if driver == "postgres" {
		// PostgreSQL DSN 格式: host=host port=port user=user password=password dbname=database
		re := regexp.MustCompile(`dbname=([^\s]+)`)
		matches := re.FindStringSubmatch(dsn)
		if len(matches) > 1 {
			return strings.Trim(matches[1], "'\"")
		}
	}
	return ""
}

// replaceDBName 替换 DSN 中的数据库名
func (c *Client) replaceDBName(dsn, driver, newDBName string) string {
	if driver == "mysql" {
		// MySQL DSN 格式: username:password@tcp(host:port)/database?params
		re := regexp.MustCompile(`/([^/?]+)`)
		return re.ReplaceAllString(dsn, "/"+newDBName)
	} else if driver == "postgres" {
		// PostgreSQL DSN 格式: host=host port=port user=user password=password dbname=database
		re := regexp.MustCompile(`dbname=[^\s]+`)
		if re.MatchString(dsn) {
			return re.ReplaceAllString(dsn, "dbname="+newDBName)
		}
		// 如果没有 dbname，添加一个
		if strings.Contains(dsn, "?") {
			return dsn + "&dbname=" + url.QueryEscape(newDBName)
		}
		return dsn + " dbname=" + newDBName
	}
	return dsn
}

