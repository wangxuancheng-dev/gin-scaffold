package clientreq

// FileCompleteRequest 直传或普通上传后的对象确认。
type FileCompleteRequest struct {
	Key            string `json:"key" binding:"required"`
	ExpectedSize   *int64 `json:"expected_size,omitempty"`   // 可选：与对象 Content-Length 一致
	ExpectedSHA256 string `json:"expected_sha256,omitempty"` // 可选：64 位十六进制；S3 优先比对元数据，否则在大小上限内流式计算
}
