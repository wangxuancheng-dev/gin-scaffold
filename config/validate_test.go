package config

import (
	"strings"
	"testing"
)

func TestValidate_Ok(t *testing.T) {
	cfg := &App{
		Env:  "dev",
		Name: "gin-scaffold",
		HTTP: HTTPConfig{
			Host:              "0.0.0.0",
			Port:              8080,
			ReadTimeout:       30,
			ReadHeaderTimeout: 10,
			WriteTimeout:      30,
			IdleTimeout:       120,
			ShutdownTimeout:   10,
		},
		DB: DBConfig{
			Driver: "mysql",
			DSN:    "root:root@tcp(127.0.0.1:3306)/scaffold?charset=utf8mb4&parseTime=True",
		},
		Redis: RedisConfig{
			Addr: "127.0.0.1:6379",
		},
		Asynq: AsynqConfig{
			RedisAddr:      "127.0.0.1:6379",
			RedisDB:        1,
			Concurrency:    10,
			Queue:          "default",
			Queues:         map[string]int{"critical": 6, "default": 3, "low": 1},
			MaxRetry:       5,
			TimeoutSec:     30,
			DedupWindowSec: 30,
		},
		JWT: JWTConfig{
			Secret:           "test-secret",
			AccessExpireMin:  60,
			RefreshExpireMin: 1440,
		},
		I18n: I18nConfig{
			DefaultLang: "zh",
			BundlePaths: []string{"./i18n/zh.json"},
		},
		Scheduler: SchedulerConfig{
			LogRetentionDays: 30,
			LockTTLSeconds:   120,
		},
		Storage: StorageConfig{
			Enabled:      true,
			Driver:       "local",
			LocalDir:     "./storage-test",
			SignSecret:   "unit-test-secret",
			MaxUploadMB:  5,
			AllowedMIME:  "text/plain",
			URLExpireSec: 60,
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("validate should pass, got: %v", err)
	}
}

func TestValidate_FailFast(t *testing.T) {
	cfg := &App{
		Env:  "dev",
		Name: "gin-scaffold",
		HTTP: HTTPConfig{Host: "0.0.0.0", Port: 8080},
		DB:   DBConfig{Driver: "mysql", DSN: "dsn"},
		Redis: RedisConfig{
			Addr: "127.0.0.1:6379",
		},
		Asynq: AsynqConfig{
			RedisAddr: "127.0.0.1:6379",
			Queue:     "default",
			Queues:    map[string]int{"default": 1},
		},
		JWT: JWTConfig{
			Secret:           "",
			AccessExpireMin:  60,
			RefreshExpireMin: 1440,
		},
		I18n: I18nConfig{
			DefaultLang: "zh",
			BundlePaths: []string{"./i18n/zh.json"},
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatalf("validate should fail when jwt.secret is empty")
	}
}

func TestValidate_AggregatesErrors(t *testing.T) {
	cfg := &App{
		Env:  "",
		Name: "",
		HTTP: HTTPConfig{
			Host: "",
			Port: 0,
		},
		DB: DBConfig{
			Driver: "sqlite",
			DSN:    "",
		},
		Redis: RedisConfig{
			Addr: "",
			DB:   -1,
		},
		Asynq: AsynqConfig{
			RedisAddr:      "",
			Queue:          "",
			Queues:         map[string]int{"": 0},
			MaxRetry:       -1,
			DedupWindowSec: -1,
		},
		JWT: JWTConfig{
			Secret:           "",
			AccessExpireMin:  0,
			RefreshExpireMin: 0,
		},
		I18n: I18nConfig{},
		Scheduler: SchedulerConfig{
			LogRetentionDays: -1,
			LockTTLSeconds:   -1,
		},
		Storage: StorageConfig{
			Enabled:      true,
			Driver:       "ftp",
			LocalDir:     "",
			SignSecret:   "",
			MaxUploadMB:  0,
			AllowedMIME:  ",",
			URLExpireSec: 0,
		},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatalf("validate should fail for invalid config")
	}
	msg := err.Error()
	wantParts := []string{
		"env is required",
		"name is required",
		"http.host is required",
		"db.driver must be mysql or postgres",
		"redis.addr is required",
		"asynq.redis_addr is required",
		"asynq.queues key must not be empty",
		"asynq.dedup_window_sec must be >= 0",
		"jwt.secret is required",
		"i18n.default_lang is required",
		"scheduler.log_retention_days must be >= 0",
		"storage.driver must be local, s3, or minio",
		"storage.sign_secret is required when storage.enabled=true",
		"storage.max_upload_mb must be > 0 when storage.enabled=true",
		"storage.url_expire_sec must be > 0 when storage.enabled=true",
		"storage.allowed_mime must not contain empty entries",
	}
	for _, part := range wantParts {
		if !strings.Contains(msg, part) {
			t.Fatalf("validate error should contain %q, got: %s", part, msg)
		}
	}
}
