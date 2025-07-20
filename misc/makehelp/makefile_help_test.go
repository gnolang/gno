// makefile_help_test.go
package main

import (
	"bytes"
	"errors"
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
	if err := ioutil.WriteFile(mf, []byte("all:\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, _, err := parseConfig([]string{"--invocation-dir-prefix", "base", "--dir", "sub", "--wildcard", "X", "--wildcard", "Y", "--wildcard", "2:Z", mf})
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
	if len(cfg.Wildcards) != 2 ||
		len(cfg.Wildcards[0]) != 2 ||
		len(cfg.Wildcards[1]) != 1 ||
		cfg.Wildcards[0][0] != "X" ||
		cfg.Wildcards[0][1] != "Y" ||
		cfg.Wildcards[1][0] != "Z" {
		t.Errorf("Wildcards = %v, want [[X,Y],[Z]]", cfg.Wildcards)
	}
}

// Test extractMakefileTargets includes and excludes correctly.
func TestExtractMakefileTargets(t *testing.T) {
	content := `
foo: # first target
bar: do stuff ## double hash desc
baz: # description with % wildcard %
legacy: # @LEGACY should skip
immediate_var:=skip me value
`
	mf := filepath.Join(t.TempDir(), "Makefile")
	if err := ioutil.WriteFile(mf, []byte(content), 0o644); err != nil {
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
	if _, ok := tMap["immediate_var"]; ok {
		t.Error("expected variable assignment to be skipped")
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
	if err := ioutil.WriteFile(filepath.Join(dir, "README.md"), []byte(markdown), 0o644); err != nil {
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
	w := [][]string{{"X", "YY"}}
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
			if err := os.WriteFile(mf, []byte(tc.content), 0o644); err != nil {
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
		{
			name:    "empty readme",
			content: "",
			want:    "",
		},
		{
			name:    "simple header",
			content: "# Hello World\nSecond line",
			want:    " (Hello World)",
		},
		{
			name:    "no leading hash",
			content: "Just text title\nmore",
			want:    " (Just text title)",
		},
		{
			name:    "strip dirname prefix",
			content: "# <NAME>: Important",
			want:    " (Important)",
		},
		{
			name:    "strip backtick dirname",
			content: "# `<NAME>` -- Note",
			want:    " (Note)",
		},
		{
			name:    "case‑insensitive prefix",
			content: "# <NAME>: MixedCase",
			want:    " (MixedCase)",
		},
		{
			name:    "only prefix yields empty",
			content: "#<NAME>",
			want:    "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			// replace placeholder
			content := strings.ReplaceAll(tc.content, "<NAME>", dir)
			if content != "" {
				if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(content), 0o644); err != nil {
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
		wildcards [][]string
		want      int
	}{
		{
			name:      "no wildcards",
			keys:      []string{"a", "bb", "ccc"},
			wildcards: nil,
			want:      3,
		},
		{
			name:      "single wildcard shorter",
			keys:      []string{"p%t", "xx"},
			wildcards: [][]string{{"Z"}},
			want:      3,
		},
		{
			name:      "wildcard longer than key",
			keys:      []string{"p%t"},
			wildcards: [][]string{{"YY"}},
			want:      4,
		},
		{
			name:      "multiple wildcards, picks longest",
			keys:      []string{"p%t"},
			wildcards: [][]string{{"X", "WWW"}},
			want:      5,
		},
		{
			name:      "mix wild and non‑wild",
			keys:      []string{"long", "m%x"},
			wildcards: [][]string{{"ZZZ"}},
			want:      5,
		},
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

func TestScrapeReadmeBanners(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	// Create README in dir1
	os.WriteFile(filepath.Join(dir1, "README.md"),
		[]byte("# "+filepath.Base(dir1)+": Hello\nMore"), 0o644)

	out := scrapeReadmeBanners([]string{}, []string{dir1, dir2})
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
	if banner, ok := out[dir1]; !ok || !strings.Contains(banner, "Hello") {
		t.Errorf("dir1 banner = %q, want contains Hello", banner)
	}
	if banner, ok := out[dir2]; !ok || banner != "" {
		t.Errorf("dir2 banner = %q, want empty", banner)
	}
}

// --- printTargets tests ---

func TestPrintTargets_Simple(t *testing.T) {
	targets := map[string]string{"a": "", "b": "desc"}
	var buf bytes.Buffer
	printTargets(&buf, targets, nil, nil)

	want := "" +
		"  a\n" +
		"  b   <-- desc\n"
	if buf.String() != want {
		t.Errorf("got:\n%q\nwant:\n%q", buf.String(), want)
	}
}

func TestPrintTargets_Wildcards(t *testing.T) {
	targets := map[string]string{"x%y": "hi%there"}
	wc := [][]string{{"Z"}}
	banners := map[string]string{"Z": " (ban)"}
	var buf bytes.Buffer

	printTargets(&buf, targets, wc, banners)

	want := "  xZy   <-- hiZthere (ban)\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

// --- printSubdirs tests ---

func TestPrintSubdirs(t *testing.T) {
	tmp := t.TempDir()
	sub1 := filepath.Join(tmp, "sub1")
	sub2 := filepath.Join(tmp, "sub2")
	os.MkdirAll(sub1, 0o755)
	os.MkdirAll(sub2, 0o755)

	// sub1 has a help target
	os.WriteFile(filepath.Join(sub1, "Makefile"), []byte("help:\n"), 0o644)
	// sub2 has a non-help target
	os.WriteFile(filepath.Join(sub2, "Makefile"), []byte("all:\n"), 0o644)

	// banners empty
	banners := map[string]string{sub1: "", sub2: ""}
	var buf bytes.Buffer
	printSubdirs(&buf, "base", []string{sub2, sub1}, banners)

	out := buf.String()
	lines := strings.Split(out, "\n")
	// Basic sanity checks:
	if !strings.Contains(lines[1], "Sub‑directories") {
		t.Errorf("missing header, got %q", lines[1])
	}
	// sub1: should be marked "*"
	if !strings.Contains(out, " *  make -C "+filepath.ToSlash(filepath.Join("base", sub1))) {
		t.Errorf("sub1 line missing star: %q", out)
	}
	// sub2: should not be marked "*"
	if !strings.Contains(out, "    make -C "+filepath.ToSlash(filepath.Join("base", sub2))) {
		t.Errorf("sub2 line missing invocation: %q", out)
	}
	// final note
	if !strings.Contains(out, "Is documented with a `help` target") {
		t.Errorf("missing final note: %q", out)
	}
}

// --- run() tests (covers main) ---

func TestRun_Success(t *testing.T) {
	dir := t.TempDir()
	mf := filepath.Join(dir, "Makefile")
	os.WriteFile(mf, []byte("a:\n"), 0o644)

	var out, errb bytes.Buffer
	err := run([]string{mf}, &out, &errb)
	if err != nil {
		t.Fatalf("error = %#v; want nil", err)
	}
	if !strings.Contains(out.String(), "Available make targets:") {
		t.Errorf("stdout missing header: %q", out.String())
	}
	if errb.Len() != 0 {
		t.Errorf("stderr unexpectedly got %q", errb.String())
	}
}

func TestRun_ParseError(t *testing.T) {
	var out, errb bytes.Buffer
	err := run([]string{}, &out, &errb)
	if err == nil {
		t.Fatal("expected non-nil error on parse error")
	}
	if !strings.Contains(err.Error(), "must specify exactly one Makefile") {
		t.Errorf("wrong error: err = %#v", err)
	}
}

func Test_ErrorToExitCode(t *testing.T) {
	cases := []struct {
		name             string
		errIn            error
		expectedStderr   string
		expectedExitCode int
	}{
		{
			name:             "no error",
			errIn:            nil,
			expectedStderr:   "",
			expectedExitCode: 0,
		},
		{
			name:             "string error",
			errIn:            errors.New("Some error"),
			expectedStderr:   "Error: Some error\n",
			expectedExitCode: 1,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var errb bytes.Buffer
			code := ErrorToExitCode(tc.errIn, &errb)
			if code != tc.expectedExitCode {
				t.Errorf("Expected %#v exit code, got %#v", tc.expectedExitCode, code)
			}
			if errb.String() != tc.expectedStderr {
				t.Errorf("Expected %#v stderr; stderr was %#v", tc.expectedStderr, errb.String())
			}
		})
	}
}
