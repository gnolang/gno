package store

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// fakeClient answers api/v1 endpoints from a map and counts realm calls.
type fakeClient struct {
	responses map[string]string // args → body
	calls     atomic.Int32
}

func (c *fakeClient) Realm(_ context.Context, _, args string) ([]byte, error) {
	c.calls.Add(1)
	body, ok := c.responses[args]
	if !ok {
		return []byte(`{"version":1,"error":"not_found"}`), nil
	}
	return []byte(body), nil
}

// listingJSON builds one api/v1 listing at gno.land/{r,p}/<ns>/<slug>;
// extra overrides or adds fields.
func listingJSON(slug, ns, kind, title, extra string) string {
	tree := "r"
	if kind == kindPackage {
		tree = "p"
	}
	s := fmt.Sprintf(`"slug":%q,"kind":%q,"path":"gno.land/%s/%s/%s","title":%q,`+
		`"tagline":"Does things","category":"defi","palette":1,"stars":0,"earned":false,`+
		`"icon":"ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG","cover":""`,
		slug, kind, tree, ns, slug, title)
	if extra != "" {
		s += "," + extra
	}
	return "{" + s + "}"
}

var (
	trustedApp   = listingJSON("blog", "gnoland", kindApp, "Blog", `"stars":12`)
	communityApp = listingJSON("pixels", "nym-acmex123", kindApp, "Pixels", "")
	hostileApp   = listingJSON("spoof", "nym-x", kindApp, "Gno\u202eSwap", "")
	// A package kind on a realm path does not hold together.
	mismatched = `{"slug":"gnoswap","kind":"package","path":"gno.land/r/nym-evil/swap","title":"GnoSwap","tagline":"x","category":"defi","palette":0}`
)

func homeBody(listings ...string) string {
	return `{"version":1,"height":1000,"time":1759538400,` +
		`"pulse":{"listed_7d":0,"stars_7d":41,"updated_7d":0,"listed_30d":9,"stars_30d":120,"updated_30d":0},` +
		`"categories":[{"key":"defi","label":"DeFi","count":4},{"key":"games","label":"Games","count":0}],` +
		`"shelves":[{"title":"New","pinned":true,"slugs":["blog","pixels","spoof","gnoswap","blog","two","three"]}],` +
		`"activity":[{"kind":"milestone","slug":"blog","stars":10},{"kind":"listed","slug":"spoof"}],` +
		`"listings":[` + strings.Join(listings, ",") + `]}`
}

func newTestHandler(t *testing.T, responses map[string]string) (*Handler, *fakeClient) {
	t.Helper()
	c := &fakeClient{responses: responses}
	h := New(Deps{
		Client:    c,
		RealmPath: "/r/gnoland/store",
		Domain:    "gno.land",
		Trusted:   func(pkg string) bool { return strings.HasPrefix(pkg, "gnoland/") },
	})
	return h, c
}

// render runs the store view for a gnoweb URL of the store realm, e.g.
// ":c/defi?page=2". It returns "" when the store hands the page back to the
// realm's own markdown.
func render(t *testing.T, h *Handler, target string) string {
	t.Helper()
	u, err := weburl.Parse("/r/gnoland/store" + target)
	require.NoError(t, err)
	view, _ := h.View(context.Background(), u)
	if view == nil {
		return ""
	}
	var buf bytes.Buffer
	require.NoError(t, view.Render(&buf))
	return buf.String()
}

