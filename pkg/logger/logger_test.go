package logger

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"gin-scaffold/config"
)

func resetLoggerForTest() {
	appL = nil
	accessL = nil
	errL = nil
	channelLogs = nil
	once = sync.Once{}
}

func TestInitAndWriteLogFiles(t *testing.T) {
	resetLoggerForTest()
	t.Cleanup(resetLoggerForTest)

	oldLocal := time.Local
	time.Local = time.UTC
	t.Cleanup(func() { time.Local = oldLocal })

	logDir, err := os.MkdirTemp("", "logger-unit-*")
	if err != nil {
		t.Fatalf("create temp dir failed: %v", err)
	}
	cfg := &config.LogConfig{
		Level:      "debug",
		Dir:        logDir,
		AppFile:    "app.log",
		AccessFile: "access.log",
		ErrorFile:  "error.log",
		RotationMode: "size",
		MaxSizeMB:  10,
		MaxBackups: 3,
		MaxAgeDays: 7,
		Compress:   false,
		Console:    false,
	}

	if err := Init(cfg); err != nil {
		t.Fatalf("Init logger failed: %v", err)
	}

	InfoX("logger unit test app message")
	Access().Info("logger unit test access message")
	ErrorL().Error("logger unit test error message", zap.String("scope", "unit"))
	Sync()

	appContent, err := os.ReadFile(filepath.Join(logDir, cfg.AppFile))
	if err != nil {
		t.Fatalf("read app log failed: %v", err)
	}
	accessContent, err := os.ReadFile(filepath.Join(logDir, cfg.AccessFile))
	if err != nil {
		t.Fatalf("read access log failed: %v", err)
	}
	errorContent, err := os.ReadFile(filepath.Join(logDir, cfg.ErrorFile))
	if err != nil {
		t.Fatalf("read error log failed: %v", err)
	}

	appText := string(appContent)
	accessText := string(accessContent)
	errorText := string(errorContent)

	if !strings.Contains(appText, "logger unit test app message") {
		t.Fatalf("app log missing expected message, content=%s", appText)
	}
	if !strings.Contains(accessText, "logger unit test access message") {
		t.Fatalf("access log missing expected message, content=%s", accessText)
	}
	if !strings.Contains(errorText, "logger unit test error message") {
		t.Fatalf("error log missing expected message, content=%s", errorText)
	}
	if !strings.Contains(appText, "\"ts\":\"") {
		t.Fatalf("app log missing ts field, content=%s", appText)
	}
	if !strings.Contains(appText, "Z\"") && !strings.Contains(appText, "+0000\"") {
		t.Fatalf("app log not in UTC offset format, content=%s", appText)
	}
}

func TestDailyRotationByConfiguredTimeZone(t *testing.T) {
	resetLoggerForTest()
	t.Cleanup(resetLoggerForTest)

	oldLocal := time.Local
	time.Local = time.UTC
	t.Cleanup(func() { time.Local = oldLocal })

	oldNow := nowFunc
	var fakeNow time.Time
	nowFunc = func() time.Time { return fakeNow }
	t.Cleanup(func() { nowFunc = oldNow })

	logDir, err := os.MkdirTemp("", "logger-daily-*")
	if err != nil {
		t.Fatalf("create temp dir failed: %v", err)
	}
	cfg := &config.LogConfig{
		Level:        "info",
		Dir:          logDir,
		AppFile:      "app.log",
		AccessFile:   "access.log",
		ErrorFile:    "error.log",
		RotationMode: "daily",
		MaxAgeDays:   7,
		Console:      false,
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init logger failed: %v", err)
	}

	fakeNow = time.Date(2026, 4, 15, 23, 59, 50, 0, time.UTC)
	InfoX("before midnight")
	Sync()

	fakeNow = time.Date(2026, 4, 16, 0, 0, 10, 0, time.UTC)
	InfoX("after midnight")
	Sync()

	day1Path := filepath.Join(logDir, "app-2026-04-15.log")
	day2Path := filepath.Join(logDir, "app-2026-04-16.log")
	if _, err := os.Stat(day1Path); err != nil {
		t.Fatalf("expected day1 log file not found: %v", err)
	}
	if _, err := os.Stat(day2Path); err != nil {
		t.Fatalf("expected day2 log file not found: %v", err)
	}

	day1Content, err := os.ReadFile(day1Path)
	if err != nil {
		t.Fatalf("read day1 log failed: %v", err)
	}
	day2Content, err := os.ReadFile(day2Path)
	if err != nil {
		t.Fatalf("read day2 log failed: %v", err)
	}
	if !strings.Contains(string(day1Content), "before midnight") {
		t.Fatalf("day1 file missing message: %s", string(day1Content))
	}
	if !strings.Contains(string(day2Content), "after midnight") {
		t.Fatalf("day2 file missing message: %s", string(day2Content))
	}
}

