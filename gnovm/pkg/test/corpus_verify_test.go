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
