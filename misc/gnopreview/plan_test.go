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
