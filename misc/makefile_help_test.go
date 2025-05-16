// makefile_help_test.go
package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// Test parseConfig with no args and with valid Makefile.
func TestParseConfig(t *testing.T) {
	_, _, err := parseConfig([]string{})
	if err == nil {
		t.Fatal("expected error when no Makefile provided")
	}

	// create a temp Makefile
	dir := t.TempDir()
	mf := filepath.Join(dir, "Makefile")
	if err := ioutil.WriteFile(mf, []byte("all:\n"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, _, err := parseConfig([]string{"-r", "base", "-d", "sub", "-w", "X", mf})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Makefile != mf {
		t.Errorf("Makefile = %q, want %q", cfg.Makefile, mf)
	}
	if cfg.RelativeTo != "base" {
		t.Errorf("RelativeTo = %q, want %q", cfg.RelativeTo, "base")
	}
	if len(cfg.Dirs) != 1 || cfg.Dirs[0] != "sub" {
		t.Errorf("Dirs = %v, want [sub]", cfg.Dirs)
	}
	if len(cfg.Wildcards) != 1 || cfg.Wildcards[0] != "X" {
		t.Errorf("Wildcards = %v, want [X]", cfg.Wildcards)
	}
}

// Test extractMakefileTargets includes and excludes correctly.
func TestExtractMakefileTargets(t *testing.T) {
	content := `
foo: # first target
bar: do stuff ## double hash desc
baz: # description with % wildcard %
legacy: # @LEGACY should skip
`
	mf := filepath.Join(t.TempDir(), "Makefile")
	if err := ioutil.WriteFile(mf, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tMap, err := extractMakefileTargets(mf)
	if err != nil {
		t.Fatal(err)
	}
	if d, ok := tMap["foo"]; !ok || d != "first target" {
		t.Errorf("foo => %q, want %q", d, "first target")
	}
	if d, ok := tMap["bar"]; !ok || d != "double hash desc" {
		t.Errorf("bar => %q, want %q", d, "double hash desc")
	}
	if d, ok := tMap["baz"]; !ok || !strings.Contains(d, "description") {
		t.Errorf("baz => %q, must contain 'description'", d)
	}
	if _, ok := tMap["legacy"]; ok {
		t.Error("expected legacy target to be skipped")
	}
}

// Test readReadmeBanner with and without file.
func TestReadReadmeBanner(t *testing.T) {
	dir := t.TempDir()
	// no README.md → empty banner
	b, err := readReadmeBanner(dir)
	if err != nil {
		t.Fatal(err)
	}
	if b != "" {
		t.Errorf("banner = %q, want empty", b)
	}

	// write a README.md
	markdown := "# " + filepath.Base(dir) + ": Hello World\nSecond line"
	if err := ioutil.WriteFile(filepath.Join(dir, "README.md"), []byte(markdown), 0644); err != nil {
		t.Fatal(err)
	}
	b, err = readReadmeBanner(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b, "Hello World") {
		t.Errorf("banner = %q, want to contain 'Hello World'", b)
	}
}

// Test maxKeyLength with and without wildcards.
func TestMaxKeyLength(t *testing.T) {
	keys := []string{"a", "longer", "p%t"}
	w := []string{"X", "YY"}
	// "longer" length 6
	if got := maxKeyLength(keys, nil); got != 6 {
		t.Errorf("maxKeyLength(keys,nil) = %d; want 6", got)
	}
	// p%t expands to length of key + maxWildcard-1 = 3+(2-1)=4, but longer=6
	if got := maxKeyLength(keys, w); got != 6 {
		t.Errorf("maxKeyLength(keys,w) = %d; want 6", got)
	}
}

// Test maxStringLength for empty and non‑empty.
func TestMaxStringLength(t *testing.T) {
	if l := maxStringLength(nil); l != 0 {
		t.Errorf("maxStringLength(nil) = %d; want 0", l)
	}
	if l := maxStringLength([]string{"", "ab", "xyz"}); l != 3 {
		t.Errorf("maxStringLength = %d; want 3", l)
	}
}


// Table‑driven tests for extractMakefileTargets, covering legacy skips,
// double‑hash, non‑letter starts, missing colons, wildcard targets, etc.
func TestExtractMakefileTargets_Table(t *testing.T) {
	cases := []struct {
		name     string
		content  string
		expected map[string]string
	}{
		{
			name:    "simple target no desc",
			content: "build:\n",
			expected: map[string]string{
				"build": "",
			},
		},
		{
			name:    "target with desc",
			content: "run: # execute application\n",
			expected: map[string]string{
				"run": "execute application",
			},
		},
		{
			name:    "double hash absorbed trimining whitespace",
			content: "test: ##  a real desc  \n",
			expected: map[string]string{
				"test": "a real desc",
			},
		},
		{
			name:    "legacy directive skipped",
			content: "old: # @LEGACY preserve\nnew: # active\n",
			expected: map[string]string{
				"new": "active",
			},
		},
		{
			name:    "wildcard target desc with %",
			content: "install-%: # install package %\n",
			expected: map[string]string{
				"install-%": "install package %",
			},
		},
		{
			name:    "non-letter start skipped",
			content: "_hidden: # skip me\nVisible: # ok\n",
			expected: map[string]string{
				"Visible": "ok",
			},
		},
		{
			name:    "no colon line ignored",
			content: "foobar # comment only\nbaz: #desc here\n",
			expected: map[string]string{
				"baz": "desc here",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			mf := filepath.Join(dir, "Makefile")
			if err := os.WriteFile(mf, []byte(tc.content), 0644); err != nil {
				t.Fatalf("failed to write temp Makefile: %v", err)
			}

			got, err := extractMakefileTargets(mf)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("got targets %v, want %v", got, tc.expected)
			}
		})
	}
}

// Table‑driven tests for readReadmeBanner, proving out the prefix‑stripping regex.
func TestReadReadmeBanner_Table(t *testing.T) {
	cases := []struct {
		name    string
		content string // raw README.md content; may include <NAME> placeholder
		want    string
	}{
		{"empty readme", "", ""},
		{"simple header", "# Hello World\nSecond line", " (Hello World)"},
		{"no leading hash", "Just text title\nmore", " (Just text title)"},
		{"strip dirname prefix", "# <NAME>: Important", " (Important)"},
		{"strip backtick dirname", "# `<NAME>` -- Note", " (Note)"},
		{"case‑insensitive prefix", "# <NAME>: MixedCase", " (MixedCase)"},
		{"only prefix yields empty", "#<NAME>", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			// replace placeholder
			content := strings.ReplaceAll(tc.content, "<NAME>", dir)
			if content != "" {
				if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write README.md: %v", err)
				}
			}

			got, err := readReadmeBanner(dir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("banner = %q, want %q", got, tc.want)
			}
		})
	}
}

