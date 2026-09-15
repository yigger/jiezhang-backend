package filestore

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type brokenReader struct{}

func (brokenReader) Read(p []byte) (int, error) { copy(p, "partial"); return 7, io.ErrUnexpectedEOF }
func TestLocalStorageAndCleanup(t *testing.T) {
	root := t.TempDir()
	store := Local{Root: root}
	path, err := store.Save(context.Background(), "private/images", "photo.PNG", strings.NewReader("image"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, path))
	if err != nil || string(data) != "image" || filepath.Ext(path) != ".png" {
		t.Fatalf("path=%s data=%s err=%v", path, data, err)
	}
	if _, err := store.Save(context.Background(), "../escape", "photo.png", strings.NewReader("x")); err == nil {
		t.Fatal("accepted traversal")
	}
	if _, err := store.Save(context.Background(), "failed", "photo.png", brokenReader{}); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatal(err)
	}
	files, err := os.ReadDir(filepath.Join(root, "failed"))
	if err != nil || len(files) != 0 {
		t.Fatalf("partial file left behind: %v %v", files, err)
	}
}
