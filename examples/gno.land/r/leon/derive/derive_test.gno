package derive

import (
	"strings"
	"testing"
)

func TestParsePkgPath(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		expected string
	}{
		// r/ namespace
		{"full r path", "gno.land/r/leon/home", "gno.land/r/leon/home"},
		{"relative r path with slash", "/r/leon/home", "gno.land/r/leon/home"},
		{"relative r path no slash", "r/leon/home", "gno.land/r/leon/home"},

		// p/ namespace
		{"full p path", "gno.land/p/leon/home", "gno.land/p/leon/home"},
		{"relative p path with slash", "/p/leon/home", "gno.land/p/leon/home"},
		{"relative p path no slash", "p/leon/home", "gno.land/p/leon/home"},

		// invalid cases
		{"random text", "randomtext", ""},
		{"random text", "randomtext with space", ""},
		{"empty", "", ""},
		{"wrong namespace q/", "q/leon/home", ""},
		{"no pkgpath query param", "", ""},
		{"malformed url", "::::", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			url := "gno.land/r/leon/derive?pkgpath=" + c.path
			result := parsePkgPath(url)
			if result != c.expected {
				t.Errorf("for input %q expected %q, got %q", c.path, c.expected, result)
			}
		})
	}
}

func TestDerive(t *testing.T) {
	got := Render("?pkgpath=gno.land/r/leon/home")
	if !strings.Contains(got, "g1n0z9kw63c6ze8fgle93s2e86hfm2qz025cgkey") {
		// manually checked the above address
		t.Fatal("Failed derivation!")
	}
}
