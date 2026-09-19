package games

import (
	"testing"
)

func TestRenderAuthors(t *testing.T) {
	assertEqual(t, RenderAuthors([]string{}), "*?*")
	assertEqual(t, RenderAuthors([]string{"tim"}), "tim")
	assertEqual(t, RenderAuthors([]string{"tim", "jim"}), "tim and jim")
	assertEqual(t, RenderAuthors([]string{"tim", "jim", "kim"}), "tim, jim and kim")
	assertEqual(t, RenderAuthors([]string{"tim", "jim", "kim", "nim"}), "tim, jim, kim and nim")
	assertEqual(t, RenderAuthors([]string{}, "@"), "*?*")
	assertEqual(t, RenderAuthors([]string{"tim"}, "@"), "@tim")
	assertEqual(t, RenderAuthors([]string{"tim", "jim"}, "@"), "@tim and @jim")
	assertEqual(t, RenderAuthors([]string{"tim", "jim", "kim"}, "@"), "@tim, @jim and @kim")
	assertEqual(t, RenderAuthors([]string{"tim", "jim", "kim", "nim"}, "@"), "@tim, @jim, @kim and @nim")
}

func assertEqual(t *testing.T, got, expected string) {
	t.Helper()
	if expected != got {
		t.Errorf("expected %s, got %s", expected, got)
	}
}
