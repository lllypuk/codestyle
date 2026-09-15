package codestyle

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// ErrNoModule — go.mod не найден ни в каталоге, ни выше по дереву.
var ErrNoModule = errors.New("go.mod не найден выше по дереву")

// Imports — импорты каждого пакета модуля по пути от корня; тесты не считаются,
// им подмена адаптером и нужна. Пакет без импортов присутствует с nil. Импорты
// полные, собственные отрезает [Own]: адаптер из чужого модуля иначе выпал бы молча.
func Imports(root string) (map[string][]string, error) {
	found := map[string][]string{}

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if path != root && slices.Contains(DefaultSkipDirs(), entry.Name()) {
				return fs.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, parseErr := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return parseErr
		}

		rel, _ := filepath.Rel(root, filepath.Dir(path))
		rel = filepath.ToSlash(rel)

		for _, imp := range file.Imports {
			found[rel] = append(found[rel], strings.Trim(imp.Path.Value, `"`))
		}

		if _, seen := found[rel]; !seen {
			found[rel] = nil
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("обход %s: %w", root, err)
	}

	return found, nil
}

// Own оставляет импорты внутри модуля и отрезает его префикс: `module/internal/x` → `internal/x`.
func Own(module string, imports []string) []string {
	prefix := strings.TrimSuffix(module, "/") + "/"

	var own []string

	for _, imp := range imports {
		if after, ok := strings.CutPrefix(imp, prefix); ok {
			own = append(own, after)
		}
	}

	return own
}

// ModuleRoot ищет go.mod от start вверх: тест запускают и из корня, и из пакета.
func ModuleRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("абсолютный путь %s: %w", start, err)
	}

	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNoModule
		}

		dir = parent
	}
}
