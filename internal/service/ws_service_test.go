package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	websocketpkg "gin-scaffold/internal/pkg/websocket"
)

func TestNewWSService_Hub(t *testing.T) {
	h := websocketpkg.NewHub()
	s := NewWSService(h)
	require.Equal(t, h, s.Hub())
}

func TestWSService_BroadcastJSON(t *testing.T) {
	h := websocketpkg.NewHub()
	s := NewWSService(h)
	require.NoError(t, s.BroadcastJSON(map[string]int{"a": 1}))
}

func TestWSService_BroadcastJSON_marshalError(t *testing.T) {
	h := websocketpkg.NewHub()
	s := NewWSService(h)
	ch := make(chan int)
	require.Error(t, s.BroadcastJSON(ch))
}
