package jwt

import (
	"testing"

	"gin-scaffold/config"
)

func TestManager_IssueParseAccess(t *testing.T) {
	m := NewManager(&config.JWTConfig{
		Secret:           "unit-jwt-secret-32chars-minimum!!",
		AccessExpireMin:  15,
		RefreshExpireMin: 60,
		Issuer:           "unit",
	})
	tok, err := m.IssueAccess(42, "admin", "t1")
	if err != nil {
		t.Fatal(err)
	}
	cl, err := m.ParseAccess(tok)
	if err != nil {
		t.Fatal(err)
	}
	if cl.UserID != 42 || cl.Role != "admin" || cl.TenantID != "t1" {
		t.Fatalf("%+v", cl)
	}
}

func TestManager_IssueParseRefresh(t *testing.T) {
	m := NewManager(&config.JWTConfig{
		Secret:           "unit-jwt-secret-32chars-minimum!!",
		AccessExpireMin:  15,
		RefreshExpireMin: 60,
		Issuer:           "unit",
	})
	rt, err := m.IssueRefresh(7)
	if err != nil {
		t.Fatal(err)
	}
	uid, jti, exp, err := m.ParseRefresh(rt)
	if err != nil {
		t.Fatal(err)
	}
	if uid != 7 || jti == "" || exp.IsZero() {
		t.Fatalf("uid=%d jti=%q exp=%v", uid, jti, exp)
	}
}
