package codestyle_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/lllypuk/codestyle"
)

// recorder — testing.TB, у которого отказы копятся, а не роняют тест.
type recorder struct {
	testing.TB

	failures []string
	logs     int
}

func (r *recorder) Helper()           {}
func (r *recorder) Log(...any)        { r.logs++ }
func (r *recorder) Fatal(args ...any) { r.failures = append(r.failures, fmt.Sprint(args...)) }
func (r *recorder) Errorf(format string, args ...any) {
	r.failures = append(r.failures, fmt.Sprintf(format, args...))
}

// TestRatchet — валится и выше, и ниже бюджета, молчит ровно на нём, печатает не больше shown.
func TestRatchet(t *testing.T) {
	t.Parallel()

	three := []codestyle.Violation{{Lines: 5}, {Lines: 4}, {Lines: 4}}

	cases := []struct {
		name    string
		found   []codestyle.Violation
		checked int
		budget  int
		want    string
	}{
		{"ровно", three, 10, 3, ""},
		{"выше", three, 10, 2, "при бюджете 2"},
		{"ниже", three, 10, 4, "впишите новое число"},
		{"пусто", nil, 0, 0, "не нашёл ни одного файла"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rec := &recorder{}
			codestyle.Ratchet(rec, tc.found, tc.checked, tc.budget, 2)

			got := strings.Join(rec.failures, "\n")
			if (tc.want == "") != (got == "") || !strings.Contains(got, tc.want) {
				t.Errorf("отказы %q, ожидалось %q", got, tc.want)
			}

			if tc.checked > 0 && rec.logs != 2 {
				t.Errorf("напечатано %d нарушений, ожидались два худших", rec.logs)
			}
		})
	}
}