func TestHome(t *testing.T) {
	h, _ := newTestHandler(t, map[string]string{"api/v1/home": homeBody(
		trustedApp, communityApp, hostileApp, mismatched,
		listingJSON("two", "nym-acmex123", kindApp, "Two", ""),
		listingJSON("three", "nym-acmex123", kindApp, "Three", ""),
	)})
	html := render(t, h, "")
	require.NotEmpty(t, html)

	assert.NotContains(t, html, "spoof", "a bidi override in a title drops the listing")
	assert.NotContains(t, html, "GnoSwap", "a kind that does not match the path drops the listing")
	assert.Equal(t, 1, strings.Count(html, `class="icon" src=`), "only the trusted listing shows its owner icon")
	assert.Contains(t, html, "In the spotlight today", "the one eligible app is the hero")
	assert.NotContains(t, html, `id="store-spotlight"`, "and no Spotlight is left")
	assert.Contains(t, html, `class="community"`, "community apps say so on their path line")
	assert.Contains(t, html, `<use href="#ico-check"></use></svg>Core</span>`, "trusted apps carry a positive badge")
	assert.NotContains(t, html, "slug=pixels", "no star count for an app nobody starred")
	assert.NotContains(t, html, "c/games", "empty categories are not offered")

	// Cards lead to the listing's own page, never to a store copy of it,
	// and the star count opens the store's prefilled Star action.
	assert.Contains(t, html, `<h2 id="store-hero"><a href="/r/gnoland/blog">Blog</a></h2>`)
	assert.Contains(t, html, `href="/r/gnoland/store$help&amp;func=Star&amp;slug=blog"`)
	assert.Contains(t, html, `href="/r/gnoland/store:c/defi"`)
	assert.Contains(t, html, `href="/r/gnoland/store:build"`)

	// Pulse: the shortest non-zero window per figure, nothing invented.
	assert.Contains(t, html, "Block 1000")
	assert.Contains(t, html, "<strong>9</strong> apps &amp; packages listed this month")
	assert.Contains(t, html, "<strong>41</strong> stars given this week")

	// Activity is phrased from fixed templates, and dropped listings drop
	// their events.
	assert.Contains(t, html, `<a href="/r/gnoland/blog">Blog crossed 10 stars</a>`)
	assert.NotContains(t, html, "was listed")

	// The lens is the page; its lead is the head description.
	assert.Contains(t, html, `aria-current="page">Apps</a>`)
	assert.Contains(t, html, "<p>"+introLine+"</p>")
}

// meta runs the store view and returns what it sets around the page.
func meta(t *testing.T, h *Handler, target string) Meta {
	t.Helper()
	u, err := weburl.Parse("/r/gnoland/store" + target)
	require.NoError(t, err)
	_, m := h.View(context.Background(), u)
	return m
}

func TestCategory(t *testing.T) {
	home := `{"version":1,"categories":[{"key":"defi","label":"DeFi","count":30},{"key":"games","label":"Games","count":0}],"shelves":[],"listings":[]}`
	h, c := newTestHandler(t, map[string]string{
		"api/v1/home":             home,
		"api/v1/category/defi/2":  `{"version":1,"category":{"key":"defi","label":"DeFi","count":30},"page":2,"pages":2,"listings":[` + communityApp + `]}`,
		"api/v1/category/games/1": `{"version":1,"category":{"key":"games","label":"Games","count":0},"page":1,"pages":1,"listings":[]}`,
	})
	html := render(t, h, ":c/defi?page=2")
	assert.Contains(t, html, "<h1>DeFi</h1>")
	assert.Contains(t, html, `<h3><a href="/r/nym-acmex123/pixels">Pixels</a></h3>`)
	assert.Contains(t, html, `rel="prev" href="?page=1"`)
	assert.NotContains(t, html, `rel="next"`)
	assert.Contains(t, html, `aria-current="page">DeFi`)

	// An empty category keeps its own entry in the strip.
	games := render(t, h, ":c/games")
	assert.Contains(t, games, `aria-current="page">Games`)
	assert.Contains(t, games, `href="/r/gnoland/store#list-your-app">List your app</a>`, "the empty box links to the snippet")
	assert.NotContains(t, games, `id="list-your-app"`, "no full call to list on an empty category")

	// The page names its category, not its cards.
	assert.NotContains(t, html, `class="b-tag--secondary" href="/r/gnoland/store:c/defi"`)
	assert.Equal(t, Meta{Status: http.StatusOK, Title: "DeFi apps · Explore", Description: "DeFi apps on gno.land, each with its code public: try one, or read how it works first."}, meta(t, h, ":c/defi?page=2"))

	// Under the category strip, the lens holds the page without being it,
	// and the grid's cards sit under a heading of their own.
	assert.Contains(t, html, `aria-current="true">Apps</a>`)
	assert.Contains(t, html, `<h2 class="u-sr-only">Apps</h2>`)

	// Keys and pages the taxonomy does not have never reach the node: they
	// are a store page saying so, whether or not they could be a key.
	calls := c.calls.Load()
	for _, target := range []string{
		":c/unknown", ":c/Defi", ":c/defi/x", ":c/", ":c/<script>",
		":c/defi?page=3", ":c/defi?page=0", ":c/defi?page=x",
	} {
		html := render(t, h, target)
		assert.Contains(t, html, "<h1>Not found</h1>", target)
		assert.NotContains(t, html, `aria-current="page"`, "%s: nothing is current on a page that does not exist", target)
		assert.NotContains(t, html, "script", target)
		assert.Equal(t, http.StatusNotFound, meta(t, h, target).Status, target)
	}
	assert.Equal(t, calls, c.calls.Load(), "no query for a category or page that does not exist")
}

