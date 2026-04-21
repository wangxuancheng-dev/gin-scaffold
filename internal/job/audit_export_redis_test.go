package job

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"

	"gin-scaffold/internal/config"
	"gin-scaffold/pkg/redis"
)

func TestAuditExportStatus_redisRoundtrip(t *testing.T) {
	mr := miniredis.RunT(t)
	t.Cleanup(func() {
		_ = redis.Close()
		mr.Close()
	})

	cfg := &config.RedisConfig{
		Addr:         mr.Addr(),
		Password:     "",
		DB:           0,
		PoolSize:     4,
		MinIdleConns: 0,
		DialTimeout:  2,
		ReadTimeout:  2,
		WriteTimeout: 2,
	}
	require.NoError(t, redis.Init(cfg))

	ctx := context.Background()
	st := &AuditExportStatus{TaskID: "audit-1", State: "queued", Filter: "a=b"}
	require.NoError(t, SetAuditExportStatus(ctx, st))

	got, err := GetAuditExportStatus(ctx, "audit-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "audit-1", got.TaskID)
	require.Equal(t, "queued", got.State)
}
