// Package validator 封装 go-playground/validator 与中文翻译。
package validator

import (
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	v        *validator.Validate
	uni      *ut.UniversalTranslator
	transZH  ut.Translator
	transEN  ut.Translator
	once     sync.Once
	bindOnce sync.Once
)

func init() {
	InitGinBindingValidator()
}

// V 返回全局校验器（懒加载）。
func V() *validator.Validate {
	once.Do(func() {
		v = validator.New()
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			tag := fld.Tag.Get("json")
			if tag == "" || tag == "-" {
				return fld.Name
			}
			name := strings.Split(tag, ",")[0]
			if name == "" || name == "-" {
				return fld.Name
			}
			return name
		})
		zhLoc := zh.New()
		enLoc := en.New()
		uni = ut.New(zhLoc, zhLoc, enLoc)
		var ok bool
		transZH, ok = uni.GetTranslator("zh")
		if !ok {
			panic("validator: zh translator not found")
		}
		transEN, ok = uni.GetTranslator("en")
		if !ok {
			panic("validator: en translator not found")
		}
		if err := zhTranslations.RegisterDefaultTranslations(v, transZH); err != nil {
			panic(err)
		}
		if err := enTranslations.RegisterDefaultTranslations(v, transEN); err != nil {
			panic(err)
		}
		registerCustomRules(v, transZH, transEN)
	})
	return v
}

// Translator 返回中文翻译器。
func Translator() ut.Translator {
	_ = V()
	return transZH
}

// TranslatorForLang 根据 Accept-Language 返回对应翻译器（当前支持 zh/en）。
func TranslatorForLang(acceptLanguage string) ut.Translator {
	_ = V()
	lang := strings.ToLower(strings.TrimSpace(acceptLanguage))
	if strings.HasPrefix(lang, "en") {
		return transEN
	}
	return transZH
}

// NormalizeFieldName converts Go-style field names to client-friendly names.
func NormalizeFieldName(field string) string {
	field = strings.TrimSpace(field)
	if field == "" {
		return ""
	}
	runes := []rune(field)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// InitGinBindingValidator wires project custom rules into Gin binding validator engine.
func InitGinBindingValidator() {
	_ = V()
	bindOnce.Do(func() {
		engine, ok := binding.Validator.Engine().(*validator.Validate)
		if !ok || engine == nil {
			return
		}
		registerCustomRules(engine, transZH, transEN)
	})
}

func registerCustomRules(v *validator.Validate, zhTrans, enTrans ut.Translator) {
	if v == nil {
		return
	}
	registerBuiltInCustomRules(v, map[string]ut.Translator{
		"zh": zhTrans,
		"en": enTrans,
	})
}