func TestRealmFieldsAreValidated(t *testing.T) {
	h, _ := newTestHandler(t, map[string]string{"api/v1/home": `{"version":1,` +
		`"categories":[{"key":"defi","label":"DeFi","count":1},{"key":"a/b","label":"Bad","count":1},{"key":"x","label":"Gno\u202e","count":1},{"key":"y","label":"Y","count":-1}],` +
		`"shelves":[{"title":"Fine","slugs":[]},{"title":"Sp\u200boof","slugs":[]}],"listings":[]}`})
	res, err := h.home(context.Background())
	require.NoError(t, err)
	require.Len(t, res.Categories, 1)
	assert.Equal(t, "defi", res.Categories[0].Key)
	require.Len(t, res.Shelves, 1)
	assert.Equal(t, "Fine", res.Shelves[0].Title)
}

func TestEmptyPulseIsNotDrawn(t *testing.T) {
	h, _ := newTestHandler(t, map[string]string{"api/v1/home": `{"version":1,"categories":[],"shelves":[],"listings":[]}`})
	assert.NotContains(t, render(t, h, ""), "b-store-pulse")
}

func TestSnippetsUseTheConfiguredRealm(t *testing.T) {
	h := New(Deps{
		Client:    &fakeClient{responses: map[string]string{"api/v1/home": `{"version":1,"categories":[],"shelves":[],"listings":[]}`}},
		RealmPath: "/r/acme/store",
		Domain:    "test.land",
		Trusted:   func(string) bool { return false },
	})
	u, err := weburl.Parse("/r/acme/store")
	require.NoError(t, err)
	var buf bytes.Buffer
	view, _ := h.View(context.Background(), u)
	require.NoError(t, view.Render(&buf))
	assert.Contains(t, buf.String(), `import &#34;test.land/r/acme/store&#34;`)
	assert.NotContains(t, buf.String(), "gnoland/store")
}

func TestBuild(t *testing.T) {
	pkg := listingJSON("mylib", "nym-acmex123", kindPackage, "My lib", "")
	h, c := newTestHandler(t, map[string]string{
		"api/v1/build": `{"version":1,"shelves":[{"title":"New to build with","slugs":["mylib","mylib","mylib","mylib"]}],` +
			`"builders":{"top":[{"namespace":"acme","listings":3,"stars":40},{"namespace":"<bad>","listings":1,"stars":1}],"new":[]},` +
			`"listings":[` + pkg + `]}`,
	})
	html := render(t, h, ":build")

	assert.Contains(t, html, `aria-current="page">Build`)
	assert.Contains(t, html, "<p>"+buildLine+"</p>", "the lead is the head description")
	assert.Contains(t, html, "Top builders")
	assert.Contains(t, html, `<a class="b-pkg-list__link" href="/u/acme">`)
	assert.NotContains(t, html, "&lt;bad&gt;", "invalid builder namespaces are dropped")
	assert.NotContains(t, html, "New builders", "an empty section is not drawn")
	assert.Contains(t, html, "List your package")
	assert.Equal(t, int32(1), c.calls.Load(), "build needs no category labels")
}

