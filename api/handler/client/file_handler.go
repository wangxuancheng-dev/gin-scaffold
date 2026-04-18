package clienthandler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gin-scaffold/api/handler"
	clientreq "gin-scaffold/api/request/client"
	"gin-scaffold/api/response"
	"gin-scaffold/config"
	"gin-scaffold/internal/pkg/snowflake"
	"gin-scaffold/pkg/storage"
	"gin-scaffold/pkg/strutil"
)

// FileHandler 文件上传与签名下载。
type FileHandler struct {
	cfg *config.StorageConfig
}

func NewFileHandler(cfg *config.StorageConfig) *FileHandler {
	return &FileHandler{cfg: cfg}
}

// Upload 上传文件（multipart/form-data, field=file）。
// @Summary 上传文件
// @Tags client-file
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "上传文件"
// @Success 200 {object} response.Body
// @Router /api/v1/client/files/upload [post]
func (h *FileHandler) Upload(c *gin.Context) {
	provider, err := storage.Require()
	if err != nil {
		handler.FailServiceUnavailable(c, err, "storage not configured")
		return
	}
	file, hdr, err := c.Request.FormFile("file")
	if err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	defer file.Close()
	if !h.allowExt(hdr.Filename) {
		handler.FailInvalidParam(c, fmt.Errorf("file extension not allowed"))
		return
	}
	if hdr.Size > h.maxUploadBytes() {
		handler.FailInvalidParam(c, fmt.Errorf("file size exceeds limit"))
		return
	}
	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		handler.FailInvalidParam(c, err)
		return
	}
	head = head[:n]
	detected := http.DetectContentType(head)
	if !h.allowMIME(detected) {
		handler.FailInvalidParam(c, fmt.Errorf("content type not allowed: %s", detected))
		return
	}
	body := io.MultiReader(bytes.NewReader(head), file)
	key, err := h.buildKey(hdr.Filename)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	if pc, ok := provider.(storage.PutContentTyper); ok {
		if err := pc.PutContentType(c.Request.Context(), key, detected, body); err != nil {
			handler.FailInternal(c, err)
			return
		}
	} else if err := provider.Put(c.Request.Context(), key, body); err != nil {
		handler.FailInternal(c, err)
		return
	}
	response.OK(c, gin.H{"key": key, "filename": hdr.Filename})
}

// PresignPut 生成 S3/MinIO 直传 PUT 预签名 URL（仅对象存储驱动可用）。
// @Summary 预签名直传 PUT
// @Tags client-file
// @Accept json
// @Produce json
// @Param body body clientreq.FilePresignPutRequest true "预签名参数"
// @Param expire_sec query int false "预签名有效期（秒），默认使用 storage.url_expire_sec"
// @Success 200 {object} response.Body
// @Router /api/v1/client/files/presign [post]
func (h *FileHandler) PresignPut(c *gin.Context) {
	provider, err := storage.Require()
	if err != nil {
		handler.FailServiceUnavailable(c, err, "storage not configured")
		return
	}
	pp, ok := provider.(storage.PresignPutProvider)
	if !ok {
		handler.FailInvalidParam(c, fmt.Errorf("presign upload is only supported for s3/minio driver"))
		return
	}
	var req clientreq.FilePresignPutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	if !h.allowExt(req.Filename) {
		handler.FailInvalidParam(c, fmt.Errorf("file extension not allowed"))
		return
	}
	if !h.allowMIME(req.ContentType) {
		handler.FailInvalidParam(c, fmt.Errorf("content type not allowed"))
		return
	}
	expireSec := h.defaultExpireSec()
	if v := strings.TrimSpace(c.Query("expire_sec")); v != "" {
		n, convErr := strconv.Atoi(v)
		if convErr != nil || n <= 0 {
			handler.FailInvalidParam(c, fmt.Errorf("expire_sec must be > 0"))
			return
		}
		expireSec = n
	}
	key, err := h.buildKey(req.Filename)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	var presignOpts *storage.PresignPutOptions
	if req.ContentLength != nil {
		if *req.ContentLength <= 0 {
			handler.FailInvalidParam(c, fmt.Errorf("content_length must be > 0 when set"))
			return
		}
		if *req.ContentLength > h.maxUploadBytes() {
			handler.FailInvalidParam(c, fmt.Errorf("content_length exceeds upload limit"))
			return
		}
		presignOpts = &storage.PresignPutOptions{ContentLength: *req.ContentLength}
	}
	if raw := strings.TrimSpace(req.Sha256); raw != "" {
		sum, decErr := hex.DecodeString(raw)
		if decErr != nil || len(sum) != 32 {
			handler.FailInvalidParam(c, fmt.Errorf("sha256 must be 64 hex characters"))
			return
		}
		sha := hex.EncodeToString(sum)
		if presignOpts == nil {
			presignOpts = &storage.PresignPutOptions{}
		}
		if presignOpts.Metadata == nil {
			presignOpts.Metadata = map[string]string{}
		}
		presignOpts.Metadata["sha256"] = sha
	}
	method, url, headers, err := pp.PresignPutURL(c.Request.Context(), key, req.ContentType, time.Duration(expireSec)*time.Second, presignOpts)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	out := gin.H{
		"key":          key,
		"method":       method,
		"url":          url,
		"headers":      headers,
		"expire_sec":   expireSec,
		"filename":     req.Filename,
		"content_type": strings.TrimSpace(strings.Split(req.ContentType, ";")[0]),
	}
	if req.ContentLength != nil {
		out["content_length"] = *req.ContentLength
	}
	if presignOpts != nil && presignOpts.Metadata != nil {
		if v := presignOpts.Metadata["sha256"]; v != "" {
			out["sha256"] = v
		}
	}
	response.OK(c, out)
}

