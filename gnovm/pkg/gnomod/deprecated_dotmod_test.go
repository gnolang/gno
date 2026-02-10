package gnomod

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/mod/modfile"
)

var modulePathTests = []struct {
	input    []byte
	expected string
}{
	{input: []byte("module \"github.com/rsc/vgotest\""), expected: "github.com/rsc/vgotest"},
	{input: []byte("module github.com/rsc/vgotest"), expected: "github.com/rsc/vgotest"},
	{input: []byte("module  \"github.com/rsc/vgotest\""), expected: "github.com/rsc/vgotest"},
	{input: []byte("module  github.com/rsc/vgotest"), expected: "github.com/rsc/vgotest"},
	{input: []byte("module `github.com/rsc/vgotest`"), expected: "github.com/rsc/vgotest"},
	{input: []byte("module \"github.com/rsc/vgotest/v2\""), expected: "github.com/rsc/vgotest/v2"},
	{input: []byte("module github.com/rsc/vgotest/v2"), expected: "github.com/rsc/vgotest/v2"},
	{input: []byte("module \"gopkg.in/yaml.v2\""), expected: "gopkg.in/yaml.v2"},
	{input: []byte("module gopkg.in/yaml.v2"), expected: "gopkg.in/yaml.v2"},
	{input: []byte("module \"gopkg.in/check.v1\"\n"), expected: "gopkg.in/check.v1"},
	{input: []byte("module \"gopkg.in/check.v1\n\""), expected: ""},
	{input: []byte("module gopkg.in/check.v1\n"), expected: "gopkg.in/check.v1"},
	{input: []byte("module \"gopkg.in/check.v1\"\r\n"), expected: "gopkg.in/check.v1"},
	{input: []byte("module gopkg.in/check.v1\r\n"), expected: "gopkg.in/check.v1"},
	{input: []byte("module \"gopkg.in/check.v1\"\n\n"), expected: "gopkg.in/check.v1"},
	{input: []byte("module gopkg.in/check.v1\n\n"), expected: "gopkg.in/check.v1"},
	{input: []byte("module \n\"gopkg.in/check.v1\"\n\n"), expected: ""},
	{input: []byte("module \ngopkg.in/check.v1\n\n"), expected: ""},
	{input: []byte("module \"gopkg.in/check.v1\"asd"), expected: ""},
	{input: []byte("module \n\"gopkg.in/check.v1\"\n\n"), expected: ""},
	{input: []byte("module \ngopkg.in/check.v1\n\n"), expected: ""},
	{input: []byte("module \"gopkg.in/check.v1\"asd"), expected: ""},
	{input: []byte("module  \nmodule a/b/c "), expected: "a/b/c"},
	{input: []byte("module \"   \""), expected: "   "},
	{input: []byte("module   "), expected: ""},
	{input: []byte("module \"  a/b/c  \""), expected: "  a/b/c  "},
	{input: []byte("module \"github.com/rsc/vgotest1\" // with a comment"), expected: "github.com/rsc/vgotest1"},
}

func TestModulePath(t *testing.T) {
	for _, test := range modulePathTests {
		t.Run(string(test.input), func(t *testing.T) {
			result := modulePath(test.input)
			if result != test.expected {
				t.Fatalf("ModulePath(%q): %s, want %s", string(test.input), result, test.expected)
			}
		})
	}
}

func TestParseVersions(t *testing.T) {
	tests := []struct {
		desc, input string
		ok          bool
	}{
		// go lines
		{desc: "empty", input: "module m\ngno \n", ok: false},
		{desc: "one", input: "module m\ngno 1\n", ok: false},
		{desc: "two", input: "module m\ngno 1.22\n", ok: true},
		{desc: "two go", input: "module m\ngo 1.22\n", ok: false},
		{desc: "three", input: "module m\ngno 1.22.333", ok: true},
		{desc: "before", input: "module m\ngno v1.2\n", ok: false},
		{desc: "after", input: "module m\ngno 1.2rc1\n", ok: true},
		{desc: "space", input: "module m\ngno 1.2 3.4\n", ok: false},
		{desc: "alt1", input: "module m\ngno 1.2.3\n", ok: true},
		{desc: "alt2", input: "module m\ngno 1.2rc1\n", ok: true},
		{desc: "alt3", input: "module m\ngno 1.2beta1\n", ok: true},
		{desc: "alt4", input: "module m\ngno 1.2.beta1\n", ok: false},
	}
	t.Run("Strict", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.desc, func(t *testing.T) {
				if _, err := parseDeprecatedDotModBytes("gno.mod", []byte(test.input)); err == nil && !test.ok {
					t.Error("unexpected success")
				} else if err != nil && test.ok {
					t.Errorf("unexpected error: %v", err)
				}
			})
		}
	})
}