// Everything the store does not own stays the realm's own page. Other tabs
// never reach View: gnoweb calls it on the Content tab only.
func TestFallsBackToTheRealm(t *testing.T) {
	h, _ := newTestHandler(t, map[string]string{"api/v1/home": `{"version":2}`})
	for _, target := range []string{
		"",             // backend failure: the realm's markdown, not an error page
		":app/blog",    // a render path the store does not own
		":api/v1/home", // raw API
		":c/<script>",  // not a category key
		":c/unknown",   // unknown category (not_found)
		":buildx",      // near miss
	} {
		assert.Empty(t, render(t, h, target), target)
	}

	u := &weburl.GnoURL{Path: "/r/gnoland/blog", WebQuery: url.Values{}}
	view, _ := h.View(context.Background(), u)
	assert.Nil(t, view, "other realms are never touched")
}

func TestCacheCoalesces(t *testing.T) {
	h, c := newTestHandler(t, map[string]string{"api/v1/home": homeBody(trustedApp)})
	for range 3 {
		render(t, h, "")
	}
	assert.Equal(t, int32(1), c.calls.Load())
}

func TestTiers(t *testing.T) {
	h, _ := newTestHandler(t, nil)
	cases := []struct {
		path   string
		earned bool
		want   tier
		rich   bool
	}{
		{"/r/gnoland/blog", false, tierTrusted, true},
		{"/p/gnoland/lib", false, tierTrusted, true},
		{"/r/nym-acmex123/pixels", false, tierRegistered, false},
		{"/r/nym-acmex123/pixels", true, tierRegistered, true},
		{"/r/g17zyd8dhz9v7ndecx4wxc4z3ufmc6e43nex9cxg/poll", true, tierAnonymous, false},
		// Looks like an address but fails the checksum: a registered name.
		{"/r/g1lpzvdyl5f4pqyyw5dd7h6p0klz5yejqmt9z2az/poll", false, tierRegistered, false},
	}
	for _, c := range cases {
		l := &listing{webPath: c.path}
		got := h.tierOf(l)
		assert.Equal(t, c.want, got, c.path)
		assert.Equal(t, c.rich, got.rich(c.earned), c.path)
	}
}

func TestAssetURL(t *testing.T) {
	assert.Equal(t, "https://ipfs.io/ipfs/QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG",
		assetURL("ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG"))
	for _, s := range []string{
		"https://imgur.com/a.png",            // mutable URL
		"data:image/png;base64,iVBORw0KGgo=", // inline data bloats every page
		"ipfs://short",
		"ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG/../x",
	} {
		assert.Empty(t, assetURL(s), s)
	}
}

func TestDisplayableRejectsSpoofing(t *testing.T) {
	for _, s := range []string{
		"Acme Swap",
		"Ελληνικά",
		"Café",
		"P\u0430yments", // look-alikes are the realm's policy, not a rendering hazard
	} {
		assert.True(t, displayable(s, maxTitleRunes), "%q", s)
	}
	for _, s := range []string{
		"Gno\u202eSwap", "a\u200bb", "\u3164", "x\u2800", "tab\tname", "no\u00a0break",
		"Gno\uE000", // private use
		"Gno  Swap", // doubled space
	} {
		assert.False(t, displayable(s, maxTitleRunes), "%q", s)
	}
}

