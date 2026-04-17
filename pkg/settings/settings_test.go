package settings

import "testing"

func TestParseBool(t *testing.T) {
	t.Parallel()
	trueCases := []string{"1", "true", "YES", "On"}
	for _, c := range trueCases {
		v, err := parseBool(c)
		if err != nil || !v {
			t.Fatalf("expected true for %q, got v=%v err=%v", c, v, err)
		}
	}
	falseCases := []string{"0", "false", "NO", "off"}
	for _, c := range falseCases {
		v, err := parseBool(c)
		if err != nil || v {
			t.Fatalf("expected false for %q, got v=%v err=%v", c, v, err)
		}
	}
	if _, err := parseBool("maybe"); err == nil {
		t.Fatalf("expected error for invalid bool")
	}
}
