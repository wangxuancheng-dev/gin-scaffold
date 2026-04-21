package clienthandler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	clientreq "gin-scaffold/internal/api/request/client"
	"gin-scaffold/internal/config"
	"gin-scaffold/pkg/storage"
)

type memFileProvider struct {
	putErr     error
	openData   string
	signErr    error
	verifyOK   bool
	stat       *storage.ObjectStat
	statErr    error
	presignErr error
}

func (m *memFileProvider) Put(ctx context.Context, key string, r io.Reader) error {
	_, _ = io.Copy(io.Discard, r)
	return m.putErr
}

func (m *memFileProvider) PutContentType(ctx context.Context, key, contentType string, r io.Reader) error {
	_, _ = io.Copy(io.Discard, r)
	return m.putErr
}

func (m *memFileProvider) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(m.openData)), nil
}

func (m *memFileProvider) Delete(ctx context.Context, key string) error { return nil }

func (m *memFileProvider) Sign(key string, expireSec int64) (string, error) {
	if m.signErr != nil {
		return "", m.signErr
	}
	return "goodsig", nil
}

func (m *memFileProvider) Verify(key string, expires int64, sig string) bool {
	return m.verifyOK && sig == "goodsig"
}

func (m *memFileProvider) StatObject(ctx context.Context, key string) (*storage.ObjectStat, error) {
	return m.stat, m.statErr
}

func (m *memFileProvider) PresignPutURL(ctx context.Context, key, contentType string, expire time.Duration, opts *storage.PresignPutOptions) (string, string, map[string]string, error) {
	if m.presignErr != nil {
		return "", "", nil, m.presignErr
	}
	return "PUT", "https://example.test/upload", map[string]string{"X-Amz-Acl": "private"}, nil
}

// putOnlyProvider implements storage.Provider only (no presign / stat).
type putOnlyProvider struct {
	openData string
}

func (p *putOnlyProvider) Put(ctx context.Context, key string, r io.Reader) error {
	_, _ = io.Copy(io.Discard, r)
	return nil
}

func (p *putOnlyProvider) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(p.openData)), nil
}

func (p *putOnlyProvider) Delete(ctx context.Context, key string) error { return nil }

func (p *putOnlyProvider) Sign(key string, expireSec int64) (string, error) { return "goodsig", nil }

func (p *putOnlyProvider) Verify(key string, expires int64, sig string) bool {
	return sig == "goodsig"
}

func jpegBody() []byte {
	b := []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}
	out := make([]byte, 64)
	copy(out, b)
	return out
}

func TestFileHandler_allowExt_allowMIME(t *testing.T) {
	h := &FileHandler{cfg: &config.StorageConfig{AllowedExt: ".jpg", AllowedMIME: "image/jpeg"}}
	if !h.allowExt("a.jpg") {
		t.Fatal("jpg ext")
	}
	if h.allowExt("a.png") {
		t.Fatal("png blocked")
	}
	if !h.allowMIME("image/jpeg") {
		t.Fatal("jpeg mime")
	}
	if h.allowMIME("image/png") {
		t.Fatal("png mime blocked")
	}
}

func TestFileHandler_maxUploadBytes_default(t *testing.T) {
	var h *FileHandler
	if h.maxUploadBytes() != 10<<20 {
		t.Fatalf("default max=%d", h.maxUploadBytes())
	}
	h = &FileHandler{cfg: &config.StorageConfig{MaxUploadMB: 2}}
	if h.maxUploadBytes() != 2<<20 {
		t.Fatalf("max=%d", h.maxUploadBytes())
	}
}