func TestShortPathKeepsNamespace(t *testing.T) {
	assert.Equal(t, "acme/swap", shortPath("acme/swap"))
	assert.Equal(t, "nym-gnoswapx123/…/pool", shortPath("nym-gnoswapx123/v1/deep/pool"))
	assert.Equal(t, "nt/avl/v0", shortPath("nt/avl/v0"), "three segments stay whole")
	assert.Equal(t, "gnoland/boards/v0", shortPath("gnoland/boards/v0"))
	assert.Equal(t, "acme/…/swap/v2", shortPath("acme/apps/swap/v2"), "a version keeps its name")
	assert.Equal(t, "g17zyd…9cxg/poll", shortPath("g17zyd8dhz9v7ndecx4wxc4z3ufmc6e43nex9cxg/poll"))
	assert.Equal(t, "g1lpzvdyl5f4pqyyw5dd7h6p0klz5yejqmt9z2az/poll", shortPath("g1lpzvdyl5f4pqyyw5dd7h6p0klz5yejqmt9z2az/poll"), "not an address: kept whole")
}

func TestGeneratedVisualsAreDeterministicAndSafe(t *testing.T) {
	l := &listing{Slug: "acme-swap", Title: `<b>"x"</b> Swap`, Category: "defi", Palette: 11}
	a, b := cover(l), cover(l)
	assert.Equal(t, a, b)
	assert.True(t, strings.HasPrefix(string(a), `<svg class="cover"`))
	assert.NotContains(t, string(icon(l)), "<b>", "initials only, escaped")

	l.Slug = "other-app"
	assert.NotEqual(t, a, cover(l))

	// The slug, not the category, picks the motif: a category page varies.
	shapes := make(map[string]bool)
	for _, slug := range []string{"aaa", "bbb", "ccc", "ddd", "eee", "fff", "ggg", "hhh"} {
		svg := string(cover(&listing{Slug: slug, Category: "defi"}))
		shapes[svg[strings.Index(svg, `fill-opacity=".1"/>`)+19:][:5]] = true
	}
	assert.Greater(t, len(shapes), 2, "one category, several motifs")
}

func TestInitials(t *testing.T) {
	for name, want := range map[string]string{
		"GnoSwap":       "GS",
		"GovDAO":        "GD",
		"DAOForge":      "DF",
		"DAO Forge":     "DF",
		"Gno.land Blog": "GB",
		"Gno.land":      "GL",
		"Gnotify":       "GN",
		"Gnoverse":      "GN",
		"Gno":           "GN",
		"Threads":       "TH",
		"X":             "X",
		"pixel arena":   "PA",
		"Web3 Tools":    "WT",
		"★ Swap":        "SW",
		"Ελληνικά":      "ΕΛ",
		"★★★":           "★",
		"\u200b":        noGlyph,
	} {
		assert.Equal(t, want, initials(name), name)
	}
}

func TestEmptyBuilderListsAreLeftOut(t *testing.T) {
	h, _ := newTestHandler(t, nil)
	res := &buildResponse{}
	assert.Empty(t, h.buildBuild(res).Builders, "no section at all")

	res.Builders.New = []builder{{Namespace: "acme", Listings: 1}}
	got := h.buildBuild(res).Builders
	require.Len(t, got, 1)
	assert.Equal(t, "New builders", got[0].Title)
}

func TestValidRef(t *testing.T) {
	assert.True(t, validRef("defi", "DeFi", 2))
	for _, c := range []category{
		{Key: "defi", Label: "De\u202eFi", Count: 2},
		{Key: "defi", Label: "", Count: 2},
		{Key: "defi", Label: "DeFi", Count: -1},
		{Key: "De Fi", Label: "DeFi", Count: 2},
	} {
		assert.False(t, validRef(c.Key, c.Label, c.Count), "%+v", c)
	}
}

