package clientreq

// UserRegisterRequest 用户注册请求。
type UserRegisterRequest struct {
	Username string `json:"username" form:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" form:"password" binding:"required,min=6,max=64"`
	Nickname string `json:"nickname" form:"nickname" binding:"max=64"`
}

// UserIDURI 路径参数用户 ID。
type UserIDURI struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}
