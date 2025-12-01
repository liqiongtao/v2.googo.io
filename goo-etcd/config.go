package gooetcd

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// Config etcd 配置
type Config struct {
	// 端点列表，格式: ["localhost:2379"]
	Endpoints []string

	// 用户名
	Username string

	// 密码
	Password string

	// 连接超时时间，默认 5 秒
	DialTimeout time.Duration

	// 自动同步间隔，默认 0（不自动同步）
	AutoSyncInterval time.Duration

	// 是否启用自动同步，默认 false
	AutoSync bool

	// 最大调用发送大小（字节），默认 2MB
	MaxCallSendMsgSize int

	// 最大调用接收大小（字节），默认 2MB
	MaxCallRecvMsgSize int

	// 是否启用压缩，默认 false
	EnableCompression bool

	// 是否启用 gRPC 调试日志，默认 false
	EnableGRPCDebugLog bool

	// 是否启用客户端 TLS，默认 false
	EnableTLS bool

	// TLS 配置（当 EnableTLS 为 true 时使用）
	TLSConfig *TLSConfig
}

// TLSConfig TLS 配置
type TLSConfig struct {
	// 证书文件路径
	CertFile string

	// 密钥文件路径
	KeyFile string

	// CA 证书文件路径
	CAFile string

	// 是否跳过证书验证，默认 false
	InsecureSkipVerify bool
}

// DefaultConfig 返回默认配置
func DefaultConfig(opts ...FuncOption) *Config {
	c := &Config{
		Endpoints:          []string{"localhost:2379"},
		DialTimeout:        5 * time.Second,
		AutoSyncInterval:   0,
		AutoSync:           false,
		MaxCallSendMsgSize: 2 * 1024 * 1024, // 2MB
		MaxCallRecvMsgSize: 2 * 1024 * 1024, // 2MB
		EnableCompression:  false,
		EnableGRPCDebugLog: false,
		EnableTLS:          false,
	}
	for _, opt := range opts {
		opt.Apply(c)
	}
	return c
}

// toEtcdConfig 转换为 etcd client v3 的 Config
func (c *Config) toEtcdConfig() (clientv3.Config, error) {
	cfg := clientv3.Config{
		Endpoints:          c.Endpoints,
		Username:           c.Username,
		Password:           c.Password,
		DialTimeout:        c.DialTimeout,
		AutoSyncInterval:   c.AutoSyncInterval,
		MaxCallSendMsgSize: c.MaxCallSendMsgSize,
		MaxCallRecvMsgSize: c.MaxCallRecvMsgSize,
		EnableCompression:  c.EnableCompression,
		EnableGRPCDebugLog: c.EnableGRPCDebugLog,
	}

	// 设置默认值
	if len(cfg.Endpoints) == 0 {
		cfg.Endpoints = []string{"localhost:2379"}
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 5 * time.Second
	}
	if cfg.MaxCallSendMsgSize == 0 {
		cfg.MaxCallSendMsgSize = 2 * 1024 * 1024 // 2MB
	}
	if cfg.MaxCallRecvMsgSize == 0 {
		cfg.MaxCallRecvMsgSize = 2 * 1024 * 1024 // 2MB
	}

	// TLS 配置
	if c.EnableTLS && c.TLSConfig != nil {
		tlsConfig, err := c.createTLSConfig()
		if err != nil {
			return cfg, err
		}
		cfg.TLS = tlsConfig
	}

	return cfg, nil
}

// createTLSConfig 创建 TLS 配置
func (c *Config) createTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.TLSConfig.InsecureSkipVerify,
	}

	// 加载 CA 证书
	if c.TLSConfig.CAFile != "" {
		caCert, err := os.ReadFile(c.TLSConfig.CAFile)
		if err != nil {
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, err
		}
		tlsConfig.RootCAs = caCertPool
	}

	// 加载客户端证书和密钥
	if c.TLSConfig.CertFile != "" && c.TLSConfig.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(c.TLSConfig.CertFile, c.TLSConfig.KeyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}