func TestLists(t *testing.T) {
	home := `{"version":1,"categories":[{"key":"defi","label":"DeFi","count":30}],` +
		`"shelves":[{"title":"New","slugs":["pixels","two","three","four"],"more":[{"key":"latest","title":"New","count":30},{"key":"Bad Key","title":"x","count":1}]}],` +
		`"listings":[` + strings.Join([]string{communityApp,
		listingJSON("two", "nym-acmex123", kindApp, "Two", ""),
		listingJSON("three", "nym-acmex123", kindApp, "Three", ""),
		listingJSON("four", "nym-acmex123", kindApp, "Four", "")}, ",") + `]}`
	h, c := newTestHandler(t, map[string]string{
		"api/v1/home":          home,
		"api/v1/list/latest/2": `{"version":1,"list":{"key":"latest","title":"New","count":30},"page":2,"pages":2,"listings":[` + communityApp + `]}`,
	})

	// A shelf links to its list, named for screen readers; a bad list
	// reference only loses its link.
	front := render(t, h, "")
	assert.Contains(t, front, `<a class="b-inline-btn" href="/r/gnoland/store:latest">See all<span class="u-sr-only"> New</span></a>`)
	assert.Contains(t, front, `<ul class="b-store-grid">`)
	assert.NotContains(t, front, "Bad Key")

	page := render(t, h, ":latest?page=2")
	assert.Contains(t, page, "<h1>New apps</h1>", "gnoweb names its lists")
	assert.Contains(t, page, `<a href="/r/gnoland/store:latest" aria-current="page">New</a>`, "the list switcher marks the list")
	assert.Contains(t, page, `aria-current="true">Apps</a>`, "the lens holds the list")
	assert.NotContains(t, page, `aria-current="page">All`, "no category is current on a list")
	assert.Equal(t, "New apps · Explore", meta(t, h, ":latest?page=2").Title)
	assert.Contains(t, page, "<p>30 apps</p>")
	assert.Contains(t, page, `<h3><a href="/r/nym-acmex123/pixels">Pixels</a></h3>`)
	assert.Contains(t, page, `rel="prev" href="?page=1"`)

	// Lists and pages the front page does not have never reach the node.
	calls := c.calls.Load()
	for _, target := range []string{":top", ":latest?page=3", ":latest?page=0"} {
		assert.Contains(t, render(t, h, target), "<h1>Not found</h1>", target)
	}
	for _, target := range []string{":nope", ":Bad Key", ":latest/x"} {
		assert.Empty(t, render(t, h, target), target)
	}
	assert.Equal(t, calls, c.calls.Load(), "no query for a list or page that does not exist")
}

func TestFeatured(t *testing.T) {
	trusted := func(slug string) listing { return listing{Slug: slug, Kind: kindApp, tier: tierTrusted} }
	earned := func(slug string) listing {
		return listing{Slug: slug, Kind: kindApp, tier: tierRegistered, Earned: true}
	}
	plain := listing{Slug: "plain", Kind: kindApp, tier: tierRegistered}
	ls := []listing{plain, trusted("t1"), trusted("t2"), trusted("t3"), earned("e1"), earned("e2"), earned("e3"), earned("e4")}
	bySlug := indexListings(ls)
	var all []string
	for _, l := range ls {
		all = append(all, l.Slug)
	}
	slugs := func(ls []*listing) (out []string) {
		for _, l := range ls {
			out = append(out, l.Slug)
		}
		return out
	}

	res := &homeResponse{Shelves: []shelf{{Slugs: all}}}
	hero, label, spot := featured(res, bySlug)
	assert.Equal(t, "t1", hero.Slug)
	assert.Equal(t, "In the spotlight today", label)
	assert.Equal(t, []string{"e1", "e2", "e3"}, slugs(spot), "eligible only, the operator's apps one place at most, hero included")

	res.Height = blocksPerDay
	hero2, _, _ := featured(res, bySlug)
	assert.NotEqual(t, hero.Slug, hero2.Slug, "next day, next hero")

	// Momentum wins over the rotation, among eligible apps only; it is
	// "trending" only if it leads the trending shelf, all apps counted.
	trendingShelf := shelf{Slugs: []string{"plain", "e3"}, Stars7d: []int{50, 7}, More: []listRef{{Key: "trending"}}}
	res.Shelves = append(res.Shelves, trendingShelf)
	hero, label, spot = featured(res, bySlug)
	assert.Equal(t, "e3", hero.Slug, "plain has more stars but is not eligible")
	assert.Equal(t, labelPopular, label, "plain leads the week")
	assert.NotContains(t, slugs(spot), "e3", "the Spotlight never repeats the hero")
	res.Shelves[len(res.Shelves)-1].Stars7d = []int{7, 7}
	_, label, _ = featured(res, bySlug)
	assert.Equal(t, labelTrending, label, "a tie for first leads")

	// GovDAO's pick wins over everything.
	res.Pick = "plain"
	hero, label, _ = featured(res, bySlug)
	assert.Equal(t, "plain", hero.Slug)
	assert.Equal(t, "Picked by GovDAO", label)

	// At launch, the operator's apps fill the page; with none eligible, no hero.
	launch := indexListings([]listing{trusted("t1"), trusted("t2"), trusted("t3"), trusted("t4")})
	hero, _, spot = featured(&homeResponse{Shelves: []shelf{{Slugs: []string{"t1", "t2", "t3", "t4"}}}}, launch)
	assert.Equal(t, "t1", hero.Slug)
	assert.Equal(t, []string{"t2", "t3", "t4"}, slugs(spot))
	hero, _, spot = featured(&homeResponse{Shelves: []shelf{{Slugs: []string{"plain"}}}}, indexListings([]listing{plain}))
	assert.Nil(t, hero)
	assert.Nil(t, spot)
}

