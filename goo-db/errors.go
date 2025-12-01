package goodb

import "errors"

var (
	// ErrClientNotFound 客户端未找到
	ErrClientNotFound = errors.New("database client not found")
	// ErrInvalidDriver 无效的数据库驱动
	ErrInvalidDriver = errors.New("invalid database driver, only support mysql and postgres")
	// ErrEmptyDSN 空的 DSN
	ErrEmptyDSN = errors.New("empty DSN")
)

