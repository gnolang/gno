package test_test

import (
	"testing"
	"github.com/gnolang/gno/gnovm/pkg/test"
)

func TestTestImportError_Error(t *testing.T) {
	err := test.TestImportError{PkgPath: "gno.land/foo/bar"}
	want := `unknown package path "gno.land/foo/bar"`

	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestTestImportError_String(t *testing.T) {
	err := test.TestImportError{PkgPath: "gno.land/foo/bar"}
	want := `TestImportError("unknown package path \"gno.land/foo/bar\"")`

	if got := err.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
