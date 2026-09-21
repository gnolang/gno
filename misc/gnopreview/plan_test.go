package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// fakeRepo writes a miniature monorepo: examples/gno.land holds the packages,
// examples/quarantined mirrors the layout for packages that must never be
// previewed.
func fakeRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(p, body string) {
		full := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("examples/gnowork.toml", "")

	pkg := func(dir, module, body string) {
		write(dir+"/gnomod.toml", "module = \""+module+"\"\n")
		write(dir+"/lib.gno", body)
	}
	// p/base <- p/mid <- r/leaf ; r/other imports p/base directly
	pkg("examples/gno.land/p/x/base/v0", "gno.land/p/x/base/v0", "package base\n")
	pkg("examples/gno.land/p/x/mid/v0", "gno.land/p/x/mid/v0",
		"package mid\nimport \"gno.land/p/x/base/v0\"\n")
	pkg("examples/gno.land/r/x/leaf", "gno.land/r/x/leaf",
		"package leaf\nimport \"gno.land/p/x/mid/v0\"\n")
	pkg("examples/gno.land/r/x/other", "gno.land/r/x/other",
		"package other\nimport \"gno.land/p/x/base/v0\"\n")
	pkg("examples/gno.land/r/x/lonely", "gno.land/r/x/lonely", "package lonely\n")
	// VM fixtures: reachable from p/x/base but never previewed indirectly
	pkg("examples/gno.land/r/tests/vm/fixture", "gno.land/r/tests/vm/fixture",
		"package fixture\nimport \"gno.land/p/x/base/v0\"\n")
	// ignored + quarantined must not show up at all
	write("examples/gno.land/r/x/ignored/gnomod.toml",
		"module = \"gno.land/r/x/ignored\"\nignore = true\n")
	write("examples/gno.land/r/x/ignored/lib.gno", "package ignored\nimport \"gno.land/p/x/base/v0\"\n")
	// same, with the flag carrying a trailing comment
	write("examples/gno.land/r/x/ignored2/gnomod.toml",
		"module = \"gno.land/r/x/ignored2\"\nignore = true # quarantined for now\n")
	write("examples/gno.land/r/x/ignored2/lib.gno", "package ignored2\n")
	pkg("examples/quarantined/gno.land/r/x/quar", "gno.land/r/x/quar",
		"package quar\nimport \"gno.land/p/x/base/v0\"\n")
	return root
}

