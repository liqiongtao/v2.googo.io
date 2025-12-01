package gooconsul

import (
	"context"
	"time"

	"github.com/hashicorp/consul/api"
)

// ServiceRegistration 服务注册信息
type ServiceRegistration struct {
	ID      string
	Name    string
	Tags    []string
	Address string
	Port    int
	Meta    map[string]string
	Check   *api.AgentServiceCheck
	Checks  api.AgentServiceChecks
}

// Client Consul 客户端封装
type Client struct {
	name   string
	config *Config
	client *api.Client
	agent  *api.Agent
	catalog *api.Catalog
	health *api.Health
}

// NewClient 创建新的 Consul 客户端
func NewClient(name string, config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证地址
	if config.Address == "" {
		return nil, ErrEmptyAddress
	}

	// 转换为 Consul API 的配置
	consulConfig, err := config.toConsulConfig()
	if err != nil {
		return nil, err
	}

	// 创建客户端
	client, err := api.NewClient(consulConfig)
	if err != nil {
		return nil, err
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Agent().Self()
	if err != nil {
		return nil, err
	}

	c := &Client{
		name:   name,
		config: config,
		client: client,
		agent:  client.Agent(),
		catalog: client.Catalog(),
		health: client.Health(),
	}

	return c, nil
}

// Name 获取客户端名称
func (c *Client) Name() string {
	return c.name
}

// Client 获取 Consul 客户端
func (c *Client) Client() *api.Client {
	return c.client
}

// Close 关闭客户端连接（Consul 客户端不需要显式关闭）
func (c *Client) Close() error {
	// Consul API 客户端不需要显式关闭
	return nil
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.agent.Self()
	return err
}

// RegisterService 注册服务
func (c *Client) RegisterService(registration *ServiceRegistration) error {
	if registration == nil {
		return ErrInvalidServiceID
	}

	serviceRegistration := &api.AgentServiceRegistration{
		ID:      registration.ID,
		Name:    registration.Name,
		Tags:    registration.Tags,
		Address: registration.Address,
		Port:    registration.Port,
		Meta:    registration.Meta,
	}

	if registration.Check != nil {
		serviceRegistration.Check = registration.Check
	}

	if len(registration.Checks) > 0 {
		serviceRegistration.Checks = registration.Checks
	}

	return c.agent.ServiceRegister(serviceRegistration)
}

// DeregisterService 注销服务
func (c *Client) DeregisterService(serviceID string) error {
	if serviceID == "" {
		return ErrInvalidServiceID
	}
	return c.agent.ServiceDeregister(serviceID)
}

// Service 获取服务信息
func (c *Client) Service(serviceID string, q *api.QueryOptions) (*api.AgentService, *api.QueryMeta, error) {
	if serviceID == "" {
		return nil, nil, ErrInvalidServiceID
	}
	return c.agent.Service(serviceID, q)
}

// Services 获取所有服务
func (c *Client) Services(q *api.QueryOptions) (map[string]*api.AgentService, error) {
	return c.agent.ServicesWithFilter("", q)
}

// ServiceHealth 获取服务健康状态
func (c *Client) ServiceHealth(serviceName string, passingOnly bool, q *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error) {
	if serviceName == "" {
		return nil, nil, ErrServiceNotFound
	}
	return c.health.Service(serviceName, "", passingOnly, q)
}

// ServiceNodes 获取服务节点列表
func (c *Client) ServiceNodes(serviceName string, tag string, passingOnly bool, q *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error) {
	if serviceName == "" {
		return nil, nil, ErrServiceNotFound
	}
	return c.health.Service(serviceName, tag, passingOnly, q)
}

// CatalogServices 获取目录服务列表
func (c *Client) CatalogServices(q *api.QueryOptions) (map[string][]string, *api.QueryMeta, error) {
	return c.catalog.Services(q)
}

// CatalogService 获取目录服务详情
func (c *Client) CatalogService(serviceName string, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error) {
	if serviceName == "" {
		return nil, nil, ErrServiceNotFound
	}
	return c.catalog.Service(serviceName, tag, q)
}

// WatchService 监听服务变化
func (c *Client) WatchService(ctx context.Context, serviceName string, tag string, passingOnly bool, handler func([]*api.ServiceEntry, error)) error {
	if serviceName == "" {
		return ErrServiceNotFound
	}

	// 使用 Health Service 进行 watch
	opts := &api.QueryOptions{
		WaitTime:  10 * time.Second,
		WaitIndex: 0,
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			entries, meta, err := c.health.Service(serviceName, tag, passingOnly, opts)
			if err != nil {
				handler(nil, err)
				return err
			}

			// 调用处理函数
			handler(entries, nil)

			// 更新 WaitIndex 以便下次获取更新
			opts.WaitIndex = meta.LastIndex
		}
	}
}

// WatchKey 监听键值变化
func (c *Client) WatchKey(ctx context.Context, key string, handler func(*api.KVPair, error)) error {
	if key == "" {
		return ErrServiceNotFound
	}

	kv := c.client.KV()
	opts := &api.QueryOptions{
		WaitTime:  10 * time.Second,
		WaitIndex: 0,
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			pair, meta, err := kv.Get(key, opts)
			if err != nil {
				handler(nil, err)
				return err
			}

			// 调用处理函数
			handler(pair, nil)

			// 更新 WaitIndex 以便下次获取更新
			if meta != nil {
				opts.WaitIndex = meta.LastIndex
			}
		}
	}
}

// WatchKeyPrefix 监听键前缀变化
func (c *Client) WatchKeyPrefix(ctx context.Context, prefix string, handler func(api.KVPairs, error)) error {
	if prefix == "" {
		return ErrServiceNotFound
	}

	kv := c.client.KV()
	opts := &api.QueryOptions{
		WaitTime:  10 * time.Second,
		WaitIndex: 0,
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			pairs, meta, err := kv.List(prefix, opts)
			if err != nil {
				handler(nil, err)
				return err
			}

			// 调用处理函数
			handler(pairs, nil)

			// 更新 WaitIndex 以便下次获取更新
			if meta != nil {
				opts.WaitIndex = meta.LastIndex
			}
		}
	}
}

// GetKV 获取键值
func (c *Client) GetKV(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error) {
	return c.client.KV().Get(key, q)
}

// PutKV 设置键值
func (c *Client) PutKV(pair *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error) {
	return c.client.KV().Put(pair, q)
}

// DeleteKV 删除键值
func (c *Client) DeleteKV(key string, q *api.WriteOptions) (*api.WriteMeta, error) {
	_, err := c.client.KV().Delete(key, q)
	if err != nil {
		return nil, err
	}
	return &api.WriteMeta{}, nil
}

// ListKV 列出键值（前缀匹配）
func (c *Client) ListKV(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error) {
	return c.client.KV().List(prefix, q)
}

