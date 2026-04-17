package storage

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Provider 定义文件存储能力。
type Provider interface {
	Put(ctx context.Context, key string, reader io.Reader) error
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Sign(key string, expireSec int64) (string, error)
	Verify(key string, expires int64, sig string) bool
}

// PutContentTyper 可选：上传时写入 Content-Type（S3 等对象存储）。
type PutContentTyper interface {
	PutContentType(ctx context.Context, key string, contentType string, reader io.Reader) error
}

// PresignPutProvider 可选：生成直传 PUT 预签名 URL（S3/MinIO）。
type PresignPutProvider interface {
	PresignPutURL(ctx context.Context, key string, contentType string, expire time.Duration) (method string, url string, headers map[string]string, err error)
}

var global Provider

// InitDefault 初始化全局 provider。
func InitDefault(p Provider) {
	global = p
}

// Default 返回全局 provider。
func Default() Provider {
	return global
}

// Require 获取全局 provider，不存在则返回错误。
func Require() (Provider, error) {
	if global == nil {
		return nil, fmt.Errorf("storage: provider not initialized")
	}
	return global, nil
}
