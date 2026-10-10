# `codestyle` v0.2.0: импорты по файлам

## Overview

Выпуск только добавляет API.

`Imports` отдаёт импорты по пакетам, а правилу слоёв DPP нужны файлы — исключения записаны парой
«файл → импорт» (`importExceptions`). Поэтому DPP держит свой обход
(`internal/codestyle/imports_test.go:293-346`, `scanImports`). Остальные 400 строк файла — политика
слоёв DPP, она остаётся у него.

Не делаем:
- **режим ратчета для `codestyle.Ratchet` — строгое равенство остаётся** (решение подтверждено codex
  и Fable 09.10.2026): гонка двух веток, сдвинувших бюджет, ловится проверкой merge-result в
  процессе диспетчера, а мягкий ратчет замораживал бы бюджет на первом значении.

## Решения (приняты 2026-10-09, не переигрывать)

1. **`codestyle.FileImports(root string) ([]FileImport, error)`**, `FileImport{File, Package, Path}` —
   пути от `root` через `/`, тесты не считаются, обход и пропуски — как у `Imports`. Оба строятся на
   общем приватном обходчике: поверх `FileImports` `Imports` потерял бы пакет без импортов (`TestImports`
   ждёт `internal/store/postgres` с nil).
2. **CHANGELOG** заводится: `## v0.2.0` и строка о v0.1.0; раздел `Ломает:` —
   с первой ломающей версии (правило — как в `llm`).

## Context (from discovery)

- `imports.go:18-64` — `Imports` (обход `filepath.WalkDir`, `DefaultSkipDirs`,
  `parser.ImportsOnly`); тест на фикстуре `testdata` — `imports_test.go`.
- DPP `scanImports` отличается от `Imports` только тем, что отдаёт файл, обходит `internal/` и
  пропускает каталоги `helperPackages` — последние два фильтруются потребителем по `Package`.
- `go 1.26`, stdlib only, CI — `.github/workflows/ci.yml`, общий `.golangci.yml`.

## Development Approach

- **testing approach**: TDD.
- Перед каждым коммитом: `go vet ./... && go test -race ./...` и
  `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run` — 0 замечаний.
- Коммиты на русском: `feat:`, `docs:`; комментарии — по `~/.claude/CLAUDE.md`.

## Testing Strategy

- Юнит-тесты без сети и файлов вне `testdata`.

## Progress Tracking

`[x]` — сделано, ➕ — найдено по ходу, ⚠️ — блокер.

## Implementation Steps

### Task 1: `FileImports`

**Files:** `imports.go`, `imports_test.go`

- [ ] `FileImport`, `FileImports` по Решению 1; `Imports` и `FileImports` — на общем обходчике
- [ ] тесты на фикстуре: файл, пакет и путь каждого импорта; `_test.go` не попадает; `TestImports`
  проходит без правки ожиданий
- [ ] `go test -race ./...` — зелёный
- [ ] коммит `feat: FileImports — импорты по файлам`

### Task 2: Verify acceptance criteria

- [ ] `Ratchet` не менялся (строгое равенство — «Не делаем»)
- [ ] `go vet`, тесты, линтер — зелёные

### Task 3: [Final] Документы

**Files:** `README.md`, `CHANGELOG.md`

- [ ] README, раздел «Импорты между пакетами»: `FileImports` — когда исключения правила пишутся по файлам
- [ ] `CHANGELOG.md` по Решению 2
- [ ] коммит `docs: v0.2.0 — FileImports в README и CHANGELOG`
