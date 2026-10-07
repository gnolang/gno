package gnoweb_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

func serve(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func newMapHandler(t *testing.T, paths ...string) http.Handler {
	t.Helper()
	pkgs := make([]*gnoweb.MockPackage, len(paths))
	for i, p := range paths {
		pkgs[i] = &gnoweb.MockPackage{Path: p, Files: map[string]string{"a.gno": "package a"}}
	}
	logger := slog.New(slog.NewTextHandler(&testingLogger{t}, &slog.HandlerOptions{}))
	h, err := gnoweb.NewHTTPHandler(logger, newTestHandlerConfig(t, gnoweb.NewMockClient(pkgs...)))
	if err != nil {
		t.Fatalf("NewHTTPHandler: %v", err)
	}
	return h
}

var (
	listItemRe = regexp.MustCompile(`<span class="name">([^<]+)</span>`)
	mapTileRe  = regexp.MustCompile(`<a class="b-map__tile[^"]*" href="([^"]+)"`)
)

// The map is the list drawn: the same packages, no more, no fewer.
func TestHTTPHandler_MapAndListShowTheSameListing(t *testing.T) {
	t.Parallel()

	h := newMapHandler(t, append(manyPaths("/r/demo/p", 10), "/r/other/x", "/p/demo/lib")...)

	list := serve(t, h, "/r/")
	mapped := serve(t, h, "/r/$map")
	if list.Code != http.StatusOK || mapped.Code != http.StatusOK {
		t.Fatalf("status list=%d map=%d, want 200", list.Code, mapped.Code)
	}

	var listed, tiles []string
	for _, m := range listItemRe.FindAllStringSubmatch(list.Body.String(), -1) {
		listed = append(listed, strings.TrimSpace(m[1]))
	}
	for _, m := range mapTileRe.FindAllStringSubmatch(mapped.Body.String(), -1) {
		tiles = append(tiles, m[1])
	}
	slices.Sort(listed)
	slices.Sort(tiles)
	want := append(manyPaths("/r/demo/p", 10), "/r/other/x")
	if !slices.Equal(listed, want) || !slices.Equal(tiles, want) {
		t.Fatalf("list = %v, map = %v, want both %v", listed, tiles, want)
	}
}

// The two renderings are tabs of the same page, where Content/Source/Actions
// would be on a package.
func TestHTTPHandler_ExplorerHasListAndMapTabs(t *testing.T) {
	t.Parallel()

	h := newMapHandler(t, manyPaths("/r/demo/p", 10)...)
	for path, active := range map[string]string{"/r/demo": "Directory", "/r/demo$map": "Map"} {
		body := serve(t, h, path).Body.String()
		if !strings.Contains(body, `href="/r/demo"`) || !strings.Contains(body, `href="/r/demo$map"`) {
			t.Errorf("%s: missing the List or Map tab", path)
		}
		activeRe := regexp.MustCompile(`link--is-active[^"]*">\s*<svg>\s*<use href="#ico-[a-z]+"></use>\s*</svg>\s*<span class="link-label">` + active + `<`)
		if !activeRe.MatchString(body) {
			t.Errorf("%s: %s is not the active tab", path, active)
		}
		if strings.Contains(body, `<span class="link-label">Actions<`) {
			t.Errorf("%s: a listing has no Actions tab", path)
		}
	}
}

// qpaths has no cursor. A listing that stops at the cap says so, in both
// renderings, and one that does not stop there says nothing.
func TestHTTPHandler_ListingSaysWhenItStops(t *testing.T) {
	t.Parallel()

	const listCap = 1000 // gnoweb's maxListedPaths
	paths := func(n int) []string { return manyPaths("/r/many/p", n) }

	full := newMapHandler(t, paths(listCap)...)
	over := newMapHandler(t, paths(listCap+1)...)
	const notice = "Only the first 1000 paths are listed"
	for _, path := range []string{"/r/many/", "/r/many/$map"} {
		if body := serve(t, full, path).Body.String(); strings.Contains(body, notice) {
			t.Errorf("%s: exactly the cap is complete, yet the page says it stopped", path)
		}
		body := serve(t, over, path).Body.String()
		if !strings.Contains(body, notice) {
			t.Errorf("%s: past the cap, the page must say it stopped", path)
		}
	}
	if n := len(mapTileRe.FindAllString(serve(t, over, "/r/many/$map").Body.String(), -1)); n != listCap {
		t.Errorf("map drew %d tiles, want the %d listed", n, listCap)
	}
}

func TestHTTPHandler_MapStatuses(t *testing.T) {
	t.Parallel()

	h := newMapHandler(t, append(manyPaths("/r/demo/p", 10), "/r/demo/a")...)
	cases := map[string]int{
		"/r/demo$map":    http.StatusOK,
		"/r/demo/a$map":  http.StatusNotFound, // nothing below the package
		"/r/nothing$map": http.StatusNotFound,
		"/$map":          http.StatusBadRequest,
	}
	for path, want := range cases {
		if got := serve(t, h, path).Code; got != want {
			t.Errorf("%s: status %d, want %d", path, got, want)
		}
	}
}

// A map exists only where the list does: on a package that has packages
// below it, the map must not draw them as if the path were a listing.
func TestHTTPHandler_NoMapOnAPackageWithSubpackages(t *testing.T) {
	t.Parallel()

	h := newMapHandler(t, append(manyPaths("/p/demo/lib/sub", 10), "/p/demo/lib")...)
	if got := serve(t, h, "/p/demo/lib$map").Code; got != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 on a package path", got)
	}
	if got := serve(t, h, "/p/demo$map").Code; got != http.StatusOK {
		t.Fatalf("status = %d, want 200 on the listing above it", got)
	}
}

