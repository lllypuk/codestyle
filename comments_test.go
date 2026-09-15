package codestyle_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/lllypuk/codestyle"
)

const fixture = "testdata/module"

// TestScan — длинный док, длинный блок в теле и дырявый блок ловятся, директивы,
// тесты и кэши CI нет; худшее нарушение первым.
func TestScan(t *testing.T) {
	t.Parallel()

	found, checked, err := codestyle.Comments{Root: fixture, SkipSuffixes: []string{"_templ.go"}}.Scan()
	if err != nil {
		t.Fatal(err)
	}

	if checked != 3 {
		t.Errorf("файлов просмотрено %d, ожидались app.go, postgres.go и main.go", checked)
	}

	if len(found) != 3 {
		t.Fatalf("нарушений %d, ожидались три: %v", len(found), found)
	}

	if found[0].Lines != 4 || found[0].Limit != 3 || found[0].Why != "док-комментарий" {
		t.Errorf("первым ожидался док на четыре строки, а не %s", found[0])
	}

	if found[1].Lines != 3 || found[1].Limit != 2 || !strings.Contains(found[1].Why, "внутри тела") {
		t.Errorf("вторым ожидался блок в теле на три строки, а не %s", found[1])
	}

	if found[2].Lines != 3 || !strings.HasSuffix(found[2].Why, "с пустой строкой //") {
		t.Errorf("третьим ожидался дырявый док, а не %s", found[2])
	}

	if filepath.Base(found[0].File) != "app.go" || found[0].Line != 9 {
		t.Errorf("адрес нарушения %s:%d, ожидался app.go:9", found[0].File, found[0].Line)
	}
}

// TestScanWithoutSuffixSkip — без исключения суффикса сгенерированный файл считается.
func TestScanWithoutSuffixSkip(t *testing.T) {
	t.Parallel()

	found, checked, err := codestyle.Comments{Root: fixture}.Scan()
	if err != nil {
		t.Fatal(err)
	}

	if checked != 4 || len(found) != 4 {
		t.Errorf("файлов %d, нарушений %d — ожидались 4 и 4 с gen_templ.go", checked, len(found))
	}
}

// TestScanLimits — пределы настраиваются: при 4 и 3 нарушением остаётся только дыра.
func TestScanLimits(t *testing.T) {
	t.Parallel()

	found, _, err := codestyle.Comments{Root: fixture, DocLines: 4, BodyLines: 3}.Scan()
	if err != nil {
		t.Fatal(err)
	}

	if len(found) != 1 || !strings.Contains(found[0].Why, "пустой строкой") {
		t.Errorf("ожидалась одна дыра, получено %v", found)
	}
}

// TestScanMissingRoot — несуществующий корень — ошибка, а не нулевой скан.
func TestScanMissingRoot(t *testing.T) {
	t.Parallel()

	if _, _, err := (codestyle.Comments{Root: "testdata/nowhere"}).Scan(); err == nil {
		t.Error("ожидалась ошибка обхода")
	}
}

// TestSelf — сама библиотека держит правило с нулевым бюджетом.
func TestSelf(t *testing.T) {
	t.Parallel()

	root, err := codestyle.ModuleRoot(".")
	if err != nil {
		t.Fatal(err)
	}

	found, checked, err := codestyle.Comments{Root: root}.Scan()
	if err != nil {
		t.Fatal(err)
	}

	codestyle.Ratchet(t, found, checked, 0, 25)
}
