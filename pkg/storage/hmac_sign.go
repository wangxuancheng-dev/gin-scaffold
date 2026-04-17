package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func signDownload(secret []byte, key string, expires int64) (string, error) {
	if expires <= 0 {
		return "", fmt.Errorf("storage: expires must be > 0")
	}
	if len(secret) == 0 {
		return "", fmt.Errorf("storage: sign secret is empty")
	}
	normalized := normalizeKey(key)
	payload := fmt.Sprintf("%s:%d", normalized, expires)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func verifyDownload(secret []byte, key string, expires int64, sig string) bool {
	if expires <= 0 || sig == "" {
		return false
	}
	expected, err := signDownload(secret, key, expires)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(expected), []byte(sig))
}
