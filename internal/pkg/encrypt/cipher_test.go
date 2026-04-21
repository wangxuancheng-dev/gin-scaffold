package encrypt

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestService_EncryptDecryptString(t *testing.T) {
	key := "base64:" + base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
	svc, err := NewService(key, CipherAES256CBC)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	src := "hello-compatible"
	token, err := svc.EncryptString(src)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	got, err := svc.DecryptString(token)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if got != src {
		t.Fatalf("decrypt mismatch: got=%q want=%q", got, src)
	}
}

func TestService_InvalidMAC(t *testing.T) {
	key := "12345678901234567890123456789012"
	svc, err := NewService(key, "")
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	token, err := svc.EncryptString("abc")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	broken := token[:len(token)-1] + "A"
	_, err = svc.DecryptString(broken)
	if err == nil {
		t.Fatal("expect error for broken payload")
	}
}

func TestNewService_InvalidKey(t *testing.T) {
	_, err := NewService("short-key", "")
	if err == nil || !strings.Contains(err.Error(), "key length") {
		t.Fatalf("expect key length error, got: %v", err)
	}
}
