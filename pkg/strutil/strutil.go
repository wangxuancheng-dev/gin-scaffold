// Package strutil holds small string helpers common in HTTP handlers and serializers.
package strutil

import "strings"

// SplitClean splits s by sep, trims spaces on each part, and drops empty segments.
// If sep is empty, returns a single segment of strings.TrimSpace(s) (or nil if empty).
func SplitClean(s, sep string) []string {
	if sep == "" {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		return []string{s}
	}
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// JoinClean trims each element, skips empties, then strings.Join with sep.
func JoinClean(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	b := make([]string, 0, len(elems))
	for _, e := range elems {
		e = strings.TrimSpace(e)
		if e != "" {
			b = append(b, e)
		}
	}
	return strings.Join(b, sep)
}

// StringValue returns the pointed string, or "" if s is nil.
func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
