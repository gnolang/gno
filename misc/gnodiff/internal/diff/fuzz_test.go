package diff

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/gnolang/gno/misc/gnodiff/internal/gen"
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
	// The seed corpus is not a fuzz input. FuzzGenerated is the entry that
	// spends time on programs both sides can run.
	f.Add("package main\n\nfunc main() {\n\tprintln(1)\n}\n")
	opt := Options{Timeout: defaultTimeout}
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
		lintRes := RunLint(src, opt)
		class, reason := Classify(Program{Source: src}, goRes, gnoRes, lintRes)
		if class == ClassInternal {
			t.Fatalf("runner failed (%s)\n--- go ---\n%s\n--- gno ---\n%s\n--- lint ---\n%s",
				reason, goRes.Summary(), gnoRes.Summary(), lintRes.Summary())
		}
		if class != ClassReal {
			return
		}
		reduced := minimizeDivergence(src, goRes, gnoRes, opt, 12)
		path := writeFuzzRepro(t, reduced)
		t.Errorf("divergence (%s)\n--- go ---\n%s\n--- gno ---\n%s\nrepro: %s\n%s",
			reason, goRes.Summary(), gnoRes.Summary(), path, reduced)
	})
}

// FuzzGenerated drives the AST generator. Byte mutation of seeds mostly
// fails to parse; this target spends its time on type-correct programs.
func FuzzGenerated(f *testing.F) {
	if !fuzzActive() {
		f.Skip("native fuzz is opt-in: go test -fuzz=FuzzGenerated ./internal/diff/")
	}
	f.Add(uint64(1))
	f.Add(uint64(2))
	opt := Options{Timeout: defaultTimeout}
	f.Fuzz(func(t *testing.T, seed uint64) {
		src := gen.Generate(seed)
		if Unsupported(src) != "" || !parses(src) {
			t.Fatalf("generator produced an out-of-subset program\n%s", src)
		}
		goRes := RunGo(context.Background(), src, opt)
		gnoRes := RunGno(src, opt)
		lintRes := RunLint(src, opt)
		class, reason := Classify(Program{Source: src}, goRes, gnoRes, lintRes)
		if class == ClassInternal {
			t.Fatalf("runner failed (%s)\n--- go ---\n%s\n--- gno ---\n%s\n--- lint ---\n%s",
				reason, goRes.Summary(), gnoRes.Summary(), lintRes.Summary())
		}
		if class != ClassReal {
			return
		}
		reduced := minimizeDivergence(src, goRes, gnoRes, opt, 12)
		path := writeFuzzRepro(t, reduced)
		t.Errorf("real divergence seed %d (%s)\n--- go ---\n%s\n--- gno ---\n%s\nrepro: %s\n%s",
			seed, reason, goRes.Summary(), gnoRes.Summary(), path, reduced)
	})
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