func TestShelfFieldsAreValidated(t *testing.T) {
	got := validShelves([]shelf{{
		Title:   "Fine",
		Note:    "Sp\u200boof",
		Slugs:   []string{"a", "b"},
		Stars7d: []int{1},
		More:    []listRef{{Key: "latest", Title: "New", Count: 1}, {Key: "x", Title: "X", Count: -1}},
	}})
	require.Len(t, got, 1)
	assert.Empty(t, got[0].Note, "a spoofing note is dropped")
	assert.Nil(t, got[0].Stars7d, "counts that do not match the slugs are dropped")
	assert.Len(t, got[0].More, 1, "a bad list reference is dropped")
}

func TestShelfRules(t *testing.T) {
	h, _ := newTestHandler(t, nil)
	var ls []listing
	for _, s := range []string{"aaa", "bbb", "ccc", "ddd", "eee", "fff", "ggg", "hhh", "iii"} {
		ls = append(ls, listing{Slug: s, Title: s, Category: "defi", webPath: "/r/nym-a/" + s})
	}
	nine := []string{"aaa", "bbb", "ccc", "ddd", "eee", "fff", "ggg", "hhh", "iii"}
	got := h.shelves([]shelf{
		{Title: "A", Slugs: append([]string{"aaa"}, nine...)}, // a repeat inside a shelf
		{Title: "B", Slugs: nine, Stars7d: []int{9, 9, 9, 9, 9, 9, 9, 9, 9}},
		{Title: "C", Slugs: []string{"ggg", "hhh", "iii", "zzz"}},
	}, indexListings(ls), nil, map[string]bool{"bbb": true})

	require.Len(t, got, 1, "B and C only hold apps shown above, below minShelf")
	var titles []string
	for _, c := range got[0].Cards {
		titles = append(titles, c.Title)
	}
	assert.Equal(t, []string{"aaa", "ccc", "ddd", "eee", "fff", "ggg", "hhh", "iii"}, titles, "gridSize cards, the Spotlight's left out, each once")
}

func TestShelvesDrawFullRows(t *testing.T) {
	h, _ := newTestHandler(t, nil)
	var ls []listing
	all := []string{"aaa", "bbb", "ccc", "ddd", "eee", "fff", "ggg", "hhh"}
	for _, s := range all {
		ls = append(ls, listing{Slug: s, Title: s, Category: "defi", webPath: "/r/nym-a/" + s})
	}
	got := h.shelves([]shelf{
		{Title: "Top", Slugs: all[:6]},                  // 6: one row of 4
		{Title: "Rediscover", Slugs: all[4:]},           // eee, fff were trimmed above: 4
		{Title: "Pinned", Pinned: true, Slugs: all[:3]}, // below a row: kept whole
		{Title: "Pinned5", Pinned: true, Slugs: all[:5]},
	}, indexListings(ls), nil, nil)

	sizes := make(map[string]int)
	for _, s := range got {
		sizes[s.Title] = len(s.Cards)
	}
	assert.Equal(t, map[string]int{"Top": 4, "Rediscover": 4, "Pinned": 3, "Pinned5": 4}, sizes)
	assert.Equal(t, "eee", got[1].Cards[0].Title, "a trimmed app is not counted as shown")
}

