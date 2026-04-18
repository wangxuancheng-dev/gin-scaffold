// Package service 实现业务编排：事务、缓存、任务入队等。
package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"gin-scaffold/config"
	"gin-scaffold/internal/dao"
	"gin-scaffold/internal/job"
	"gin-scaffold/internal/model"
	"gin-scaffold/internal/pkg/errcode"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
	"gin-scaffold/internal/pkg/tenant"
	"gin-scaffold/pkg/logger"
	appredis "gin-scaffold/pkg/redis"
)

// UserRepo 定义 UserService 依赖的数据访问接口，便于单元测试注入 mock。
type UserRepo interface {
	Create(ctx context.Context, u *model.User) error
	CreateTx(ctx context.Context, tx *gorm.DB, u *model.User) error
	BindRole(ctx context.Context, userID int64, role string) error
	BindRoleTx(ctx context.Context, tx *gorm.DB, userID int64, role string) error
	Restore(ctx context.Context, id int64, hashedPassword, nickname string) (*model.User, error)
	RestoreTx(ctx context.Context, tx *gorm.DB, id int64, hashedPassword, nickname string) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByUsername(ctx context.Context, name string) (*model.User, error)
	GetByUsernameWithDeleted(ctx context.Context, name string) (*model.User, error)
	List(ctx context.Context, q model.UserQuery, offset, limit int) ([]model.User, int64, error)
	ListAfterID(ctx context.Context, q model.UserQuery, lastID int64, limit int) ([]model.User, error)
	GetPrimaryRole(ctx context.Context, userID int64) (string, error)
	GetPrimaryRoles(ctx context.Context, userIDs []int64) (map[int64]string, error)
	Update(ctx context.Context, id int64, nickname *string, hashedPassword *string) (*model.User, error)
	SoftDelete(ctx context.Context, id int64) error
	SetRole(ctx context.Context, userID int64, role string) error
}

// UserService 用户业务。
type UserService struct {
	dao              UserRepo
	queue            *job.Client
	jwt              *jwtpkg.Manager
	superAdminUserID int64
	db               *gorm.DB
	outbox           *dao.OutboxDAO
	outboxCfg        config.OutboxConfig
}

// NewUserService 构造。
func NewUserService(
	d UserRepo,
	q *job.Client,
	j *jwtpkg.Manager,
	superAdminUserID int64,
	db *gorm.DB,
	outbox *dao.OutboxDAO,
	outboxCfg config.OutboxConfig,
) *UserService {
	return &UserService{
		dao:              d,
		queue:            q,
		jwt:              j,
		superAdminUserID: superAdminUserID,
		db:               db,
		outbox:           outbox,
		outboxCfg:        outboxCfg,
	}
}

// AdminCreate 后台创建用户并绑定角色。
func (s *UserService) AdminCreate(ctx context.Context, username, password, nickname, role string) (*model.User, error) {
	if role == "" {
		role = "user"
	}
	if _, err := s.dao.GetByUsername(ctx, username); err == nil {
		return nil, errcode.New(errcode.UserExists, errcode.KeyUserExists)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
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
	if err := s.dao.SetRole(ctx, u.ID, role); err != nil {
		return nil, err
	}
	u.Password = ""
	return u, nil
}

// AdminUpdate 后台更新用户（昵称/密码/角色）。
func (s *UserService) AdminUpdate(ctx context.Context, id int64, nickname, password, role *string) (*model.User, error) {
	if _, err := s.dao.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.New(errcode.UserNotFound, errcode.KeyUserNotFound)
		}
		return nil, err
	}
	var hashed *string
	if password != nil {
		h, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		hv := string(h)
		hashed = &hv
	}
	u, err := s.dao.Update(ctx, id, nickname, hashed)
	if err != nil {
		return nil, err
	}
	if role != nil && *role != "" {
		if err := s.dao.SetRole(ctx, id, *role); err != nil {
			return nil, err
		}
	}
	u.Password = ""
	return u, nil
}

