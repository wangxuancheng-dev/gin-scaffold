package middleware

import (
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// I18n 注册 gin-contrib/i18n，从 ./i18n 目录加载 zh.yaml / en.yaml。
func I18n() gin.HandlerFunc {
	return ginI18n.Localize(
		ginI18n.WithBundle(&ginI18n.BundleCfg{
			DefaultLanguage:  language.MustParse("zh"),
			FormatBundleFile: "yaml",
			RootPath:         "./i18n",
			AcceptLanguage: []language.Tag{
				language.MustParse("zh"),
				language.MustParse("en"),
			},
			UnmarshalFunc: yaml.Unmarshal,
			Loader:        nil,
		}),
	)
}
