package gooetcd

import "errors"

var (
	// ErrClientNotFound 客户端未找到
	ErrClientNotFound = errors.New("etcd client not found")
	// ErrEmptyEndpoints 空的端点列表
	ErrEmptyEndpoints = errors.New("empty endpoints")
)
