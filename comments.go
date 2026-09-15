// Package codestyle — тест-сканы стиля, которых нет ни в одном линтере: длина
// комментариев и импорты между пакетами модуля. Бюджет нарушений, исключения
// путей и правила слоёв принадлежат проекту.
package codestyle

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

// Comments — правило длины комментариев: док до DocLines, внутри тела до BodyLines,
// голых строк `//` внутри блока не бывает. Тесты скан не трогает: у теста док часто
// единственное место, где написано, какой инвариант он защищает.
type Comments struct {
	// Root — корень модуля; правило про весь код, а не про один пакет.
	Root string
	// DocLines и BodyLines — пределы; ноль означает умолчания 3 и 2.
	DocLines  int
	BodyLines int
	// SkipDirs — имена каталогов сверх умолчаний (.git, testdata, кэши CI).
	SkipDirs []string
	// SkipSuffixes — суффиксы файлов сверх `_test.go`: например `_templ.go`.
	SkipSuffixes []string
}

// Violation — один слишком длинный или дырявый блок.
type Violation struct {
	File  string
	Line  int
	Lines int
	Limit int
	Why   string
}

// String — строка отчёта: файл:строка, причина, длина и предел.
func (v Violation) String() string {
	return fmt.Sprintf("%s:%d — %s: %d строк при пределе %d", v.File, v.Line, v.Why, v.Lines, v.Limit)
}

// DefaultSkipDirs — каталоги, которые не исходники: git, кэши CI, зависимости, фикстуры.
func DefaultSkipDirs() []string {
	return []string{".git", ".gopath", ".gocache", ".tools", "node_modules", "testdata", "vendor"}
}

// Scan обходит модуль и отдаёт нарушения, худшие первыми, и число просмотренных файлов.
// Ноль файлов — сломанный корень, и проверять его обязан вызывающий.
func (c Comments) Scan() ([]Violation, int, error) {
	docLimit, bodyLimit := c.limits()
	skip := append(DefaultSkipDirs(), c.SkipDirs...)

	var (
		found   []Violation
		checked int
	)

	err := filepath.WalkDir(c.Root, func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			if name != c.Root && slices.Contains(skip, entry.Name()) {
				return fs.SkipDir
			}

			return nil
		}

		if !c.scannable(entry.Name()) {
			return nil
		}

		violations, err := inspect(name, docLimit, bodyLimit)
		if err != nil {
			return err
		}

		checked++
		found = append(found, violations...)

		return nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("обход %s: %w", c.Root, err)
	}

	slices.SortStableFunc(found, func(a, b Violation) int { return b.Lines - a.Lines })

	return found, checked, nil
}

func (c Comments) limits() (int, int) {
	docLimit, bodyLimit := c.DocLines, c.BodyLines
	if docLimit == 0 {
		docLimit = 3
	}

	if bodyLimit == 0 {
		bodyLimit = 2
	}

	return docLimit, bodyLimit
}

func (c Comments) scannable(name string) bool {
	if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
		return false
	}

	return !slices.ContainsFunc(c.SkipSuffixes, func(suffix string) bool { return strings.HasSuffix(name, suffix) })
}

func inspect(path string, docLimit, bodyLimit int) ([]Violation, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("разбор %s: %w", path, err)
	}

	docs, bodies := shape(file)

	var found []Violation

	for _, group := range file.Comments {
		if directive(group) {
			continue
		}

		limit, why := bodyLimit, "комментарий внутри тела функции"

		switch {
		case docs[group.Pos()]:
			limit, why = docLimit, "док-комментарий"
		case !within(bodies, group.Pos()):
			limit, why = docLimit, "комментарий верхнего уровня"
		}

		lines := height(group)
		if lines <= limit && !blank(group) {
			continue
		}

		if blank(group) {
			why += " с пустой строкой //"
		}

		found = append(found, Violation{
			File:  path,
			Line:  fset.Position(group.Pos()).Line,
			Lines: lines,
			Limit: limit,
			Why:   why,
		})
	}

	return found, nil
}

// shape — позиции док-комментариев и границы тел функций одним обходом.
func shape(file *ast.File) (map[token.Pos]bool, [][2]token.Pos) {
	docs := map[token.Pos]bool{}
	bodies := [][2]token.Pos{}

	mark := func(doc *ast.CommentGroup) {
		if doc != nil {
			docs[doc.Pos()] = true
		}
	}

	mark(file.Doc)

	ast.Inspect(file, func(node ast.Node) bool {
		switch decl := node.(type) {
		case *ast.FuncDecl:
			mark(decl.Doc)

			if decl.Body != nil {
				bodies = append(bodies, [2]token.Pos{decl.Body.Pos(), decl.Body.End()})
			}
		case *ast.FuncLit:
			bodies = append(bodies, [2]token.Pos{decl.Body.Pos(), decl.Body.End()})
		case *ast.GenDecl:
			mark(decl.Doc)
		case *ast.TypeSpec:
			mark(decl.Doc)
		case *ast.ValueSpec:
			mark(decl.Doc)
		case *ast.ImportSpec:
			mark(decl.Doc)
		case *ast.Field:
			mark(decl.Doc)
		}

		return true
	})

	return docs, bodies
}

func within(bodies [][2]token.Pos, pos token.Pos) bool {
	return slices.ContainsFunc(bodies, func(span [2]token.Pos) bool {
		return pos >= span[0] && pos < span[1]
	})
}

// directive — //go:, //nolint и прочее, что читает не человек.
func directive(group *ast.CommentGroup) bool {
	head := group.List[0].Text

	for _, prefix := range []string{"//go:", "//nolint", "//lint:", "//line ", "// +build", "//sys "} {
		if strings.HasPrefix(head, prefix) {
			return true
		}
	}

	return false
}

func height(group *ast.CommentGroup) int {
	var lines int

	for _, comment := range group.List {
		lines += strings.Count(comment.Text, "\n") + 1
	}

	return lines
}

func blank(group *ast.CommentGroup) bool {
	return slices.ContainsFunc(group.List, func(comment *ast.Comment) bool {
		return strings.TrimSpace(comment.Text) == "//"
	})
}