func TestPerFileRotationModeOverride(t *testing.T) {
	resetLoggerForTest()
	t.Cleanup(resetLoggerForTest)

	oldLocal := time.Local
	time.Local = time.UTC
	t.Cleanup(func() { time.Local = oldLocal })

	oldNow := nowFunc
	fakeNow := time.Date(2026, 4, 16, 0, 0, 10, 0, time.UTC)
	nowFunc = func() time.Time { return fakeNow }
	t.Cleanup(func() { nowFunc = oldNow })

	logDir, err := os.MkdirTemp("", "logger-mode-*")
	if err != nil {
		t.Fatalf("create temp dir failed: %v", err)
	}
	cfg := &config.LogConfig{
		Level:              "info",
		Dir:                logDir,
		AppFile:            "app.log",
		AccessFile:         "access.log",
		ErrorFile:          "error.log",
		RotationMode:       "size",
		AppRotationMode:    "none",
		AccessRotationMode: "daily",
		ErrorRotationMode:  "size",
		MaxAgeDays:         7,
		Console:            false,
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init logger failed: %v", err)
	}

	InfoX("app plain file")
	Access().Info("access daily file")
	ErrorL().Error("error size file")
	Sync()

	if _, err := os.Stat(filepath.Join(logDir, "app.log")); err != nil {
		t.Fatalf("expected app.log for none mode, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(logDir, "access-2026-04-16.log")); err != nil {
		t.Fatalf("expected dated access log for daily mode, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(logDir, "error.log")); err != nil {
		t.Fatalf("expected error.log for size mode, got err=%v", err)
	}
}

func TestCustomChannelDailyAndLevel(t *testing.T) {
	resetLoggerForTest()
	t.Cleanup(resetLoggerForTest)

	oldLocal := time.Local
	time.Local = time.UTC
	t.Cleanup(func() { time.Local = oldLocal })

	oldNow := nowFunc
	fakeNow := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)
	nowFunc = func() time.Time { return fakeNow }
	t.Cleanup(func() { nowFunc = oldNow })

	logDir, err := os.MkdirTemp("", "logger-channel-*")
	if err != nil {
		t.Fatalf("create temp dir failed: %v", err)
	}
	cfg := &config.LogConfig{
		Level:        "debug",
		Dir:          logDir,
		AppFile:      "app.log",
		AccessFile:   "access.log",
		ErrorFile:    "error.log",
		RotationMode: "size",
		MaxAgeDays:   7,
		Console:      false,
		Channels: map[string]config.LogChannelConfig{
			"del_user": {
				File:         "del_user.log",
				Level:        "warn",
				RotationMode: "daily",
			},
		},
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init logger failed: %v", err)
	}

	Channel("del_user").Info("should be filtered")
	Channel("del_user").Warn("delete user warn log")
	Sync()

	path := filepath.Join(logDir, "del_user-2026-04-16.log")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read del_user channel log failed: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "should be filtered") {
		t.Fatalf("channel level filtering failed, content=%s", text)
	}
	if !strings.Contains(text, "delete user warn log") {
		t.Fatalf("channel message missing, content=%s", text)
	}
}

func TestUsageExample_DeleteUserAuditChannel(t *testing.T) {
	resetLoggerForTest()
	t.Cleanup(resetLoggerForTest)

	oldLocal := time.Local
	time.Local = time.UTC
	t.Cleanup(func() { time.Local = oldLocal })

	oldNow := nowFunc
	fakeNow := time.Date(2026, 4, 17, 9, 30, 0, 0, time.UTC)
	nowFunc = func() time.Time { return fakeNow }
	t.Cleanup(func() { nowFunc = oldNow })

	logDir, err := os.MkdirTemp("", "logger-usage-*")
	if err != nil {
		t.Fatalf("create temp dir failed: %v", err)
	}
	cfg := &config.LogConfig{
		Level:        "info",
		Dir:          logDir,
		AppFile:      "app.log",
		AccessFile:   "access.log",
		ErrorFile:    "error.log",
		RotationMode: "size",
		Console:      false,
		Channels: map[string]config.LogChannelConfig{
			"del_user": {
				File:         "del_user.log",
				Level:        "info",
				RotationMode: "daily",
			},
		},
	}
	if err := Init(cfg); err != nil {
		t.Fatalf("Init logger failed: %v", err)
	}

	// 业务示例：删除用户审计日志
	Channel("del_user").Info("admin delete user",
		zap.Int64("operator_id", 1),
		zap.Int64("target_user_id", 10086),
		zap.String("result", "success"),
		zap.String("request_id", "req-123"),
		zap.String("ip", "127.0.0.1"),
	)
	Sync()

	path := filepath.Join(logDir, "del_user-2026-04-17.log")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read del_user usage log failed: %v", err)
	}
	text := string(content)
	for _, expect := range []string{
		"admin delete user",
		"\"operator_id\":1",
		"\"target_user_id\":10086",
		"\"result\":\"success\"",
		"\"request_id\":\"req-123\"",
		"\"ip\":\"127.0.0.1\"",
	} {
		if !strings.Contains(text, expect) {
			t.Fatalf("usage log missing %s, content=%s", expect, text)
		}
	}
}
