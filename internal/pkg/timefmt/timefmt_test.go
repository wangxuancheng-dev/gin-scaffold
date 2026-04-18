package timefmt

import (
	"testing"
	"time"
)

func TestParseRFC3339_TrimSpace(t *testing.T) {
	want := time.Date(2026, 4, 18, 12, 0, 0, 0, time.UTC)
	got, err := ParseRFC3339("  2026-04-18T12:00:00Z  ")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParseRFC3339_Invalid(t *testing.T) {
	_, err := ParseRFC3339("not-a-time")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFormatPtr(t *testing.T) {
	if FormatPtr(nil) != "" {
		t.Fatal()
	}
	tt := time.Date(2026, 4, 18, 0, 0, 0, 0, time.UTC)
	if FormatPtr(&tt) != "2026-04-18T00:00:00Z" {
		t.Fatalf("got %q", FormatPtr(&tt))
	}
}
