package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

// ErrObjectNotExist 表示对象不存在（用于 Stat / 完成校验等）。
var ErrObjectNotExist = errors.New("storage: object not found")

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

// PresignPutOptions 预签名 PUT 的可选约束（客户端需按返回的 headers 上传，完成后可校验）。
type PresignPutOptions struct {
	ContentLength int64             // >0 时约束 Content-Length（需与 PUT 一致）
	Metadata      map[string]string // 写入 x-amz-meta-*（键名不含前缀，建议小写）
}

// PresignPutProvider 可选：生成直传 PUT 预签名 URL（S3/MinIO）。
type PresignPutProvider interface {
	PresignPutURL(ctx context.Context, key string, contentType string, expire time.Duration, opts *PresignPutOptions) (method string, url string, headers map[string]string, err error)
}

// ObjectStat 对象元信息（Head/Stat）。
type ObjectStat struct {
	Size         int64
	ContentType  string
	Metadata     map[string]string // 键已转成小写便于比对
	ETag         string
	DeleteMarker bool
}

// ObjectStatProvider 可选：Head/Stat 对象（用于上传完成确认）。
type ObjectStatProvider interface {
	StatObject(ctx context.Context, key string) (*ObjectStat, error)
}

// ReadinessChecker 可选：就绪探活（如 S3 HeadBucket、本地目录 stat）。
type ReadinessChecker interface {
	Ready(ctx context.Context) error
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
