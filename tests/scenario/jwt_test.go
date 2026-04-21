package unit_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gin-scaffold/internal/config"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
)

func TestJWT_IssueAndParseAccess(t *testing.T) {
	t.Parallel()

	m := jwtpkg.NewManager(&config.JWTConfig{
		Secret:           "unit-test-secret",
		AccessExpireMin:  30,
		RefreshExpireMin: 60,
		Issuer:           "unit-test",
	})

	token, err := m.IssueAccess(42, "user", "")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := m.ParseAccess(token)
	require.NoError(t, err)
	require.Equal(t, int64(42), claims.UserID)
	require.Equal(t, "user", claims.Role)
}
