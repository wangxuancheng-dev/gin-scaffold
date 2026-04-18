package limiter

import "testing"

func TestStore_AllowIPAndRoute(t *testing.T) {
	s := NewStore(100, 10, 200, 20)
	for i := 0; i < 5; i++ {
		if !s.AllowIP("10.0.0.1") {
			t.Fatalf("burst ip at %d", i)
		}
	}
	if !s.AllowRoute("/api/v1/x") {
		t.Fatal()
	}
}
