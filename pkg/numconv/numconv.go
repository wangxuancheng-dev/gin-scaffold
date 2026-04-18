// Package numconv parses numbers from query/body strings with safe defaults on error.
package numconv

import (
	"strconv"
	"strings"
)

// ParseInt parses base-10 int after trimming space; empty or invalid returns def.
func ParseInt(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

// ParseInt64 parses base-10 int64; empty or invalid returns def.
func ParseInt64(s string, def int64) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return v
}

// ParseUint64 parses base-10 uint64; empty or invalid returns def.
func ParseUint64(s string, def uint64) uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return def
	}
	return v
}

// ParseFloat64 parses float64; empty or invalid returns def.
func ParseFloat64(s string, def float64) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return v
}
