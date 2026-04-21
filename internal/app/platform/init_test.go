package platform

import (
	"testing"

	"gin-scaffold/internal/config"
	"gin-scaffold/pkg/notify"
)

func TestInit_nilUsesLogNotifier(t *testing.T) {
	t.Cleanup(func() { notify.SetDefault(nil) })
	Init(nil)
	if _, ok := notify.Default().(notify.LogNotifier); !ok {
		t.Fatalf("expected LogNotifier, got %T", notify.Default())
	}
}

func TestInit_noopDriver(t *testing.T) {
	t.Cleanup(func() { notify.SetDefault(nil) })
	cfg := &config.App{
		Platform: config.PlatformConfig{
			Notify: config.NotifyConfig{Driver: "noop"},
		},
	}
	Init(cfg)
	if _, ok := notify.Default().(notify.Noop); !ok {
		t.Fatalf("expected Noop, got %T", notify.Default())
	}
}
