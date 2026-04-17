// Package config 定义应用全量配置结构体（嵌套子配置），供 Viper 反序列化与业务读取。
package config

// App 根配置，对应 configs/*.yaml 顶层字段。
type App struct {
	Env       string          `mapstructure:"env"`
	Name      string          `mapstructure:"name"`
	Debug     bool            `mapstructure:"debug"`
	HTTP      HTTPConfig      `mapstructure:"http"`
	Log       LogConfig       `mapstructure:"log"`
	DB        DBConfig        `mapstructure:"db"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Asynq     AsynqConfig     `mapstructure:"asynq"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
	Trace     TraceConfig     `mapstructure:"trace"`
	I18n      I18nConfig      `mapstructure:"i18n"`
	Limiter   LimiterConfig   `mapstructure:"limiter"`
	Snowflake SnowflakeConfig `mapstructure:"snowflake"`
	CORS      CORSConfig      `mapstructure:"cors"`
	RBAC      RBACConfig      `mapstructure:"rbac"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Outbound  OutboundConfig  `mapstructure:"outbound"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Platform  PlatformConfig  `mapstructure:"platform"`
}

// PlatformConfig 横切能力：审计、幂等、缓存前缀、通知驱动等。
type PlatformConfig struct {
	Audit       AuditConfig       `mapstructure:"audit"`
	Idempotency IdempotencyConfig `mapstructure:"idempotency"`
	Cache       CacheConfig       `mapstructure:"cache"`
	Notify      NotifyConfig      `mapstructure:"notify"`
}

// AuditConfig 写操作 HTTP 审计落库（异步插入，失败仅打日志）。
type AuditConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	ExportDefaultDays int  `mapstructure:"export_default_days"` // 导出默认时间窗（天），默认 7
	ExportMaxDays     int  `mapstructure:"export_max_days"`     // 导出最大允许时间窗（天），默认 31
}

// IdempotencyConfig POST 幂等（Redis，需 X-Idempotency-Key）。
type IdempotencyConfig struct {
	Enabled                 bool  `mapstructure:"enabled"`
	TTLSeconds              int   `mapstructure:"ttl_seconds"`
	LockSeconds             int   `mapstructure:"lock_seconds"`
	MaxBodyBytes            int64 `mapstructure:"max_body_bytes"`
	MaxCachedResponseBytes  int64 `mapstructure:"max_cached_response_bytes"`
}

// CacheConfig 业务缓存键前缀（pkg/cache）。
type CacheConfig struct {
	KeyPrefix string `mapstructure:"key_prefix"`
}

// NotifyConfig 通知通道驱动。
type NotifyConfig struct {
	Driver string `mapstructure:"driver"` // log | noop
}

// HTTPConfig HTTP 服务监听与超时配置。
type HTTPConfig struct {
	Host              string `mapstructure:"host"`
	Port              int    `mapstructure:"port"`
	ReadTimeout       int    `mapstructure:"read_timeout_sec"`
	ReadHeaderTimeout int    `mapstructure:"read_header_timeout_sec"`
	WriteTimeout      int    `mapstructure:"write_timeout_sec"`
	IdleTimeout       int    `mapstructure:"idle_timeout_sec"`
	ShutdownTimeout   int    `mapstructure:"shutdown_timeout_sec"`
	MaxBodyBytes      int64  `mapstructure:"max_body_bytes"`
}

