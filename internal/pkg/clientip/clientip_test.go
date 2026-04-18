package clientip

import (
	"context"
	"testing"
)

func TestWith_FromContext(t *testing.T) {
	if FromContext(context.Background()) != "" {
		t.Fatal()
	}
	ctx := With(context.Background(), "203.0.113.1")
	if FromContext(ctx) != "203.0.113.1" {
		t.Fatal()
	}
}

func TestWith_NilContext(t *testing.T) {
	ctx := With(nil, "1.1.1.1")
	if FromContext(ctx) != "1.1.1.1" {
		t.Fatal()
	}
}
