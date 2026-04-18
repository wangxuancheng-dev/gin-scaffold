package middleware

import (
	"net/http/httptest"
	"testing"
)

func TestWebSocketCheckOrigin_Wildcard(t *testing.T) {
	chk := WebSocketCheckOrigin([]string{"*"})
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Origin", "https://evil.example")
	if !chk(r) {
		t.Fatal()
	}
}

func TestWebSocketCheckOrigin_List(t *testing.T) {
	chk := WebSocketCheckOrigin([]string{"https://app.example.com"})
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Origin", "https://app.example.com")
	if !chk(r) {
		t.Fatal()
	}
	r2 := httptest.NewRequest("GET", "/", nil)
	r2.Header.Set("Origin", "https://other.example")
	if chk(r2) {
		t.Fatal()
	}
	r3 := httptest.NewRequest("GET", "/", nil)
	if !chk(r3) {
		t.Fatal("no Origin should be allowed for native clients")
	}
}
