package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"

	"gin-scaffold/config"
)

func TestInit_nilConfig(t *testing.T) {
	require.Error(t, Init(nil))
}

func TestClient_uninitialized(t *testing.T) {
	prev := client
	t.Cleanup(func() { client = prev })
	client = nil
	require.Nil(t, Client())
}

func TestOps_uninitialized(t *testing.T) {
	prev := client
	t.Cleanup(func() { client = prev })
	client = nil
	ctx := context.Background()
	require.Error(t, Ping(ctx))
	_, err := Get(ctx, "k")
	require.Error(t, err)
	require.Error(t, Set(ctx, "k", "v", time.Second))
	ok, err := SetNX(ctx, "k", "v", time.Second)
	require.False(t, ok)
	require.Error(t, err)
	_, err = Incr(ctx, "c")
	require.Error(t, err)
	require.Error(t, Del(ctx, "k"))
	require.Error(t, Expire(ctx, "k", time.Second))
	require.Error(t, HSet(ctx, "h", "f", "v"))
	_, err = HGet(ctx, "h", "f")
	require.Error(t, err)
}

func TestInitAndOps_miniredis(t *testing.T) {
	prev := client
	mr := miniredis.RunT(t)
	t.Cleanup(func() {
		client = prev
		mr.Close()
	})
	client = nil
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
	require.NoError(t, Init(cfg))
	ctx := context.Background()
	require.NoError(t, Ping(ctx))
	require.NoError(t, Set(ctx, "k1", "v1", time.Minute))
	v, err := Get(ctx, "k1")
	require.NoError(t, err)
	require.Equal(t, "v1", v)
	ok, err := SetNX(ctx, "k2", "v2", time.Minute)
	require.NoError(t, err)
	require.True(t, ok)
	n, err := Incr(ctx, "ctr")
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
	require.NoError(t, HSet(ctx, "hash", "f", "hv"))
	hv, err := HGet(ctx, "hash", "f")
	require.NoError(t, err)
	require.Equal(t, "hv", hv)
	require.NoError(t, Expire(ctx, "k1", time.Hour))
	require.NoError(t, Del(ctx, "k2"))
	_, err = Get(ctx, "k2")
	require.Error(t, err)
}
