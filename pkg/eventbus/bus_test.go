package eventbus

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestBusEmitOrder(t *testing.T) {
	t.Parallel()
	var n int32
	b := New()
	b.On("e1", func(ctx context.Context, e Event) {
		_ = ctx
		if e.Name != "e1" {
			t.Fatalf("name: %s", e.Name)
		}
		atomic.AddInt32(&n, 1)
	})
	b.On("e1", func(ctx context.Context, e Event) {
		_ = ctx
		if atomic.LoadInt32(&n) != 1 {
			t.Fatalf("expected first handler run first")
		}
		atomic.AddInt32(&n, 10)
	})
	b.Emit(context.Background(), Event{Name: "e1", Payload: nil})
	if atomic.LoadInt32(&n) != 11 {
		t.Fatalf("counter: %d", n)
	}
}
