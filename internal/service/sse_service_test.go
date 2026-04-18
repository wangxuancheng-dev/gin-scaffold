package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSSEService_TickMessages_cancelBeforeFirstTick(t *testing.T) {
	s := NewSSEService()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ch := s.TickMessages(ctx, 50*time.Millisecond)
	_, ok := <-ch
	require.False(t, ok)
}

func TestSSEService_TickMessages_receivesEvent(t *testing.T) {
	s := NewSSEService()
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	ch := s.TickMessages(ctx, 20*time.Millisecond)
	msg, ok := <-ch
	require.True(t, ok)
	require.Contains(t, msg, "event ")
}
