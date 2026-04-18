package port

import (
	"context"

	"gin-scaffold/internal/model"
)

// MenuService 定义后台菜单查询与维护能力（前端侧栏 + 管理端 CRUD）。
type MenuService interface {
	ListByRole(ctx context.Context, role string) ([]model.Menu, error)
	ListAllByTenant(ctx context.Context) ([]model.Menu, error)
	GetByID(ctx context.Context, id int64) (*model.Menu, error)
	Create(ctx context.Context, name, path, permCode string, sort int, parentID *int64) (*model.Menu, error)
	Update(ctx context.Context, id int64, name, path, permCode *string, sort *int, parentID *int64) (*model.Menu, error)
	Delete(ctx context.Context, id int64) error
}
