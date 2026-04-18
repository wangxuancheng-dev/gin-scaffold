package job

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"

	"gin-scaffold/config"
	"gin-scaffold/pkg/redis"
)

func TestUserExportStatus_redisRoundtrip(t *testing.T) {
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
	st := &UserExportStatus{TaskID: "task-1", State: "queued", Filter: "k=v"}
	require.NoError(t, SetUserExportStatus(ctx, st))

	got, err := GetUserExportStatus(ctx, "task-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, "task-1", got.TaskID)
	require.Equal(t, "queued", got.State)
	require.Equal(t, "k=v", got.Filter)
	require.NotEmpty(t, got.CreatedAt)
	require.NotEmpty(t, got.UpdatedAt)

	missing, err := GetUserExportStatus(ctx, "no-such-task")
	require.NoError(t, err)
	require.Nil(t, missing)
}
