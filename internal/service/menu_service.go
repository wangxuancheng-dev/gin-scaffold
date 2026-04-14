package service

import (
	"context"

	"gin-scaffold/internal/model"
)

// MenuRepo 定义菜单服务依赖的数据访问接口。
type MenuRepo interface {
	ListByRole(ctx context.Context, role string) ([]model.Menu, error)
}

// MenuService 后台菜单业务服务。
type MenuService struct {
	dao MenuRepo
}

func NewMenuService(d MenuRepo) *MenuService {
	return &MenuService{dao: d}
}

func (s *MenuService) ListByRole(ctx context.Context, role string) ([]model.Menu, error) {
	return s.dao.ListByRole(ctx, role)
}
