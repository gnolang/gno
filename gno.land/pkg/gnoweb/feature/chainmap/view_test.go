package chainmap

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"
)

// render draws a listing's map, figure and key, as the listing page would.
func render(t *testing.T, h *Handler, l Listing) string {
	t.Helper()
	parts := h.Map(context.Background(), l)
	var b strings.Builder
	for _, c := range []interface{ Render(io.Writer) error }{parts.Figure, parts.Key} {
		if err := c.Render(&b); err != nil {
			t.Fatalf("render: %v", err)
		}
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
	for _, want := range []string{`href="/r/a/x"`, `href="/r/a/y"`, `href="/r/b"`, `href="/r/a/$map"`, `data-listing-item data-name="/r/b"`} {
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
		">10–20<",
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
	if !strings.Contains(out, "nothing calls a pure package") {
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

// An activity refresh that failed is unavailable, not still being counted.
func TestMapSaysUnavailableNotPendingForAFailedRefresh(t *testing.T) {
	t.Parallel()

	h := New(Deps{Indexer: &fakeIndexer{}, Imports: fakeImports{}, Domain: "gno.land"})
	h.activity.store("activity", nil, context.DeadlineExceeded)
	out := render(t, h, Listing{Path: "/r/a", Paths: []string{"/r/a/x"}})
	if strings.Contains(out, "still being counted") || !strings.Contains(out, "Calls unavailable") {
		t.Error("a failed refresh must read as unavailable")
	}
}

// The key lists the busiest realms on the map, most called first, and none
// that nobody called.
func TestMapKeyListsTheBusiestRealms(t *testing.T) {
	t.Parallel()

	h := New(Deps{Indexer: &fakeIndexer{}, Imports: fakeImports{}, Domain: "gno.land"})
	h.activity.store("activity", &Activity{
		Calls: map[string]int{"gno.land/r/a/busy": 100, "gno.land/r/a/some": 3},
		From:  10, To: 20, Since: time.Unix(0, 0),
	}, nil)
	out := render(t, h, Listing{Path: "/r/a", Paths: []string{"/r/a/busy", "/r/a/some", "/r/a/idle"}})
	busy, some := strings.Index(out, `href="/r/a/busy"><span`), strings.Index(out, `href="/r/a/some"><span`)
	if busy < 0 || some < 0 || busy > some {
		t.Error("the busiest list must hold busy then some")
	}
	if strings.Contains(out, `href="/r/a/idle"><span`) {
		t.Error("a realm nobody called is not among the busiest")
	}
}

// The key's steps cover every count from 1 to the busiest exactly once, in
// order, and name only shades some count reaches.
func TestScaleTilesTheCounts(t *testing.T) {
	t.Parallel()

	for _, busiest := range []int{1, 2, 3, 7, 100, 1000, 12345} {
		next := 1
		for _, st := range scale(busiest, false, MetricCalls.format)[1:] {
			lo, hi, _ := strings.Cut(st.Range, "–")
			if hi == "" {
				hi = lo
			}
			l, _ := strconv.Atoi(lo)
			h, _ := strconv.Atoi(hi)
			if l != next || h < l {
				t.Fatalf("busiest %d: step %q after %d", busiest, st.Range, next-1)
			}
			for c := l; c <= h; c++ {
				if got := shadeClasses[level(c, busiest)]; got != st.Class {
					t.Fatalf("busiest %d: %d calls drawn %s, key says %s", busiest, c, got, st.Class)
				}
			}
			next = h + 1
		}
		if next != busiest+1 {
			t.Errorf("busiest %d: the key stops at %d", busiest, next-1)
		}
	}
	if s := scale(10, true, MetricCalls.format); s[len(s)-1].Class != unknownShadeClass {
		t.Error("a partial window's key must show the unknown shade")
	}
}

// The figure carries what the layout fitted to, so the stylesheet need not
// repeat it; keyboard readers can skip it, and the key names each shade.
func TestMapFigureCarriesItsMetrics(t *testing.T) {
	t.Parallel()

	h := New(Deps{Indexer: &fakeIndexer{}, Imports: fakeImports{}, Domain: "gno.land"})
	h.activity.store("activity", &Activity{
		Calls: map[string]int{"gno.land/r/a/busy": 100}, From: 10, To: 20, Since: time.Unix(0, 0),
	}, nil)
	out := render(t, h, Listing{Path: "/r/a", Paths: []string{"/r/a/busy", "/r/a/idle"}})
	for _, want := range []string{
		fmt.Sprintf(`viewBox="0 0 %g %g"`, mapWidth, mapHeight),
		fmt.Sprintf(`font-size="%g"`, tileFont),
		`href="#map-end"`,
		`id="map-end"`,
		`b-map-swatch b-map__tile--l4`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("map lacks %q", want)
		}
	}
	if strings.Contains(out, "aria-live") {
		t.Error("the hover status must not be a live region")
	}
}

// A gas map shades and ranks by gas, keeps its metric on every map link, and
// offers the way back to calls; the call map offers gas.
func TestMapColorsByGas(t *testing.T) {
	t.Parallel()

	h := New(Deps{Indexer: &fakeIndexer{}, Imports: fakeImports{}, Domain: "gno.land"})
	h.activity.store("activity", &Activity{
		Calls: map[string]int{"gno.land/r/a/x/busy": 900, "gno.land/r/a/x/heavy": 3},
		Gas:   map[string]int64{"gno.land/r/a/x/busy": 4_000_000, "gno.land/r/a/x/heavy": 77_000_000_000},
		From:  10, To: 20, Since: time.Unix(0, 0),
	}, nil)
	paths := []string{"/r/a/x/busy", "/r/a/x/heavy", "/r/a/x/idle", "/r/a/y/1", "/r/a/y/2", "/r/a/y/3"}

	out := render(t, h, Listing{Path: "/r/", Up: "", Paths: paths, Metric: MetricGas})
	for _, want := range []string{
		`aria-current="true" title="Gas used`,
		`href="/r/$map"`,                 // back to calls
		`href="/r/a/$map&amp;color=gas"`, // a zoom keeps the metric
		"Most gas · 7 days",
		"77.0 B",
		"/r/a/x/heavy · 77.0 B gas over 3 calls",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("gas map lacks %q", want)
		}
	}
	// heavy, with 3 calls, outranks busy, with 900, on gas.
	if i, j := strings.Index(out, "<small>77.0 B</small>"), strings.Index(out, "<small>4 M</small>"); i < 0 || j < 0 || i > j {
		t.Errorf("busiest list is not ranked by gas")
	}

	calls := render(t, h, Listing{Path: "/r/", Paths: paths})
	if !strings.Contains(calls, `href="/r/$map&amp;color=gas"`) || !strings.Contains(calls, "Most called · 7 days") {
		t.Error("the call map must offer gas and keep its own list")
	}
}

func TestParseMetric(t *testing.T) {
	t.Parallel()

	for in, want := range map[string]Metric{"gas": MetricGas, "calls": MetricCalls, "": MetricCalls, "GAS": MetricCalls, "x": MetricCalls} {
		if got := ParseMetric(in); got != want {
			t.Errorf("ParseMetric(%q) = %q, want %q", in, got, want)
		}
	}
}
