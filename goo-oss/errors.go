package goooss

import "errors"

var (
	// ErrClientNotFound 客户端未找到
	ErrClientNotFound = errors.New("oss client not found")
	// ErrEmptyAccessKeyID AccessKeyID 为空
	ErrEmptyAccessKeyID = errors.New("access key id is empty")
	// ErrEmptyAccessKeySecret AccessKeySecret 为空
	ErrEmptyAccessKeySecret = errors.New("access key secret is empty")
	// ErrEmptyEndpoint Endpoint 为空
	ErrEmptyEndpoint = errors.New("endpoint is empty")
	// ErrEmptyBucket Bucket 为空
	ErrEmptyBucket = errors.New("bucket is empty")
)

