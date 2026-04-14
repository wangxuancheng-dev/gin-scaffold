package service

import (
	"encoding/json"

	websocketpkg "gin-scaffold/internal/pkg/websocket"
)

// WSService WebSocket 业务封装。
type WSService struct {
	hub *websocketpkg.Hub
}

// NewWSService 构造。
func NewWSService(h *websocketpkg.Hub) *WSService {
	return &WSService{hub: h}
}

// Hub 返回底层连接中心。
func (s *WSService) Hub() *websocketpkg.Hub {
	return s.hub
}

// BroadcastJSON 广播 JSON 消息。
func (s *WSService) BroadcastJSON(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	s.hub.Broadcast(b)
	return nil
}
