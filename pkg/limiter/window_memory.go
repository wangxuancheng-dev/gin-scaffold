package limiter

import "time"

type slotCount struct {
	slot  int64
	count int
}

func (s *Store) currentSlot() int64 {
	if s.windowSec <= 0 {
		return time.Now().Unix()
	}
	return time.Now().Unix() / int64(s.windowSec)
}

// allowWindowLocked 在已持锁情况下，对 key 做固定窗口计数；max 为每窗口允许次数（含第 max 次）。
func (s *Store) allowWindowLocked(m map[string]*slotCount, key string, max int) bool {
	if max < 1 {
		max = 1
	}
	slot := s.currentSlot()
	sc := m[key]
	if sc == nil {
		sc = &slotCount{slot: slot, count: 0}
		m[key] = sc
	}
	if sc.slot != slot {
		sc.slot = slot
		sc.count = 0
	}
	if sc.count >= max {
		return false
	}
	sc.count++
	return true
}
