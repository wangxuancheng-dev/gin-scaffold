// Package timefmt centralizes API-facing time string parsing and optional formatting helpers.
package timefmt

import (
	"strings"
	"time"
)

// ParseRFC3339 parses s as RFC3339 after trimming leading and trailing ASCII space.
// Use this for query/body fields so accidental spaces do not fail the whole request.
func ParseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(s))
}

// FormatPtr returns RFC3339 for non-nil t, or "" when t is nil.
// The formatted instant uses t's location (no forced UTC); use t.UTC() before calling if needed.
func FormatPtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
