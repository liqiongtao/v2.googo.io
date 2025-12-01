package gooredis

import "errors"

var (
	// ErrClientNotFound 客户端未找到
	ErrClientNotFound = errors.New("redis client not found")
)

