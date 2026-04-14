package port

import (
	"context"

	"gin-scaffold/internal/model"
)

// UserService 定义用户业务能力，供客户端与后台 handler 复用。
type UserService interface {
	Register(ctx context.Context, username, password, nickname string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	Login(ctx context.Context, username, password string) (string, error)
	LoginWithRefresh(ctx context.Context, username, password string) (string, string, error)
	RefreshAccess(ctx context.Context, refreshToken string) (string, string, error)
	List(ctx context.Context, page, pageSize int) ([]model.User, int64, error)
}
