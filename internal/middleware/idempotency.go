package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	goredis "github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"gin-scaffold/internal/api/response"
	"gin-scaffold/internal/config"
	"gin-scaffold/internal/pkg/errcode"
	"gin-scaffold/pkg/logger"
	"gin-scaffold/pkg/redis"
)

const headerIdempotencyKey = "X-Idempotency-Key"

type idemCached struct {
	Status      int    `json:"status"`
	ContentType string `json:"content_type"`
	Body        []byte `json:"body"`
}

type bufferedWriter struct {
	gin.ResponseWriter
	buf    bytes.Buffer
	status int
}

func (w *bufferedWriter) WriteHeader(code int) {
	w.status = code
}

func (w *bufferedWriter) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}

func (w *bufferedWriter) statusCode() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *bufferedWriter) flush() {
	code := w.statusCode()
	w.ResponseWriter.WriteHeader(code)
	if w.buf.Len() > 0 {
		_, _ = w.ResponseWriter.Write(w.buf.Bytes())
	}
}

// Idempotency 对带 X-Idempotency-Key 的 JSON POST 做 Redis 幂等（需 platform.idempotency.enabled=true）。
func Idempotency() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get()
		if cfg == nil || !cfg.Platform.Idempotency.Enabled {
			c.Next()
			return
		}
		idem := cfg.Platform.Idempotency
		if c.Request.Method != http.MethodPost {
			c.Next()
			return
		}
		idemKey := strings.TrimSpace(c.GetHeader(headerIdempotencyKey))
		if idemKey == "" {
			c.Next()
			return
		}
		if strings.Contains(strings.ToLower(c.GetHeader("Content-Type")), "multipart/") {
			c.Next()
			return
		}
		path := c.Request.URL.Path
		if !strings.HasPrefix(path, "/api/v1/") {
			c.Next()
			return
		}
		if c.Request.ContentLength < 0 || c.Request.ContentLength > idem.MaxBodyBytes {
			c.Next()
			return
		}
		body, err := io.ReadAll(io.LimitReader(c.Request.Body, idem.MaxBodyBytes+1))
		if err != nil {
			response.FailHTTP(c, http.StatusBadRequest, errcode.BadRequest, errcode.KeyInvalidParam, err.Error())
			c.Abort()
			return
		}
		if int64(len(body)) > idem.MaxBodyBytes {
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(body))

		prefix := cacheKeyPrefix(cfg)
		fp := idempotencyFingerprint(actorKey(c), idemKey, path, body)
		respKey := prefix + "idem:v1:" + fp
		lockKey := prefix + "idem:lock:v1:" + fp

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		raw, err := redis.Get(ctx, respKey)
		if err == nil && raw != "" {
			var cached idemCached
			if json.Unmarshal([]byte(raw), &cached) == nil && len(cached.Body) > 0 {
				c.Data(cached.Status, cached.ContentType, cached.Body)
				c.Abort()
				return
			}
		}
		if err != nil && err != goredis.Nil {
			logger.L().Warn("idempotency redis get", zap.Error(err))
			c.Next()
			return
		}

		lockTTL := time.Duration(idem.LockSeconds) * time.Second
		ok, err := redis.SetNX(ctx, lockKey, "1", lockTTL)
		if err != nil {
			logger.L().Warn("idempotency redis lock", zap.Error(err))
			c.Next()
			return
		}
		if !ok {
			response.FailHTTP(c, http.StatusConflict, errcode.Conflict, errcode.KeyIdempotencyConflict, "duplicate request in flight")
			c.Abort()
			return
		}

		orig := c.Writer
		bw := &bufferedWriter{ResponseWriter: orig}
		c.Writer = bw
		defer func() { c.Writer = orig }()

		c.Next()

		code := bw.statusCode()
		ct := orig.Header().Get("Content-Type")
		if code >= 200 && code < 300 &&
			strings.Contains(strings.ToLower(ct), "application/json") &&
			int64(bw.buf.Len()) <= idem.MaxCachedResponseBytes &&
			bw.buf.Len() > 0 {
			payload, mErr := json.Marshal(&idemCached{Status: code, ContentType: ct, Body: bw.buf.Bytes()})
			if mErr == nil {
				ttl := time.Duration(idem.TTLSeconds) * time.Second
				if setErr := redis.Set(ctx, respKey, string(payload), ttl); setErr != nil {
					logger.L().Warn("idempotency cache set", zap.Error(setErr))
				}
			}
		}
		if delErr := redis.Del(context.Background(), lockKey); delErr != nil {
			logger.L().Warn("idempotency unlock", zap.Error(delErr))
		}
		bw.flush()
	}
}

func cacheKeyPrefix(cfg *config.App) string {
	p := strings.TrimSpace(cfg.Platform.Cache.KeyPrefix)
	if p == "" {
		p = "app:"
	}
	if !strings.HasSuffix(p, ":") {
		p += ":"
	}
	return p
}

func actorKey(c *gin.Context) string {
	if cl, ok := Claims(c); ok && cl != nil {
		return "u:" + strconv.FormatInt(cl.UserID, 10)
	}
	return "anon"
}

func idempotencyFingerprint(actor, idemKey, path string, body []byte) string {
	h := sha256.New()
	h.Write([]byte(actor))
	h.Write([]byte{0})
	h.Write([]byte(idemKey))
	h.Write([]byte{0})
	h.Write([]byte(path))
	h.Write([]byte{0})
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
