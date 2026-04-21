package middleware

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gin-scaffold/internal/config"
	ginI18n "github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

type conciseMessage struct {
	ID          string `json:"id"`
	Translation string `json:"translation"`
}

// I18n 注册 gin-contrib/i18n，支持通过配置指定默认语言与可用语言包。
func I18n(cfg *config.I18nConfig) gin.HandlerFunc {
	defaultLang := language.MustParse("zh")
	rootPath := resolveI18nRootPath()
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
				// Expected file naming: <lang>.json or <name>.<lang>.json.
				if len(parts) >= 2 {
					langPart := parts[len(parts)-2]
					if len(parts) == 2 {
						langPart = parts[0]
					}
					if parsed, err := language.Parse(langPart); err == nil {
						acceptLang = append(acceptLang, parsed)
					}
				}
			}
			dir := filepath.Dir(cfg.BundlePaths[0])
			if dir != "." && dir != "" {
				rootPath = resolveI18nRootPath(dir)
			}
			if len(acceptLang) == 0 {
				acceptLang = []language.Tag{defaultLang}
			}
		}
	}

	return ginI18n.Localize(
		ginI18n.WithBundle(&ginI18n.BundleCfg{
			DefaultLanguage:  defaultLang,
			FormatBundleFile: "json",
			RootPath:         rootPath,
			AcceptLanguage:   acceptLang,
			UnmarshalFunc:    unmarshalI18nJSON,
			Loader:           nil,
		}),
	)
}

func resolveI18nRootPath(paths ...string) string {
	// Support running from repo root and from nested package test directories.
	candidates := make([]string, 0, len(paths)+4)
	for _, p := range paths {
		if strings.TrimSpace(p) != "" {
			candidates = append(candidates, p)
		}
	}
	candidates = append(candidates,
		"./i18n",
		"../i18n",
		"../../i18n",
		"../../../i18n",
	)
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
	}
	return "./i18n"
}

func unmarshalI18nJSON(data []byte, v any) error {
	trimmed := strings.TrimSpace(string(data))
	if !strings.HasPrefix(trimmed, "{") {
		return fmt.Errorf("i18n json must be object format, e.g. {\"success\":\"ok\"}")
	}
	kv := map[string]string{}
	if err := json.Unmarshal(data, &kv); err != nil {
		return err
	}
	items := make([]conciseMessage, 0, len(kv))
	keys := make([]string, 0, len(kv))
	for k := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		items = append(items, conciseMessage{
			ID:          k,
			Translation: kv[k],
		})
	}
	b, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
