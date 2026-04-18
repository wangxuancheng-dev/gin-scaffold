package strutil

import "testing"

func TestSplitClean(t *testing.T) {
	got := SplitClean(" a , b ,  , c ", ",")
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

func TestSplitClean_EmptySep(t *testing.T) {
	got := SplitClean("  hello  ", "")
	if len(got) != 1 || got[0] != "hello" {
		t.Fatalf("got %#v", got)
	}
}

func TestJoinClean(t *testing.T) {
	got := JoinClean([]string{" a ", "", "b", " c "}, "|")
	if got != "a|b|c" {
		t.Fatalf("got %q", got)
	}
}

func TestStringValue(t *testing.T) {
	s := "x"
	if StringValue(nil) != "" || StringValue(&s) != "x" {
		t.Fatal()
	}
}
