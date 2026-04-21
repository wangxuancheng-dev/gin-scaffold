package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	// CipherAES256CBC 为默认可逆加密算法。
	CipherAES256CBC = "aes-256-cbc"
)

var (
	ErrInvalidPayload = errors.New("invalid encrypted payload")
	ErrInvalidMAC     = errors.New("invalid payload mac")
)

type encryptedPayload struct {
	IV    string `json:"iv"`
	Value string `json:"value"`
	MAC   string `json:"mac"`
	Tag   string `json:"tag,omitempty"`
}

// Service 提供加解密能力（base64(json{iv,value,mac,tag?})）。
type Service struct {
	key    []byte
	cipher string
}

// NewService 创建加密器。
// key 支持 "base64:<base64-encoded-32bytes>" 或明文 32 字节。
func NewService(key, cipherName string) (*Service, error) {
	raw, err := parseKey(key)
	if err != nil {
		return nil, err
	}
	c := strings.ToLower(strings.TrimSpace(cipherName))
	if c == "" {
		c = CipherAES256CBC
	}
	if c != CipherAES256CBC {
		return nil, fmt.Errorf("unsupported cipher %q", cipherName)
	}
	if len(raw) != 32 {
		return nil, fmt.Errorf("key length must be 32 bytes for %s", c)
	}
	return &Service{key: raw, cipher: c}, nil
}

// EncryptString 加密明文并返回密文。
func (s *Service) EncryptString(plain string) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return "", err
	}
	padded := pkcs7Pad([]byte(plain), aes.BlockSize)
	out := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, padded)

	ivB64 := base64.StdEncoding.EncodeToString(iv)
	valueB64 := base64.StdEncoding.EncodeToString(out)
	mac := signMAC(s.key, ivB64+valueB64)
	p := encryptedPayload{
		IV:    ivB64,
		Value: valueB64,
		MAC:   mac,
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(raw), nil
}

// DecryptString 解密密文。
func (s *Service) DecryptString(payloadB64 string) (string, error) {
	p, err := decodePayload(payloadB64)
	if err != nil {
		return "", err
	}
	expect := signMAC(s.key, p.IV+p.Value)
	if !hmac.Equal([]byte(expect), []byte(strings.ToLower(p.MAC))) {
		return "", ErrInvalidMAC
	}
	iv, err := base64.StdEncoding.DecodeString(p.IV)
	if err != nil || len(iv) != aes.BlockSize {
		return "", ErrInvalidPayload
	}
	cipherText, err := base64.StdEncoding.DecodeString(p.Value)
	if err != nil || len(cipherText) == 0 || len(cipherText)%aes.BlockSize != 0 {
		return "", ErrInvalidPayload
	}
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}
	out := make([]byte, len(cipherText))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, cipherText)
	unpadded, err := pkcs7Unpad(out, aes.BlockSize)
	if err != nil {
		return "", ErrInvalidPayload
	}
	return string(unpadded), nil
}

func parseKey(key string) ([]byte, error) {
	k := strings.TrimSpace(key)
	if k == "" {
		return nil, errors.New("empty encryption key")
	}
	if strings.HasPrefix(k, "base64:") {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(k, "base64:"))
		if err != nil {
			return nil, fmt.Errorf("invalid base64 key: %w", err)
		}
		return decoded, nil
	}
	return []byte(k), nil
}

func decodePayload(raw string) (*encryptedPayload, error) {
	blob, err := base64.StdEncoding.DecodeString(strings.TrimSpace(raw))
	if err != nil {
		return nil, ErrInvalidPayload
	}
	var p encryptedPayload
	if err := json.Unmarshal(blob, &p); err != nil {
		return nil, ErrInvalidPayload
	}
	if strings.TrimSpace(p.IV) == "" || strings.TrimSpace(p.Value) == "" || strings.TrimSpace(p.MAC) == "" {
		return nil, ErrInvalidPayload
	}
	return &p, nil
}

func signMAC(key []byte, data string) string {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write([]byte(data))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	pad := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(data, pad...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("invalid block size")
	}
	padLen := int(data[len(data)-1])
	if padLen <= 0 || padLen > blockSize || padLen > len(data) {
		return nil, errors.New("invalid padding")
	}
	for _, b := range data[len(data)-padLen:] {
		if int(b) != padLen {
			return nil, errors.New("invalid padding")
		}
	}
	return data[:len(data)-padLen], nil
}
