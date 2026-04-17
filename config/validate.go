package config

import (
	"fmt"
	"strings"
)

// Validate 校验关键配置，启动阶段失败即退出（fail fast）。
func (a *App) Validate() error {
	if a == nil {
		return fmt.Errorf("config is nil")
	}
	var errs []string
	if strings.TrimSpace(a.Env) == "" {
		errs = append(errs, "env is required")
	}
	if strings.TrimSpace(a.Name) == "" {
		errs = append(errs, "name is required")
	}
	errs = append(errs, a.validateHTTP()...)
	errs = append(errs, a.validateDB()...)
	errs = append(errs, a.validateRedis()...)
	errs = append(errs, a.validateAsynq()...)
	errs = append(errs, a.validateJWT()...)
	errs = append(errs, a.validateI18n()...)
	errs = append(errs, a.validateScheduler()...)
	errs = append(errs, a.validateCORS()...)
	errs = append(errs, a.validateOutbound()...)
	errs = append(errs, a.validateStorage()...)
	if len(errs) > 0 {
		return fmt.Errorf("invalid config: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (a *App) validateHTTP() []string {
	var errs []string
	if strings.TrimSpace(a.HTTP.Host) == "" {
		errs = append(errs, "http.host is required")
	}
	if a.HTTP.Port <= 0 || a.HTTP.Port > 65535 {
		errs = append(errs, "http.port must be between 1 and 65535")
	}
	if a.HTTP.ReadTimeout < 0 || a.HTTP.ReadHeaderTimeout < 0 || a.HTTP.WriteTimeout < 0 || a.HTTP.IdleTimeout < 0 || a.HTTP.ShutdownTimeout < 0 {
		errs = append(errs, "http timeout values must be >= 0")
	}
	if a.HTTP.MaxBodyBytes < 0 {
		errs = append(errs, "http.max_body_bytes must be >= 0")
	}
	return errs
}

func (a *App) validateDB() []string {
	var errs []string
	driver := strings.ToLower(strings.TrimSpace(a.DB.Driver))
	if driver != "mysql" && driver != "postgres" {
		errs = append(errs, "db.driver must be mysql or postgres")
	}
	if strings.TrimSpace(a.DB.DSN) == "" {
		errs = append(errs, "db.dsn is required")
	}
	if a.DB.MaxOpenConns < 0 || a.DB.MaxIdleConns < 0 || a.DB.ConnMaxLifetimeSec < 0 || a.DB.ConnMaxIdleTimeSec < 0 || a.DB.SlowThresholdMS < 0 {
		errs = append(errs, "db pool/threshold values must be >= 0")
	}
	return errs
}

func (a *App) validateRedis() []string {
	var errs []string
	if strings.TrimSpace(a.Redis.Addr) == "" {
		errs = append(errs, "redis.addr is required")
	}
	if a.Redis.DB < 0 || a.Redis.PoolSize < 0 || a.Redis.MinIdleConns < 0 || a.Redis.DialTimeout < 0 || a.Redis.ReadTimeout < 0 || a.Redis.WriteTimeout < 0 {
		errs = append(errs, "redis values must be >= 0")
	}
	return errs
}

func (a *App) validateJWT() []string {
	var errs []string
	if strings.TrimSpace(a.JWT.Secret) == "" {
		errs = append(errs, "jwt.secret is required")
	}
	if a.JWT.AccessExpireMin <= 0 || a.JWT.RefreshExpireMin <= 0 {
		errs = append(errs, "jwt.access_expire_min and jwt.refresh_expire_min must be > 0")
	}
	return errs
}

func (a *App) validateAsynq() []string {
	var errs []string
	if strings.TrimSpace(a.Asynq.RedisAddr) == "" {
		errs = append(errs, "asynq.redis_addr is required")
	}
	if a.Asynq.RedisDB < 0 || a.Asynq.Concurrency < 0 {
		errs = append(errs, "asynq.redis_db and asynq.concurrency must be >= 0")
	}
	if strings.TrimSpace(a.Asynq.Queue) == "" {
		if len(a.Asynq.Queues) == 0 {
			errs = append(errs, "asynq.queue is required (or configure asynq.queues)")
		}
	}
	for q, w := range a.Asynq.Queues {
		if strings.TrimSpace(q) == "" {
			errs = append(errs, "asynq.queues key must not be empty")
			continue
		}
		if w <= 0 {
			errs = append(errs, "asynq.queues weight must be > 0")
		}
	}
	if a.Asynq.MaxRetry < 0 {
		errs = append(errs, "asynq.max_retry must be >= 0")
	}
	if a.Asynq.TimeoutSec < 0 {
		errs = append(errs, "asynq.timeout_sec must be >= 0")
	}
	if a.Asynq.DedupWindowSec < 0 {
		errs = append(errs, "asynq.dedup_window_sec must be >= 0")
	}
	return errs
}

func (a *App) validateI18n() []string {
	var errs []string
	if strings.TrimSpace(a.I18n.DefaultLang) == "" {
		errs = append(errs, "i18n.default_lang is required")
	}
	if len(a.I18n.BundlePaths) == 0 {
		errs = append(errs, "i18n.bundle_paths is required")
	}
	return errs
}

func (a *App) validateScheduler() []string {
	var errs []string
	if a.Scheduler.LogRetentionDays < 0 {
		errs = append(errs, "scheduler.log_retention_days must be >= 0")
	}
	if a.Scheduler.LockTTLSeconds < 0 {
		errs = append(errs, "scheduler.lock_ttl_seconds must be >= 0")
	}
	return errs
}

func (a *App) validateCORS() []string {
	var errs []string
	if a.CORS.AllowCredentials {
		for _, origin := range a.CORS.AllowOrigins {
			if strings.TrimSpace(origin) == "*" {
				errs = append(errs, "cors.allow_credentials cannot be true when allow_origins contains *")
				break
			}
		}
	}
	return errs
}

func (a *App) validateOutbound() []string {
	var errs []string
	if a.Outbound.TimeoutMS < 0 || a.Outbound.RetryMax < 0 || a.Outbound.RetryBackoffMS < 0 || a.Outbound.CircuitThreshold < 0 || a.Outbound.CircuitOpenSec < 0 {
		errs = append(errs, "outbound values must be >= 0")
	}
	return errs
}

func (a *App) validateStorage() []string {
	if !a.Storage.Enabled {
		return nil
	}
	var errs []string
	driver := strings.ToLower(strings.TrimSpace(a.Storage.Driver))
	if driver == "" {
		driver = "local"
	}
	switch driver {
	case "local":
		if strings.TrimSpace(a.Storage.LocalDir) == "" {
			errs = append(errs, "storage.local_dir is required when storage.driver=local")
		}
	case "s3", "minio":
		if strings.TrimSpace(a.Storage.S3Endpoint) == "" {
			errs = append(errs, "storage.s3_endpoint is required when storage.driver is s3/minio")
		}
		if strings.TrimSpace(a.Storage.S3Bucket) == "" {
			errs = append(errs, "storage.s3_bucket is required when storage.driver is s3/minio")
		}
		if strings.TrimSpace(a.Storage.S3AccessKey) == "" {
			errs = append(errs, "storage.s3_access_key is required when storage.driver is s3/minio")
		}
		if strings.TrimSpace(a.Storage.S3SecretKey) == "" {
			errs = append(errs, "storage.s3_secret_key is required when storage.driver is s3/minio")
		}
	default:
		errs = append(errs, "storage.driver must be local, s3, or minio")
	}
	if strings.TrimSpace(a.Storage.SignSecret) == "" {
		errs = append(errs, "storage.sign_secret is required when storage.enabled=true")
	}
	if a.Storage.MaxUploadMB <= 0 {
		errs = append(errs, "storage.max_upload_mb must be > 0 when storage.enabled=true")
	}
	if a.Storage.URLExpireSec <= 0 {
		errs = append(errs, "storage.url_expire_sec must be > 0 when storage.enabled=true")
	}
	if strings.TrimSpace(a.Storage.AllowedMIME) == "" {
		errs = append(errs, "storage.allowed_mime is required when storage.enabled=true")
	} else {
		for _, part := range strings.Split(a.Storage.AllowedMIME, ",") {
			if strings.TrimSpace(part) == "" {
				errs = append(errs, "storage.allowed_mime must not contain empty entries")
				break
			}
		}
	}
	return errs
}
