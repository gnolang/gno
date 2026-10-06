package chainmap

import (
	"context"
	"strings"
	"testing"
	"time"
)

func render(t *testing.T, h *Handler, l Listing) string {
	t.Helper()
	var b strings.Builder
	if err := h.MapView(context.Background(), l).Render(&b); err != nil {
		t.Fatalf("render: %v", err)
	}
	return b.String()
}

// With no indexer the map says nothing about activity: no legend line, no
// footer, no shade, so nothing reads as "no calls".
func TestMapWithoutIndexerSaysNothingAboutActivity(t *testing.T) {
	t.Parallel()

	out := render(t, New(Deps{Imports: fakeImports{}}), Listing{Path: "/r", Paths: []string{"/r/a/x", "/r/a/y", "/r/b"}})
	for _, absent := range []string{"calls", "indexer", "b-map__tile--l"} {
		if strings.Contains(out, absent) {
			t.Errorf("map without an indexer mentions %q", absent)
		}
	}
	for _, want := range []string{`href="/r/a/x"`, `href="/r/a/y"`, `href="/r/b"`, `href="/r/a/$map"`, "Map · 3 Packages"} {
		if !strings.Contains(out, want) {
			t.Errorf("map lacks %q", want)
		}
	}
}

// The geometry must ride on attributes: gnoweb's CSP drops inline styles,
// which is how the first version of this map rendered as an empty box.
func TestMapUsesNoInlineStyle(t *testing.T) {
	t.Parallel()

	out := render(t, New(Deps{Imports: fakeImports{}}), Listing{Path: "/r", Paths: []string{"/r/a/x", "/r/b"}})
	if strings.Contains(out, "style=") {
		t.Fatal("the map carries a style attribute, which the Content-Security-Policy blocks")
	}
}

// Path segments come from the chain. Whatever they hold reaches the page as
// text, in labels, titles and links alike.
func TestMapEscapesPaths(t *testing.T) {
	t.Parallel()

	out := render(t, New(Deps{Imports: fakeImports{}}), Listing{Path: "/r", Paths: []string{`/r/x/"><script>alert(1)</script>`}})
	if strings.Contains(out, "<script>") {
		t.Fatal("a path reached the page unescaped")
	}
}

func TestMapSaysWhenTheListingIsTruncated(t *testing.T) {
	t.Parallel()

	out := render(t, New(Deps{Imports: fakeImports{}}), Listing{Path: "/r", Paths: []string{"/r/a"}, Truncated: true})
	if !strings.Contains(out, "Only the first 1 paths are listed") {
		t.Error("a truncated listing must say so on the map too")
	}
}

func TestMapColoursRealmsByCalls(t *testing.T) {
	t.Parallel()

	h := New(Deps{Indexer: &fakeIndexer{}, Imports: fakeImports{}, Domain: "gno.land"})
	h.activity.store("activity", &Activity{
		Calls:   map[string]int{"gno.land/r/a/busy": 100, "gno.land/r/a/some": 3},
		Callers: map[string]int{"gno.land/r/a/busy": 7, "gno.land/r/a/some": 1},
		From:    10, To: 20, Since: time.Unix(0, 0),
	}, nil)

	out := render(t, h, Listing{Path: "/r/a", Paths: []string{"/r/a/busy", "/r/a/some", "/r/a/idle"}})
	for _, want := range []string{
		"b-map__tile b-map__tile--l4",
		"/r/a/busy · 100 calls by 7 accounts",
		"/r/a/some · 3 calls by 1 account",
		"/r/a/idle · no calls",
		"from block 10",
		"last indexed block 20",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("activity map lacks %q", want)
		}
	}
}

// Over a partially read window a zero is not a zero, and a count is a floor.
func TestMapPartialWindowNeverShowsZero(t *testing.T) {
	t.Parallel()

	h := New(Deps{Indexer: &fakeIndexer{}, Imports: fakeImports{}, Domain: "gno.land"})
	h.activity.store("activity", &Activity{
		Calls: map[string]int{"gno.land/r/a/x": 2}, Callers: map[string]int{"gno.land/r/a/x": 1},
		Partial: true,
	}, nil)

	out := render(t, h, Listing{Path: "/r/a", Paths: []string{"/r/a/x", "/r/a/y"}})
	if !strings.Contains(out, "b-map__tile--unknown") || !strings.Contains(out, "calls unknown") {
		t.Error("an uncounted realm must be drawn as unknown")
	}
	if strings.Contains(out, "· no calls") {
		t.Error("a partial window must never claim a realm had no calls")
	}
	if !strings.Contains(out, "at least 2 calls") {
		t.Error("a count over a partial window is a lower bound")
	}
}

func TestMapOfPurePackagesExplainsTheMissingActivity(t *testing.T) {
	t.Parallel()

	h := New(Deps{Indexer: &fakeIndexer{}, Imports: fakeImports{}, Domain: "gno.land"})
	out := render(t, h, Listing{Path: "/p/nt", Paths: []string{"/p/nt/avl/v0"}})
	if !strings.Contains(out, "Pure packages are not called directly") {
		t.Error("a pure-package map must say why it shows no activity")
	}
}

func TestLevelIsLogarithmic(t *testing.T) {
	t.Parallel()

	cases := []struct{ calls, busiest, want int }{
		{0, 100, 0},
		{1, 1, 4},
		{2, 2, 4},
		{100, 100, 4},
		{1, 1000, 1},
		{30, 1000, 2},
		{300, 1000, 4},
	}
	for _, c := range cases {
		if got := level(c.calls, c.busiest); got != c.want {
			t.Errorf("level(%d, %d) = %d, want %d", c.calls, c.busiest, got, c.want)
		}
	}
}
