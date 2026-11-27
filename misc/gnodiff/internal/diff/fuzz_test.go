package diff

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// FuzzPrograms mutates seed sources with go test -fuzz. It is skipped during
// an ordinary go test so the corpus does not shell out to gc on every CI run.
// Inputs that fall outside the Gno subset, or that fail to parse, are
// discarded: both toolchains agree those are not programs. A parsed program
// that behaves differently is a failure, and a line-reduced repro is written
// under gnodiff-out/fuzz (or $GNODIFF_OUT).
func FuzzPrograms(f *testing.F) {
	if !fuzzActive() {
		f.Skip("native fuzz is opt-in: go test -fuzz=FuzzPrograms ./internal/diff/")
	}
	progs, err := LoadDir(CorpusDir())
	if err != nil {
		f.Fatal(err)
	}
	for _, prog := range progs {
		f.Add(prog.Source)
	}
	opt := Options{Timeout: 8 * time.Second}
	f.Fuzz(func(t *testing.T, src string) {
		if len(src) == 0 || len(src) > 8<<10 || !utf8.ValidString(src) {
			return
		}
		if !strings.Contains(src, "package main") || !strings.Contains(src, "func main") {
			return
		}
		if Unsupported(src) != "" || !parses(src) {
			return
		}
		goRes := RunGo(context.Background(), src, opt)
		gnoRes := RunGno(src, opt)
		if goRes.Kind == KindInternal || gnoRes.Kind == KindInternal {
			t.Fatalf("runner failed\n--- go ---\n%s\n--- gno ---\n%s", goRes.Summary(), gnoRes.Summary())
		}
		match, reason := Compare(goRes, gnoRes)
		if match {
			return
		}
		reduced := Minimize(src, 12, func(candidate string) bool {
			if Unsupported(candidate) != "" || !parses(candidate) {
				return false
			}
			left := RunGo(context.Background(), candidate, opt)
			right := RunGno(candidate, opt)
			ok, _ := Compare(left, right)
			return !ok && left.Kind != KindInternal && right.Kind != KindInternal
		})
		path := writeFuzzRepro(t, reduced)
		t.Errorf("divergence (%s)\n--- go ---\n%s\n--- gno ---\n%s\nrepro: %s\n%s",
			reason, goRes.Summary(), gnoRes.Summary(), path, reduced)
	})
}

func parses(src string) bool {
	_, err := parser.ParseFile(token.NewFileSet(), "main.go", src, parser.SkipObjectResolution)
	return err == nil
}

func fuzzActive() bool {
	if os.Getenv("GNODIFF_FUZZ") == "1" {
		return true
	}
	for _, arg := range os.Args {
		if arg == "-test.fuzz" || strings.HasPrefix(arg, "-test.fuzz=") || arg == "-fuzz" || strings.HasPrefix(arg, "-fuzz=") {
			return true
		}
	}
	return false
}

func writeFuzzRepro(t *testing.T, src string) string {
	t.Helper()
	dir := os.Getenv("GNODIFF_OUT")
	if dir == "" {
		dir = "gnodiff-out"
	}
	dir = filepath.Join(dir, "fuzz")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Logf("mkdir repro: %v", err)
		return ""
	}
	sum := sha256.Sum256([]byte(src))
	path := filepath.Join(dir, hex.EncodeToString(sum[:6])+".go")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Logf("write repro: %v", err)
		return ""
	}
	return path
}