// AdminDelete 后台删除用户（软删除）。
func (s *UserService) AdminDelete(ctx context.Context, id int64) error {
	if s.superAdminUserID > 0 && id == s.superAdminUserID {
		return errcode.New(errcode.Forbidden, errcode.KeySuperAdminProtected)
	}
	if _, err := s.dao.GetByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.New(errcode.UserNotFound, errcode.KeyUserNotFound)
		}
		return err
	}
	return s.dao.SoftDelete(ctx, id)
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
	useOutbox := s.outboxCfg.Enabled && s.db != nil && s.outbox != nil

	existed, err := s.dao.GetByUsernameWithDeleted(ctx, username)
	if err == nil {
		if !existed.DeletedAt.Valid {
			return nil, errcode.New(errcode.UserExists, errcode.KeyUserExists)
		}
		var u *model.User
		if useOutbox {
			err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				row, e := s.dao.RestoreTx(ctx, tx, existed.ID, string(hash), nickname)
				if e != nil {
					return e
				}
				if e := s.dao.BindRoleTx(ctx, tx, row.ID, "user"); e != nil {
					return e
				}
				payload, e := json.Marshal(map[string]any{"user_id": row.ID, "username": row.Username})
				if e != nil {
					return e
				}
				maxAttempts := s.outboxCfg.MaxAttempts
				if maxAttempts <= 0 {
					maxAttempts = 10
				}
				if _, e := s.outbox.EnqueueTx(ctx, tx, "user.registered", string(payload), maxAttempts); e != nil {
					return e
				}
				u = row
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			u, err = s.dao.Restore(ctx, existed.ID, string(hash), nickname)
			if err != nil {
				return nil, err
			}
			if err := s.dao.BindRole(ctx, u.ID, "user"); err != nil {
				return nil, err
			}
		}
		if s.queue != nil {
			if err := s.queue.EnqueueWelcome(ctx, u.ID, u.Username); err != nil {
				logger.WarnX("enqueue welcome failed", zap.Int64("uid", u.ID), zap.Error(err))
			}
		}
		u.Password = ""
		return u, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	u := &model.User{
		Username: username,
		Password: string(hash),
		Nickname: nickname,
	}
	if useOutbox {
		err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if e := s.dao.CreateTx(ctx, tx, u); e != nil {
				return e
			}
			if e := s.dao.BindRoleTx(ctx, tx, u.ID, "user"); e != nil {
				return e
			}
			payload, e := json.Marshal(map[string]any{"user_id": u.ID, "username": u.Username})
			if e != nil {
				return e
			}
			maxAttempts := s.outboxCfg.MaxAttempts
			if maxAttempts <= 0 {
				maxAttempts = 10
			}
			if _, e := s.outbox.EnqueueTx(ctx, tx, "user.registered", string(payload), maxAttempts); e != nil {
				return e
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := s.dao.Create(ctx, u); err != nil {
			return nil, err
		}
		if err := s.dao.BindRole(ctx, u.ID, "user"); err != nil {
			return nil, err
		}
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
func (s *UserService) List(ctx context.Context, q model.UserQuery, page, pageSize int) ([]model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	rows, total, err := s.dao.List(ctx, q, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	for i := range rows {
		rows[i].Password = ""
	}
	return rows, total, nil
}

// StreamExport 流式导出，按批次扫描并回调消费导出行。
func (s *UserService) StreamExport(
	ctx context.Context,
	q model.UserQuery,
	page, pageSize, limit, batchSize int,
	pageOnly, withRole bool,
	consume func(model.UserExportRow) error,
) error {
	if batchSize <= 0 || batchSize > 5000 {
		batchSize = 1000
	}
	if pageOnly {
		rows, _, err := s.List(ctx, q, page, pageSize)
		if err != nil {
			return err
		}
		return s.emitExportRows(ctx, rows, withRole, consume)
	}

	if limit <= 0 || limit > 10_000_000 {
		limit = 100_000
	}

	var (
		lastID int64
		sent   int
	)
	for sent < limit {
		size := batchSize
		remain := limit - sent
		if remain < size {
			size = remain
		}
		rows, err := s.dao.ListAfterID(ctx, q, lastID, size)
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		if err := s.emitExportRows(ctx, rows, withRole, consume); err != nil {
			return err
		}
		lastID = rows[len(rows)-1].ID
		sent += len(rows)
	}
	return nil
}

func (s *UserService) emitExportRows(
	ctx context.Context,
	users []model.User,
	withRole bool,
	consume func(model.UserExportRow) error,
) error {
	roles := map[int64]string{}
	if withRole {
		ids := make([]int64, 0, len(users))
		for _, u := range users {
			ids = append(ids, u.ID)
		}
		m, err := s.dao.GetPrimaryRoles(ctx, ids)
		if err != nil {
			return err
		}
		roles = m
	}
	for _, u := range users {
		row := model.UserExportRow{
			ID:        u.ID,
			Username:  u.Username,
			Nickname:  u.Nickname,
			CreatedAt: u.CreatedAt,
			Role:      roles[u.ID],
		}
		if row.Role == "" {
			row.Role = "user"
		}
		if err := consume(row); err != nil {
			return err
		}
	}
	return nil
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
	token, err := s.jwt.IssueAccess(u.ID, role, resolveTenantID(ctx, u))
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
	access, err := s.jwt.IssueAccess(u.ID, role, resolveTenantID(ctx, u))
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
	access, err := s.jwt.IssueAccess(u.ID, role, resolveTenantID(ctx, u))
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

func resolveTenantID(ctx context.Context, u *model.User) string {
	if u != nil && u.TenantID != "" {
		return u.TenantID
	}
	if tid := tenant.FromContext(ctx); tid != "" {
		return tid
	}
	return "default"
}
