package mdlist

import (
	"strings"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestUL(t *testing.T) {
	list := List{
		{Title: "First", Body: "Body1", Links: []string{"[link1](url1)"}},
		{Title: "Second", Body: "Body2", Links: []string{"[link2](url2)", "[link3](url3)"}},
		{Title: "Third", Body: "Body3", Link: "https://example.com"},
		{Title: "Fourth", Body: "Body4", Link: "https://example.com", Links: []string{"[link4](url4)", "[link5](url5)"}},
	}
	col1 := "- **First**: *Body1* [link1](url1)"
	col2 := "- **Second**: *Body2* [link2](url2) [link3](url3)"
	col3 := "- **[Third](https://example.com)**: *Body3*"
	col4 := "- **[Fourth](https://example.com)**: *Body4* [link4](url4) [link5](url5)"
	expected := strings.Join([]string{col1, col2, col3, col4}, "\n") + "\n"
	uassert.Equal(t, expected, list.UL())
}

func TestGrid(t *testing.T) {
	list := List{
		{Title: "First", Body: "Body1", Links: []string{"[link1](url1)"}},
		{Title: "Second", Body: "Body2", Links: []string{"[link2](url2)", "[link3](url3)"}},
		{Title: "Third", Body: "Body3", Link: "https://example.com"},
		{Title: "Fourth", Body: "Body4", Link: "https://example.com", Links: []string{"[link4](url4)", "[link5](url5)"}},
	}
	col1 := "## First\nBody1\n\n[link1](url1)\n"
	col2 := "## Second\nBody2\n\n[link2](url2) [link3](url3)\n"
	col3 := "## [Third](https://example.com)\nBody3\n"
	col4 := "## [Fourth](https://example.com)\nBody4\n\n[link4](url4) [link5](url5)\n"
	expected := "<gno-columns>\n" + col1 + "|||\n" + col2 + "|||\n" + col3 + "|||\n" + col4 + "</gno-columns>\n"
	uassert.Equal(t, expected, list.Grid())
}

func TestTruncateText(t *testing.T) {
	long := "abcd\nefghijklmnopqrs\ntuvwxyz01\n23456789"
	short := "short"
	urequire.Equal(t, "abcd ef...", truncateText(long, 10, false))
	urequire.Equal(t, short, truncateText(short, 10, false))
	urequire.Equal(t, "abcd\nef...", truncateText(long, 10, true))
	urequire.Equal(t, "short", truncateText(short, 10, true))
}
