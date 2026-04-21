// Package i18n 提供从 gin Context 读取本地化消息的辅助函数（依赖 gin-contrib/i18n）。
package i18n

import (
	"encoding/json"
	"strings"
	"sync"

	i18nres "gin-scaffold/i18n"

	"github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
)

var (
	onceCatalog sync.Once
	catalogs    map[string]map[string]string
)

// T 读取消息模板；若 i18n 未注入则返回 defaultMsg。
func T(c *gin.Context, messageID, defaultMsg string) string {
	if c != nil {
		if _, ok := c.Get("i18n"); ok {
			s, err := i18n.GetMessage(c, messageID)
			if err == nil && s != "" {
				return s
			}
		}
		if s := catalogMessage(languageFromContext(c), messageID); s != "" {
			return s
		}
		return defaultMsg
	}
	if s := catalogMessage("zh", messageID); s != "" {
		return s
	}
	return defaultMsg
}

// TFormat reads localized template and replaces {key} with params[key].
func TFormat(c *gin.Context, messageID, defaultTemplate string, params map[string]string) string {
	s := T(c, messageID, defaultTemplate)
	if len(params) == 0 {
		return s
	}
	for k, v := range params {
		s = strings.ReplaceAll(s, "{"+k+"}", v)
		s = strings.ReplaceAll(s, ":"+k, v)
	}
	return s
}

func languageFromContext(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return "zh"
	}
	lang := strings.ToLower(strings.TrimSpace(c.GetHeader("Accept-Language")))
	if strings.HasPrefix(lang, "en") {
		return "en"
	}
	return "zh"
}

func catalogMessage(lang, key string) string {
	loadCatalogs()
	lang = strings.ToLower(strings.TrimSpace(lang))
	if strings.HasPrefix(lang, "en") {
		lang = "en"
	} else {
		lang = "zh"
	}
	if m := catalogs[lang]; m != nil {
		return strings.TrimSpace(m[key])
	}
	return ""
}

func loadCatalogs() {
	onceCatalog.Do(func() {
		catalogs = map[string]map[string]string{
			"zh": {},
			"en": {},
		}
		loadOne := func(lang, name string) {
			b, err := i18nres.Files.ReadFile(name)
			if err != nil {
				return
			}
			tmp := map[string]string{}
			if err := json.Unmarshal(b, &tmp); err != nil {
				return
			}
			catalogs[lang] = tmp
		}
		loadOne("zh", "zh.json")
		loadOne("en", "en.json")
	})
}
