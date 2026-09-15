package codestyle

import "testing"

// Ratchet сверяет число нарушений с бюджетом в обе стороны: выросло — комментарий
// объясняет видное из кода, упало — впишите новое число, иначе бюджет замрёт.
// Ноль просмотренных файлов — сломанный корень, а не чистый модуль.
func Ratchet(t testing.TB, found []Violation, checked, budget, shown int) {
	t.Helper()

	if checked == 0 {
		t.Fatal("скан не нашёл ни одного файла — проверка сломана")
	}

	for _, v := range found[:min(len(found), shown)] {
		t.Log(v.String())
	}

	switch {
	case len(found) > budget:
		t.Errorf(
			"нарушений %d при бюджете %d (файлов просмотрено %d): комментарий объясняет то, что видно из кода",
			len(found), budget, checked,
		)
	case len(found) < budget:
		t.Errorf("нарушений %d, а бюджет %d — впишите новое число, иначе ратчет встанет", len(found), budget)
	}
}
