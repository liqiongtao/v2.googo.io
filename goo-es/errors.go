package gooes

import "errors"

var (
	// ErrClientNotFound 客户端未找到
	ErrClientNotFound = errors.New("elasticsearch client not found")
	// ErrEmptyAddresses 空的地址列表
	ErrEmptyAddresses = errors.New("empty addresses")
)
