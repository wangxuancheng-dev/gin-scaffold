package clientreq

// FilePresignPutRequest 预签名直传 PUT 请求体。
type FilePresignPutRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
}
