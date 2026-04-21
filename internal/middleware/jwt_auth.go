package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"gin-scaffold/internal/api/response"
	"gin-scaffold/internal/pkg/errcode"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
)

const ctxClaims = "jwt_claims"

var (
	rbacMu            sync.RWMutex
	permissionChecker PermissionChecker
	superAdminUserID  int64
)

// PermissionChecker 支持将权限判断委托到数据库/权限服务。
type PermissionChecker interface {
	HasPermission(ctx context.Context, userID int64, role, permission string) (bool, error)
}

// SetPermissionChecker 设置自定义权限检查器（如基于数据表的 RBAC）。
func SetPermissionChecker(checker PermissionChecker) {
	rbacMu.Lock()
	defer rbacMu.Unlock()
	permissionChecker = checker
}

// SetSuperAdminUserID 设置内置超管用户 ID（0 表示关闭）。
func SetSuperAdminUserID(id int64) {
	rbacMu.Lock()
	defer rbacMu.Unlock()
	superAdminUserID = id
}

func isSuperAdminUser(userID int64) bool {
	rbacMu.RLock()
	id := superAdminUserID
	rbacMu.RUnlock()
	return id > 0 && userID == id
}

// JWTAuth 校验 Authorization Bearer 访问令牌。
func JWTAuth(m *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if m == nil {
			c.Next()
			return
		}
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			response.FailHTTP(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.KeyUnauthorized, "missing token")
			c.Abort()
			return
		}
		raw := strings.TrimSpace(h[7:])
		claims, err := m.ParseAccess(raw)
		if err != nil {
			response.FailHTTP(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.KeyUnauthorized, err.Error())
			c.Abort()
			return
		}
		if jwtpkg.IsAccessTokenRevoked(c.Request.Context(), raw) {
			response.FailHTTP(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.KeyUnauthorized, "token revoked")
			c.Abort()
			return
		}
		c.Set(ctxClaims, claims)
		c.Set("jwt_raw_token", raw)
		c.Next()
	}
}

// Claims 读取已鉴权的 JWT 声明。
func Claims(c *gin.Context) (*jwtpkg.Claims, bool) {
	v, ok := c.Get(ctxClaims)
	if !ok {
		return nil, false
	}
	cl, ok := v.(*jwtpkg.Claims)
	return cl, ok
}

// RequireRoles 要求 JWT 声明中的 role 在允许集合内（RBAC 基础能力）。
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := Claims(c)
		if !ok || claims == nil {
			response.FailHTTP(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.KeyUnauthorized, "missing claims")
			c.Abort()
			return
		}
		role := claims.Role
		if isSuperAdminUser(claims.UserID) {
			c.Next()
			return
		}
		for _, allowed := range roles {
			if role == allowed {
				c.Next()
				return
			}
		}
		response.FailHTTP(c, http.StatusForbidden, errcode.Forbidden, errcode.KeyForbidden, "forbidden")
		c.Abort()
	}
}

// RequirePermission 基于角色策略校验权限（RBAC 细粒度）。
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := Claims(c)
		if !ok || claims == nil {
			response.FailHTTP(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.KeyUnauthorized, "missing claims")
			c.Abort()
			return
		}
		rbacMu.RLock()
		checker := permissionChecker
		rbacMu.RUnlock()
		if isSuperAdminUser(claims.UserID) {
			c.Next()
			return
		}

		if checker == nil {
			response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, "permission checker not configured")
			c.Abort()
			return
		}

		allowed, err := checker.HasPermission(c.Request.Context(), claims.UserID, claims.Role, permission)
		if err != nil {
			response.FailHTTP(c, http.StatusInternalServerError, errcode.InternalError, errcode.KeyInternal, err.Error())
			c.Abort()
			return
		}

		if !allowed {
			response.FailHTTP(c, http.StatusForbidden, errcode.Forbidden, errcode.KeyForbidden, "forbidden")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RawToken 获取当前请求中通过鉴权的 access token 原文。
func RawToken(c *gin.Context) (string, bool) {
	v, ok := c.Get("jwt_raw_token")
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
