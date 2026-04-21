package validator

import (
	"reflect"
	"strings"
	"time"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

type customRuleSpec struct {
	tag      string
	validate validator.Func
}

func registerBuiltInCustomRules(v *validator.Validate, transByLang map[string]ut.Translator) {
	_ = transByLang // kept for forward compatibility
	for _, spec := range builtInRuleSpecs() {
		if strings.TrimSpace(spec.tag) == "" || spec.validate == nil {
			continue
		}
		_ = v.RegisterValidation(spec.tag, spec.validate)
	}
}

func builtInRuleSpecs() []customRuleSpec {
	specs := make([]customRuleSpec, 0, 8)
	specs = append(specs, stringRuleSpecs()...)
	specs = append(specs, numericRuleSpecs()...)
	specs = append(specs, crossFieldRuleSpecs()...)
	return specs
}

func stringRuleSpecs() []customRuleSpec {
	return []customRuleSpec{
		{
			tag: "not_admin",
			validate: func(fl validator.FieldLevel) bool {
				if fl.Field().Kind() != reflect.String {
					return true
				}
				s := strings.TrimSpace(fl.Field().String())
				if s == "" {
					return true
				}
				return !strings.EqualFold(s, "admin")
			},
		},
	}
}

func numericRuleSpecs() []customRuleSpec {
	return nil
}

func crossFieldRuleSpecs() []customRuleSpec {
	return []customRuleSpec{
		{
			tag: "same_field",
			validate: func(fl validator.FieldLevel) bool {
				otherFieldName := strings.TrimSpace(fl.Param())
				if otherFieldName == "" {
					return true
				}
				other := fl.Parent().FieldByName(otherFieldName)
				if !other.IsValid() {
					return false
				}
				return compareEqual(fl.Field(), other)
			},
		},
		{
			tag: "after_field",
			validate: func(fl validator.FieldLevel) bool {
				otherFieldName := strings.TrimSpace(fl.Param())
				if otherFieldName == "" {
					return true
				}
				other := fl.Parent().FieldByName(otherFieldName)
				if !other.IsValid() {
					return false
				}
				return compareAfter(fl.Field(), other)
			},
		},
	}
}

func compareEqual(current, other reflect.Value) bool {
	if !current.IsValid() || !other.IsValid() {
		return false
	}
	if current.Kind() == reflect.Ptr {
		if current.IsNil() {
			return other.Kind() == reflect.Ptr && other.IsNil()
		}
		current = current.Elem()
	}
	if other.Kind() == reflect.Ptr {
		if other.IsNil() {
			return false
		}
		other = other.Elem()
	}
	if current.Type() != other.Type() {
		return false
	}
	switch current.Kind() {
	case reflect.String:
		return strings.TrimSpace(current.String()) == strings.TrimSpace(other.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return current.Int() == other.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return current.Uint() == other.Uint()
	case reflect.Float32, reflect.Float64:
		return current.Float() == other.Float()
	case reflect.Bool:
		return current.Bool() == other.Bool()
	case reflect.Struct:
		if current.Type() == reflect.TypeOf(time.Time{}) {
			return current.Interface().(time.Time).Equal(other.Interface().(time.Time))
		}
	}
	return reflect.DeepEqual(current.Interface(), other.Interface())
}

func compareAfter(current, other reflect.Value) bool {
	if !current.IsValid() || !other.IsValid() {
		return false
	}
	if current.Kind() == reflect.Ptr {
		if current.IsNil() {
			return true
		}
		current = current.Elem()
	}
	if other.Kind() == reflect.Ptr {
		if other.IsNil() {
			return true
		}
		other = other.Elem()
	}

	if current.Type() == reflect.TypeOf(time.Time{}) && other.Type() == reflect.TypeOf(time.Time{}) {
		c := current.Interface().(time.Time)
		o := other.Interface().(time.Time)
		if c.IsZero() || o.IsZero() {
			return true
		}
		return c.After(o)
	}

	if current.Kind() == reflect.String && other.Kind() == reflect.String {
		cs := strings.TrimSpace(current.String())
		os := strings.TrimSpace(other.String())
		if cs == "" || os == "" {
			return true
		}
		ct, err1 := time.Parse(time.RFC3339, cs)
		ot, err2 := time.Parse(time.RFC3339, os)
		if err1 == nil && err2 == nil {
			return ct.After(ot)
		}
		return cs > os
	}

	switch current.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if other.Kind() >= reflect.Int && other.Kind() <= reflect.Int64 {
			return current.Int() > other.Int()
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if other.Kind() >= reflect.Uint && other.Kind() <= reflect.Uint64 {
			return current.Uint() > other.Uint()
		}
	case reflect.Float32, reflect.Float64:
		if other.Kind() == reflect.Float32 || other.Kind() == reflect.Float64 {
			return current.Float() > other.Float()
		}
	}
	return false
}
