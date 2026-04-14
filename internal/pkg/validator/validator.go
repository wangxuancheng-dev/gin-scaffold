// Package validator 封装 go-playground/validator 与中文翻译。
package validator

import (
	"sync"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	v     *validator.Validate
	trans ut.Translator
	once  sync.Once
)

// V 返回全局校验器（懒加载）。
func V() *validator.Validate {
	once.Do(func() {
		v = validator.New()
		zhLoc := zh.New()
		uni := ut.New(zhLoc, zhLoc)
		var found bool
		trans, found = uni.GetTranslator("zh")
		if !found {
			panic("validator: zh translator not found")
		}
		if err := zhTranslations.RegisterDefaultTranslations(v, trans); err != nil {
			panic(err)
		}
	})
	return v
}

// Translator 返回中文翻译器。
func Translator() ut.Translator {
	_ = V()
	return trans
}
