package docs

import (
	"io/fs"
	"path"
	"regexp"
	"strings"
	"testing"
)

// TestEmbedKnownFiles guards against the embed glob silently dropping a
// well-known page. The list intentionally mirrors the top-level navigation
// in README.md so adding a category here makes the test fail loudly until
// someone updates both.
func TestEmbedKnownFiles(t *testing.T) {
	must := []string{
		"README.md",
		"CONSTITUTION.md",
		"LAWS.md",
		"MANIFESTO.md",
		"builders/getting-started.md",
		"users/discover-gnoland.md",
		"resources/effective-gno.md",
	}
	for _, name := range must {
		if _, err := fs.Stat(FS(), name); err != nil {
			t.Errorf("expected embedded file %q to exist: %v", name, err)
		}
	}
}

// linkRE captures Markdown links and images: [label](target) and ![alt](src).
// Submatch 1 is the target. Multi-line targets are not legal in CommonMark.
var linkRE = regexp.MustCompile(`!?\[[^\]]*\]\(([^)\s]+)\)`)

// TestInternalLinksResolve walks every embedded .md file and verifies that
// every relative .md link resolves to an embedded file. This is what
// docs.gno.land relies on its linter for today; here we get it at PR time
// for free via `go test`.
func TestInternalLinksResolve(t *testing.T) {
	var checked int
	err := fs.WalkDir(FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return err
		}
		src, err := fs.ReadFile(FS(), p)
		if err != nil {
			return err
		}
		dir := path.Dir(p)
		for _, m := range linkRE.FindAllSubmatch(src, -1) {
			target := string(m[1])
			if !isInternalMarkdownLink(target) {
				continue
			}
			// Strip anchor / query.
			if i := strings.IndexAny(target, "#?"); i >= 0 {
				target = target[:i]
			}
			if target == "" {
				continue
			}
			resolved := path.Clean(path.Join(dir, target))
			// Cross-repo references that escape docs/ are intentional;
			// they get rewritten to GitHub URLs at render time.
			if strings.HasPrefix(resolved, "..") {
				continue
			}
			if _, err := fs.Stat(FS(), resolved); err != nil {
				t.Errorf("%s: broken internal link %q (resolved to %q)", p, m[1], resolved)
			}
			checked++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if checked == 0 {
		t.Fatal("no internal links inspected; regex probably broken")
	}
	t.Logf("verified %d internal links", checked)
}

// isInternalMarkdownLink returns true for targets that point at another
// .md file in the embed (or an asset under images/_assets). External URLs,
// fragment-only links, mailto:, and absolute /docs/ URLs are excluded
// since they are not the responsibility of this test.
func isInternalMarkdownLink(target string) bool {
	switch {
	case target == "":
		return false
	case strings.HasPrefix(target, "http://"),
		strings.HasPrefix(target, "https://"),
		strings.HasPrefix(target, "//"),
		strings.HasPrefix(target, "#"),
		strings.HasPrefix(target, "mailto:"),
		strings.HasPrefix(target, "data:"):
		return false
	case strings.HasPrefix(target, "/"):
		// Absolute paths are out of scope; gnoweb resolves them at request time.
		return false
	}
	// Only check .md links to keep the test focused on prose pages. Image
	// and asset links are visually noisy when broken and can be added in
	// a follow-up.
	return strings.HasSuffix(strings.SplitN(target, "#", 2)[0], ".md")
}
