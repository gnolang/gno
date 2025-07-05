package test_test

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/test"
)

func TestDirectivesFileTest(t *testing.T) {
	d := test.Directives{
		{
			Name:     "",
			Content:  "Empty name",
			Complete: false,
			LastLine: "---a line---",
		},
		{
			Name:     "ALL CAPS NAME",
			Content:  "other content",
			Complete: false,
			LastLine: "---a line---",
		},
		{
			Name:     "default name",
			Content:  "Has a\n\nblank line",
			Complete: false,
			LastLine: "---a line---",
		},
		{
			Name:     "other default name",
			Content:  "", // Empty content
			Complete: false,
			LastLine: "---a line---",
		},
	}
	result := d.FileTest()
	expected := "Empty name\n\n// ALL CAPS NAME: other content\n// default name:\n// Has a\n//\n// blank line\n\n"
	if result != expected {
		t.Errorf("d.FileTest() gave result = %#v, expected %#v", result, expected)
	}
}
