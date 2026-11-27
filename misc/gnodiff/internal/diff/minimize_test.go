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
	got := Minimize(src, 10, func(string) bool { return false })
	if got != src && got != src+"\n" && !strings.Contains(got, "println(1)") {
		t.Fatalf("reducer deleted a line the predicate still needed:\n%s", got)
	}
	if !strings.Contains(got, "println(1)") {
		t.Fatalf("lost println:\n%s", got)
	}
}