func TestMomentum(t *testing.T) {
	h, _ := newTestHandler(t, nil)
	var ls []listing
	all := []string{"aaa", "bbb", "ccc", "ddd"}
	for _, s := range all {
		ls = append(ls, listing{Slug: s, Title: s, Category: "defi", webPath: "/r/nym-a/" + s})
	}
	got := h.shelves([]shelf{{Title: "Trending", Slugs: all, Stars7d: []int{9, 4, 0, 5}}}, indexListings(ls), nil, nil)
	require.Len(t, got, 1)
	var weeks []int
	for _, c := range got[0].Cards {
		weeks = append(weeks, c.Momentum)
	}
	assert.Equal(t, []int{9, 0, 0, 5}, weeks, "momentum from momentumMin up")
}

func TestPinnedShelves(t *testing.T) {
	h, _ := newTestHandler(t, nil)
	var ls []listing
	for _, s := range []string{"aaa", "bbb", "ccc", "ddd", "eee"} {
		ls = append(ls, listing{Slug: s, Title: s, Category: "defi", webPath: "/r/nym-a/" + s})
	}
	got := h.shelves([]shelf{
		{Title: "New", Pinned: true, Slugs: []string{"aaa", "bbb", "ccc", "ddd"}},
		{Title: "Top", Pinned: true, Slugs: []string{"bbb", "eee"}, Empty: "None yet."},
		{Title: "Trending", Slugs: []string{"aaa", "bbb", "ccc", "ddd"}},
		{Title: "Quiet", Pinned: true, Empty: "Nothing to show."},
		{Title: "Featured", Pinned: true, Slugs: []string{"eee"}, Empty: "None listed."},
		{Title: "Blank", Pinned: true}, // validShelves dropped its empty text
	}, indexListings(ls), nil, map[string]bool{"eee": true})

	var titles []string
	for _, s := range got {
		titles = append(titles, s.Title)
	}
	assert.Equal(t, []string{"New", "Top", "Quiet"}, titles, "Trending only holds apps shown above; Featured only the Spotlight's; Blank has nothing to say")
	require.Len(t, got[1].Cards, 1, "a pinned shelf shows its true head, below the minimum, but for the Spotlight's apps")
	assert.Equal(t, "bbb", got[1].Cards[0].Title)
	assert.Empty(t, got[2].Cards)
	assert.Equal(t, "Nothing to show.", got[2].Empty)
}

func TestPickIsTheHero(t *testing.T) {
	home := `{"version":1,"categories":[],"shelves":[],"pick":"blog","listings":[` + trustedApp + `]}`
	h, _ := newTestHandler(t, map[string]string{"api/v1/home": home})
	html := render(t, h, "")
	assert.Contains(t, html, `<section class="b-store-hero" aria-labelledby="store-hero">`)
	assert.Contains(t, html, "Picked by GovDAO")
	assert.Contains(t, html, `<h2 id="store-hero"><a href="/r/gnoland/blog">Blog</a></h2>`)
	assert.Contains(t, html, `href="/r/gnoland/blog$source"`)

	// A pick that sanitize dropped, or none, draws no hero.
	h, _ = newTestHandler(t, map[string]string{"api/v1/home": `{"version":1,"categories":[],"shelves":[],"pick":"spoof","listings":[` + hostileApp + `]}`})
	assert.NotContains(t, render(t, h, ""), "Picked by GovDAO")
}
