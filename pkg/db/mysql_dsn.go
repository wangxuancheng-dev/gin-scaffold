package db

import (
	"fmt"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
)

// NormalizeMySQLDSN 解析 DSN，统一设置 parseTime 与驱动 Loc（与 db.time_zone / TIME_ZONE 解析结果一致），
// 不再要求在连接串里手写 loc=（避免与 TIME_ZONE 双源配置）。
func NormalizeMySQLDSN(dsn string, loc *time.Location) (string, error) {
	if loc == nil {
		loc = time.UTC
	}
	cfg, err := mysqldriver.ParseDSN(dsn)
	if err != nil {
		return "", fmt.Errorf("mysql dsn: %w", err)
	}
	cfg.ParseTime = true
	cfg.Loc = loc
	return cfg.FormatDSN(), nil
}