// Complete 确认对象已写入（Head/Stat），可选校验大小与 SHA256。
// @Summary 确认上传完成
// @Tags client-file
// @Accept json
// @Produce json
// @Param body body clientreq.FileCompleteRequest true "完成确认参数"
// @Success 200 {object} response.Body
// @Router /api/v1/client/files/complete [post]
func (h *FileHandler) Complete(c *gin.Context) {
	provider, err := storage.Require()
	if err != nil {
		handler.FailServiceUnavailable(c, err, "storage not configured")
		return
	}
	st, ok := provider.(storage.ObjectStatProvider)
	if !ok {
		handler.FailServiceUnavailable(c, fmt.Errorf("stat unsupported"), "storage driver does not support object stat")
		return
	}
	var req clientreq.FileCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.FailInvalidParam(c, err)
		return
	}
	key := strings.TrimSpace(req.Key)
	stat, err := st.StatObject(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			handler.FailInvalidParam(c, fmt.Errorf("object not found"))
			return
		}
		handler.FailInternal(c, err)
		return
	}
	if req.ExpectedSize != nil && *req.ExpectedSize != stat.Size {
		handler.FailInvalidParam(c, fmt.Errorf("size mismatch"))
		return
	}
	if exp := strings.TrimSpace(req.ExpectedSHA256); exp != "" {
		sum, decErr := hex.DecodeString(exp)
		if decErr != nil || len(sum) != 32 {
			handler.FailInvalidParam(c, fmt.Errorf("expected_sha256 must be 64 hex characters"))
			return
		}
		want := hex.EncodeToString(sum)
		got := strings.TrimSpace(stat.Metadata["sha256"])
		if got != "" {
			if got != want {
				handler.FailInvalidParam(c, fmt.Errorf("sha256 mismatch"))
				return
			}
		} else {
			if stat.Size > h.maxUploadBytes() {
				handler.FailInvalidParam(c, fmt.Errorf("object too large to verify sha256 without sha256 metadata; presign with sha256 in request body"))
				return
			}
			rc, openErr := provider.Open(c.Request.Context(), key)
			if openErr != nil {
				handler.FailInternal(c, openErr)
				return
			}
			defer rc.Close()
			hash := sha256.New()
			lr := &io.LimitedReader{R: rc, N: h.maxUploadBytes() + 1}
			if _, copyErr := io.Copy(hash, lr); copyErr != nil {
				handler.FailInternal(c, copyErr)
				return
			}
			if lr.N <= 0 {
				handler.FailInvalidParam(c, fmt.Errorf("object too large to verify"))
				return
			}
			if hex.EncodeToString(hash.Sum(nil)) != want {
				handler.FailInvalidParam(c, fmt.Errorf("sha256 mismatch"))
				return
			}
		}
	}
	response.OK(c, gin.H{"key": key, "size": stat.Size})
}

