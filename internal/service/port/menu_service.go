package port

import (
	"context"

	"gin-scaffold/internal/model"
)

// MenuService 定义后台菜单查询能力。
type MenuService interface {
	ListByRole(ctx context.Context, role string) ([]model.Menu, error)
}
