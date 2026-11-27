package diff

import (
	"strings"
	"testing"
)

func TestMinimizeDropsNoise(t *testing.T) {
	src := "" +
		"package main\n" +
		"\n" +
		"func unused() int { return 0 }\n" +
		"\n" +
		"func main() {\n" +
		"\tprintln(1)\n" +
		"}\n"
	got := Minimize(src, 20, func(s string) bool {
		return strings.Contains(s, "println(1)") && strings.Contains(s, "package main")
	})
	if !strings.Contains(got, "println(1)") {
		t.Fatalf("minimized program lost the interesting line:\n%s", got)
	}
	if strings.Contains(got, "unused") {
		t.Fatalf("unused function survived:\n%s", got)
	}
	if strings.Contains(got, "package ") && !strings.Contains(got, "package main") {
		t.Fatalf("package clause changed:\n%s", got)
	}
}

func TestMinimizeStopsWhenPredicateRejects(t *testing.T) {
	src := "package main\nfunc main() { println(1) }\n"
	calls := 0
	got := Minimize(src, 10, func(string) bool {
		calls++
		return false
	})
	if got != src {
		t.Fatalf("reducer changed program when predicate rejected:\n%s", got)
	}
	// Only 1 non-package line exists; trailing newline must not waste budget.
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}