// SignURL 生成签名下载地址。
// @Summary 生成签名下载地址
// @Tags client-file
// @Produce json
// @Param key query string true "文件 key"
// @Param expire_sec query int false "过期秒数，默认使用 storage.url_expire_sec"
// @Success 200 {object} response.Body
// @Router /api/v1/client/files/url [get]
func (h *FileHandler) SignURL(c *gin.Context) {
	provider, err := storage.Require()
	if err != nil {
		handler.FailServiceUnavailable(c, err, "storage not configured")
		return
	}
	key := strings.TrimSpace(c.Query("key"))
	if key == "" {
		handler.FailInvalidParam(c, fmt.Errorf("key is required"))
		return
	}
	expireSec := h.defaultExpireSec()
	if v := strings.TrimSpace(c.Query("expire_sec")); v != "" {
		n, convErr := strconv.Atoi(v)
		if convErr != nil || n <= 0 {
			handler.FailInvalidParam(c, fmt.Errorf("expire_sec must be > 0"))
			return
		}
		expireSec = n
	}
	expires := time.Now().Unix() + int64(expireSec)
	sig, err := provider.Sign(key, expires)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	dl := fmt.Sprintf("/api/v1/client/files/download?key=%s&e=%d&sig=%s", url.QueryEscape(key), expires, url.QueryEscape(sig))
	response.OK(c, gin.H{"url": dl, "expires": expires})
}

// Download 通过签名 URL 下载文件。
// @Summary 下载文件（签名）
// @Tags client-file
// @Produce application/octet-stream
// @Param key query string true "文件 key"
// @Param e query int true "过期时间（Unix 秒）"
// @Param sig query string true "签名"
// @Success 200 {file} file
// @Router /api/v1/client/files/download [get]
func (h *FileHandler) Download(c *gin.Context) {
	provider, err := storage.Require()
	if err != nil {
		handler.FailServiceUnavailable(c, err, "storage not configured")
		return
	}
	key := strings.TrimSpace(c.Query("key"))
	if key == "" {
		handler.FailInvalidParam(c, fmt.Errorf("key is required"))
		return
	}
	expires, err := strconv.ParseInt(strings.TrimSpace(c.Query("e")), 10, 64)
	if err != nil || expires <= 0 || time.Now().Unix() > expires {
		handler.FailUnauthorized(c, "download link expired")
		return
	}
	sig := strings.TrimSpace(c.Query("sig"))
	if !provider.Verify(key, expires, sig) {
		handler.FailUnauthorized(c, "invalid signature")
		return
	}
	rc, err := provider.Open(c.Request.Context(), key)
	if err != nil {
		handler.FailInternal(c, err)
		return
	}
	defer rc.Close()
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", strutil.AttachmentFilename(key)))
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, rc)
}

func (h *FileHandler) buildKey(filename string) (string, error) {
	id, err := snowflake.NextID()
	if err != nil {
		return "", err
	}
	ext := strings.ToLower(path.Ext(filename))
	return time.Now().Format("20060102") + "/" + strconv.FormatInt(id, 10) + ext, nil
}

func (h *FileHandler) maxUploadBytes() int64 {
	if h == nil || h.cfg == nil || h.cfg.MaxUploadMB <= 0 {
		return 10 << 20
	}
	return h.cfg.MaxUploadMB << 20
}

func (h *FileHandler) defaultExpireSec() int {
	if h == nil || h.cfg == nil || h.cfg.URLExpireSec <= 0 {
		return 300
	}
	return h.cfg.URLExpireSec
}

func (h *FileHandler) allowExt(filename string) bool {
	if h == nil || h.cfg == nil || strings.TrimSpace(h.cfg.AllowedExt) == "" {
		return true
	}
	ext := strings.ToLower(path.Ext(filename))
	for _, part := range strings.Split(h.cfg.AllowedExt, ",") {
		if ext == strings.ToLower(strings.TrimSpace(part)) {
			return true
		}
	}
	return false
}

func (h *FileHandler) allowMIME(detected string) bool {
	if h == nil || h.cfg == nil || strings.TrimSpace(h.cfg.AllowedMIME) == "" {
		return true
	}
	detected = strings.TrimSpace(strings.Split(detected, ";")[0])
	detected = strings.ToLower(detected)
	for _, part := range strings.Split(h.cfg.AllowedMIME, ",") {
		want := strings.ToLower(strings.TrimSpace(strings.Split(part, ";")[0]))
		if want == "" {
			continue
		}
		if detected == want {
			return true
		}
	}
	return false
}