func TestComments(t *testing.T) {
	for _, test := range []struct {
		desc, input, want string
	}{
		{
			desc: "comment_only",
			input: `
// a
// b
`,
			want: `
comments before "// a"
comments before "// b"
`,
		}, {
			desc: "line",
			input: `
// a

// b
module m // c
// d

// e
`,
			want: `
comments before "// a"
line before "// b"
line suffix "// c"
comments before "// d"
comments before "// e"
`,
		}, {
			desc:  "cr_removed",
			input: "// a\r\r\n",
			want:  `comments before "// a\r"`,
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			f, err := parseDeprecatedDotModBytes("gno.mod", []byte(test.input))
			if err != nil {
				t.Fatal(err)
			}

			if test.desc == "block" {
				panic("hov")
			}

			buf := &bytes.Buffer{}
			printComments := func(prefix string, cs *modfile.Comments) {
				for _, c := range cs.Before {
					fmt.Fprintf(buf, "%s before %q\n", prefix, c.Token)
				}
				for _, c := range cs.Suffix {
					fmt.Fprintf(buf, "%s suffix %q\n", prefix, c.Token)
				}
				for _, c := range cs.After {
					fmt.Fprintf(buf, "%s after %q\n", prefix, c.Token)
				}
			}

			printComments("file", &f.Syntax.Comments)
			for _, stmt := range f.Syntax.Stmt {
				switch stmt := stmt.(type) {
				case *modfile.CommentBlock:
					printComments("comments", stmt.Comment())
				case *modfile.Line:
					printComments("line", stmt.Comment())
				}
			}

			got := strings.TrimSpace(buf.String())
			want := strings.TrimSpace(test.want)
			if got != want {
				t.Errorf("got:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

var setGnoTests = []struct {
	desc    string
	in      string
	version string
	out     string
}{
	{
		`existing`,
		`
		gno 0.0
		`,
		"0.9",
		`
		gno 0.9
		`,
	},
	{
		`new`,
		``,
		"0.9",
		`
		gno 0.9
		`,
	},
}

var setModuleTests = []struct {
	desc string
	in   string
	path string
	out  string
}{
	{
		`existing`,
		`
		module m
		`,
		"n",
		`
		module n
		`,
	},
	{
		`new`,
		``,
		"m",
		`
		module m
		`,
	},
}

var addReplaceTests = []struct {
	desc    string
	in      string
	oldPath string
	oldVers string
	newPath string
	newVers string
	out     string
}{
	{
		`replace_with_module`,
		`
		module m
		`,
		"x.y/z",
		"v1.5.6",
		"a.b/c",
		"v1.5.6",
		`
		module m
		replace x.y/z v1.5.6 => a.b/c v1.5.6
		`,
	},
	{
		`replace_with_dir`,
		`
		module m
		`,
		"x.y/z",
		"v1.5.6",
		"/path/to/dir",
		"",
		`
		module m
		replace x.y/z v1.5.6 => /path/to/dir
		`,
	},
}

var dropReplaceTests = []struct {
	desc string
	in   string
	path string
	vers string
	out  string
}{
	{
		`existing`,
		`
		module m

		replace x.y/z v1.2.3 => a.b/c v1.5.6
		`,
		"x.y/z",
		"v1.2.3",
		`
		module m
		`,
	},
	{
		`not_exists`,
		`
		module m

		replace x.y/z v1.2.3 => a.b/c v1.5.6
		`,
		"a.b/c",
		"v3.2.1",
		`
		module m

		replace x.y/z v1.2.3 => a.b/c v1.5.6
		`,
	},
}

func testEdit(t *testing.T, in, want string, transform func(f *deprecatedModFile) error) *deprecatedModFile {
	t.Helper()
	f, err := parseDeprecatedDotModBytes("in", []byte(in))
	if err != nil {
		t.Fatal(err)
	}
	g, err := parseDeprecatedDotModBytes("out", []byte(want))
	if err != nil {
		t.Fatal(err)
	}
	golden := modfile.Format(g.Syntax)
	if err := transform(f); err != nil {
		t.Fatal(err)
	}
	out := modfile.Format(f.Syntax)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, golden) {
		t.Errorf("have:\n%s\nwant:\n%s", out, golden)
	}

	return f
}
