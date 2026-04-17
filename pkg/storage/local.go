package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalProvider 本地文件系统实现。
type LocalProvider struct {
	baseDir string
	secret  []byte
}

func NewLocalProvider(baseDir, signSecret string) (*LocalProvider, error) {
	if strings.TrimSpace(baseDir) == "" {
		return nil, fmt.Errorf("storage local: base dir is empty")
	}
	if strings.TrimSpace(signSecret) == "" {
		return nil, fmt.Errorf("storage local: sign secret is empty")
	}
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, err
	}
	return &LocalProvider{baseDir: baseDir, secret: []byte(signSecret)}, nil
}

func (p *LocalProvider) Put(_ context.Context, key string, reader io.Reader) error {
	path, err := p.absPath(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, reader)
	return err
}

func (p *LocalProvider) Open(_ context.Context, key string) (io.ReadCloser, error) {
	path, err := p.absPath(key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (p *LocalProvider) Delete(_ context.Context, key string) error {
	path, err := p.absPath(key)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func (p *LocalProvider) Sign(key string, expireSec int64) (string, error) {
	return signDownload(p.secret, key, expireSec)
}

func (p *LocalProvider) Verify(key string, expires int64, sig string) bool {
	return verifyDownload(p.secret, key, expires, sig)
}

func (p *LocalProvider) absPath(key string) (string, error) {
	normalized := normalizeKey(key)
	if normalized == "" {
		return "", fmt.Errorf("storage local: empty key")
	}
	root := filepath.Clean(p.baseDir)
	target := filepath.Clean(filepath.Join(root, normalized))
	if target != root && !strings.HasPrefix(target, root+string(filepath.Separator)) {
		return "", fmt.Errorf("storage local: invalid key path")
	}
	return target, nil
}

func normalizeKey(key string) string {
	k := strings.TrimSpace(key)
	k = strings.TrimPrefix(k, "/")
	k = strings.ReplaceAll(k, "\\", "/")
	k = filepath.ToSlash(filepath.Clean(k))
	if k == "." {
		return ""
	}
	return k
}
