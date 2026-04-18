package tenant

import (
	"context"
	"testing"
)

func TestWithContext_FromContext(t *testing.T) {
	ctx := context.Background()
	if FromContext(ctx) != "" {
		t.Fatal()
	}
	ctx = WithContext(ctx, "  acme  ")
	if FromContext(ctx) != "acme" {
		t.Fatalf("got %q", FromContext(ctx))
	}
	ctx = WithContext(ctx, "")
	if FromContext(ctx) != "acme" {
		t.Fatal("empty WithContext should not clear")
	}
}

func TestWithContext_NilContext(t *testing.T) {
	ctx := WithContext(nil, "t1")
	if FromContext(ctx) != "t1" {
		t.Fatal()
	}
}
