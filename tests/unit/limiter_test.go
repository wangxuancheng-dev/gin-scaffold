package unit_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gin-scaffold/pkg/limiter"
)

func TestLimiter_BurstBehavior(t *testing.T) {
	t.Parallel()

	store := limiter.NewStore(100, 1, 100, 1)

	require.True(t, store.AllowIP("127.0.0.1"))
	require.False(t, store.AllowIP("127.0.0.1"))

	require.True(t, store.AllowRoute("GET /api/v1/client/ping"))
	require.False(t, store.AllowRoute("GET /api/v1/client/ping"))
}
