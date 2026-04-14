// Package websocket 提供简易连接中心：注册、广播、按用户单播。
package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub 管理所有活跃连接。
type Hub struct {
	mu    sync.RWMutex
	conns map[int64]*websocket.Conn // userID -> conn（演示：单端在线）
	all   []*websocket.Conn
}

// NewHub 创建 Hub。
func NewHub() *Hub {
	return &Hub{conns: make(map[int64]*websocket.Conn)}
}

// Register 注册用户连接（覆盖旧连接）。
func (h *Hub) Register(uid int64, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[uid] = c
	h.all = append(h.all, c)
}

// Unregister 移除连接。
func (h *Hub) Unregister(uid int64, c *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns, uid)
	for i, v := range h.all {
		if v == c {
			h.all = append(h.all[:i], h.all[i+1:]...)
			break
		}
	}
	_ = c.Close()
}

// Broadcast 向所有连接发送文本帧（忽略错误）。
func (h *Hub) Broadcast(payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.all {
		if c != nil {
			_ = c.WriteMessage(websocket.TextMessage, payload)
		}
	}
}

// SendToUser 单播。
func (h *Hub) SendToUser(uid int64, payload []byte) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	c, ok := h.conns[uid]
	if !ok || c == nil {
		return false
	}
	_ = c.WriteMessage(websocket.TextMessage, payload)
	return true
}

// OnlineCount 在线连接数。
func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.all)
}