// LogConfig Zap + Lumberjack 日志配置。
type LogConfig struct {
	Level              string                      `mapstructure:"level"`
	Dir                string                      `mapstructure:"dir"`
	AppFile            string                      `mapstructure:"app_file"`
	AccessFile         string                      `mapstructure:"access_file"`
	ErrorFile          string                      `mapstructure:"error_file"`
	RotationMode       string                      `mapstructure:"rotation_mode"`        // 全局默认: size | daily | none，默认 size
	AppRotationMode    string                      `mapstructure:"app_rotation_mode"`    // app 日志单独模式，空则用 rotation_mode
	AccessRotationMode string                      `mapstructure:"access_rotation_mode"` // access 日志单独模式，空则用 rotation_mode
	ErrorRotationMode  string                      `mapstructure:"error_rotation_mode"`  // error 日志单独模式，空则用 rotation_mode
	MaxSizeMB          int                         `mapstructure:"max_size_mb"`
	MaxBackups         int                         `mapstructure:"max_backups"`
	MaxAgeDays         int                         `mapstructure:"max_age_days"`
	Compress           bool                        `mapstructure:"compress"`
	Console            bool                        `mapstructure:"console"`
	Channels           map[string]LogChannelConfig `mapstructure:"channels"` // 自定义日志通道
}

// LogChannelConfig 单个日志通道的配置。
type LogChannelConfig struct {
	File         string `mapstructure:"file"`          // 目标文件名（相对 log.dir，可为空；调用 Channel 时动态传入）
	Level        string `mapstructure:"level"`         // debug|info|warn|error，默认继承 log.level
	RotationMode string `mapstructure:"rotation_mode"` // size|daily|none，默认继承 log.rotation_mode
	MaxSizeMB    int    `mapstructure:"max_size_mb"`   // <=0 则继承全局
	MaxBackups   int    `mapstructure:"max_backups"`   // <0 则继承全局
	MaxAgeDays   int    `mapstructure:"max_age_days"`  // <=0 则继承全局
	Compress     *bool  `mapstructure:"compress"`      // nil 则继承全局
}

// DBConfig 数据库连接、连接池与慢查询配置。
type DBConfig struct {
	Driver             string   `mapstructure:"driver"` // mysql | postgres
	DSN                string   `mapstructure:"dsn"`
	TimeZone           string   `mapstructure:"time_zone"` // MySQL: SET time_zone；PostgreSQL: SET TIME ZONE；空则 UTC；可用环境变量 TIME_ZONE 覆盖
	Replicas           []string `mapstructure:"replicas"`
	MaxOpenConns       int      `mapstructure:"max_open_conns"`
	MaxIdleConns       int      `mapstructure:"max_idle_conns"`
	ConnMaxLifetimeSec int      `mapstructure:"conn_max_lifetime_sec"`
	ConnMaxIdleTimeSec int      `mapstructure:"conn_max_idle_time_sec"`
	SlowThresholdMS    int      `mapstructure:"slow_threshold_ms"`
	LogLevel           string   `mapstructure:"log_level"` // silent|error|warn|info
}

// RedisConfig Redis 客户端与连接池配置。
type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	DialTimeout  int    `mapstructure:"dial_timeout_sec"`
	ReadTimeout  int    `mapstructure:"read_timeout_sec"`
	WriteTimeout int    `mapstructure:"write_timeout_sec"`
}

// AsynqConfig 异步任务队列配置。
type AsynqConfig struct {
	RedisAddr      string         `mapstructure:"redis_addr"`
	RedisPassword  string         `mapstructure:"redis_password"`
	RedisDB        int            `mapstructure:"redis_db"`
	Concurrency    int            `mapstructure:"concurrency"`
	StrictPriority bool           `mapstructure:"strict_priority"`
	Queue          string         `mapstructure:"queue"`
	Queues         map[string]int `mapstructure:"queues"`
	MaxRetry       int            `mapstructure:"max_retry"`
	TimeoutSec     int            `mapstructure:"timeout_sec"`
	DedupWindowSec int            `mapstructure:"dedup_window_sec"`
}

// JWTConfig JWT 签发与校验配置。
type JWTConfig struct {
	Secret           string `mapstructure:"secret"`
	AccessExpireMin  int    `mapstructure:"access_expire_min"`
	RefreshExpireMin int    `mapstructure:"refresh_expire_min"`
	Issuer           string `mapstructure:"issuer"`
}

