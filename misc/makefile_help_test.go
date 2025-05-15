package main

import (
	// "bytes"
	// "io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	cfg, _, err := parseArgs([]string{"-relative-to", "foo/bar", "-dir", "dir1", "-dir", "dir2", "-wildcard", "val1", "-wildcard", "val2", "Makefile"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.RelativeTo != "foo/bar" {
		t.Errorf("RelativeTo = %q; want %q", cfg.RelativeTo, "foo/bar")
	}
	if len(cfg.Dirs.vals) != 2 || cfg.Dirs.vals[0] != "dir1" || cfg.Dirs.vals[1] != "dir2" {
		t.Errorf("Dirs = %v; want [dir1 dir2]", cfg.Dirs)
	}
	if len(cfg.Wildcards.vals) != 2 || cfg.Wildcards.vals[0] != "val1" || cfg.Wildcards.vals[1] != "val2" {
		t.Errorf("Wildcards = %v; want [val1 val2]", cfg.Wildcards)
	}
	if cfg.Makefile != "Makefile" {
		t.Errorf("Makefile = %q; want \"Makefile\"", cfg.Makefile)
	}
}

func TestExtractTargets(t *testing.T) {
	content := `
foo: #desc1
bar:desc2
baz: #desc3 with spaces
legacy: value # some stuff@LEGACY
NotA: #valid
` +
		"1no: #skip" + "\n"
	tmp, err := ioutil.TempFile("", "mk")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	targets, err := extractTargets(tmp.Name())
	if err != nil {
		t.Fatalf("extractTargets error: %v", err)
	}
	expects := map[string]string{
		"foo":  "desc1",
		"bar":  "",
		"baz":  "desc3 with spaces",
		"NotA": "valid",
	}
	for name, want := range expects {
		got, ok := targets[name]
		if !ok {
			t.Errorf("missing target %q", name)
			continue
		}
		if got != want {
			t.Errorf("desc[%q] = %q; want %q", name, got, want)
		}
	}
	if _, exists := targets["legacy"]; exists {
		t.Error("legacy target should be skipped")
	}
	if _, exists := targets["1no"]; exists {
		t.Error("non-letter-leading target should be skipped")
	}
}

func TestMaxKeyLength(t *testing.T) {
	keys := []string{"a", "abcd", "xyz"}
	if got := maxKeyLength(keys, []string{}); got != 4 {
		t.Errorf("maxKeyLength = %d; want %d", got, 4)
	}
	// empty slice
	if got := maxKeyLength([]string{}, []string{}); got != 0 {
		t.Errorf("maxKeyLength(empty) = %d; want %d", got, 0)
	}
}

func TestReadReadmeBanner(t *testing.T) {
	dir, err := ioutil.TempDir("", "d")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	// no README
	if b, err := readReadmeBanner(dir); err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if b != "" {
		t.Errorf("banner = %q; want empty", b)
	}
	// with README
	path := filepath.Join(dir, "README.md")
	ioutil.WriteFile(path, []byte("# Hello Banner\nSecond line"), 0644)
	b, err := readReadmeBanner(dir)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(b, "Hello Banner") {
		t.Errorf("banner %q missing content", b)
	}
	if !strings.HasPrefix(b, " (") || !strings.HasSuffix(b, ")") {
		t.Errorf("banner format = %q; want parenthesized", b)
	}
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	out, _ := ioutil.ReadAll(r)
	return string(out)
}

func TestDisplayTargets(t *testing.T) {
	targets := map[string]string{
		"a":   "one",
		"b%":  "two",
		"long": "three",
	}
	wilds := []string{"X"}
	out := captureOutput(func() {
		displayTargets(targets, wilds,map[string]string{})
	})
	lines := strings.Split(strings.TrimSpace(out), "\n")
	// expect line for "a", "long", and expanded "bX"
	if len(lines) != 3 {
		t.Fatalf("got %d lines; want %d", len(lines), 3)
	}
	if !strings.Contains(lines[0], "<-- one") {
		t.Errorf("missing one: %q", lines[0])
	}
	if !strings.Contains(lines[2], "<-- three") {
		t.Errorf("missing three: %q", lines[1])
	}
	if !strings.Contains(lines[1], "bX") || !strings.Contains(lines[1], "<-- two") {
		t.Errorf("missing expanded bX two: %q", lines[2])
	}
}

func TestDisplayDirs(t *testing.T) {
	// create dirs
	d1, _ := ioutil.TempDir("", "d1")
	d2, _ := ioutil.TempDir("", "d2")
	defer os.RemoveAll(d1)
	defer os.RemoveAll(d2)
	// write Makefile in d1 with help
	ioutil.WriteFile(filepath.Join(d1, "Makefile"), []byte("help: #desc"), 0644)
	// write Makefile in d2 without help
	ioutil.WriteFile(filepath.Join(d2, "Makefile"), []byte("foo: #bar"), 0644)
	// write README in d1
	ioutil.WriteFile(filepath.Join(d1, "README.md"), []byte("Title1"), 0644)

	scraped := scrapeReadmeBanners([]string{d1},[]string{d1, d2})

	out := captureOutput(func() {
		displayDirs("", []string{d1, d2},scraped)
	})
	if !strings.Contains(out, "*") {
		t.Error("expected '*' for help in output:", out)
	}
	if !strings.Contains(out, "Title1") {
		t.Error("expected README banner in output:", out)
	}
}
