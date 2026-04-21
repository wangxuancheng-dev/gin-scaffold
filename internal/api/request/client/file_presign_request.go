package clientreq

// FilePresignPutRequest 预签名直传 PUT 请求体。
type FilePresignPutRequest struct {
	Filename      string `json:"filename" binding:"required"`
	ContentType   string `json:"content_type" binding:"required"`
	ContentLength *int64 `json:"content_length,omitempty"` // 可选：与 PUT 的 Content-Length 一致（需 <= max_upload_mb）
	Sha256        string `json:"sha256,omitempty"`         // 可选：64 位十六进制，写入 x-amz-meta-sha256 便于 /files/complete 校验
}
