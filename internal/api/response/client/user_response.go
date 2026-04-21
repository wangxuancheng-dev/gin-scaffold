package clientresp

import (
	"time"

	"github.com/jinzhu/copier"

	"gin-scaffold/internal/model"
)

// UserVO 用户对外响应模型。
type UserVO struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// FromUser 将 model.User 转换为对外响应结构体。
func FromUser(u *model.User) UserVO {
	if u == nil {
		return UserVO{}
	}
	var vo UserVO
	_ = copier.Copy(&vo, u)
	if !u.CreatedAt.IsZero() {
		vo.CreatedAt = u.CreatedAt.Format(time.DateTime)
	}
	if !u.UpdatedAt.IsZero() {
		vo.UpdatedAt = u.UpdatedAt.Format(time.DateTime)
	}
	return vo
}
