package googrpc

import "errors"

var (
	// ErrClientNotFound 客户端未找到
	ErrClientNotFound = errors.New("grpc client not found")
	// ErrServerNotFound 服务端未找到
	ErrServerNotFound = errors.New("grpc server not found")
	// ErrEmptyAddress 空的地址
	ErrEmptyAddress = errors.New("empty address")
	// ErrEmptyServiceName 空的服务名称
	ErrEmptyServiceName = errors.New("empty service name")
	// ErrRegistryNotSupported 不支持的注册中心类型
	ErrRegistryNotSupported = errors.New("registry not supported")
)