// manyPaths returns n package paths prefix0000, prefix0001, …: enough for a
// listing to offer a map.
func manyPaths(prefix string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("%s%04d", prefix, i)
	}
	return out
}

// A listing too small for a map offers no tabs, and asking for its map draws
// the list: a map of a few packages shows nothing the list does not.
func TestHTTPHandler_SmallListingHasNoMap(t *testing.T) {
	t.Parallel()

	h := newMapHandler(t, "/r/demo/a", "/r/demo/b")
	for _, path := range []string{"/r/demo", "/r/demo$map"} {
		rr := serve(t, h, path)
		body := rr.Body.String()
		if rr.Code != http.StatusOK || !strings.Contains(body, `class="b-list"`) {
			t.Errorf("%s: status %d, want the list", path, rr.Code)
		}
		if strings.Contains(body, `<span class="link-label">Map<`) || strings.Contains(body, "b-map") {
			t.Errorf("%s: a small listing must offer no map", path)
		}
	}
}

// The listing sits on the realm grid: a rail with the filter, beside the
// content, in both renderings.
func TestHTTPHandler_ListingHasARailWithAFilter(t *testing.T) {
	t.Parallel()

	h := newMapHandler(t, manyPaths("/r/demo/p", 10)...)
	for _, path := range []string{"/r/demo", "/r/demo$map"} {
		body := serve(t, h, path).Body.String()
		if !strings.Contains(body, `class="b-sidebar sidebar"`) || !strings.Contains(body, `data-filter-items-value="[data-listing-item]"`) {
			t.Errorf("%s: no listing rail with its filter", path)
		}
	}
	if body := serve(t, h, "/r/demo$map").Body.String(); !strings.Contains(body, "Map key") {
		t.Error("the map's key belongs in the rail")
	}
}

