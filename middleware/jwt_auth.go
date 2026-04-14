package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gin-scaffold/api/response"
	"gin-scaffold/internal/pkg/errcode"
	jwtpkg "gin-scaffold/internal/pkg/jwt"
)

const ctxClaims = "jwt_claims"

var rolePermissions = map[string]map[string]bool{
	"admin": {
		"db:ping": true,
		"user:rw": true,
	},
	"user": {
		"user:read": true,
	},
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
		policy, ok := rolePermissions[claims.Role]
		if !ok || !policy[permission] {
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
