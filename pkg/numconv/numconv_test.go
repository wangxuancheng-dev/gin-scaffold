package numconv

import "testing"

func TestParseInt(t *testing.T) {
	if ParseInt(" 42 ", 0) != 42 {
		t.Fatal()
	}
	if ParseInt("x", 7) != 7 {
		t.Fatal()
	}
	if ParseInt("", 9) != 9 {
		t.Fatal()
	}
}

func TestParseInt64(t *testing.T) {
	if ParseInt64("-1", 0) != -1 {
		t.Fatal()
	}
	if ParseInt64("bad", 99) != 99 {
		t.Fatal()
	}
}

func TestParseUint64(t *testing.T) {
	if ParseUint64("10", 0) != 10 {
		t.Fatal()
	}
}

func TestParseFloat64(t *testing.T) {
	if ParseFloat64("3.5", 0) != 3.5 {
		t.Fatal()
	}
}
