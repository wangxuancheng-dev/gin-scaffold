package unit_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"

	"gin-scaffold/internal/config"
	"gin-scaffold/pkg/logger"
)

func TestLoggerBasicFeatures(t *testing.T) {
	logDir := filepath.Join(".", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("create logs dir failed: %v", err)
	}
	dailyPath := filepath.Join(logDir, "audit-"+time.Now().In(time.Local).Format("2006-01-02")+".log")
	for _, p := range []string{
		filepath.Join(logDir, "app.log"),
		filepath.Join(logDir, "access.log"),
		filepath.Join(logDir, "error.log"),
		filepath.Join(logDir, "del_user.log"),
		filepath.Join(logDir, "fixed.log"),
		dailyPath,
	} {
		_ = os.Remove(p)
	}

	cfg := &config.LogConfig{
		Level:        "debug",
		Dir:          logDir,
		AppFile:      "app.log",
		AccessFile:   "access.log",
		ErrorFile:    "error.log",
		RotationMode: "size",
		MaxSizeMB:    10,
		MaxBackups:   3,
		MaxAgeDays:   7,
		Compress:     false,
		Console:      false,
		Channels: map[string]config.LogChannelConfig{
			"plain": {
				Level:        "info",
				RotationMode: "none",
			},
			"daily": {
				// 不显式配置 level，验证默认回退为 info。
				RotationMode: "daily",
			},
			"fixed": {
				File:         "fixed.log",
				RotationMode: "none",
			},
		},
	}

	if err := logger.Init(cfg); err != nil {
		t.Fatalf("init logger failed: %v", err)
	}

	logger.InfoX("unit app log")
	logger.Access().Info("unit access log")
	logger.ErrorL().Error("unit error log", zap.String("scope", "unit"))
	logger.Channel("plain", "del_user.log").Info("unit delete user log", zap.Int64("target_user_id", 1001))
	logger.Channel("daily", "audit.log").Debug("unit daily debug should be filtered")
	logger.Channel("daily", "audit.log").Info("unit daily log")
	logger.Channel("fixed").Info("unit fixed log")
	logger.Sync()

	assertContains(t, filepath.Join(logDir, "app.log"), "unit app log")
	assertContains(t, filepath.Join(logDir, "access.log"), "unit access log")
	assertContains(t, filepath.Join(logDir, "error.log"), "unit error log")
	assertContains(t, filepath.Join(logDir, "del_user.log"), "unit delete user log")
	assertContains(t, filepath.Join(logDir, "del_user.log"), "\"target_user_id\":1001")
	assertContains(t, dailyPath, "unit daily log")
	assertNotContains(t, dailyPath, "unit daily debug should be filtered")
	assertContains(t, filepath.Join(logDir, "fixed.log"), "unit fixed log")
}

func assertContains(t *testing.T, filePath string, expected string) {
	t.Helper()
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read log file failed (%s): %v", filePath, err)
	}
	if !strings.Contains(string(content), expected) {
		t.Fatalf("log file %s missing expected content %q, got: %s", filePath, expected, string(content))
	}
}

func assertNotContains(t *testing.T, filePath string, expected string) {
	t.Helper()
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read log file failed (%s): %v", filePath, err)
	}
	if strings.Contains(string(content), expected) {
		t.Fatalf("log file %s unexpectedly contains %q, got: %s", filePath, expected, string(content))
	}
}
