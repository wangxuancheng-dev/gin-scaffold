package clientreq

// LoginRequest 登录请求体。
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=64"`
}

// RefreshTokenRequest 刷新访问令牌请求体。
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