func TestBuildPlan(t *testing.T) {
	t.Parallel()
	root := fakeRepo(t)
	for _, tc := range []struct {
		name       string
		changed    []string
		maxRealms  int
		wantGnoweb bool
		wantRealms []string
		wantMode   string
		wantDrop   int
	}{
		{
			name:       "nothing relevant",
			changed:    []string{"docs/x.md", "gnovm/pkg/gnolang/y.go"},
			wantRealms: []string{},
			wantMode:   "none",
		},
		{
			name:       "one realm",
			changed:    []string{"examples/gno.land/r/x/leaf/lib.gno"},
			wantRealms: []string{"gno.land/r/x/leaf"},
			wantMode:   "realms",
		},
		{
			name:       "transitive dependency pulls both dependents",
			changed:    []string{"examples/gno.land/p/x/base/v0/lib.gno"},
			wantRealms: []string{"gno.land/r/x/leaf", "gno.land/r/x/other"},
			wantMode:   "realms",
		},
		{
			name:       "tests fixtures are not pulled in indirectly",
			changed:    []string{"examples/gno.land/p/x/mid/v0/lib.gno"},
			wantRealms: []string{"gno.land/r/x/leaf"},
			wantMode:   "realms",
		},
		{
			name:       "a changed test realm is still previewed",
			changed:    []string{"examples/gno.land/r/tests/vm/fixture/lib.gno"},
			wantRealms: []string{"gno.land/r/tests/vm/fixture"},
			wantMode:   "realms",
		},
		{
			name:       "quarantined packages never appear",
			changed:    []string{"examples/quarantined/gno.land/r/x/quar/lib.gno"},
			wantRealms: []string{},
			wantMode:   "none",
		},
		{
			name:       "tests and filetests do not trigger a preview",
			changed:    []string{"examples/gno.land/r/x/leaf/lib_test.gno", "examples/gno.land/r/x/leaf/z_filetest.gno"},
			wantRealms: []string{},
			wantMode:   "none",
		},
		{
			name:       "gnoweb alone",
			changed:    []string{"gno.land/pkg/gnoweb/app.go"},
			wantGnoweb: true,
			wantRealms: []string{}, // the seed realms do not exist in the fake repo
			wantMode:   "gnoweb",
		},
		{
			name:       "gnoweb and a realm",
			changed:    []string{"gno.land/pkg/gnoweb/public/main.css", "examples/gno.land/r/x/leaf/lib.gno"},
			wantGnoweb: true,
			wantRealms: []string{"gno.land/r/x/leaf"},
			wantMode:   "both",
		},
		{
			name:       "an ignored realm is not previewed even when it changed",
			changed:    []string{"examples/gno.land/r/x/ignored/lib.gno"},
			wantRealms: []string{},
			wantMode:   "none",
		},
		{
			name:       "ignore = true still holds with a comment after it",
			changed:    []string{"examples/gno.land/r/x/ignored2/lib.gno"},
			wantRealms: []string{},
			wantMode:   "none",
		},
		{
			name:       "the cap keeps the changed realm and restores alphabetical order",
			changed:    []string{"examples/gno.land/p/x/base/v0/lib.gno", "examples/gno.land/r/x/other/lib.gno"},
			maxRealms:  2,
			wantRealms: []string{"gno.land/r/x/leaf", "gno.land/r/x/other"},
			wantMode:   "realms",
		},
		{
			name:       "cap keeps the changed realm and reports the rest",
			changed:    []string{"examples/gno.land/p/x/base/v0/lib.gno", "examples/gno.land/r/x/other/lib.gno"},
			maxRealms:  1,
			wantRealms: []string{"gno.land/r/x/other"},
			wantMode:   "realms",
			wantDrop:   1,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			limit := tc.maxRealms
			if limit == 0 {
				limit = defaultMaxRealms
			}
			got, err := BuildPlan(root, tc.changed, limit)
			if err != nil {
				t.Fatal(err)
			}
			if got.Gnoweb != tc.wantGnoweb {
				t.Errorf("Gnoweb = %v; want %v", got.Gnoweb, tc.wantGnoweb)
			}
			if !reflect.DeepEqual(got.Realms, tc.wantRealms) {
				t.Errorf("Realms = %v; want %v", got.Realms, tc.wantRealms)
			}
			if got.Mode() != tc.wantMode {
				t.Errorf("Mode = %q; want %q", got.Mode(), tc.wantMode)
			}
			if got.Dropped != tc.wantDrop {
				t.Errorf("Dropped = %d; want %d", got.Dropped, tc.wantDrop)
			}
			if got.Empty() != (tc.wantMode == "none") {
				t.Errorf("Empty = %v; want %v", got.Empty(), tc.wantMode == "none")
			}
			// render pairs plan.Dirs[i] with plan.Realms[i] to find the realm
			// in the merge-base checkout, so the two slices have to stay
			// index-aligned: a drift renders one realm under another's name.
			if len(got.Dirs) != len(got.Realms) {
				t.Fatalf("Dirs = %d; want one per realm (%d)", len(got.Dirs), len(got.Realms))
			}
			for i, r := range got.Realms {
				if want := "examples/" + r; got.Dirs[i] != want {
					t.Errorf("Dirs[%d] = %q; want %q", i, got.Dirs[i], want)
				}
			}
		})
	}
}

