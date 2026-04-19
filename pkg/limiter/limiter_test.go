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

func TestStoreWithOptions_IPWindow(t *testing.T) {
	// 大窗口同一 slot 内：每 IP 仅 2 次
	s := NewStoreWithOptions(StoreOptions{
		WindowSec:         3600,
		IPMaxPerWindow:    2,
		RouteMaxPerWindow: 0,
		IPRPS:             100, IPBurst: 10,
		RouteRPS: 100, RouteBurst: 10,
	})
	if !s.AllowIP("192.0.2.1") {
		t.Fatal("first")
	}
	if !s.AllowIP("192.0.2.1") {
		t.Fatal("second")
	}
	if s.AllowIP("192.0.2.1") {
		t.Fatal("third should be denied in same window")
	}
	if !s.AllowIP("192.0.2.2") {
		t.Fatal("other ip independent")
	}
}

func TestStoreWithOptions_routeWindow_ipTB(t *testing.T) {
	s := NewStoreWithOptions(StoreOptions{
		WindowSec:         3600,
		IPMaxPerWindow:    0,
		RouteMaxPerWindow: 1,
		IPRPS:             1000, IPBurst: 1000,
		RouteRPS:          0, RouteBurst: 0,
	})
	if !s.AllowIP("192.0.2.1") {
		t.Fatal("ip tb")
	}
	if !s.AllowRoute("GET /a") {
		t.Fatal("first route")
	}
	if s.AllowRoute("GET /a") {
		t.Fatal("second same route key same window")
	}
	if !s.AllowRoute("GET /b") {
		t.Fatal("other route key")
	}
}
