package notify

import (
	"context"
	"errors"
)

// Chain 将通知并行投递到多个 Notifier（任一失败会 errors.Join 汇总）。
type Chain []Notifier

func (c Chain) Notify(ctx context.Context, msg Message) error {
	var joined error
	for _, n := range c {
		if n == nil {
			continue
		}
		if err := n.Notify(ctx, msg); err != nil {
			joined = errors.Join(joined, err)
		}
	}
	return joined
}
