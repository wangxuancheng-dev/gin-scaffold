package unit_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gin-scaffold/api/response"
	"gin-scaffold/internal/model"
)

func TestFromUser_ShouldMapFields(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 14, 12, 30, 0, 0, time.Local)
	u := &model.User{
		ID:        1001,
		Username:  "tester",
		Nickname:  "Tester",
		Password:  "hidden",
		CreatedAt: now,
		UpdatedAt: now,
	}

	vo := response.FromUser(u)

	require.Equal(t, int64(1001), vo.ID)
	require.Equal(t, "tester", vo.Username)
	require.Equal(t, "Tester", vo.Nickname)
	require.NotEmpty(t, vo.CreatedAt)
	require.NotEmpty(t, vo.UpdatedAt)
}