// Table‑driven tests for maxStringLength.
func TestMaxStringLength_Table(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want int
	}{
		{"nil slice", nil, 0},
		{"empty strings", []string{"", ""}, 0},
		{"mixed lengths", []string{"a", "abc", "ab"}, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := maxStringLength(tc.in); got != tc.want {
				t.Errorf("maxStringLength(%v) = %d; want %d", tc.in, got, tc.want)
			}
		})
	}
}

// Table‑driven tests for maxKeyLength, exercising '%' expansion logic.
func TestMaxKeyLength_Table(t *testing.T) {
	cases := []struct {
		name      string
		keys      []string
		wildcards []string
		want      int
	}{
		{"no wildcards", []string{"a", "bb", "ccc"}, nil, 3},
		{"single wildcard shorter", []string{"p%t", "xx"}, []string{"Z"}, 3},
		{"wildcard longer than key", []string{"p%t"}, []string{"YY"}, 4},
		{"multiple wildcards, picks longest", []string{"p%t"}, []string{"X", "WWW"}, 5},
		{"mix wild and non‑wild", []string{"long", "m%x"}, []string{"ZZZ"}, 5},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := maxKeyLength(tc.keys, tc.wildcards)
			if got != tc.want {
				t.Errorf("maxKeyLength(keys=%v, wildcards=%v) = %d; want %d",
					tc.keys, tc.wildcards, got, tc.want)
			}
		})
	}
}