func TestRenderRelevant(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"examples/gno.land/r/x/y/a.gno", true},
		{"examples/gno.land/r/x/y/gnomod.toml", true},
		{"examples/gno.land/r/x/y/a_test.gno", false},
		{"examples/gno.land/r/x/y/a_filetest.gno", false},
		{"examples/gno.land/r/x/y/filetests/a.gno", false},
		{"examples/gno.land/r/x/y/README.md", false},
	} {
		if got := renderRelevant(tc.in); got != tc.want {
			t.Errorf("renderRelevant(%q) = %v; want %v", tc.in, got, tc.want)
		}
	}
}

func TestCommentEmptyIsSilent(t *testing.T) {
	t.Parallel()
	// A PR with nothing to preview must produce no comment at all.
	if got := Comment(&Plan{}, "https://example.test/pr-1", "1"); got != "" {
		t.Errorf("Comment(empty plan) = %q; want empty", got)
	}
}

func TestCommentRealms(t *testing.T) {
	t.Parallel()
	p := &Plan{
		ChangedRealms: []string{"gno.land/r/x/leaf"},
		ChangedPkgs:   []string{"gno.land/p/x/base/v0"},
		Realms:        []string{"gno.land/r/x/leaf", "gno.land/r/x/other"},
		Dropped:       3,
	}
	got := Comment(p, "https://example.test/pr-7/", "7")
	for _, want := range []string{
		CommentMarker,
		"**Changed realms (1)**",
		"https://example.test/pr-7/r/x/leaf/",
		"https://example.test/pr-7/r/x/leaf/_t/source/",
		"**Realms affected through a changed package (1)**",
		"gno.land/p/x/base/v0",
		"3 more affected realm(s) were **not** rendered",
		"PR #7 closes",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("comment missing %q\n---\n%s", want, got)
		}
	}
	// Screenshots are a gnoweb-change affordance only.
	if strings.Contains(got, "<table>") {
		t.Errorf("realm-only comment should carry no screenshots:\n%s", got)
	}
}

func TestCommentGnowebHasShots(t *testing.T) {
	t.Parallel()
	p := &Plan{
		Gnoweb: true,
		Realms: []string{"gno.land/r/gnoland/home"},
		Shots:  []Shot{{File: "_shots/home.png", Label: "Home"}},
	}
	got := Comment(p, "https://example.test/pr-9", "9")
	for _, want := range []string{
		"changes **gnoweb itself**",
		`<img src="https://example.test/pr-9/_shots/home.png"`,
		"https://example.test/pr-9/r/gnoland/home/",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("comment missing %q\n---\n%s", want, got)
		}
	}
}

func TestCommentBeforeAfter(t *testing.T) {
	t.Parallel()
	p := &Plan{
		ChangedRealms: []string{"gno.land/r/x/leaf", "gno.land/r/x/fresh"},
		Realms:        []string{"gno.land/r/x/fresh", "gno.land/r/x/leaf"},
		Pairs: []ShotPair{
			{Realm: "gno.land/r/x/leaf", Before: "_shots/r-x-leaf-before.png", After: "_shots/r-x-leaf-after.png", URL: "r/x/leaf/"},
			{Realm: "gno.land/r/x/fresh", After: "_shots/r-x-fresh-after.png", URL: "r/x/fresh/", New: true},
		},
	}
	got := Comment(p, "https://example.test/pr-5", "5")
	for _, want := range []string{
		`<img src="https://example.test/pr-5/_shots/r-x-leaf-before.png"`,
		"before — merge base",
		"<b>after — this PR</b>",
		// a realm the PR adds has no before, and the comment says why
		"New in this PR — nothing to compare against.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("comment missing %q\n---\n%s", want, got)
		}
	}
	// The gnoweb sample and the before/after pairs are alternatives, never both.
	if strings.Contains(got, "Home — rendered markdown") {
		t.Errorf("realm change should not carry the gnoweb sample:\n%s", got)
	}
}