func TestFileHandler_SignURL_and_Download(t *testing.T) {
	t.Cleanup(func() { storage.InitDefault(nil) })
	storage.InitDefault(&memFileProvider{verifyOK: true, openData: "hello"})

	cfg := &config.StorageConfig{URLExpireSec: 600}
	h := NewFileHandler(cfg)

	t.Run("sign_missing_key", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/url", nil)
		h.SignURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("code=%d", w.Code)
		}
	})

	t.Run("sign_ok", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/url?key=k1", nil)
		h.SignURL(c)
		if w.Code != http.StatusOK {
			t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("download_expired", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/dl?key=k&e=1&sig=x", nil)
		h.Download(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("code=%d", w.Code)
		}
	})

	t.Run("download_ok", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		exp := time.Now().Unix() + 3600
		c.Request = httptest.NewRequest(http.MethodGet, "http://localhost/dl?key=k1&sig=goodsig&e="+strconv.FormatInt(exp, 10), nil)
		h.Download(c)
		if w.Code != http.StatusOK {
			t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "hello") {
			t.Fatalf("body=%q", w.Body.String())
		}
	})
}

func TestFileHandler_Upload_noStorage(t *testing.T) {
	t.Cleanup(func() { storage.InitDefault(nil) })
	h := NewFileHandler(&config.StorageConfig{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "http://localhost/upload", nil)
	h.Upload(c)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestFileHandler_Upload_ok(t *testing.T) {
	t.Cleanup(func() { storage.InitDefault(nil) })
	storage.InitDefault(&memFileProvider{})

	cfg := &config.StorageConfig{AllowedExt: ".jpg", AllowedMIME: "image/jpeg", MaxUploadMB: 1}
	h := NewFileHandler(cfg)

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", "photo.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(jpegBody()); err != nil {
		t.Fatal(err)
	}
	if err := mw.Close(); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	c.Request = req
	h.Upload(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestFileHandler_PresignPut_notSupported(t *testing.T) {
	t.Cleanup(func() { storage.InitDefault(nil) })
	storage.InitDefault(&putOnlyProvider{})

	h := NewFileHandler(&config.StorageConfig{AllowedExt: ".jpg", AllowedMIME: "image/jpeg"})
	body, _ := json.Marshal(clientreq.FilePresignPutRequest{Filename: "a.jpg", ContentType: "image/jpeg"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/presign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	h.PresignPut(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestFileHandler_PresignPut_badExpireQuery(t *testing.T) {
	t.Cleanup(func() { storage.InitDefault(nil) })
	storage.InitDefault(&memFileProvider{})

	h := NewFileHandler(&config.StorageConfig{AllowedExt: ".jpg", AllowedMIME: "image/jpeg"})
	body, _ := json.Marshal(clientreq.FilePresignPutRequest{Filename: "a.jpg", ContentType: "image/jpeg"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/presign?expire_sec=0", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	h.PresignPut(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code=%d", w.Code)
	}
}

func TestFileHandler_PresignPut_ok(t *testing.T) {
	t.Cleanup(func() { storage.InitDefault(nil) })
	storage.InitDefault(&memFileProvider{})

	h := NewFileHandler(&config.StorageConfig{AllowedExt: ".jpg", AllowedMIME: "image/jpeg"})
	body, _ := json.Marshal(clientreq.FilePresignPutRequest{Filename: "a.jpg", ContentType: "image/jpeg"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/presign", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	h.PresignPut(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestFileHandler_Complete_statUnsupported(t *testing.T) {
	t.Cleanup(func() { storage.InitDefault(nil) })
	storage.InitDefault(&putOnlyProvider{})

	h := NewFileHandler(&config.StorageConfig{})
	body, _ := json.Marshal(clientreq.FileCompleteRequest{Key: "k"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/complete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	h.Complete(c)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestFileHandler_Complete_ok(t *testing.T) {
	t.Cleanup(func() { storage.InitDefault(nil) })
	storage.InitDefault(&memFileProvider{
		stat: &storage.ObjectStat{Size: 42, Metadata: map[string]string{}},
	})

	h := NewFileHandler(&config.StorageConfig{})
	body, _ := json.Marshal(clientreq.FileCompleteRequest{Key: "some-key"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/complete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	h.Complete(c)
	if w.Code != http.StatusOK {
		t.Fatalf("code=%d body=%s", w.Code, w.Body.String())
	}
}
