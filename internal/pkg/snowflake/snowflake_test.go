package snowflake

import "testing"

func TestNextID_Monotonic(t *testing.T) {
	if err := Init(1); err != nil {
		t.Fatalf("init: %v", err)
	}
	a, err := NextID()
	if err != nil {
		t.Fatal(err)
	}
	b, err := NextID()
	if err != nil {
		t.Fatal(err)
	}
	if b <= a {
		t.Fatalf("expected increasing ids, got %d then %d", a, b)
	}
}
