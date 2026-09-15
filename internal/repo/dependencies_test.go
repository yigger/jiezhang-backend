package repo_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestLayerDependencies(t *testing.T) {
	const module = "github.com/yigger/jiezhang-backend/internal/"
	err := filepath.WalkDir("..", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relative, err := filepath.Rel("..", path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		var forbidden []string
		switch {
		case strings.HasPrefix(relative, "controller/"), strings.HasPrefix(relative, "middleware/"), strings.HasPrefix(relative, "mcp/"):
			forbidden = []string{module + "repo", "gorm.io/", "database/sql"}
		case strings.HasPrefix(relative, "service/"):
			forbidden = []string{module + "controller", module + "router", module + "middleware", module + "infrastructure/", module + "config", module + "repo/mysql", "github.com/gin-gonic/", "gorm.io/", "database/sql"}
		case strings.HasPrefix(relative, "repo/mysql/"):
			forbidden = []string{module + "service/", module + "controller", module + "types", "github.com/gin-gonic/"}
		case strings.HasPrefix(relative, "repo/"):
			forbidden = []string{module + "service/", module + "controller", module + "types", module + "repo/mysql", module + "infrastructure/", "github.com/gin-gonic/", "gorm.io/", "database/sql"}
		case strings.HasPrefix(relative, "model/"):
			forbidden = []string{module, "gorm.io/", "github.com/gin-gonic/"}
		case strings.HasPrefix(relative, "types/"):
			forbidden = []string{module + "repo", module + "service/", module + "controller", module + "infrastructure/", "gorm.io/", "github.com/gin-gonic/"}
		}
		for _, imp := range file.Imports {
			dep, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			for _, prefix := range forbidden {
				if strings.HasPrefix(dep, prefix) {
					t.Errorf("%s imports forbidden dependency %s", relative, dep)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
