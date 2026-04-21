package i18nres

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// Files embeds i18n json resources into the binary.
//
//go:embed *.json
var Files embed.FS

var (
	extractOnce sync.Once
	extractDir  string
	extractErr  error
)

// ExtractToTempDir extracts embedded i18n json files to a temp directory and returns that path.
func ExtractToTempDir() (string, error) {
	extractOnce.Do(func() {
		dir, err := os.MkdirTemp("", "gin-scaffold-i18n-*")
		if err != nil {
			extractErr = err
			return
		}
		entries, err := fs.Glob(Files, "*.json")
		if err != nil {
			extractErr = err
			return
		}
		for _, name := range entries {
			b, err := Files.ReadFile(name)
			if err != nil {
				extractErr = err
				return
			}
			target := filepath.Join(dir, filepath.Base(name))
			if err := os.WriteFile(target, b, 0o644); err != nil {
				extractErr = err
				return
			}
		}
		extractDir = dir
	})
	return extractDir, extractErr
}
