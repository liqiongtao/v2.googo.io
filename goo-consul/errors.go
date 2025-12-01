package gooconsul

import "errors"

var (
	// ErrClientNotFound 客户端未找到
	ErrClientNotFound = errors.New("consul client not found")
	// ErrEmptyAddress 空的地址
	ErrEmptyAddress = errors.New("empty address")
	// ErrServiceNotFound 服务未找到
	ErrServiceNotFound = errors.New("service not found")
	// ErrInvalidServiceID 无效的服务 ID
	ErrInvalidServiceID = errors.New("invalid service id")
)

