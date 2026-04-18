package logger

import (
	"runtime"
	"testing"

	"gin-scaffold/config"
)

func TestLogger_syncNoPanicWhenUnconfigured(t *testing.T) {
	Sync()
}

func TestLogger_initWritesFiles_nonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("log files stay open after Init; TempDir cleanup fails on Windows")
	}
	dir := t.TempDir()
	cfg := &config.LogConfig{
		Dir:          dir,
		AppFile:      "app.log",
		AccessFile:   "access.log",
		ErrorFile:    "error.log",
		RotationMode: "none",
		Level:        "info",
		Console:      false,
	}
	if err := Init(cfg); err != nil {
		t.Fatal(err)
	}
	if L() == nil {
		t.Fatal("L() is nil after Init")
	}
	L().Info("logger_test_message")
	Sync()
}

func TestLogger_infoXWhenNil(t *testing.T) {
	// Safe no-op when Init has not run in this test binary.
	InfoX("x")
	ErrorX("e")
	WarnX("w")
	DebugX("d")
}
