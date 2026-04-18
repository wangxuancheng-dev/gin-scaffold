package httpclient

import "testing"

func TestDefault_WithoutInit(t *testing.T) {
	c := Default()
	if c == nil {
		t.Fatal()
	}
}
