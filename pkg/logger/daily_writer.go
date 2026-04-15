package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var nowFunc = time.Now

type dailyRotateWriter struct {
	basePath        string
	maxAgeDays      int
	loc             *time.Location
	now             func() time.Time
	mu              sync.Mutex
	file            *os.File
	currentDate     string
	lastCleanupDate string
}

func newDailyRotateWriter(basePath string, maxAgeDays int, loc *time.Location) *dailyRotateWriter {
	if loc == nil {
		loc = time.Local
	}
	return &dailyRotateWriter{
		basePath:   basePath,
		maxAgeDays: maxAgeDays,
		loc:        loc,
		now:        nowFunc,
	}
}

func (w *dailyRotateWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.rotateIfNeededLocked(); err != nil {
		return 0, err
	}
	return w.file.Write(p)
}

func (w *dailyRotateWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	return w.file.Sync()
}

func (w *dailyRotateWriter) rotateIfNeededLocked() error {
	now := w.now().In(w.loc)
	date := now.Format("2006-01-02")
	if w.file != nil && date == w.currentDate {
		return nil
	}
	if w.file != nil {
		_ = w.file.Close()
		w.file = nil
	}
	targetPath := datedFilePath(w.basePath, date)
	f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	w.file = f
	w.currentDate = date

	if w.maxAgeDays > 0 && w.lastCleanupDate != date {
		w.cleanupOldFilesLocked(now)
		w.lastCleanupDate = date
	}
	return nil
}

func (w *dailyRotateWriter) cleanupOldFilesLocked(now time.Time) {
	dir := filepath.Dir(w.basePath)
	base := filepath.Base(w.basePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	prefix := name + "-"

	cutoff := dateOnly(now.AddDate(0, 0, -w.maxAgeDays))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		fn := e.Name()
		if !strings.HasPrefix(fn, prefix) || !strings.HasSuffix(fn, ext) {
			continue
		}
		datePart := strings.TrimSuffix(strings.TrimPrefix(fn, prefix), ext)
		t, err := time.ParseInLocation("2006-01-02", datePart, w.loc)
		if err != nil {
			continue
		}
		if dateOnly(t).Before(cutoff) {
			_ = os.Remove(filepath.Join(dir, fn))
		}
	}
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func datedFilePath(basePath, date string) string {
	dir := filepath.Dir(basePath)
	base := filepath.Base(basePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return filepath.Join(dir, fmt.Sprintf("%s-%s%s", name, date, ext))
}
