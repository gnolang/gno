package mdlist

import (
	"testing"
)

func TestUL(t *testing.T) {
	list := List{
		{Title: "First", Body: "Body1", Links: []string{"[link1](url1)"}},
		{Title: "Second", Body: "Body2", Links: []string{"[link2](url2)", "[link3](url3)"}},
	}
	expected := "- **First**: *Body1* [link1](url1)\n- **Second**: *Body2* [link2](url2) [link3](url3)\n"
	got := list.UL()
	if got != expected {
		t.Errorf("UL() = %q, want %q", got, expected)
	}
}

func TestGrid(t *testing.T) {
	list := List{
		{Title: "First", Body: "Body1", Links: []string{"[link1](url1)"}},
		{Title: "Second", Body: "Body2", Links: []string{"[link2](url2)", "[link3](url3)"}},
	}
	col1 := "## First\nBody1\n\n[link1](url1)\n"
	col2 := "## Second\nBody2\n\n[link2](url2) [link3](url3)\n"
	expected := "<gno-columns>\n" + col1 + "|||\n" + col2 + "</gno-columns>\n"
	got := list.Grid()
	if got != expected {
		t.Errorf("Grid() = %q, want %q", got, expected)
	}
}

func TestTruncateText(t *testing.T) {
	long := "abcdefghijklmnopqrstuvwxyz0123456789"
	short := "short"
	if got := truncateText(long, 10); got != "abcdefghij..." {
		t.Errorf("truncateText(long, 10) = %q, want %q", got, "abcdefghij...")
	}
	if got := truncateText(short, 10); got != short {
		t.Errorf("truncateText(short, 10) = %q, want %q", got, short)
	}
}
