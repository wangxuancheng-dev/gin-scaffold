// Package sliceutil holds small slice helpers for handlers and services.
package sliceutil

// UniqueStable returns elements of in in first-seen order, dropping later duplicates.
func UniqueStable[T comparable](in []T) []T {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[T]struct{}, len(in))
	out := make([]T, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// Filter returns elements for which keep returns true, preserving order.
func Filter[T any](in []T, keep func(T) bool) []T {
	if len(in) == 0 {
		return nil
	}
	out := make([]T, 0, len(in))
	for _, v := range in {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

// Coalesce returns the first value not equal to T's zero value, or zero if none.
// Useful for query defaults (e.g. page size). Not suitable for float when NaN matters.
func Coalesce[T comparable](vals ...T) T {
	var zero T
	for _, v := range vals {
		if v != zero {
			return v
		}
	}
	return zero
}
