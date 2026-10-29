package test

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

// TestVerifyCorpusVerdicts runs .go corpus probes through verify (no
// sync) and pins what verify refuses: a feature-gap message under a
// Clean verdict is a stale verdict, never a skip; a second verdict line
// cannot hide behind an Unsupported first one; a Known* verdict names
// the file's mode.
func TestVerifyCorpusVerdicts(t *testing.T) {
	rootDir, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	opts := &TestOptions{RootDir: rootDir, Output: io.Discard, Error: io.Discard}
	opts.BaseStore, opts.TestStore = StoreWithOptions(rootDir, opts.WriterForStore(), StoreOptions{Testing: true})

	const clean = "package main\n\nfunc main() {\n\tprintln(\"x\")\n}\n"
	const gap = "package main\n\nfunc main() {\n\t_ = Version\n\tprintln(\"x\")\n}\n"
	const goldens = "\n// GnoOutput:\n// x\n\n// GoOutput:\n// x\n\n"
	for _, tc := range []struct {
		name, src, wantErr string
	}{
		{"clean passes", clean + goldens + "// Verdict: Clean run\n", ""},
		{"unsupported skips", gap + "\n// Verdict: Unsupported stdlib-symbol: Version\n", "skipped"},
		{"feature gap under Clean is stale, not a skip", gap + goldens + "// Verdict: Clean run\n", "stale verdict"},
		{"second verdict behind Unsupported", clean + "\n// Verdict: Unsupported channel: x\n\n// Verdict: Clean run\n", "exactly one"},
		{"Known* verdict names the mode", clean + goldens + "// Verdict: KnownIssue compile: #1 x\n", `mode "compile"`},
		// Unsupported is the one verdict that skips, so it is the one a
		// bad tag or a placeholder reason can hide behind.
		{"Unsupported tag must be in the vocabulary", gap + "\n// Verdict: Unsupported wibble-not-a-tag: I'd rather not look\n", "tag vocabulary"},
		{"Unsupported needs a real reason", gap + "\n// Verdict: Unsupported stdlib-symbol: TODO: explain\n", "needs a reason"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, _, err := opts.RunFiletest("gocorpus/testdata/run/probe.go", []byte(tc.src), opts.TestStore)
			var skip *SkipError
			if errors.As(err, &skip) != (tc.wantErr == "skipped") {
				t.Fatalf("skip = %v, want %v (err %v)", skip != nil, tc.wantErr == "skipped", err)
			}
			switch {
			case tc.wantErr == "" && err != nil:
				t.Fatalf("unexpected error: %v", err)
			case tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
				t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

// TestSyncRefusesDowngradeToUnsupported pins that sync will not quietly
// move a file that reached a verdict into `Unsupported`. That transition
// is a regression — Gno used to get a result here — and Unsupported is
// skipped, so the file and its evidence would leave the ledger with CI
// still green.
func TestSyncRefusesDowngradeToUnsupported(t *testing.T) {
	rootDir, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	opts := &TestOptions{RootDir: rootDir, Output: io.Discard, Error: io.Discard, Sync: true}
	opts.BaseStore, opts.TestStore = StoreWithOptions(rootDir, opts.WriterForStore(), StoreOptions{Testing: true})

	// A file Gno now rejects for a feature gap, carrying a prior verdict.
	const gap = "package main\n\nfunc main() {\n\t_ = Version\n\tprintln(\"x\")\n}\n"
	const goldens = "\n// GnoOutput:\n// x\n\n// GoOutput:\n// x\n\n"
	for _, prior := range []string{"Clean run", "KnownIssue run: #1 x", "KnownDivergence run: #1 x"} {
		t.Run(prior, func(t *testing.T) {
			src := gap + goldens + "// Verdict: " + prior + "\n"
			_, _, _, err := opts.RunFiletest("testdata/run/probe.go", []byte(src), opts.TestStore)
			if err == nil || !strings.Contains(err.Error(), "refusing to downgrade") {
				t.Fatalf("sync silently downgraded `%s` to Unsupported (err %v)", prior, err)
			}
		})
	}
	t.Run("a file already Unsupported is still normalised", func(t *testing.T) {
		src := gap + "\n// Verdict: Unsupported stdlib-symbol: Version\n"
		if _, _, _, err := opts.RunFiletest("testdata/run/probe.go", []byte(src), opts.TestStore); err != nil {
			var skip *SkipError
			if !errors.As(err, &skip) {
				t.Fatalf("an already-Unsupported file must still sync: %v", err)
			}
		}
	})
}

// TestVerdictArmsTheGoComparison pins that EVERY verdict opts a .gno file
// into the Gno-vs-Go comparison, not just KnownDivergence. A verdict is a
// claim about what gc does; without a `// GoOutput:` beside it nothing
// ever checks that claim, and the file leaves the ledger silently.
func TestVerdictArmsTheGoComparison(t *testing.T) {
	rootDir, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	opts := &TestOptions{RootDir: rootDir, Output: io.Discard, Error: io.Discard}
	opts.BaseStore, opts.TestStore = StoreWithOptions(rootDir, opts.WriterForStore(), StoreOptions{Testing: true})

	const body = "package main\n\nfunc main() {\n\tprintln(\"x\")\n}\n\n// Output:\n// x\n"
	for _, kind := range []string{"Clean run", "KnownIssue run: #1 x", "KnownDivergence run: #1 x"} {
		t.Run(kind, func(t *testing.T) {
			src := body + "\n// Verdict: " + kind + "\n"
			_, _, _, err := opts.RunFiletest("probe.gno", []byte(src), opts.TestStore)
			if err == nil {
				t.Fatalf("`// Verdict: %s` with no `// GoOutput:` passed verify", kind)
			}
			t.Logf("refused with: %v", err)
		})
	}
	t.Run("no verdict is still pure-Gno", func(t *testing.T) {
		if _, _, _, err := opts.RunFiletest("probe.gno", []byte(body), opts.TestStore); err != nil {
			t.Fatalf("a .gno file with no verdict must keep its pure-Gno behaviour: %v", err)
		}
	})
}
