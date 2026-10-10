package codestyle_test

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"github.com/lllypuk/codestyle"
)

// TestImports — импорты полные, тесты не считаются, пакет без импортов присутствует.
func TestImports(t *testing.T) {
	t.Parallel()

	got, err := codestyle.Imports(fixture)
	if err != nil {
		t.Fatal(err)
	}

	if want := []string{"example.com/mod/internal/store/postgres", "fmt"}; !slices.Equal(got["internal/app"], want) {
		t.Errorf("internal/app импортирует %v, ожидалось %v", got["internal/app"], want)
	}

	if imports, ok := got["internal/store/postgres"]; !ok || imports != nil {
		t.Errorf("адаптер без импортов: присутствует %v, импорты %v", ok, imports)
	}

	own := codestyle.Own("example.com/mod", got["cmd/server"])
	if want := []string{"internal/app", "internal/store/postgres"}; !slices.Equal(own, want) {
		t.Errorf("собственные импорты cmd/server %v, ожидалось %v", own, want)
	}

	if !slices.Contains(got["cmd/server"], "github.com/lllypuk/other/postgres") {
		t.Error("внешний адаптер выпал из списка — правило слоёв его не увидит")
	}
}

// TestModuleRoot — go.mod находится и из корня, и из вложенного каталога; без него — ErrNoModule.
func TestModuleRoot(t *testing.T) {
	t.Parallel()

	want, _ := filepath.Abs(fixture)

	for _, start := range []string{fixture, filepath.Join(fixture, "internal", "app")} {
		got, err := codestyle.ModuleRoot(start)
		if err != nil || got != want {
			t.Errorf("корень от %s: %q, %v; ожидался %q", start, got, err, want)
		}
	}

	if _, err := codestyle.ModuleRoot(t.TempDir()); !errors.Is(err, codestyle.ErrNoModule) {
		t.Errorf("без go.mod ошибка %v, ожидалась ErrNoModule", err)
	}
}

// TestFileImports — каждый импорт с файлом и пакетом, тесты не считаются.
func TestFileImports(t *testing.T) {
	t.Parallel()

	got, err := codestyle.FileImports(fixture)
	if err != nil {
		t.Fatal(err)
	}

	want := []codestyle.FileImport{
		{File: "cmd/server/main.go", Package: "cmd/server", Path: "example.com/mod/internal/app"},
		{File: "cmd/server/main.go", Package: "cmd/server", Path: "example.com/mod/internal/store/postgres"},
		{File: "cmd/server/main.go", Package: "cmd/server", Path: "github.com/lllypuk/other/postgres"},
		{File: "internal/app/app.go", Package: "internal/app", Path: "example.com/mod/internal/store/postgres"},
		{File: "internal/app/app.go", Package: "internal/app", Path: "fmt"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("импорты по файлам:\n%v\nожидалось:\n%v", got, want)
	}
}
