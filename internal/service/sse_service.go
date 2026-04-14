package service

import (
	"context"
	"fmt"
	"time"
)

// SSEService 服务端推送演示逻辑。
type SSEService struct{}

// NewSSEService 构造。
func NewSSEService() *SSEService {
	return &SSEService{}
}

// TickMessages 按间隔生成文本事件，直到上下文取消。
func (s *SSEService) TickMessages(ctx context.Context, interval time.Duration) <-chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		t := time.NewTicker(interval)
		defer t.Stop()
		i := 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				i++
				select {
				case ch <- fmt.Sprintf("event %d @ %s", i, time.Now().Format(time.RFC3339)):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ch
}
