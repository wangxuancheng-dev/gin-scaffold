package loginthrottle

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"gin-scaffold/config"
)

func TestRedisKeyPrefix_loginSecurityPrefix(t *testing.T) {
	cfg := config.LoginSecurityConfig{RedisKeyPrefix: "myapp"}
	require.Equal(t, "myapp:", redisKeyPrefix(cfg, "ignored"))
}

func TestRedisKeyPrefix_loginSecurityPrefixAlreadyHasColon(t *testing.T) {
	cfg := config.LoginSecurityConfig{RedisKeyPrefix: "myapp:"}
	require.Equal(t, "myapp:", redisKeyPrefix(cfg, ""))
}

func TestRedisKeyPrefix_fallbackCachePrefix(t *testing.T) {
	cfg := config.LoginSecurityConfig{}
	require.Equal(t, "cache:", redisKeyPrefix(cfg, "cache"))
}

func TestRedisKeyPrefix_defaultApp(t *testing.T) {
	cfg := config.LoginSecurityConfig{}
	require.Equal(t, "app:", redisKeyPrefix(cfg, ""))
}

func TestIdentityHashStable(t *testing.T) {
	a := identityHash("T1", "User", "1.2.3.4")
	b := identityHash(" t1 ", " USER ", "1.2.3.4")
	require.Equal(t, a, b)
	require.Len(t, a, 24)
}

func TestFailKeyLockKeyFormat(t *testing.T) {
	p := "app:"
	fk := failKey(p, "t", "u", "127.0.0.1")
	lk := lockKey(p, "t", "u", "127.0.0.1")
	id := identityHash("t", "u", "127.0.0.1")
	require.Equal(t, p+"login:v1:fail:t:"+id, fk)
	require.Equal(t, p+"login:v1:lock:t:"+id, lk)
}

func TestIsLocked_disabled(t *testing.T) {
	ok, err := IsLocked(context.Background(), config.LoginSecurityConfig{Enabled: false}, "", "t", "ip", "u")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestRegisterFailure_disabled(t *testing.T) {
	locked, err := RegisterFailure(context.Background(), config.LoginSecurityConfig{Enabled: false}, "", "t", "ip", "u")
	require.NoError(t, err)
	require.False(t, locked)
}

func TestClear_disabled(t *testing.T) {
	Clear(context.Background(), config.LoginSecurityConfig{Enabled: false}, "", "t", "ip", "u")
}
