// Package strutil holds small string helpers common in HTTP handlers and serializers.
package strutil

import (
	"path"
	"strings"
)

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

// AttachmentFilename returns a safe single-token name for Content-Disposition from a storage key.
// It strips path segments, CR/LF, and double-quotes to reduce response-header injection risk.
func AttachmentFilename(key string) string {
	k := strings.TrimSpace(key)
	k = strings.ReplaceAll(k, "\\", "/")
	s := path.Base(k)
	if s == "." || s == "/" || s == "" {
		return "download"
	}
	s = strings.ReplaceAll(s, "\"", "'")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.TrimSpace(s)
	if s == "" {
		return "download"
	}
	return s
}
