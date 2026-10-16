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
		body := strings.ReplaceAll(stripTrailingGoldenRegion(string(src)), "\r\n", "\n")
		body = strings.TrimRight(body, "\n") + "\n"
		sum := sha256.Sum256([]byte(body))
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
