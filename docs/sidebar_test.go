package docs

import (
	"strings"
	"testing"
)

func TestParseSidebar(t *testing.T) {
	t.Parallel()

	src := []byte(`# Gno.land

Welcome to the official documentation of Gno.land.

intro prose, not in the sidebar.

## Use Gno.land

Some lead paragraph.

- [Discover Gno.land](users/discover-gnoland.md) - Discover the ecosystem.
- [Using ` + "`gnoweb`" + `](users/explore-with-gnoweb.md) - Browse realms.

## Resources

- [Effective Gno](resources/effective-gno.md) - Best practices.
- [Gno Examples](https://github.com/gnolang/gno/tree/master/examples) - A library.
`)

	got := parseSidebar(src)
	if len(got) != 2 {
		t.Fatalf("want 2 sections, got %d", len(got))
	}
	if got[0].Title != "Use Gno.land" {
		t.Errorf("section[0].Title = %q", got[0].Title)
	}
	if len(got[0].Items) != 2 {
		t.Fatalf("section[0] want 2 items, got %d", len(got[0].Items))
	}
	if got[0].Items[0].Href != "users/discover-gnoland.md" {
		t.Errorf("href[0] = %q", got[0].Items[0].Href)
	}
	if got[0].Items[1].Title != "Using `gnoweb`" {
		t.Errorf("title with backticks preserved, got %q", got[0].Items[1].Title)
	}
	if got[1].Items[1].Href != "https://github.com/gnolang/gno/tree/master/examples" {
		t.Errorf("external href = %q", got[1].Items[1].Href)
	}
	if !got[1].Items[1].External {
		t.Error("external item should be flagged External")
	}
}

func TestSidebarFromReadme(t *testing.T) {
	t.Parallel()

	// The real README.md must yield the three canonical categories. If a
	// category is renamed in README, update this test (and the gnoweb
	// component that depends on the labels).
	want := []string{"Use Gno.land", "Build on Gno.land", "Resources"}
	got := Sidebar()
	if len(got) != len(want) {
		t.Fatalf("want %d sections, got %d", len(want), len(got))
	}
	for i, w := range want {
		if got[i].Title != w {
			t.Errorf("section[%d] = %q, want %q", i, got[i].Title, w)
		}
		if len(got[i].Items) == 0 {
			t.Errorf("section %q has no items", got[i].Title)
		}
	}

	// Each non-external item must point at an embedded .md file.
	for _, sec := range got {
		for _, it := range sec.Items {
			if it.External {
				continue
			}
			if !strings.HasSuffix(it.Href, ".md") {
				t.Errorf("%q: non-external item href %q has no .md suffix", it.Title, it.Href)
			}
		}
	}
}