func TestCommentNoBaseMakesNoClaim(t *testing.T) {
	t.Parallel()
	// Before is empty because no merge-base checkout was supplied, not because
	// the realm is new. The comment must not say it is new.
	p := &Plan{
		ChangedRealms: []string{"gno.land/r/x/leaf"},
		Realms:        []string{"gno.land/r/x/leaf"},
		Pairs:         []ShotPair{{Realm: "gno.land/r/x/leaf", After: "_shots/a.png", URL: "r/x/leaf/"}},
	}
	got := Comment(p, "https://example.test/pr-1", "1")
	if strings.Contains(got, "New in this PR") {
		t.Errorf("comment claims the realm is new when no base was rendered:\n%s", got)
	}
	if !strings.Contains(got, "_shots/a.png") {
		t.Errorf("comment dropped the after screenshot:\n%s", got)
	}
}

// A trailing comment is not part of the value. Without the cut, a package
// carrying `ignore = true # why` reads as neither true nor false and is
// previewed as if it were live.
func TestModFlagsTrailingComment(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	for _, tc := range []struct {
		name               string
		body               string
		wantDraft, wantIgn bool
	}{
		{"plain", "module = \"x\"\nignore = true\n", false, true},
		{"commented", "module = \"x\"\nignore = true # quarantined\n", false, true},
		{"draft commented", "module = \"x\"\ndraft = true  # wip\n", true, false},
		{"false stays false", "module = \"x\"\nignore = false # not yet\n", false, false},
		{"neither", "module = \"x\"\n", false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := filepath.Join(dir, strings.ReplaceAll(tc.name, " ", "_")+".toml")
			if err := os.WriteFile(f, []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			draft, ign := modFlags(f)
			if draft != tc.wantDraft || ign != tc.wantIgn {
				t.Errorf("modFlags = (draft %v, ignore %v); want (%v, %v)", draft, ign, tc.wantDraft, tc.wantIgn)
			}
		})
	}
}

// The landing page skips realms the crawl never captured, so its header has to
// count the rows it wrote and not the realms that were planned.
func TestIndexCountsOnlyCapturedRealms(t *testing.T) {
	t.Parallel()
	p := &Plan{Realms: []string{"gno.land/r/x/leaf", "gno.land/r/x/other"}}
	c := &Crawler{pages: map[string]*page{
		"/r/x/leaf": {File: "r/x/leaf/index.html"},
	}}
	got := Index(p, c)
	if !strings.Contains(got, "1 realm(s) rendered") {
		t.Errorf("header does not count the captured realm alone:\n%s", got)
	}
	if strings.Contains(got, "gno.land/r/x/other") {
		t.Error("an uncaptured realm got a row")
	}
}

// A realm the crawl never captured must not get a link: the snapshot holds no
// page for it, so the bullet would point at a 404 on the previews site.
func TestCommentMissedRealmIsNotLinked(t *testing.T) {
	t.Parallel()
	p := &Plan{
		Realms:        []string{"gno.land/r/x/leaf", "gno.land/r/x/other"},
		ChangedRealms: []string{"gno.land/r/x/leaf", "gno.land/r/x/other"},
		Missed:        []string{"gno.land/r/x/other"},
	}
	got := Comment(p, "https://example.test/pr-1", "1")
	if !strings.Contains(got, "[`gno.land/r/x/leaf`](https://example.test/pr-1/r/x/leaf/)") {
		t.Errorf("the captured realm lost its link:\n%s", got)
	}
	if strings.Contains(got, "(https://example.test/pr-1/r/x/other/)") {
		t.Errorf("the uncaptured realm was linked:\n%s", got)
	}
	if !strings.Contains(got, "`gno.land/r/x/other` (not rendered") {
		t.Errorf("the uncaptured realm was not called out:\n%s", got)
	}
}
