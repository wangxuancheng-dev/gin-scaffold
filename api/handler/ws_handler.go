package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"gin-scaffold/internal/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// WSHandler WebSocket 入口。
type WSHandler struct {
	svc *service.WSService
}

// NewWSHandler 构造。
func NewWSHandler(s *service.WSService) *WSHandler {
	return &WSHandler{svc: s}
}

// Handle 使用 gorilla/websocket 升级并处理心跳和回显。
// @Summary WebSocket 演示
// @Tags realtime
// @Param uid query int true "用户ID"
// @Router /api/v1/ws [get]
func (h *WSHandler) Handle(c *gin.Context) {
	uid, err := strconv.ParseInt(c.Query("uid"), 10, 64)
	if err != nil || uid <= 0 {
		c.Status(http.StatusBadRequest)
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	hub := h.svc.Hub()
	hub.Register(uid, conn)
	defer hub.Unregister(uid, conn)

	conn.SetReadLimit(1024 * 8)
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			msgType, payload, readErr := conn.ReadMessage()
			if readErr != nil {
				return
			}
			if msgType == websocket.TextMessage || msgType == websocket.BinaryMessage {
				_ = conn.WriteMessage(msgType, payload)
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err = conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}
}
