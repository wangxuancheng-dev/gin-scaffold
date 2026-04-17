package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLocalProviderPutOpenVerify(t *testing.T) {
	dir := t.TempDir()
	p, err := NewLocalProvider(dir, "test-secret")
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	key := "20260416/demo.txt"
	if err := p.Put(context.Background(), key, strings.NewReader("hello")); err != nil {
		t.Fatalf("put: %v", err)
	}
	rc, err := p.Open(context.Background(), key)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(b) != "hello" {
		t.Fatalf("unexpected content: %s", string(b))
	}
	expires := time.Now().Unix() + 60
	sig, err := p.Sign(key, expires)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if !p.Verify(key, expires, sig) {
		t.Fatalf("verify should pass")
	}
	if p.Verify(key, expires, "bad-signature") {
		t.Fatalf("verify should fail for bad signature")
	}
}

func TestLocalProviderRejectPathTraversal(t *testing.T) {
	dir := t.TempDir()
	p, err := NewLocalProvider(dir, "test-secret")
	if err != nil {
		t.Fatalf("new provider: %v", err)
	}
	err = p.Put(context.Background(), "../../etc/passwd", strings.NewReader("x"))
	if err == nil {
		t.Fatalf("expected path traversal error")
	}
	_, statErr := os.Stat(filepath.Join(dir, "..", "..", "etc", "passwd"))
	if statErr == nil {
		t.Fatalf("unexpected file created outside base dir")
	}
}
