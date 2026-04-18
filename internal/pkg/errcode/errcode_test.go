package errcode

import (
	"errors"
	"testing"
)

func TestBizError_Error(t *testing.T) {
	e := New(1, "k")
	if e.Error() != "code=1 key=k" {
		t.Fatalf("got %q", e.Error())
	}
	inner := errors.New("inner")
	w := Wrap(2, "wrap", inner)
	if w.Unwrap() == nil {
		t.Fatal()
	}
	if !errors.Is(w, inner) {
		t.Fatal("Unwrap chain")
	}
}
