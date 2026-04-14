// Package service 实现业务编排：事务、缓存、任务入队等。
package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"gin-scaffold/internal/job"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/pkg/logger"
	appredis "gin-scaffold/pkg/redis"
)

// UserRepo 定义 UserService 依赖的数据访问接口，便于单元测试注入 mock。
type UserRepo interface {
	Create(ctx context.Context, u *model.User) error
	BindRole(ctx context.Context, userID int64, role string) error
	Restore(ctx context.Context, id int64, hashedPassword, nickname string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByUsername(ctx context.Context, name string) (*model.User, error)
	GetByUsernameWithDeleted(ctx context.Context, name string) (*model.User, error)
	List(ctx context.Context, offset, limit int) ([]model.User, int64, error)
	GetPrimaryRole(ctx context.Context, userID int64) (string, error)
}

// UserService 用户业务。
type UserService struct {
	dao   UserRepo
	queue *job.Client
	jwt   *jwtpkg.Manager
}

// NewUserService 构造。
func NewUserService(d UserRepo, q *job.Client, j *jwtpkg.Manager) *UserService {
	return &UserService{dao: d, queue: q, jwt: j}
}

// Register 注册并异步发送欢迎任务。
func (s *UserService) Register(ctx context.Context, username, password, nickname string) (*model.User, error) {
	if _, err := s.dao.GetByUsername(ctx, username); err == nil {
		return nil, errcode.New(errcode.UserExists, errcode.KeyUserExists)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	if existed, err := s.dao.GetByUsernameWithDeleted(ctx, username); err == nil {
		// Same username exists as soft-deleted record: restore it.
		if existed.DeletedAt.Valid {
			restored, err := s.dao.Restore(ctx, existed.ID, string(hash), nickname)
			if err != nil {
				return nil, err
			}
			if err := s.dao.BindRole(ctx, restored.ID, "user"); err != nil {
				return nil, err
			}
			restored.Password = ""
			return restored, nil
		}
		return nil, errcode.New(errcode.UserExists, errcode.KeyUserExists)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	u := &model.User{
		Username: username,
		Password: string(hash),
		Nickname: nickname,
	}
	if err := s.dao.Create(ctx, u); err != nil {
		return nil, err
	}
	if err := s.dao.BindRole(ctx, u.ID, "user"); err != nil {
		return nil, err
	}
	if s.queue != nil {
		if err := s.queue.EnqueueWelcome(ctx, u.ID, u.Username); err != nil {
			logger.WarnX("enqueue welcome failed", zap.Int64("uid", u.ID), zap.Error(err))
		}
	}
	u.Password = ""
	return u, nil
}

// GetByID 查询用户；带简单缓存。
func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	key := "user:" + strconv.FormatInt(id, 10)
	var u model.User
	err := appredis.CacheGetOrSet(ctx, key, 5*time.Minute, &u, func(c context.Context) (interface{}, error) {
		row, err := s.dao.GetByID(c, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil
			}
			return nil, err
		}
		return row, nil
	})
	if err != nil {
		if errors.Is(err, appredis.ErrCacheNull) {
			return nil, errcode.New(errcode.UserNotFound, errcode.KeyUserNotFound)
		}
		return nil, err
	}
	u.Password = ""
	return &u, nil
}

// List 用户分页。
func (s *UserService) List(ctx context.Context, page, pageSize int) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	rows, total, err := s.dao.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for i := range rows {
		rows[i].Password = ""
	}
	return rows, total, nil
}

// Login 校验密码并签发访问令牌。
func (s *UserService) Login(ctx context.Context, username, password string) (access string, err error) {
	u, err := s.dao.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errcode.New(errcode.Unauthorized, errcode.KeyUnauthorized)
		}
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return "", errcode.New(errcode.Unauthorized, errcode.KeyUnauthorized)
	}
	if s.jwt == nil {
		return "", errcode.New(errcode.InternalError, errcode.KeyInternal)
	}
	role, err := s.getPrimaryRole(ctx, u.ID)
	if err != nil {
		return "", err
	}
	token, err := s.jwt.IssueAccess(u.ID, role)
	if err != nil {
		return "", err
	}
	return token, nil
}

// LoginWithRefresh 登录后同时签发 access 与 refresh token。
func (s *UserService) LoginWithRefresh(ctx context.Context, username, password string) (string, string, error) {
	u, err := s.dao.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errcode.New(errcode.Unauthorized, errcode.KeyUnauthorized)
		}
		return "", "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return "", "", errcode.New(errcode.Unauthorized, errcode.KeyUnauthorized)
	}
	if s.jwt == nil {
		return "", "", errcode.New(errcode.InternalError, errcode.KeyInternal)
	}
	role, err := s.getPrimaryRole(ctx, u.ID)
	if err != nil {
		return "", "", err
	}
	access, err := s.jwt.IssueAccess(u.ID, role)
	if err != nil {
		return "", "", err
	}
	refresh, err := s.jwt.IssueRefresh(u.ID)
	if err != nil {
		return "", "", err
	}
	_, jti, exp, err := s.jwt.ParseRefresh(refresh)
	if err != nil {
		return "", "", err
	}
	if err = jwtpkg.SaveRefreshJTI(ctx, u.ID, jti, exp); err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

// RefreshAccess 通过 refresh token 刷新 access token。
func (s *UserService) RefreshAccess(ctx context.Context, refreshToken string) (string, string, error) {
	if s.jwt == nil {
		return "", "", errcode.New(errcode.InternalError, errcode.KeyInternal)
	}
	uid, jti, _, err := s.jwt.ParseRefresh(refreshToken)
	if err != nil {
		return "", "", errcode.New(errcode.Unauthorized, errcode.KeyUnauthorized)
	}
	if err = jwtpkg.ValidateRefreshJTI(ctx, uid, jti); err != nil {
		return "", "", errcode.New(errcode.Unauthorized, errcode.KeyUnauthorized)
	}
	u, err := s.dao.GetByID(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errcode.New(errcode.Unauthorized, errcode.KeyUnauthorized)
		}
		return "", "", err
	}
	role, err := s.getPrimaryRole(ctx, u.ID)
	if err != nil {
		return "", "", err
	}
	access, err := s.jwt.IssueAccess(u.ID, role)
	if err != nil {
		return "", "", err
	}
	newRefresh, err := s.jwt.IssueRefresh(u.ID)
	if err != nil {
		return "", "", err
	}
	_, newJTI, newExp, err := s.jwt.ParseRefresh(newRefresh)
	if err != nil {
		return "", "", err
	}
	if err = jwtpkg.SaveRefreshJTI(ctx, u.ID, newJTI, newExp); err != nil {
		return "", "", err
	}
	return access, newRefresh, nil
}

func (s *UserService) getPrimaryRole(ctx context.Context, userID int64) (string, error) {
	role, err := s.dao.GetPrimaryRole(ctx, userID)
	if err == nil && role != "" {
		return role, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Keep backward compatibility for users without role binding yet.
		return "user", nil
	}
	if err != nil {
		return "", err
	}
	return "user", nil
}
