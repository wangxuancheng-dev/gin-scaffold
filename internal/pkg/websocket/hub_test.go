package websocket

import "testing"

func TestNewHub_InitiallyEmpty(t *testing.T) {
	h := NewHub()
	if h.OnlineCount() != 0 {
		t.Fatalf("count=%d", h.OnlineCount())
	}
	h.Broadcast([]byte("noop"))
}
