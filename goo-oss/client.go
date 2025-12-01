package goooss

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// Client OSS 客户端封装
type Client struct {
	name   string
	config *Config
	client *oss.Client
	bucket *oss.Bucket
}

// NewClient 创建新的 OSS 客户端
func NewClient(name string, config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证必填字段
	if config.AccessKeyID == "" {
		return nil, ErrEmptyAccessKeyID
	}
	if config.AccessKeySecret == "" {
		return nil, ErrEmptyAccessKeySecret
	}
	if config.Endpoint == "" {
		return nil, ErrEmptyEndpoint
	}
	if config.Bucket == "" {
		return nil, ErrEmptyBucket
	}

	// 获取认证信息
	accessKeyID, accessKeySecret, securityToken := config.getCredentials()

	// 创建客户端选项
	clientOptions := []oss.ClientOption{}
	if config.Timeout > 0 {
		clientOptions = append(clientOptions, oss.Timeout(config.Timeout))
	}

	// 如果使用 STS 临时密钥
	if securityToken != "" {
		clientOptions = append(clientOptions, oss.SecurityToken(securityToken))
	}

	// 创建 OSS 客户端
	client, err := oss.New(config.Endpoint, accessKeyID, accessKeySecret, clientOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OSS client: %w", err)
	}

	// 获取存储桶
	bucket, err := client.Bucket(config.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}

	// 测试连接（使用一个不存在的对象键来测试连接）
	exist, err := bucket.IsObjectExist("test-connection-check")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to OSS: %w", err)
	}
	_ = exist // 忽略结果，仅用于测试连接

	c := &Client{
		name:   name,
		config: config,
		client: client,
		bucket: bucket,
	}

	return c, nil
}

// Name 获取客户端名称
func (c *Client) Name() string {
	return c.name
}

// Client 获取 OSS 客户端
func (c *Client) Client() *oss.Client {
	return c.client
}

// Bucket 获取存储桶对象
func (c *Client) Bucket() *oss.Bucket {
	return c.bucket
}

// Close 关闭客户端连接（OSS 客户端通常不需要显式关闭）
func (c *Client) Close() error {
	// OSS SDK 的 HTTP 客户端会自动管理连接，无需显式关闭
	return nil
}

// Ping 测试连接
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.bucket.IsObjectExist("test-connection-check")
	return err
}

// PutObject 上传对象
func (c *Client) PutObject(ctx context.Context, objectKey string, reader io.Reader, options ...oss.Option) error {
	return c.bucket.PutObject(objectKey, reader, options...)
}

// GetObject 获取对象
func (c *Client) GetObject(ctx context.Context, objectKey string, options ...oss.Option) (io.ReadCloser, error) {
	return c.bucket.GetObject(objectKey, options...)
}

// DeleteObject 删除对象
func (c *Client) DeleteObject(ctx context.Context, objectKey string, options ...oss.Option) error {
	return c.bucket.DeleteObject(objectKey, options...)
}

// HeadObject 获取对象元信息
func (c *Client) HeadObject(ctx context.Context, objectKey string, options ...oss.Option) (oss.GetObjectMetaResult, error) {
	return c.bucket.GetObjectMeta(objectKey, options...)
}

// CopyObject 复制对象
func (c *Client) CopyObject(ctx context.Context, destObjectKey, srcObjectKey string, options ...oss.Option) (oss.CopyObjectResult, error) {
	return c.bucket.CopyObject(srcObjectKey, destObjectKey, options...)
}

// ListObjects 列出对象
func (c *Client) ListObjects(ctx context.Context, options ...oss.Option) (oss.ListObjectsResult, error) {
	return c.bucket.ListObjects(options...)
}

// PutObjectACL 设置对象 ACL
func (c *Client) PutObjectACL(ctx context.Context, objectKey string, objectACL oss.ACLType, options ...oss.Option) error {
	return c.bucket.SetObjectACL(objectKey, objectACL, options...)
}

// GetObjectACL 获取对象 ACL
func (c *Client) GetObjectACL(ctx context.Context, objectKey string, options ...oss.Option) (oss.GetObjectACLResult, error) {
	return c.bucket.GetObjectACL(objectKey, options...)
}

// InitiateMultipartUpload 初始化分片上传
func (c *Client) InitiateMultipartUpload(ctx context.Context, objectKey string, options ...oss.Option) (oss.InitiateMultipartUploadResult, error) {
	return c.bucket.InitiateMultipartUpload(objectKey, options...)
}

// UploadPart 上传分片
func (c *Client) UploadPart(ctx context.Context, imur oss.InitiateMultipartUploadResult, partNumber int, reader io.Reader, options ...oss.Option) (oss.UploadPart, error) {
	return c.bucket.UploadPart(imur, reader, partNumber, options...)
}

// CompleteMultipartUpload 完成分片上传
func (c *Client) CompleteMultipartUpload(ctx context.Context, imur oss.InitiateMultipartUploadResult, parts []oss.UploadPart, options ...oss.Option) (oss.CompleteMultipartUploadResult, error) {
	return c.bucket.CompleteMultipartUpload(imur, parts, options...)
}

// AbortMultipartUpload 取消分片上传
func (c *Client) AbortMultipartUpload(ctx context.Context, imur oss.InitiateMultipartUploadResult, options ...oss.Option) error {
	return c.bucket.AbortMultipartUpload(imur, options...)
}

// ListMultipartUploads 列出进行中的分片上传
func (c *Client) ListMultipartUploads(ctx context.Context, options ...oss.Option) (oss.ListMultipartUploadResult, error) {
	return c.bucket.ListMultipartUploads(options...)
}

// IsObjectExist 检查对象是否存在
func (c *Client) IsObjectExist(ctx context.Context, objectKey string, options ...oss.Option) (bool, error) {
	// 注意：OSS SDK 的 IsObjectExist 不支持 context，但保留 context 参数以保持接口一致性
	return c.bucket.IsObjectExist(objectKey, options...)
}

// SignURL 生成签名 URL
func (c *Client) SignURL(ctx context.Context, objectKey string, method oss.HTTPMethod, expiredInSec int64, options ...oss.Option) (string, error) {
	return c.bucket.SignURL(objectKey, method, expiredInSec, options...)
}

