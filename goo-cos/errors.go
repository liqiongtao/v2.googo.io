package goocos

import "errors"

var (
	// ErrClientNotFound 客户端未找到
	ErrClientNotFound = errors.New("cos client not found")
	// ErrEmptySecretID SecretID 为空
	ErrEmptySecretID = errors.New("secret id is empty")
	// ErrEmptySecretKey SecretKey 为空
	ErrEmptySecretKey = errors.New("secret key is empty")
	// ErrEmptyRegion Region 为空
	ErrEmptyRegion = errors.New("region is empty")
	// ErrEmptyBucket Bucket 为空
	ErrEmptyBucket = errors.New("bucket is empty")
)

