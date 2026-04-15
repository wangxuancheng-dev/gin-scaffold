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
	List(ctx context.Context, q model.UserQuery, page, pageSize int) ([]model.User, int64, error)
	AdminCreate(ctx context.Context, username, password, nickname, role string) (*model.User, error)
	AdminUpdate(ctx context.Context, id int64, nickname, password, role *string) (*model.User, error)
	AdminDelete(ctx context.Context, id int64) error
	StreamExport(
		ctx context.Context,
		q model.UserQuery,
		page, pageSize, limit, batchSize int,
		pageOnly, withRole bool,
		consume func(model.UserExportRow) error,
	) error
}