// Zoom out leads to the map one level up, and only where there is one: not
// from a kind's root, and not to a package, which has no listing to map.
func TestHTTPHandler_MapZoomsOutToAListing(t *testing.T) {
	t.Parallel()

	h := newMapHandler(t, append(append(manyPaths("/r/a/b/c", 10), manyPaths("/r/p/x/y", 10)...), "/r/p")...)
	for path, want := range map[string]string{
		"/r/a/b$map":  `href="/r/a/$map"`,
		"/r/a$map":    `href="/r/$map"`,
		"/r/a/b/$map": `href="/r/a/$map"`,
		"/r/p/x$map":  "",
		"/r/$map":     "",
	} {
		body := serve(t, h, path).Body.String()
		if !strings.Contains(body, "b-map__tile") {
			t.Errorf("%s: no map drawn", path)
			continue
		}
		if got := strings.Contains(body, "Zoom out"); got != (want != "") {
			t.Errorf("%s: zoom out shown = %v, want %v", path, got, want != "")
		}
		if want != "" && !strings.Contains(body, want) {
			t.Errorf("%s: zoom out must link %s", path, want)
		}
	}
}

// $map&color=gas is a map of the same listing; without an indexer it is the
// plain map, with no switch to offer.
func TestHTTPHandler_GasMapWithoutIndexer(t *testing.T) {
	t.Parallel()

	h := newMapHandler(t, manyPaths("/r/demo/p", 10)...)
	rr := serve(t, h, "/r/demo$map&color=gas")
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if body := rr.Body.String(); !strings.Contains(body, "b-map__tile") || strings.Contains(body, "b-map-switch") {
		t.Error("without an indexer the gas map must be the plain map, with no switch")
	}
}

// With an indexer, $map&color=gas shades by gas: the switch marks Gas, and the
// side list ranks by gas, so a realm called rarely but heavily leads it.
func TestHTTPHandler_MapColorsByGasWithIndexer(t *testing.T) {
	t.Parallel()

	callTo := func(path string, gas int) indexer.Tx {
		tx := indexer.Tx{Height: 50, Success: true, GasUsed: gas, Messages: make([]indexer.Message, 1)}
		tx.Messages[0].Value.Type, tx.Messages[0].Value.PkgPath, tx.Messages[0].Value.Caller = "MsgCall", path, "g1caller"
		return tx
	}
	paths := manyPaths("/r/demo/p", 10)
	// The test config has no domain, so a fully qualified path is the
	// gnoweb path itself.
	idx := &stubIndexer{calls: []indexer.Tx{
		callTo("/r/demo/p0000", 1_000_000), callTo("/r/demo/p0000", 1_000_000), callTo("/r/demo/p0000", 1_000_000),
		callTo("/r/demo/p0001", 9_000_000_000),
	}}
	pkgs := make([]*gnoweb.MockPackage, len(paths))
	for i, p := range paths {
		pkgs[i] = &gnoweb.MockPackage{Path: p, Files: map[string]string{"a.gno": "package a"}}
	}
	cfg := newTestHandlerConfig(t, gnoweb.NewMockClient(pkgs...))
	cfg.Indexer = idx
	h, err := gnoweb.NewHTTPHandler(slog.New(slog.NewTextHandler(&testingLogger{t}, nil)), cfg)
	if err != nil {
		t.Fatalf("NewHTTPHandler: %v", err)
	}

	body := serve(t, h, "/r/demo$map&color=gas").Body.String()
	for _, want := range []string{`<span aria-current="true" title="Gas used`, "Most gas · 7 days"} {
		if !strings.Contains(body, want) {
			t.Errorf("gas map lacks %q", want)
		}
	}
	if i, j := strings.Index(body, "<small>9.0 bn</small>"), strings.Index(body, "<small>3.0 M</small>"); i < 0 || j < 0 || i > j {
		t.Error("the gas list must put p0001 (9 bn gas, 1 call) before p0000 (3 M gas, 3 calls)")
	}
	if calls := serve(t, h, "/r/demo$map").Body.String(); !strings.Contains(calls, "Most called · 7 days") {
		t.Error("$map without color must stay the call map")
	}
}
