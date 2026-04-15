package db

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var offsetTZ = regexp.MustCompile(`^([+-])(\d{1,2}):(\d{2})$`)

// NormalizeTimeZone 返回用于 MySQL `SET time_zone` / PG `SET TIME ZONE` 的字符串；空则默认 UTC（与 `TIME_ZONE` 环境变量约定一致）。
func NormalizeTimeZone(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "UTC"
	}
	return s
}

// LocationForTimeZone 将配置中的 `db.time_zone` 解析为 *time.Location，用于 GORM NowFunc。
func LocationForTimeZone(s string) (*time.Location, error) {
	s = NormalizeTimeZone(s)
	if strings.EqualFold(s, "UTC") || s == "Z" {
		return time.UTC, nil
	}
	if m := offsetTZ.FindStringSubmatch(s); m != nil {
		sign := 1
		if m[1] == "-" {
			sign = -1
		}
		h, err := strconv.Atoi(m[2])
		if err != nil {
			return nil, fmt.Errorf("db: invalid time_zone offset %q", s)
		}
		min, err := strconv.Atoi(m[3])
		if err != nil {
			return nil, fmt.Errorf("db: invalid time_zone offset %q", s)
		}
		if min < 0 || min > 59 || h < 0 || h > 23 {
			return nil, fmt.Errorf("db: time_zone offset out of range %q", s)
		}
		sec := sign * (h*3600 + min*60)
		return time.FixedZone("", sec), nil
	}
	return time.LoadLocation(s)
}

// ApplyTimeZone 在连接上设置数据库会话时区（影响 MySQL 的 NOW() 等；建议与服务器 default_time_zone 一致）。
func ApplyTimeZone(db *gorm.DB, driver, tz string) error {
	tz = NormalizeTimeZone(tz)
	switch driver {
	case "mysql", "":
		// 未导入 mysql 时区表时，命名时区如 UTC 会报 Error 1298；偏移 +00:00 始终可用。
		return db.Exec("SET time_zone = ?", mysqlSetTimeZoneValue(tz)).Error
	case "postgres", "pg":
		// PostgreSQL 自带时区解析，'UTC' 与 IANA 名（如 Asia/Shanghai）一般无需额外导入表。
		if dbTimeZoneIsUTCEquivalent(tz) {
			return db.Exec("SET TIME ZONE 'UTC'").Error
		}
		return db.Exec("SET TIME ZONE ?", tz).Error
	default:
		return nil
	}
}

func dbTimeZoneIsUTCEquivalent(s string) bool {
	switch strings.TrimSpace(strings.ToUpper(s)) {
	case "UTC", "Z", "+00:00", "-00:00":
		return true
	default:
		return false
	}
}

// mysqlSetTimeZoneValue 供 `SET time_zone = ?` 使用。MySQL 在未执行 mysql_tzinfo_to_sql 时，命名时区常报 1298；
// 因此 UTC 等价物统一为 +00:00；IANA 名称则转为当前日期下的 ±HH:MM 偏移（不依赖 mysql 时区表）。
func mysqlSetTimeZoneValue(tz string) string {
	st := strings.TrimSpace(tz)
	if dbTimeZoneIsUTCEquivalent(st) {
		return "+00:00"
	}
	if offsetTZ.MatchString(st) {
		return st
	}
	loc, err := time.LoadLocation(st)
	if err != nil {
		return st
	}
	_, off := time.Now().In(loc).Zone()
	return formatMySQLTimeZoneOffset(off)
}

func formatMySQLTimeZoneOffset(offSecondsEastOfUTC int) string {
	if offSecondsEastOfUTC == 0 {
		return "+00:00"
	}
	sign := '+'
	if offSecondsEastOfUTC < 0 {
		sign = '-'
		offSecondsEastOfUTC = -offSecondsEastOfUTC
	}
	totalMin := offSecondsEastOfUTC / 60
	h := totalMin / 60
	m := totalMin % 60
	return fmt.Sprintf("%c%02d:%02d", sign, h, m)
}

// SyncProcessLocalToTimeZone 将进程 `time.Local` 与 `db.time_zone` / TIME_ZONE 对齐（Gin 与普通 `time.Now()` 使用该本地时区）。
func SyncProcessLocalToTimeZone(raw string) error {
	loc, err := LocationForTimeZone(raw)
	if err != nil {
		return err
	}
	time.Local = loc
	return nil
}
