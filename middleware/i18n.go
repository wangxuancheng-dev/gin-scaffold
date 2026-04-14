package middleware

import (
	"path/filepath"
	"strings"

	"gin-scaffold/config"
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

// I18n 注册 gin-contrib/i18n，支持通过配置指定默认语言与可用语言包。
func I18n(cfg *config.I18nConfig) gin.HandlerFunc {
	defaultLang := language.MustParse("zh")
	rootPath := "./i18n"
	acceptLang := []language.Tag{language.MustParse("zh"), language.MustParse("en")}

	if cfg != nil {
		if cfg.DefaultLang != "" {
			if parsed, err := language.Parse(cfg.DefaultLang); err == nil {
				defaultLang = parsed
			}
		}
		if len(cfg.BundlePaths) > 0 {
			acceptLang = make([]language.Tag, 0, len(cfg.BundlePaths))
			for _, p := range cfg.BundlePaths {
				base := filepath.Base(p)
				parts := strings.Split(base, ".")
				// Expected file naming: <name>.<lang>.yaml, e.g. active.zh.yaml.
				if len(parts) >= 3 {
					if parsed, err := language.Parse(parts[len(parts)-2]); err == nil {
						acceptLang = append(acceptLang, parsed)
					}
				}
			}
			dir := filepath.Dir(cfg.BundlePaths[0])
			if dir != "." && dir != "" {
				rootPath = dir
			}
			if len(acceptLang) == 0 {
				acceptLang = []language.Tag{defaultLang}
			}
		}
	}

	return ginI18n.Localize(
		ginI18n.WithBundle(&ginI18n.BundleCfg{
			DefaultLanguage:  defaultLang,
			FormatBundleFile: "yaml",
			RootPath:         rootPath,
			AcceptLanguage:   acceptLang,
			UnmarshalFunc:    yaml.Unmarshal,
			Loader:           nil,
		}),
	)
}
