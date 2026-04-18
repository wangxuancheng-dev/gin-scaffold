package service

import (
	"context"
	"errors"
	"strings"

	mysql "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
)

// MenuRepo 定义菜单服务依赖的数据访问接口。
type MenuRepo interface {
	ListByRole(ctx context.Context, role string) ([]model.Menu, error)
	ListAllByTenant(ctx context.Context) ([]model.Menu, error)
	GetByID(ctx context.Context, id int64) (*model.Menu, error)
	Create(ctx context.Context, m *model.Menu) error
	Save(ctx context.Context, m *model.Menu) error
	SoftDelete(ctx context.Context, id int64) error
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

func (s *MenuService) ListAllByTenant(ctx context.Context) ([]model.Menu, error) {
	return s.dao.ListAllByTenant(ctx)
}

func (s *MenuService) GetByID(ctx context.Context, id int64) (*model.Menu, error) {
	return s.menuOrErr(ctx, id)
}

func (s *MenuService) menuOrErr(ctx context.Context, id int64) (*model.Menu, error) {
	m, err := s.dao.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.NotFound, errcode.KeyNotFound)
		}
		return nil, err
	}
	return m, nil
}

func (s *MenuService) Create(ctx context.Context, name, path, permCode string, sort int, parentID *int64) (*model.Menu, error) {
	m := &model.Menu{
		Name:     strings.TrimSpace(name),
		Path:     strings.TrimSpace(path),
		PermCode: strings.TrimSpace(permCode),
		Sort:     sort,
	}
	if parentID != nil && *parentID > 0 {
		if err := s.validateParentExists(ctx, *parentID); err != nil {
			return nil, err
		}
		pid := *parentID
		m.ParentID = &pid
	}
	if err := s.dao.Create(ctx, m); err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return nil, errcode.New(errcode.Conflict, errcode.KeyInvalidParam)
		}
		return nil, err
	}
	return m, nil
}

func (s *MenuService) Update(ctx context.Context, id int64, name, path, permCode *string, sort *int, parentID *int64) (*model.Menu, error) {
	m, err := s.menuOrErr(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		m.Name = strings.TrimSpace(*name)
	}
	if path != nil {
		m.Path = strings.TrimSpace(*path)
	}
	if permCode != nil {
		m.PermCode = strings.TrimSpace(*permCode)
	}
	if sort != nil {
		m.Sort = *sort
	}
	if parentID != nil {
		if *parentID <= 0 {
			m.ParentID = nil
		} else {
			if *parentID == id {
				return nil, errcode.New(errcode.BadRequest, errcode.KeyInvalidParam)
			}
			if err := s.validateParentExists(ctx, *parentID); err != nil {
				return nil, err
			}
			all, err := s.dao.ListAllByTenant(ctx)
			if err != nil {
				return nil, err
			}
			if menuDescendantContains(all, id, *parentID) {
				return nil, errcode.New(errcode.BadRequest, errcode.KeyInvalidParam)
			}
			pid := *parentID
			m.ParentID = &pid
		}
	}
	if err := s.dao.Save(ctx, m); err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return nil, errcode.New(errcode.Conflict, errcode.KeyInvalidParam)
		}
		return nil, err
	}
	return s.menuOrErr(ctx, id)
}

func (s *MenuService) Delete(ctx context.Context, id int64) error {
	if err := s.dao.SoftDelete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.New(errcode.NotFound, errcode.KeyNotFound)
		}
		return err
	}
	return nil
}

func (s *MenuService) validateParentExists(ctx context.Context, parentID int64) error {
	_, err := s.dao.GetByID(ctx, parentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.New(errcode.NotFound, errcode.KeyNotFound)
		}
		return err
	}
	return nil
}

// menuDescendantContains 若 candidate 在 root 的子树中（含 root 自身）返回 true。
func menuDescendantContains(menus []model.Menu, root, candidate int64) bool {
	if root == candidate {
		return true
	}
	children := make(map[int64][]int64)
	for _, m := range menus {
		if m.ParentID == nil || *m.ParentID <= 0 {
			continue
		}
		p := *m.ParentID
		children[p] = append(children[p], m.ID)
	}
	var walk func(int64) bool
	walk = func(n int64) bool {
		for _, c := range children[n] {
			if c == candidate {
				return true
			}
			if walk(c) {
				return true
			}
		}
		return false
	}
	return walk(root)
}
