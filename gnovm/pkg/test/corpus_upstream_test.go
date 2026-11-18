package test

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCorpusBodiesMatchUpstream pins that every corpus file is the
// upstream file verbatim plus the harness's trailing golden region: the
// body (region stripped, CRLF and trailing newlines normalised) must hash
// to the entry UPSTREAM.sha256 records for it, and every upstream-layout
// file must have an entry. Regenerate the manifest with
// upstream-manifest.sh against the release named in UPSTREAM.
func TestCorpusBodiesMatchUpstream(t *testing.T) {
	root := filepath.Join("..", "..", "tests", "files", "gocorpus")
	want := map[string]string{}
	if f, err := os.Open(filepath.Join(root, "UPSTREAM.sha256")); err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if sum, rel, ok := strings.Cut(sc.Text(), "  "); ok {
				want[rel] = sum
			}
		}
	}
	seen := map[string]bool{}
	testdata := filepath.Join(root, "testdata")
	err := filepath.WalkDir(testdata, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".go") {
			return err
		}
		rel := filepath.ToSlash(strings.TrimPrefix(p, testdata+string(filepath.Separator)))
		switch strings.SplitN(rel, "/", 2)[0] {
		case "run", "compile", "errorcheck", "gno": // canaries, not upstream
			return nil
		}
		seen[rel] = true
		src, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		sum := sha256.Sum256([]byte(upstreamBody(string(src))))
		got := hex.EncodeToString(sum[:])
		switch w, ok := want[rel]; {
		case !ok:
			t.Errorf("%s: not in UPSTREAM.sha256; run upstream-manifest.sh", rel)
		case w != got:
			t.Errorf("%s: body differs from upstream (sha256 %s, manifest %s)", rel, got[:12], w[:12])
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for rel := range want {
		if !seen[rel] {
			t.Errorf("%s: in UPSTREAM.sha256 but not in testdata; run upstream-manifest.sh", rel)
		}
	}
}

// upstreamBody is a corpus file as the manifest hashes it: golden region
// stripped, CRLF and trailing newlines normalised. CRLF goes first, or the
// blank line above the region leaves a stray \r behind the strip.
func upstreamBody(src string) string {
	src = strings.ReplaceAll(src, "\r\n", "\n")
	return strings.TrimRight(stripTrailingGoldenRegion(src), "\n") + "\n"
}

func TestUpstreamBody(t *testing.T) {
	t.Parallel()
	const body = "// run\n\npackage main\n\nfunc main() {}\n"
	const tail = "\n// GnoOutput:\n\n// GoOutput:\n\n// Verdict: Clean run\n"
	crlf := strings.ReplaceAll(body+tail, "\n", "\r\n")
	for name, src := range map[string]string{
		"lf":                 body + tail,
		"crlf whole file":    crlf,
		"crlf body only":     strings.ReplaceAll(body, "\n", "\r\n") + tail,
		"extra newlines":     body + "\n\n" + tail + "\n\n",
		"no trailing region": body,
	} {
		if got := upstreamBody(src); got != body {
			t.Errorf("%s: upstreamBody = %q, want %q", name, got, body)
		}
	}
}

// A trailing note that is not a golden directive is body, not region, so
// the manifest covers it.
func TestStripTrailingGoldenRegion_KeepsNotes(t *testing.T) {
	t.Parallel()
	const body = "package main\n\nfunc main() {}\n\n// Tracked: issue #1\n"
	const tail = "\n// GoOutput:\n\n// Verdict: Clean run\n"
	if got := stripTrailingGoldenRegion(body + tail); strings.TrimRight(got, "\n") != strings.TrimRight(body, "\n") {
		t.Errorf("stripTrailingGoldenRegion dropped the note: %q", got)
	}
}