// MetricsConfig Prometheus 指标暴露配置。
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// TraceConfig OpenTelemetry 链路追踪配置（OTLP HTTP，可对接 Jaeger）。
type TraceConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Endpoint    string `mapstructure:"endpoint"` // e.g. http://jaeger:4318/v1/traces
	ServiceName string `mapstructure:"service_name"`
	Insecure    bool   `mapstructure:"insecure"`
}

// I18nConfig 多语言资源路径与默认语言。
type I18nConfig struct {
	DefaultLang string   `mapstructure:"default_lang"`
	BundlePaths []string `mapstructure:"bundle_paths"`
}

// LimiterConfig 全局限流默认参数。
type LimiterConfig struct {
	IPRPS      float64 `mapstructure:"ip_rps"` // 每 IP 每秒令牌补充速率
	IPBurst    int     `mapstructure:"ip_burst"`
	RouteRPS   float64 `mapstructure:"route_rps"` // 每路由每秒
	RouteBurst int     `mapstructure:"route_burst"`
}

// SnowflakeConfig 雪花算法节点号（0-1023）。
type SnowflakeConfig struct {
	Node int64 `mapstructure:"node"`
}

// CORSConfig 跨域中间件配置。
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

// RBACConfig 权限相关配置。
type RBACConfig struct {
	SuperAdminUserID int64 `mapstructure:"super_admin_user_id"`
}

// SchedulerConfig cron 定时任务配置（robfig/cron）。
type SchedulerConfig struct {
	Enabled          bool   `mapstructure:"enabled"`            // 是否启用数据库任务调度器
	WithSeconds      bool   `mapstructure:"with_seconds"`       // 是否启用秒级字段（6 段）
	LogRetentionDays int    `mapstructure:"log_retention_days"` // 任务执行日志保留天数，<=0 表示不清理
	LockEnabled      bool   `mapstructure:"lock_enabled"`       // 多实例防重：是否启用 Redis 分布式锁
	LockTTLSeconds   int    `mapstructure:"lock_ttl_seconds"`   // 分布式锁 TTL（秒）
	LockPrefix       string `mapstructure:"lock_prefix"`        // 分布式锁 key 前缀
}

// OutboundConfig 下游 HTTP 客户端治理参数。
type OutboundConfig struct {
	TimeoutMS        int `mapstructure:"timeout_ms"`
	RetryMax         int `mapstructure:"retry_max"`
	RetryBackoffMS   int `mapstructure:"retry_backoff_ms"`
	CircuitThreshold int `mapstructure:"circuit_threshold"`
	CircuitOpenSec   int `mapstructure:"circuit_open_sec"`
}

// StorageConfig 文件存储配置（V1: local）。
type StorageConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	Driver       string `mapstructure:"driver"` // local | s3 | minio（minio 与 s3 等价）
	LocalDir     string `mapstructure:"local_dir"`
	SignSecret   string `mapstructure:"sign_secret"`
	MaxUploadMB  int64  `mapstructure:"max_upload_mb"`
	AllowedExt   string `mapstructure:"allowed_ext"`  // 逗号分隔，如 .jpg,.png,.pdf
	AllowedMIME  string `mapstructure:"allowed_mime"` // 逗号分隔，如 image/jpeg,application/pdf
	URLExpireSec int    `mapstructure:"url_expire_sec"`
	S3Endpoint   string `mapstructure:"s3_endpoint"`   // 如 https://minio.example.com
	S3Region     string `mapstructure:"s3_region"`   // 可为空，MinIO 常用 us-east-1
	S3Bucket     string `mapstructure:"s3_bucket"`
	S3AccessKey  string `mapstructure:"s3_access_key"`
	S3SecretKey  string `mapstructure:"s3_secret_key"`
	S3PathStyle  bool   `mapstructure:"s3_path_style"` // MinIO 通常 true
	S3Insecure   bool   `mapstructure:"s3_insecure"`   // 跳过 TLS 校验（仅内网/开发）
	ReadyzCheck  bool   `mapstructure:"readyz_check"`  // /readyz 是否检查存储连通性（HeadBucket 或本地目录）
}
