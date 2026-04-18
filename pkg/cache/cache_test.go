package cache

import (
	"strings"
	"testing"
)

func TestClient_Key(t *testing.T) {
	c := &Client{prefix: "unit:"}
	if got := c.Key("users", "1"); got != "unit:users:1" {
		t.Fatalf("got %q", got)
	}
}

func TestNewFromConfig_NilGlobalUsesDefaultPrefix(t *testing.T) {
	c := NewFromConfig()
	if c == nil {
		t.Fatal()
	}
	if !strings.HasPrefix(c.Key("x"), "app:") {
		t.Fatalf("prefix got %q", c.Key("x"))
	}
}
