// Package config 定义应用全量配置结构体（嵌套子配置），供 Viper 反序列化与业务读取。
package config

// App 根配置，对应 configs/*.yaml 顶层字段。
type App struct {
	Env       string          `mapstructure:"env"`
	Name      string          `mapstructure:"name"`
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
}

// HTTPConfig HTTP 服务监听与超时配置。
type HTTPConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout_sec"`
	WriteTimeout int    `mapstructure:"write_timeout_sec"`
	IdleTimeout  int    `mapstructure:"idle_timeout_sec"`
}

// LogConfig Zap + Lumberjack 日志配置。
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Dir        string `mapstructure:"dir"`
	AppFile    string `mapstructure:"app_file"`
	AccessFile string `mapstructure:"access_file"`
	ErrorFile  string `mapstructure:"error_file"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
	Compress   bool   `mapstructure:"compress"`
	Console    bool   `mapstructure:"console"`
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
	RedisAddr      string `mapstructure:"redis_addr"`
	RedisPassword  string `mapstructure:"redis_password"`
	RedisDB        int    `mapstructure:"redis_db"`
	Concurrency    int    `mapstructure:"concurrency"`
	StrictPriority bool   `mapstructure:"strict_priority"`
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
