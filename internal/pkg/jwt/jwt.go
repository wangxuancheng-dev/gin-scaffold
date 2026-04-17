// Package jwt 封装访问令牌与刷新令牌的签发、解析。
package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"gin-scaffold/config"
)

// Claims 访问令牌载荷。
type Claims struct {
	UserID    int64  `json:"uid"`
	Role      string `json:"role"`
	TenantID  string `json:"tenant_id,omitempty"`
	jwtlib.RegisteredClaims
}

var (
	ErrTokenInvalid = errors.New("jwt: token invalid")
	ErrTokenExpired = errors.New("jwt: token expired")
)

// Manager JWT 管理器。
type Manager struct {
	secret []byte
	iss    string
	acc    time.Duration
	ref    time.Duration
}

// NewManager 从配置创建。
func NewManager(cfg *config.JWTConfig) *Manager {
	return &Manager{
		secret: []byte(cfg.Secret),
		iss:    cfg.Issuer,
		acc:    time.Duration(cfg.AccessExpireMin) * time.Minute,
		ref:    time.Duration(cfg.RefreshExpireMin) * time.Minute,
	}
}

// IssueAccess 签发访问令牌。
func (m *Manager) IssueAccess(userID int64, role string, tenantID string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Role:     role,
		TenantID: tenantID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Issuer:    m.iss,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(m.acc)),
		},
	}
	t := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return t.SignedString(m.secret)
}

// IssueRefresh 签发刷新令牌（更长有效期）。
func (m *Manager) IssueRefresh(userID int64) (string, error) {
	now := time.Now()
	claims := jwtlib.RegisteredClaims{
		ID:        uuid.NewString(),
		Subject:   fmt.Sprintf("%d", userID),
		Issuer:    m.iss,
		IssuedAt:  jwtlib.NewNumericDate(now),
		ExpiresAt: jwtlib.NewNumericDate(now.Add(m.ref)),
	}
	t := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return t.SignedString(m.secret)
}

// ParseRefresh 解析刷新令牌并返回用户 ID、jti 与过期时间。
func (m *Manager) ParseRefresh(token string) (int64, string, time.Time, error) {
	parsed, err := jwtlib.ParseWithClaims(token, &jwtlib.RegisteredClaims{}, func(t *jwtlib.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return 0, "", time.Time{}, ErrTokenExpired
		}
		return 0, "", time.Time{}, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}
	c, ok := parsed.Claims.(*jwtlib.RegisteredClaims)
	if !ok || !parsed.Valid {
		return 0, "", time.Time{}, ErrTokenInvalid
	}
	uid, convErr := strconv.ParseInt(c.Subject, 10, 64)
	if convErr != nil {
		return 0, "", time.Time{}, fmt.Errorf("%w: invalid refresh subject", ErrTokenInvalid)
	}
	if c.ID == "" || c.ExpiresAt == nil {
		return 0, "", time.Time{}, fmt.Errorf("%w: invalid refresh claims", ErrTokenInvalid)
	}
	return uid, c.ID, c.ExpiresAt.Time, nil
}

// ParseAccess 解析并校验访问令牌。
func (m *Manager) ParseAccess(token string) (*Claims, error) {
	parsed, err := jwtlib.ParseWithClaims(token, &Claims{}, func(t *jwtlib.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}
	c, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, ErrTokenInvalid
	}
	return c, nil
}
