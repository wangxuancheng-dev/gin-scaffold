package storage

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalProviderStatObject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p, err := NewLocalProvider(dir, "test-secret-for-stat")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = p.StatObject(ctx, "missing.bin")
	if err == nil || !errors.Is(err, ErrObjectNotExist) {
		t.Fatalf("StatObject missing: got err=%v", err)
	}
	key := "a/b/c.txt"
	full := filepath.Join(dir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	payload := []byte("hello")
	if err := os.WriteFile(full, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := p.StatObject(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if st.Size != int64(len(payload)) {
		t.Fatalf("size: want %d got %d", len(payload), st.Size)
	}
}

func TestLocalProviderReady(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p, err := NewLocalProvider(dir, "test-secret-for-ready")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := p.Ready(context.Background()); err == nil {
		t.Fatal("expected error when base dir removed")
	}
}
