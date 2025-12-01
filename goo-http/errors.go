package goohttp

import "errors"

var (
	// ErrServerNotFound 服务器未找到
	ErrServerNotFound = errors.New("http server not found")
	// ErrEmptyAddress 地址为空
	ErrEmptyAddress = errors.New("address is empty")
	// ErrInvalidEncryptionKey 加密密钥无效
	ErrInvalidEncryptionKey = errors.New("encryption key is invalid")
)

