package settings

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"gin-scaffold/pkg/cache"
	"gin-scaffold/pkg/db"
)

var ErrNotFound = errors.New("setting not found")

type valueRow struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	ValueType string `json:"value_type"`
}

func GetString(ctx context.Context, key string) (string, error) {
	row, err := getValue(ctx, key)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(row.Value), nil
}

func GetStringDefault(ctx context.Context, key, fallback string) string {
	v, err := GetString(ctx, key)
	if err != nil {
		return fallback
	}
	return v
}

func GetInt64(ctx context.Context, key string) (int64, error) {
	row, err := getValue(ctx, key)
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(row.Value)
	if s == "" {
		return 0, fmt.Errorf("setting %s empty", key)
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("setting %s parse int: %w", key, err)
	}
	return n, nil
}

func GetBool(ctx context.Context, key string) (bool, error) {
	row, err := getValue(ctx, key)
	if err != nil {
		return false, err
	}
	return parseBool(strings.TrimSpace(row.Value))
}

func getValue(ctx context.Context, key string) (*valueRow, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, fmt.Errorf("empty setting key")
	}
	if row, ok := getCached(ctx, key); ok {
		return row, nil
	}
	gdb := db.DB()
	if gdb == nil {
		return nil, fmt.Errorf("db not initialized")
	}
	var row valueRow
	err := gdb.WithContext(ctx).
		Table("system_settings").
		Select("`key`, `value`, `value_type`").
		Where("`key` = ? AND deleted_at IS NULL", key).
		Order("id DESC").
		Limit(1).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	setCached(ctx, &row)
	return &row, nil
}

func parseBool(s string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "y", "on":
		return true, nil
	case "0", "false", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool value %q", s)
	}
}

func cacheKey(c *cache.Client, key string) string {
	if c == nil {
		return ""
	}
	return c.Key("sys_setting", key)
}

func getCached(ctx context.Context, key string) (*valueRow, bool) {
	c := cache.NewFromConfig()
	ck := cacheKey(c, key)
	if ck == "" {
		return nil, false
	}
	var row valueRow
	if err := c.GetJSON(ctx, ck, &row); err != nil {
		return nil, false
	}
	return &row, true
}

func setCached(ctx context.Context, row *valueRow) {
	if row == nil || strings.TrimSpace(row.Key) == "" {
		return
	}
	c := cache.NewFromConfig()
	ck := cacheKey(c, row.Key)
	if ck == "" {
		return
	}
	_ = c.SetJSON(ctx, ck, row, 2*time.Minute)
}
