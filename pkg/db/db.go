// Package db 封装 GORM 初始化、双驱动、连接池与慢查询日志。
package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"

	"gin-scaffold/config"
)

var global *gorm.DB

// Init 根据配置创建全局 *gorm.DB，支持 MySQL / PostgreSQL 与可选只读副本。
func Init(cfg *config.DBConfig) (*gorm.DB, error) {
	if cfg == nil || cfg.DSN == "" {
		return nil, fmt.Errorf("db: empty dsn")
	}
	gormLog := logger.New(
		log.New(os.Stdout, "\r", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Duration(cfg.SlowThresholdMS) * time.Millisecond,
			LogLevel:                  parseGormLogLevel(cfg.LogLevel),
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
	tz := NormalizeTimeZone(cfg.TimeZone)
	loc, err := LocationForTimeZone(tz)
	if err != nil {
		return nil, fmt.Errorf("db: time_zone: %w", err)
	}

	var dialector gorm.Dialector
	switch cfg.Driver {
	case "postgres", "pg":
		dialector = postgres.Open(cfg.DSN)
	case "mysql", "":
		mysqlDSN, err := NormalizeMySQLDSN(cfg.DSN, loc)
		if err != nil {
			return nil, err
		}
		dialector = mysql.Open(mysqlDSN)
	default:
		return nil, fmt.Errorf("db: unsupported driver %s", cfg.Driver)
	}
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:  gormLog,
		NowFunc: func() time.Time { return time.Now().In(loc) },
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}
	if err = ApplyTimeZone(db, cfg.Driver, tz); err != nil {
		return nil, fmt.Errorf("db: set time zone: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeSec) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTimeSec) * time.Second)

	if len(cfg.Replicas) > 0 {
		replicas := make([]gorm.Dialector, 0, len(cfg.Replicas))
		for _, rep := range cfg.Replicas {
			switch cfg.Driver {
			case "postgres", "pg":
				replicas = append(replicas, postgres.Open(rep))
			default:
				repDSN, err := NormalizeMySQLDSN(rep, loc)
				if err != nil {
					return nil, fmt.Errorf("mysql replica dsn: %w", err)
				}
				replicas = append(replicas, mysql.Open(repDSN))
			}
		}
		if err := db.Use(dbresolver.Register(dbresolver.Config{
			Replicas: replicas,
			Policy:   dbresolver.RandomPolicy{},
		})); err != nil {
			return nil, fmt.Errorf("dbresolver: %w", err)
		}
	}
	global = db
	return db, nil
}

// DB 返回全局数据库实例。
func DB() *gorm.DB {
	return global
}

func parseGormLogLevel(lv string) logger.LogLevel {
	switch lv {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}
