package handler

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"gin-scaffold/internal/service"
)

// SSEHandler SSE 演示。
type SSEHandler struct {
	svc *service.SSEService
}

// NewSSEHandler 构造。
func NewSSEHandler(s *service.SSEService) *SSEHandler {
	return &SSEHandler{svc: s}
}

// Stream 以 text/event-stream 推送事件。
// @Summary SSE 演示
// @Tags realtime
// @Router /api/v1/sse/stream [get]
func (h *SSEHandler) Stream(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		FailInternal(c, fmt.Errorf("streaming unsupported by response writer"))
		return
	}
	ch := h.svc.TickMessages(c.Request.Context(), 2*time.Second)
	for msg := range ch {
		_, err := io.WriteString(c.Writer, fmt.Sprintf("data: %s\n\n", msg))
		if err != nil {
			return
		}
		flusher.Flush()
	}
}
