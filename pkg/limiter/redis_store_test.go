package limiter

import "testing"

func TestRedisStore_ipLimit_routeLimit_maxOverride(t *testing.T) {
	s := NewRedisStore("p:", 60, 100, 10, 50, 5, 120, 80)
	if s.ipLimit() != 120 {
		t.Fatalf("ipLimit want 120 got %d", s.ipLimit())
	}
	if s.routeLimit() != 80 {
		t.Fatalf("routeLimit want 80 got %d", s.routeLimit())
	}
}

func TestRedisStore_ipLimit_formulaWhenMaxZero(t *testing.T) {
	s := NewRedisStore("p:", 10, 2, 3, 1, 1, 0, 0)
	// ceil(2*10)+3 = 23
	if s.ipLimit() != 23 {
		t.Fatalf("ipLimit want 23 got %d", s.ipLimit())
	}
	// ceil(1*10)+1 = 11
	if s.routeLimit() != 11 {
		t.Fatalf("routeLimit want 11 got %d", s.routeLimit())
	}
}
