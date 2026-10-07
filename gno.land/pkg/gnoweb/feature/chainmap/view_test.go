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
	if strings.Contains(out, "still being counted") || !strings.Contains(out, "Activity unavailable") {
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
		for _, st := range scale(busiest, false)[1:] {
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
	if s := scale(10, true); s[len(s)-1].Class != unknownShadeClass {
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
		"77.0 bn",
		"/r/a/x/heavy · 77.0 bn gas over 3 calls",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("gas map lacks %q", want)
		}
	}
	// heavy, with 3 calls, outranks busy, with 900, on gas.
	if i, j := strings.Index(out, "<small>77.0 bn</small>"), strings.Index(out, "<small>4.0 M</small>"); i < 0 || j < 0 || i > j {
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

// Gas is ranked and written as measured: 1.9 M ranks above 1.1 M (rounding up
// to whole millions would tie them), and 50 k is not "1 M". Units read as gas,
// never as bytes, and the legend's steps never overlap.
func TestGasIsRankedAndWrittenAsMeasured(t *testing.T) {
	t.Parallel()

	h := New(Deps{Indexer: &fakeIndexer{}, Imports: fakeImports{}, Domain: "gno.land"})
	h.activity.store("activity", &Activity{
		Calls: map[string]int{"gno.land/r/a/one": 1, "gno.land/r/a/two": 1, "gno.land/r/a/tiny": 1, "gno.land/r/a/huge": 1},
		Gas:   map[string]int64{"gno.land/r/a/one": 1_100_000, "gno.land/r/a/two": 1_900_000, "gno.land/r/a/tiny": 50_000, "gno.land/r/a/huge": 76_000_000_000},
		From:  10, To: 20, Since: time.Unix(0, 0),
	}, nil)
	out := render(t, h, Listing{Path: "/r/a", Paths: []string{"/r/a/one", "/r/a/two", "/r/a/tiny", "/r/a/huge"}, Metric: MetricGas})

	i1, i2 := strings.Index(out, "<small>1.9 M</small>"), strings.Index(out, "<small>1.1 M</small>")
	if i1 < 0 || i2 < 0 || i1 > i2 {
		t.Error("1.9 M must rank above 1.1 M")
	}
	for _, want := range []string{"<small>50 k</small>", "<small>76.0 bn</small>", "/r/a/tiny · 50 k gas over 1 call", "Colored by gas used over the last 7 days"} {
		if !strings.Contains(out, want) {
			t.Errorf("gas map lacks %q", want)
		}
	}
	if strings.Contains(out, " B<") || strings.Contains(out, " B–") {
		t.Error("billions must not be written B, which reads as bytes")
	}
}

// Under gas the legend gives each shade's upper bound: two ranges written
// with one decimal would print the same number where they meet.
func TestGasLegendGivesUpperBounds(t *testing.T) {
	t.Parallel()

	steps := gasScale(500, 76_000_000, false)
	if len(steps) < 3 {
		t.Fatalf("steps = %v, want several", steps)
	}
	seen := map[string]bool{}
	for _, st := range steps[1:] {
		if !strings.HasPrefix(st.Range, "≤ ") || seen[st.Range] {
			t.Errorf("step %q is not a distinct upper bound", st.Range)
		}
		seen[st.Range] = true
	}
}

// A map whose heaviest realm used under a million gas still tells its realms
// apart: whole millions would put every called tile in the top shade.
func TestLightGasMapKeepsShadesApart(t *testing.T) {
	t.Parallel()

	h := New(Deps{Indexer: &fakeIndexer{}, Imports: fakeImports{}, Domain: "gno.land"})
	h.activity.store("activity", &Activity{
		Calls: map[string]int{"gno.land/r/a/s": 1, "gno.land/r/a/m": 1, "gno.land/r/a/l": 1},
		Gas:   map[string]int64{"gno.land/r/a/s": 20_000, "gno.land/r/a/m": 300_000, "gno.land/r/a/l": 900_000},
		From:  10, To: 20, Since: time.Unix(0, 0),
	}, nil)
	out := render(t, h, Listing{Path: "/r/a", Paths: []string{"/r/a/s", "/r/a/m", "/r/a/l"}, Metric: MetricGas})
	shade := func(p string) string {
		i := strings.Index(out, `href="`+p+`"`)
		j := strings.LastIndex(out[:i], "b-map__tile--")
		return out[j : j+len("b-map__tile--l0")]
	}
	if shade("/r/a/s") == shade("/r/a/l") {
		t.Errorf("20 k and 900 k gas share the shade %s", shade("/r/a/s"))
	}
}

// Real gas sits in a narrow band (a call costs hundreds of thousands): the
// gas scale starts at the lightest realm on the map, not at zero, so realms
// between 500 k and 3.8 M spread over the shades instead of filling the top.
func TestGasScaleStartsAtTheLightestRealm(t *testing.T) {
	t.Parallel()

	h := New(Deps{Indexer: &fakeIndexer{}, Imports: fakeImports{}, Domain: "gno.land"})
	gas := map[string]int64{"gno.land/r/a/p1": 500_000, "gno.land/r/a/p2": 900_000, "gno.land/r/a/p3": 1_700_000, "gno.land/r/a/p4": 3_800_000}
	calls := map[string]int{}
	for p := range gas {
		calls[p] = 1
	}
	h.activity.store("activity", &Activity{Calls: calls, Gas: gas, From: 10, To: 20, Since: time.Unix(0, 0)}, nil)
	out := render(t, h, Listing{Path: "/r/a", Paths: []string{"/r/a/p1", "/r/a/p2", "/r/a/p3", "/r/a/p4"}, Metric: MetricGas})

	shades := map[string]bool{}
	for _, p := range []string{"/r/a/p1", "/r/a/p2", "/r/a/p3", "/r/a/p4"} {
		i := strings.Index(out, `href="`+p+`"`)
		j := strings.LastIndex(out[:i], "b-map__tile--")
		shades[out[j:j+len("b-map__tile--l0")]] = true
	}
	if len(shades) < 3 {
		t.Errorf("four realms from 500 k to 3.8 M use %d shades, want at least 3", len(shades))
	}
	if !strings.Contains(out, "≤ 3.8 M") || strings.Contains(out, "≤ 1 k") {
		t.Error("the legend must run from the lightest realm to the heaviest")
	}
}
