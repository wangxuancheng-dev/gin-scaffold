package tracer

import (
	"context"
	"testing"

	"gin-scaffold/internal/config"
)

func TestInit_disabled(t *testing.T) {
	shutdown, err := Init(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown func")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestInit_disabledEmptyEndpoint(t *testing.T) {
	shutdown, err := Init(context.Background(), &config.TraceConfig{Enabled: true, Endpoint: ""})
	if err != nil {
		t.Fatal(err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}
