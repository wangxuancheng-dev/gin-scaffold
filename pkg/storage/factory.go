package storage

import (
	"fmt"
	"strings"

	"gin-scaffold/internal/config"
)

// NewFromConfig 根据配置创建存储 Provider。
func NewFromConfig(cfg *config.StorageConfig) (Provider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("storage: nil config")
	}
	driver := strings.ToLower(strings.TrimSpace(cfg.Driver))
	if driver == "" {
		driver = "local"
	}
	switch driver {
	case "local":
		return NewLocalProvider(cfg.LocalDir, cfg.SignSecret)
	case "s3", "minio":
		return NewS3Provider(cfg)
	default:
		return nil, fmt.Errorf("storage: unsupported driver %q", cfg.Driver)
	}
}
