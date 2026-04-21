package validator

import (
	"testing"
	"time"

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

type timeRangeSample struct {
	StartAt time.Time `validate:"required"`
	EndAt   time.Time `validate:"required,after_field=StartAt"`
}

type passwordPairSample struct {
	Password string `validate:"required,min=6"`
	Confirm  string `validate:"required,same_field=Password"`
}

func TestCustomRule_AfterField(t *testing.T) {
	now := time.Now()
	okRange := &timeRangeSample{StartAt: now, EndAt: now.Add(time.Minute)}
	if err := V().Struct(okRange); err != nil {
		t.Fatalf("expected valid range, got: %v", err)
	}

	badRange := &timeRangeSample{StartAt: now, EndAt: now.Add(-time.Minute)}
	err := V().Struct(badRange)
	if err == nil {
		t.Fatal("expect validation error for after_field")
	}
	verr, ok := err.(validator.ValidationErrors)
	if !ok || len(verr) == 0 {
		t.Fatalf("want ValidationErrors got %T", err)
	}
	if verr[0].Tag() != "after_field" {
		t.Fatalf("want after_field tag, got %s", verr[0].Tag())
	}
}

func TestCustomRule_SameField(t *testing.T) {
	okPair := &passwordPairSample{Password: "secret1", Confirm: "secret1"}
	if err := V().Struct(okPair); err != nil {
		t.Fatalf("expected valid same_field pair, got: %v", err)
	}

	badPair := &passwordPairSample{Password: "secret1", Confirm: "secret2"}
	err := V().Struct(badPair)
	if err == nil {
		t.Fatal("expect validation error for same_field")
	}
	verr, ok := err.(validator.ValidationErrors)
	if !ok || len(verr) == 0 {
		t.Fatalf("want ValidationErrors got %T", err)
	}
	if verr[0].Tag() != "same_field" {
		t.Fatalf("want same_field tag, got %s", verr[0].Tag())
	}
}
