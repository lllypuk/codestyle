# codestyle

Тест-сканы стиля, которых нет ни в одном линтере golangci-lint: длина комментариев и импорты
между пакетами модуля. Библиотека даёт движок; бюджет нарушений, исключения путей и правила
слоёв принадлежат проекту и живут в его тесте.

## Длина комментариев

Правило: док-комментарий до трёх строк, комментарий внутри тела функции до двух, голых строк `//`
внутри блока не бывает. Тесты скан не трогает: у теста док часто единственное место, где написано,
какой инвариант он защищает.

```go
func TestCommentBlocksAreShort(t *testing.T) {
	root, _ := codestyle.ModuleRoot(".")
	found, checked, err := codestyle.Comments{
		Root:         root,
		SkipDirs:     []string{"sqlc"},
		SkipSuffixes: []string{"_templ.go"},
	}.Scan()
	if err != nil {
		t.Fatal(err)
	}

	codestyle.Ratchet(t, found, checked, budget, 25)
}
```

`Ratchet` валится и когда нарушений стало больше, и когда меньше: во втором случае впишите новое
число, иначе бюджет замрёт на первом значении. Ноль просмотренных файлов — сломанный корень.

## Импорты между пакетами

`Imports(root)` отдаёт импорты каждого пакета модуля по пути от корня, тесты не считаются.
Импорты полные, собственные отрезает `Own(module, imports)`: адаптер, вынесенный в чужой модуль,
иначе выпал бы из проверки молча.

```go
for pkg, imports := range packages {
	if slices.Contains(contexts, pkg) {
		for _, imp := range codestyle.Own("my-module", imports) {
			if strings.HasPrefix(imp, "internal/platform/") {
				t.Errorf("%s импортирует платформу %s", pkg, imp)
			}
		}
	}
}
```
