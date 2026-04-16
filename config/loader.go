package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	global atomic.Value // *App
	meta   atomic.Value // SourceMeta
	mu     sync.RWMutex
	v      *viper.Viper
)

type SourceMeta struct {
	Env         string
	Profile     string
	YAMLFiles   []string
	DotEnvFiles []string
}

// Load 使用 Viper 加载多层配置并注册热重载回调。
// 加载顺序（后者覆盖前者）:
// 1) app.yaml
// 2) app.{env}.yaml (test/prod)
// 3) app.{env}.{profile}.yaml (可选，多套生产系统)
func Load(env, profile string) (*App, error) {
	if env == "" {
		env = "dev"
	}
	dotEnvFiles := loadDotEnv(env, profile)
	v = viper.New()
	v.SetConfigType("yaml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	bindEnvKeys(v)
	// 开发使用 app.yaml；测试 app.test.yaml；生产 app.prod.yaml
	base := "app"
	switch env {
	case "test":
		base = "app.test"
	case "prod":
		base = "app.prod"
	default:
		base = "app"
	}
	v.SetConfigName(base)
	v.AddConfigPath("./configs")
	v.AddConfigPath("configs")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read base config: %w", err)
	}
	yamlFiles := []string{v.ConfigFileUsed()}
	if env != "dev" {
		if p, ok := resolveConfigFile(base); ok {
			v.SetConfigFile(p)
			if err := v.MergeInConfig(); err != nil {
				return nil, fmt.Errorf("merge env config: %w", err)
			}
			yamlFiles = append(yamlFiles, p)
		}
	}
	if profile != "" {
		profileName := fmt.Sprintf("%s.%s", base, profile)
		if p, ok := resolveConfigFile(profileName); ok {
			v.SetConfigFile(p)
			if err := v.MergeInConfig(); err != nil {
				return nil, fmt.Errorf("merge profile config: %w", err)
			}
			yamlFiles = append(yamlFiles, p)
		}
	}
	var cfg App
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	global.Store(&cfg)
	meta.Store(SourceMeta{
		Env:         env,
		Profile:     profile,
		YAMLFiles:   yamlFiles,
		DotEnvFiles: dotEnvFiles,
	})
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		var next App
		if err := v.Unmarshal(&next); err != nil {
			return
		}
		if err := next.Validate(); err != nil {
			return
		}
		mu.Lock()
		global.Store(&next)
		mu.Unlock()
	})
	return &cfg, nil
}

func loadDotEnv(env, profile string) []string {
	// 生产安全策略：
	// 1) 非 dev 环境默认不加载 .env 文件，避免与容器/平台注入环境变量冲突。
	// 2) 使用 Load（非 Overload），仅填充未设置变量，不覆盖进程已有变量。
	if env != "dev" {
		return nil
	}

	files := []string{
		".env",
		fmt.Sprintf(".env.%s", env),
	}
	if profile != "" {
		files = append(files, fmt.Sprintf(".env.%s.%s", env, profile))
	}
	files = append(files, ".env.local", fmt.Sprintf(".env.%s.local", env))
	loaded := make([]string, 0, len(files))
	for _, f := range files {
		if _, err := os.Stat(f); err == nil {
			_ = godotenv.Load(f)
			loaded = append(loaded, f)
		}
	}
	return loaded
}

func bindEnvKeys(v *viper.Viper) {
	keys := []string{
		"env", "name", "debug",
		"http.host", "http.port", "http.read_timeout_sec", "http.read_header_timeout_sec", "http.write_timeout_sec", "http.idle_timeout_sec", "http.shutdown_timeout_sec",
		"log.level", "log.dir", "log.app_file", "log.access_file", "log.error_file",
		"log.rotation_mode", "log.app_rotation_mode", "log.access_rotation_mode", "log.error_rotation_mode",
		"log.max_size_mb", "log.max_backups", "log.max_age_days", "log.compress", "log.console",
		"db.driver", "db.dsn", "db.max_open_conns", "db.max_idle_conns", "db.conn_max_lifetime_sec", "db.conn_max_idle_time_sec", "db.slow_threshold_ms", "db.log_level",
		"redis.addr", "redis.password", "redis.db", "redis.pool_size", "redis.min_idle_conns",
		"asynq.redis_addr", "asynq.redis_password", "asynq.redis_db", "asynq.concurrency", "asynq.strict_priority", "asynq.queue", "asynq.max_retry", "asynq.timeout_sec", "asynq.dedup_window_sec",
		"jwt.secret", "jwt.access_expire_min", "jwt.refresh_expire_min", "jwt.issuer",
		"metrics.enabled", "metrics.path",
		"trace.enabled", "trace.endpoint", "trace.service_name", "trace.insecure",
		"i18n.default_lang",
		"limiter.ip_rps", "limiter.ip_burst", "limiter.route_rps", "limiter.route_burst",
		"snowflake.node",
		"rbac.super_admin_user_id",
		"scheduler.enabled", "scheduler.with_seconds", "scheduler.log_retention_days",
		"scheduler.lock_enabled", "scheduler.lock_ttl_seconds", "scheduler.lock_prefix",
	}
	for _, k := range keys {
		_ = v.BindEnv(k)
	}
	// 与 OS 习惯一致：用 TIME_ZONE（如 UTC、Asia/Shanghai）
	_ = v.BindEnv("db.time_zone", "TIME_ZONE")
}

func resolveConfigFile(name string) (string, bool) {
	candidates := []string{
		filepath.Join("configs", name+".yaml"),
		filepath.Join(".", "configs", name+".yaml"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, true
		}
	}
	return "", false
}

// Get 返回当前内存中的配置指针（热重载后会更新）。
func Get() *App {
	v := global.Load()
	if v == nil {
		return nil
	}
	return v.(*App)
}

// Viper 返回全局 Viper 实例（用于测试或高级场景）。
func Viper() *viper.Viper {
	return v
}

// Source 返回最近一次成功加载的配置来源信息（用于启动日志排障）。
func Source() SourceMeta {
	m := meta.Load()
	if m == nil {
		return SourceMeta{}
	}
	return m.(SourceMeta)
}
