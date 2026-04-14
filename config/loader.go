package config

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	global atomic.Value // *App
	mu     sync.RWMutex
	v      *viper.Viper
)

// Load 使用 Viper 从 configs/app.{env}.yaml 加载配置，并注册热重载回调。
func Load(env string) (*App, error) {
	if env == "" {
		env = "dev"
	}
	v = viper.New()
	v.SetConfigType("yaml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
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
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg App
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	global.Store(&cfg)
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		var next App
		if err := v.Unmarshal(&next); err != nil {
			return
		}
		mu.Lock()
		global.Store(&next)
		mu.Unlock()
	})
	return &cfg, nil
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
