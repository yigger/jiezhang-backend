package filestore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Local struct{ Root string }

func (s Local) Save(ctx context.Context, dir, name string, src io.Reader) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if !filepath.IsLocal(dir) {
		return "", fmt.Errorf("invalid storage directory")
	}
	targetDir := filepath.Join(s.Root, dir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		ext = ".bin"
	}
	filename := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), os.Getpid(), ext)
	target := filepath.Join(targetDir, filename)
	dst, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return "", err
	}
	_, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	if copyErr != nil {
		_ = os.Remove(target)
		return "", copyErr
	}
	if closeErr != nil {
		_ = os.Remove(target)
		return "", closeErr
	}
	return filepath.ToSlash(filepath.Join(dir, filename)), nil
}
