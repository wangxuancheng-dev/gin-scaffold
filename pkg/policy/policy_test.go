package policy

import "testing"

func TestSameUser(t *testing.T) {
	if !SameUser(1, 1) {
		t.Fatal()
	}
	if SameUser(0, 1) || SameUser(1, 0) || SameUser(1, 2) {
		t.Fatal()
	}
}
