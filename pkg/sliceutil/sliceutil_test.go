package sliceutil

import "testing"

func TestUniqueStable(t *testing.T) {
	got := UniqueStable([]string{"a", "b", "a", "c", "b"})
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v want %#v", got, want)
		}
	}
}

func TestFilter(t *testing.T) {
	got := Filter([]int{1, 2, 3, 4}, func(n int) bool { return n%2 == 0 })
	if len(got) != 2 || got[0] != 2 || got[1] != 4 {
		t.Fatalf("got %#v", got)
	}
}

func TestCoalesce(t *testing.T) {
	if Coalesce(0, 0, 3) != 3 {
		t.Fatal()
	}
	if Coalesce(0, 0) != 0 {
		t.Fatal()
	}
}
