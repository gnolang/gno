package fqname

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input           string
		expectedPkgPath string
		expectedName    string
	}{
		{"gno.land/p/nt/avl/v0.Tree", "gno.land/p/nt/avl/v0", "Tree"},
		{"gno.land/p/nt/avl/v0", "gno.land/p/nt/avl/v0", ""},
		{"gno.land/p/nt/avl/v0.Tree.Node", "gno.land/p/nt/avl/v0", "Tree.Node"},
		{"gno.land/p/nt/avl/v0/nested.Package.Func", "gno.land/p/nt/avl/v0/nested", "Package.Func"},
		{"path/filepath.Split", "path/filepath", "Split"},
		{"path.Split", "path", "Split"},
		{"path/filepath", "path/filepath", ""},
		{"path", "path", ""},
		{"", "", ""},
	}

	for _, tt := range tests {
		pkgpath, name := Parse(tt.input)
		uassert.Equal(t, tt.expectedPkgPath, pkgpath, "Package path did not match")
		uassert.Equal(t, tt.expectedName, name, "Name did not match")
	}
}

func TestConstruct(t *testing.T) {
	tests := []struct {
		pkgpath  string
		name     string
		expected string
	}{
		{"gno.land/r/demo/foo20", "Token", "gno.land/r/demo/foo20.Token"},
		{"gno.land/r/demo/foo20", "", "gno.land/r/demo/foo20"},
		{"path", "", "path"},
		{"path", "Split", "path.Split"},
		{"path/filepath", "", "path/filepath"},
		{"path/filepath", "Split", "path/filepath.Split"},
		{"", "JustName", ".JustName"},
		{"", "", ""},
	}

	for _, tt := range tests {
		result := Construct(tt.pkgpath, tt.name)
		uassert.Equal(t, tt.expected, result, "Constructed FQName did not match expected")
	}
}

func TestRenderLink(t *testing.T) {
	tests := []struct {
		pkgPath  string
		slug     string
		expected string
	}{
		{"gno.land/p/nt/avl/v0", "Tree", "[gno.land/p/nt/avl/v0](/p/nt/avl/v0).Tree"},
		{"gno.land/p/nt/avl/v0", "", "[gno.land/p/nt/avl/v0](/p/nt/avl/v0)"},
		{"github.com/a/b", "C", "github.com/a/b.C"},
		{"example.com/pkg", "Func", "example.com/pkg.Func"},
		{"gno.land/r/demo/foo20", "Token", "[gno.land/r/demo/foo20](/r/demo/foo20).Token"},
		{"gno.land/r/demo/foo20", "", "[gno.land/r/demo/foo20](/r/demo/foo20)"},
		{"", "", ""},
	}

	for _, tt := range tests {
		result := RenderLink(tt.pkgPath, tt.slug)
		uassert.Equal(t, tt.expected, result, "Rendered link did not match expected")
	}
}
