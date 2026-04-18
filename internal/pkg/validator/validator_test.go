package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

type sample struct {
	Name string `validate:"required,min=2"`
	Age  int    `validate:"min=1"`
}

func TestV_Struct(t *testing.T) {
	err := V().Struct(&sample{Name: "ab", Age: 1})
	if err != nil {
		t.Fatalf("valid struct: %v", err)
	}
	err = V().Struct(&sample{Name: "", Age: 1})
	if err == nil {
		t.Fatal("expect validation error")
	}
	if _, ok := err.(validator.ValidationErrors); !ok {
		t.Fatalf("want ValidationErrors got %T", err)
	}
}

func TestTranslator(t *testing.T) {
	if Translator() == nil {
		t.Fatal()
	}
}
