package config

import (
	"fmt"
	"net"
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
	errs = append(errs, a.validateMetrics()...)
	errs = append(errs, a.validateLimiter()...)
	errs = append(errs, a.validatePlatform()...)
	errs = append(errs, a.validateOutbox()...)
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
	var errs []string
	if a.Storage.ReadyzCheck && !a.Storage.Enabled {
		errs = append(errs, "storage.readyz_check requires storage.enabled=true")
	}
	if !a.Storage.Enabled {
		return errs
	}
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

func (a *App) validatePlatform() []string {
	var errs []string
	p := a.Platform
	if p.Audit.ExportDefaultDays <= 0 {
		errs = append(errs, "platform.audit.export_default_days must be > 0")
	}
	if p.Audit.ExportMaxDays <= 0 {
		errs = append(errs, "platform.audit.export_max_days must be > 0")
	}
	if p.Audit.ExportDefaultDays > 0 && p.Audit.ExportMaxDays > 0 && p.Audit.ExportDefaultDays > p.Audit.ExportMaxDays {
		errs = append(errs, "platform.audit.export_default_days must be <= platform.audit.export_max_days")
	}
	if p.Idempotency.Enabled {
		if p.Idempotency.TTLSeconds < 60 {
			errs = append(errs, "platform.idempotency.ttl_seconds must be >= 60 when idempotency is enabled")
		}
		if p.Idempotency.LockSeconds < 5 || p.Idempotency.LockSeconds > 900 {
			errs = append(errs, "platform.idempotency.lock_seconds must be between 5 and 900 when idempotency is enabled")
		}
		if p.Idempotency.MaxBodyBytes <= 0 {
			errs = append(errs, "platform.idempotency.max_body_bytes must be > 0 when idempotency is enabled")
		}
		if p.Idempotency.MaxCachedResponseBytes <= 0 {
			errs = append(errs, "platform.idempotency.max_cached_response_bytes must be > 0 when idempotency is enabled")
		}
	}
	for _, part := range splitNotifyDrivers(p.Notify.Driver) {
		switch part {
		case "log", "noop", "smtp", "webhook":
		default:
			errs = append(errs, "platform.notify.driver tokens must be one of log,noop,smtp,webhook (comma-separated allowed)")
		}
	}
	errs = append(errs, validateNotifyTargets(p.Notify)...)
	if p.LoginSecurity.Enabled {
		if p.LoginSecurity.MaxFailedPerWindow <= 0 {
			errs = append(errs, "platform.login_security.max_failed_per_window must be > 0 when login_security is enabled")
		}
		if p.LoginSecurity.WindowSec <= 0 {
			errs = append(errs, "platform.login_security.window_sec must be > 0 when login_security is enabled")
		}
		if p.LoginSecurity.LockoutSec <= 0 {
			errs = append(errs, "platform.login_security.lockout_sec must be > 0 when login_security is enabled")
		}
	}
	return errs
}

func splitNotifyDrivers(driver string) []string {
	raw := strings.TrimSpace(driver)
	if raw == "" {
		return []string{"log"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return []string{"log"}
	}
	return out
}

func validateNotifyTargets(n NotifyConfig) []string {
	var errs []string
	tokens := splitNotifyDrivers(n.Driver)
	needSMTP := false
	needWH := false
	for _, t := range tokens {
		if t == "smtp" {
			needSMTP = true
		}
		if t == "webhook" {
			needWH = true
		}
	}
	if needSMTP {
		if strings.TrimSpace(n.SMTP.Host) == "" {
			errs = append(errs, "platform.notify.smtp.host is required when notify driver includes smtp")
		}
		if n.SMTP.Port <= 0 || n.SMTP.Port > 65535 {
			errs = append(errs, "platform.notify.smtp.port must be between 1 and 65535 when notify driver includes smtp")
		}
		if strings.TrimSpace(n.SMTP.From) == "" {
			errs = append(errs, "platform.notify.smtp.from is required when notify driver includes smtp")
		}
		if strings.TrimSpace(n.SMTP.ToDefault) == "" {
			errs = append(errs, "platform.notify.smtp.to_default is required when notify driver includes smtp (per-message meta[\"to\"] 可覆盖)")
		}
	}
	if needWH {
		if strings.TrimSpace(n.Webhook.URL) == "" {
			errs = append(errs, "platform.notify.webhook.url is required when notify driver includes webhook")
		}
	}
	return errs
}

func (a *App) validateMetrics() []string {
	var errs []string
	if !a.Metrics.Enabled {
		return errs
	}
	path := strings.TrimSpace(a.Metrics.Path)
	if path == "" {
		errs = append(errs, "metrics.path is required when metrics.enabled=true")
	} else if !strings.HasPrefix(path, "/") {
		errs = append(errs, "metrics.path must start with /")
	}
	for _, cidr := range a.Metrics.AllowedNetworks {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			errs = append(errs, "metrics.allowed_networks must not contain empty entries")
			continue
		}
		if _, _, err := net.ParseCIDR(cidr); err != nil {
			errs = append(errs, fmt.Sprintf("metrics.allowed_networks: invalid CIDR %q: %v", cidr, err))
		}
	}
	return errs
}

func (a *App) validateLimiter() []string {
	var errs []string
	mode := strings.ToLower(strings.TrimSpace(a.Limiter.Mode))
	if mode == "" {
		mode = "memory"
	}
	if mode != "memory" && mode != "redis" {
		errs = append(errs, "limiter.mode must be memory or redis")
	}
	if mode == "redis" {
		if a.Limiter.WindowSec <= 0 {
			errs = append(errs, "limiter.window_sec must be > 0 when limiter.mode is redis")
		}
		if a.Limiter.IPRPS < 0 || a.Limiter.RouteRPS < 0 {
			errs = append(errs, "limiter ip_rps/route_rps must be >= 0")
		}
		if a.Limiter.IPBurst < 0 || a.Limiter.RouteBurst < 0 {
			errs = append(errs, "limiter ip_burst/route_burst must be >= 0")
		}
	}
	return errs
}

func (a *App) validateOutbox() []string {
	var errs []string
	if !a.Outbox.Enabled {
		return errs
	}
	if a.Outbox.PollIntervalSec <= 0 {
		errs = append(errs, "outbox.poll_interval_sec must be > 0 when outbox is enabled")
	}
	if a.Outbox.BatchSize <= 0 {
		errs = append(errs, "outbox.batch_size must be > 0 when outbox is enabled")
	}
	if a.Outbox.MaxAttempts <= 0 {
		errs = append(errs, "outbox.max_attempts must be > 0 when outbox is enabled")
	}
	if a.Outbox.RetryBackoffSec <= 0 {
		errs = append(errs, "outbox.retry_backoff_sec must be > 0 when outbox is enabled")
	}
	pm := strings.ToLower(strings.TrimSpace(a.Outbox.PublishMode))
	if pm == "" {
		pm = "eventbus"
	}
	if pm != "eventbus" && pm != "http" {
		errs = append(errs, "outbox.publish_mode must be eventbus or http when outbox is enabled")
	}
	if pm == "http" {
		if strings.TrimSpace(a.Outbox.HTTPURL) == "" {
			errs = append(errs, "outbox.http_url is required when outbox.publish_mode is http")
		}
		if a.Outbox.HTTPTimeoutMS < 0 {
			errs = append(errs, "outbox.http_timeout_ms must be >= 0")
		}
	}
	return errs
}
