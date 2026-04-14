// Package i18n 提供从 gin Context 读取本地化消息的辅助函数（依赖 gin-contrib/i18n）。
package i18n

import (
	"github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
)

// T 读取消息模板；若 i18n 未注入则返回 defaultMsg。
func T(c *gin.Context, messageID, defaultMsg string) string {
	if c == nil {
		return defaultMsg
	}
	if _, ok := c.Get("i18n"); !ok {
		return defaultMsg
	}
	s, err := i18n.GetMessage(c, messageID)
	if err != nil || s == "" {
		return defaultMsg
	}
	return s
}
