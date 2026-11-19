package store

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

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
		`"tagline":"Does things","category":"defi","palette":1,"stars":0,`+
		`"icon":"ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG","cover":""`,
		slug, kind, tree, ns, slug, title)
	if extra != "" {
		s += "," + extra
	}
	return "{" + s + "}"
}

// confirmed marks a listing its own realm listed, as most are.
const confirmed = `"confirmed":true`

var (
	trustedApp   = listingJSON("blog", "gnoland", kindApp, "Blog", `"stars":12,"core":true,`+confirmed+fmt.Sprintf(`,"created":%d`, time.Now().Unix()))
	communityApp = listingJSON("pixels", "nym-acmex123", kindApp, "Pixels", confirmed)
	hostileApp   = listingJSON("spoof", "nym-x", kindApp, "Gno\u202eSwap", "")
	// A package kind on a realm path does not hold together.
	mismatched = `{"slug":"gnoswap","kind":"package","path":"gno.land/r/nym-evil/swap","title":"GnoSwap","tagline":"x","category":"defi","palette":0}`
)

func homeBody(listings ...string) string {
	return `{"version":1,"height":1000,"time":1759538400,"moved_to":"gno.land/r/gnoland/store/v1",` +
		`"pulse":{"listed_7d":0,"stars_7d":41,"updated_7d":0,"listed_30d":9,"stars_30d":120,"updated_30d":0},` +
		`"categories":[{"key":"defi","label":"DeFi","count":4,"confirmed":4},{"key":"games","label":"Games","count":0,"confirmed":0}],` +
		`"shelves":[{"key":"new","slugs":["blog","pixels","spoof","gnoswap","blog","two","three"]}],` +
		`"listings":[` + strings.Join(listings, ",") + `]}`
}

func newTestHandler(t *testing.T, responses map[string]string) (*Handler, *fakeClient) {
	t.Helper()
	c := &fakeClient{responses: responses}
	h := New(Deps{
		Client:      c,
		RealmPath:   "/r/gnoland/store",
		Domain:      "gno.land",
		ChainID:     "test-chain",
		Remote:      "https://rpc.example:443",
		IPFSGateway: "https://ipfs.example",
		Trusted:     func(pkg string) bool { return strings.HasPrefix(pkg, "gnoland/") },
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

	assert.Contains(t, html, `Explore has moved: <a href="/r/gnoland/store/v1">`, "a successor named by GovDAO is shown")
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

	// The latest event is told from the listings' own dates, with fixed
	// words.
	assert.Contains(t, html, `<a href="/r/gnoland/blog">Blog just joined</a>`)

	// What a listing is, is said in the open, not in a tooltip only; and
	// the store vouches for no one.
	assert.Contains(t, html, `role="note">Everything here was listed by its author: being listed is not being endorsed.`)
	assert.NotContains(t, html, "vouches")

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
	claim := listingJSON("claimed", "nym-acmex123", kindApp, "Claimed", "")
	home := `{"version":1,"categories":[{"key":"defi","label":"DeFi","count":30,"confirmed":28},{"key":"games","label":"Games","count":0,"confirmed":0}],"shelves":[],"listings":[]}`
	h, c := newTestHandler(t, map[string]string{
		"api/v1/home":             home,
		"api/v1/category/defi/2":  `{"version":1,"category":{"key":"defi","label":"DeFi","count":30,"confirmed":28},"page":2,"pages":2,"listings":[` + communityApp + `,` + claim + `]}`,
		"api/v1/category/games/1": `{"version":1,"category":{"key":"games","label":"Games","count":0,"confirmed":0},"page":1,"pages":1,"listings":[]}`,
	})
	html := render(t, h, ":c/defi?page=2")
	assert.Contains(t, html, "<h1>DeFi</h1>")

	// A category is counted by its confirmed apps; a claim nothing confirms
	// says so on its card, where a confirmed app says "community".
	assert.Contains(t, html, "<p>28 apps, and 2 claimed listings not confirmed yet</p>")
	assert.Contains(t, html, `DeFi <span class="count u-font-mono">28</span>`)
	assert.Equal(t, 1, strings.Count(html, `>unverified</span>`))
	assert.Equal(t, 1, strings.Count(html, `>community</span>`))
	assert.Contains(t, html, `<h3><a href="/r/nym-acmex123/pixels" rel="nofollow ugc">Pixels</a></h3>`)
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
		`"categories":[{"key":"defi","label":"DeFi","count":1},{"key":"a/b","label":"Bad","count":1},{"key":"x","label":"Gno\u202e","count":1},{"key":"y","label":"Y","count":-1},{"key":"z","label":"Z","count":1,"confirmed":2}],` +
		`"shelves":[{"key":"new","title":"Sp\u200boof","note":"Official","slugs":[]},{"key":"nope","slugs":[]}],"listings":[]}`})
	res, err := h.home(context.Background())
	require.NoError(t, err)
	require.Len(t, res.Categories, 1)
	assert.Equal(t, "defi", res.Categories[0].Key)
	require.Len(t, res.Shelves, 1, "a shelf gnoweb does not know is not drawn")
	assert.Equal(t, "New", res.Shelves[0].Title, "a shelf's words are gnoweb's: the realm's are not read")
	assert.Empty(t, res.Shelves[0].Note)
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
		"api/v1/build": `{"version":1,"shelves":[{"key":"new","slugs":["mylib","mylib","mylib","mylib"]}],` +
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

	// The Claim command runs as copied: enough gas for Claim (see
	// integration/testdata/store_claim_snippet.txtar), and the configured
	// chain and node rather than gnokey's defaults.
	assert.Contains(t, html, `-gas-wanted 50000000 -broadcast \
  -chainid &#34;test-chain&#34; -remote &#34;https://rpc.example:443&#34; KEY`)
	assert.Contains(t, html, "once it has 3 stars from accounts with a registered name", "a claim lists nothing until confirmed")
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

func TestCacheServesHitsWithinTTL(t *testing.T) {
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
		// Valid bech32, of another prefix: a registered name too, as the realm reads it.
		{"/r/abc1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqql9d4m7/poll", true, tierRegistered, true},
	}
	for _, c := range cases {
		l := &listing{webPath: c.path}
		got := h.tierOf(l)
		assert.Equal(t, c.want, got, c.path)
		assert.Equal(t, c.rich, got.rich(c.earned), c.path)
	}
}

func TestAssetURL(t *testing.T) {
	h, _ := newTestHandler(t, nil)
	const cid = "ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG"
	assert.Equal(t, "https://ipfs.example/ipfs/QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG", h.assetURL(cid), "the operator's gateway")
	for _, s := range []string{
		"https://imgur.com/a.png",            // mutable URL
		"data:image/png;base64,iVBORw0KGgo=", // inline data bloats every page
		"ipfs://short",
		cid + "/../x",
	} {
		assert.Empty(t, h.assetURL(s), s)
	}

	h.deps.IPFSGateway = ""
	assert.Empty(t, h.assetURL(cid), "no gateway, no owner image")
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
	assert.Equal(t, "g17zyd8dhz…ex9cxg/poll", shortPath("g17zyd8dhz9v7ndecx4wxc4z3ufmc6e43nex9cxg/poll"), "8 data and 6 checksum characters: no vanity search finds a twin")
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
		"(Beta) Swap":   "BS",
		"#1 Wallet":     "1W",
	} {
		assert.Equal(t, want, initials(name), name)
	}
}

func TestEmptyBuilderListsAreLeftOut(t *testing.T) {
	h, _ := newTestHandler(t, nil)
	res := &buildResponse{}
	assert.Empty(t, h.buildBuild(res, nil).Builders, "no section at all")

	res.Builders.New = []builder{{Namespace: "acme", Listings: 1}}
	got := h.buildBuild(res, nil).Builders
	require.Len(t, got, 1)
	assert.Equal(t, "New builders", got[0].Title)
}

func TestValidCategory(t *testing.T) {
	assert.True(t, validCategory(category{Key: "defi", Label: "DeFi", Count: 2, Confirmed: 2}))
	for _, c := range []category{
		{Key: "defi", Label: "De\u202eFi", Count: 2},
		{Key: "defi", Label: "", Count: 2},
		{Key: "defi", Label: "DeFi", Count: -1},
		{Key: "De Fi", Label: "DeFi", Count: 2},
		{Key: "defi", Label: "DeFi", Count: 2, Confirmed: 3},
		{Key: "defi", Label: "DeFi", Count: 2, Confirmed: -1},
	} {
		assert.False(t, validCategory(c), "%+v", c)
	}
}

func TestLists(t *testing.T) {
	home := `{"version":1,"categories":[{"key":"defi","label":"DeFi","count":30,"confirmed":30}],` +
		`"shelves":[{"key":"new","slugs":["pixels","two","three","four"],"more":[{"key":"latest","count":30},{"key":"Bad Key","count":1}]}],` +
		`"listings":[` + strings.Join([]string{communityApp,
		listingJSON("two", "nym-acmex123", kindApp, "Two", ""),
		listingJSON("three", "nym-acmex123", kindApp, "Three", ""),
		listingJSON("four", "nym-acmex123", kindApp, "Four", "")}, ",") + `]}`
	h, c := newTestHandler(t, map[string]string{
		"api/v1/home":          home,
		"api/v1/list/latest/2": `{"version":1,"list":{"key":"latest","count":30},"page":2,"pages":2,"listings":[` + communityApp + `]}`,
	})

	// A shelf links to its list, named for screen readers; a bad list
	// reference only loses its link.
	front := render(t, h, "")
	assert.Contains(t, front, `<a class="b-inline-btn" href="/r/gnoland/store:latest">See all<span class="u-sr-only"> New apps</span></a>`)
	assert.Contains(t, front, `<ul class="b-store-grid">`)
	assert.NotContains(t, front, "Bad Key")

	page := render(t, h, ":latest?page=2")
	assert.Contains(t, page, "<h1>New apps</h1>", "gnoweb names its lists")
	assert.Contains(t, page, `<a href="/r/gnoland/store:latest" aria-current="page">New</a>`, "the list switcher marks the list")
	assert.Contains(t, page, `aria-current="true">Apps</a>`, "the lens holds the list")
	assert.NotContains(t, page, `aria-current="page">All`, "no category is current on a list")
	assert.Equal(t, "New apps · Explore", meta(t, h, ":latest?page=2").Title)
	assert.Contains(t, page, "<p>30 apps</p>")
	assert.Contains(t, page, `<h3><a href="/r/nym-acmex123/pixels" rel="nofollow ugc">Pixels</a></h3>`)
	assert.Contains(t, page, `rel="prev" href="?page=1"`)

	// Lists and pages the front page does not have never reach the node.
	calls := c.calls.Load()
	for _, target := range []string{":top", ":latest?page=3", ":latest?page=0"} {
		assert.Contains(t, render(t, h, target), "<h1>Not found</h1>", target)
	}
	assert.Equal(t, calls, c.calls.Load(), "no query for a list or page that does not exist")
}

func TestFeatured(t *testing.T) {
	trusted := func(slug string) listing { return listing{Slug: slug, Kind: kindApp, tier: tierTrusted} }
	earned := func(slug string) listing {
		return listing{Slug: slug, Kind: kindApp, tier: tierRegistered, earned: true}
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

	res.Time = secondsPerDay
	hero2, _, _ := featured(res, bySlug)
	assert.NotEqual(t, hero.Slug, hero2.Slug, "next day, next hero")

	// Momentum wins over the rotation, among eligible apps only; it is
	// "trending" only if it leads the trending shelf, all apps counted.
	bySlug["plain"].Week, bySlug["e3"].Week = 50, 7
	res.Shelves = append(res.Shelves, shelf{Key: trendingKey, Slugs: []string{"plain", "e3"}})
	hero, label, spot = featured(res, bySlug)
	assert.Equal(t, "e3", hero.Slug, "plain has more stars but is not eligible")
	assert.Equal(t, labelPopular, label, "plain leads the week")
	assert.NotContains(t, slugs(spot), "e3", "the Spotlight never repeats the hero")
	bySlug["plain"].Week = 7
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
	got := validShelves([]shelf{
		{
			Key:   "top",
			Slugs: []string{"a", "b"},
			More:  []listRef{{Key: "latest", Count: 1}, {Key: "top", Count: -1}, {Key: "x", Count: 1}},
		},
		{Key: "unknown", Slugs: []string{"a"}},
	}, homeShelfDefs)
	require.Len(t, got, 1, "a shelf gnoweb has no words for is dropped")
	assert.Equal(t, "Top", got[0].Title)
	assert.True(t, got[0].Pinned)
	require.Len(t, got[0].More, 1, "a bad list reference is dropped")
	assert.Equal(t, "New apps", got[0].More[0].Title, "a list's title is gnoweb's")
}

// The floors a shelf shows from are gnoweb's: the realm sends its ranking,
// and a gnoweb release can ask more of an app as the chain grows.
func TestShelfFloors(t *testing.T) {
	ls := []listing{
		{Slug: "aaa", Ranked: topStars, Week: trendingStars},
		{Slug: "bbb", Ranked: topStars - 1, Week: trendingStars - 1},
		{Slug: "ccc", Ranked: 9, Week: 7},
	}
	shelves := validShelves([]shelf{
		{Key: "top", Slugs: []string{"ccc", "aaa", "bbb", "gone"}},
		{Key: "trending", Slugs: []string{"aaa", "bbb", "ccc"}},
		{Key: "new", Slugs: []string{"bbb"}},
	}, homeShelfDefs)
	applyFloors(shelves, indexListings(ls))

	assert.Equal(t, []string{"ccc", "aaa"}, shelves[0].Slugs, "Top from its floor, and only listings the response holds")
	assert.Equal(t, []string{"aaa", "ccc"}, shelves[1].Slugs, "Trending by each listing's own week")
	assert.Equal(t, []string{"bbb"}, shelves[2].Slugs, "a shelf with no floor is left as it is")
}

// The floors hold on the pages too: the front page's Top draws no app below
// its floor, and the list behind Trending none below the shelf's.
func TestFloorsOnThePages(t *testing.T) {
	low := listingJSON("low", "nym-acmex123", kindApp, "Low App", confirmed+`,"stars":2,"ranked":2,"stars_7d":1`)
	hot := listingJSON("hot", "nym-acmex123", kindApp, "Hot App", confirmed+`,"stars":9,"ranked":9,"stars_7d":9`)
	h, _ := newTestHandler(t, map[string]string{
		"api/v1/home": `{"version":1,"categories":[],"shelves":[` +
			`{"key":"top","slugs":["hot","low"]},` +
			`{"key":"trending","slugs":["hot","low"],"stars_7d":[9,1],"more":[{"key":"trending","count":2}]}],` +
			`"listings":[` + hot + `,` + low + `]}`,
		"api/v1/list/trending/1": `{"version":1,"list":{"key":"trending","count":2},"page":1,"pages":1,"listings":[` + hot + `,` + low + `]}`,
	})
	front := render(t, h, "")
	assert.Contains(t, front, "Hot App")
	assert.NotContains(t, front, "Low App", "below Top's floor, and on no other shelf")

	page := render(t, h, ":trending")
	assert.Contains(t, page, "Hot App")
	assert.NotContains(t, page, "Low App", "below Trending's floor")
	assert.Contains(t, page, "<p>1 app</p>")
}

// No word of a shelf is the realm's: its JSON cannot even set one.
func TestShelfWordsAreNotDecoded(t *testing.T) {
	var s shelf
	require.NoError(t, json.Unmarshal([]byte(`{"key":"new","Title":"Official","title":"Official","Note":"n","Empty":"e","Pinned":true,"slugs":["a"]}`), &s))
	assert.Equal(t, shelf{Key: "new", Slugs: []string{"a"}}, s)
}

// titled and pinned build the shelves of the layout tests, which gnoweb
// otherwise words by key.
func titled(title string, s shelf) shelf {
	s.Title = title
	return s
}

func pinned(title, empty string, s shelf) shelf {
	s.Title, s.Empty, s.Pinned = title, empty, true
	return s
}

func TestShelfRules(t *testing.T) {
	h, _ := newTestHandler(t, nil)
	var ls []listing
	for _, s := range []string{"aaa", "bbb", "ccc", "ddd", "eee", "fff", "ggg", "hhh", "iii"} {
		ls = append(ls, listing{Slug: s, Title: s, Category: "defi", webPath: "/r/nym-a/" + s})
	}
	nine := []string{"aaa", "bbb", "ccc", "ddd", "eee", "fff", "ggg", "hhh", "iii"}
	got := h.shelves([]shelf{
		titled("A", shelf{Slugs: append([]string{"aaa"}, nine...)}), // a repeat inside a shelf
		titled("B", shelf{Slugs: nine}),
		titled("C", shelf{Slugs: []string{"ggg", "hhh", "iii", "zzz"}}),
	}, indexListings(ls), nil, map[string]bool{"bbb": true}, nil)

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
		titled("Top", shelf{Slugs: all[:6]}),        // 6: one row of 4
		titled("Rediscover", shelf{Slugs: all[4:]}), // eee, fff were trimmed above: 4
		pinned("Pinned", "", shelf{Slugs: all[:3]}), // below a row: kept whole
		pinned("Pinned5", "", shelf{Slugs: all[:5]}),
	}, indexListings(ls), nil, nil, nil)

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
	for i, s := range all {
		ls = append(ls, listing{Slug: s, Title: s, Category: "defi", webPath: "/r/nym-a/" + s, Week: []int{9, 4, 0, 5}[i]})
	}
	weeks := func(s shelf) (out []int) {
		got := h.shelves([]shelf{s}, indexListings(ls), nil, nil, nil)
		require.Len(t, got, 1)
		for _, c := range got[0].Cards {
			out = append(out, c.Momentum)
		}
		return out
	}
	assert.Equal(t, []int{9, 0, 0, 5}, weeks(shelf{Key: trendingKey, Slugs: all, shelfDef: homeShelfDefs[trendingKey]}), "momentum from momentumMin up")
	assert.Equal(t, []int{0, 0, 0, 0}, weeks(titled("New", shelf{Slugs: all})), "and only on a shelf ranked by the week")
}

func TestPinnedShelves(t *testing.T) {
	h, _ := newTestHandler(t, nil)
	var ls []listing
	for _, s := range []string{"aaa", "bbb", "ccc", "ddd", "eee"} {
		ls = append(ls, listing{Slug: s, Title: s, Category: "defi", webPath: "/r/nym-a/" + s})
	}
	got := h.shelves([]shelf{
		pinned("New", "", shelf{Slugs: []string{"aaa", "bbb", "ccc", "ddd"}}),
		pinned("Top", "None yet.", shelf{Slugs: []string{"bbb", "eee"}}),
		titled("Trending", shelf{Slugs: []string{"aaa", "bbb", "ccc", "ddd"}}),
		pinned("Quiet", "Nothing to show.", shelf{}),
		pinned("Featured", "None listed.", shelf{Slugs: []string{"eee"}}),
		pinned("Blank", "", shelf{}), // no empty text: nothing to say
	}, indexListings(ls), nil, map[string]bool{"eee": true}, nil)

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

// A realm answer for another key or page falls back to the markdown, and
// says why in the log, as a decode failure does.
func TestPagedMismatchIsLogged(t *testing.T) {
	var logs bytes.Buffer
	h, _ := newTestHandler(t, map[string]string{
		"api/v1/home":            `{"version":1,"categories":[{"key":"defi","label":"DeFi","count":30}],"shelves":[],"listings":[]}`,
		"api/v1/category/defi/1": `{"version":1,"category":{"key":"games","label":"Games","count":30},"page":1,"pages":2,"listings":[]}`,
	})
	h.deps.Logger = slog.New(slog.NewTextHandler(&logs, nil))
	assert.Empty(t, render(t, h, ":c/defi"))
	assert.Contains(t, logs.String(), "category/defi/1 answered games/1 of 2")
}

// The front page's count decides which pages exist, so a fresher page that
// counts more never links to a page the store then refuses.
func TestNextStaysWithinTheFrontPageCount(t *testing.T) {
	h, _ := newTestHandler(t, map[string]string{
		"api/v1/home":            `{"version":1,"categories":[{"key":"defi","label":"DeFi","count":24}],"shelves":[],"listings":[]}`,
		"api/v1/category/defi/1": `{"version":1,"category":{"key":"defi","label":"DeFi","count":25},"page":1,"pages":2,"listings":[` + communityApp + `]}`,
	})
	html := render(t, h, ":c/defi")
	assert.Contains(t, html, "Pixels")
	assert.NotContains(t, html, `rel="next"`)
	assert.Contains(t, render(t, h, ":c/defi?page=2"), "<h1>Not found</h1>")
}

// Every render path the store does not own but the raw API is the store's
// own Not found page, as an unknown category is.
func TestOtherArgsAreNotFound(t *testing.T) {
	h, c := newTestHandler(t, map[string]string{"api/v1/home": homeBody(trustedApp)})
	for _, target := range []string{":nope", ":app/blog", ":buildx", ":latest/x", ":api"} {
		assert.Contains(t, render(t, h, target), "<h1>Not found</h1>", target)
		assert.Equal(t, http.StatusNotFound, meta(t, h, target).Status, target)
	}
	assert.Empty(t, render(t, h, ":api/v1/home"), "the raw API stays the realm's")
	assert.Equal(t, int32(1), c.calls.Load(), "the cached front page only")
}

// Links into a community app are not followed, wherever the store draws
// one: the hero, the pulse's latest event and "Surprise me".
func TestCommunityLinksAreNotFollowed(t *testing.T) {
	fresh := listingJSON("pixels", "nym-acmex123", kindApp, "Pixels", confirmed+fmt.Sprintf(`,"created":%d`, time.Now().Unix()))
	home := `{"version":1,"categories":[],"shelves":[{"key":"new","slugs":["pixels"]}],"pick":"pixels","listings":[` + fresh + `]}`
	h, _ := newTestHandler(t, map[string]string{"api/v1/home": home})
	html := render(t, h, "")
	for _, link := range []string{
		`<h2 id="store-hero"><a href="/r/nym-acmex123/pixels" rel="nofollow ugc">Pixels</a></h2>`,
		`href="/r/nym-acmex123/pixels" rel="nofollow ugc">Open</a>`,
		`href="/r/nym-acmex123/pixels$source" rel="nofollow ugc">`,
		// Named by its path: the top of the page shows no title anyone can register.
		`<a href="/r/nym-acmex123/pixels" rel="nofollow ugc">nym-acmex123/pixels just joined</a>`,
		`href="/r/nym-acmex123/pixels" rel="nofollow ugc">Surprise me</a>`,
	} {
		assert.Contains(t, html, link)
	}

	h, _ = newTestHandler(t, map[string]string{"api/v1/home": homeBody(trustedApp)})
	assert.NotContains(t, render(t, h, ""), "nofollow", "the operator's apps keep their links")
}

// An event older than momentAge does not close the pulse: on an idle
// store, no app "just joined" for months. The newest of a listing and an
// update is told, from the shelves that hold them.
func TestMomentIsBounded(t *testing.T) {
	now := time.Now()
	ago := func(d time.Duration) int64 { return now.Add(-d).Unix() }
	shelves := []shelf{{Key: "new", Slugs: []string{"blog", "gone"}}, {Key: "updated", Slugs: []string{"wiki"}}, {Key: "top", Slugs: []string{"blog"}}}
	moment := func(created, updated int64) *moment {
		return latestMoment(shelves, indexListings([]listing{
			{Slug: "blog", Title: "Blog", webPath: "/r/gnoland/blog", Core: true, Created: created},
			{Slug: "wiki", Title: "Wiki", webPath: "/r/gnoland/wiki", Core: true, Created: ago(1000 * time.Hour), Updated: updated},
		}), now)
	}
	require.NotNil(t, moment(ago(time.Hour), 0))
	assert.Equal(t, "Blog just joined", moment(ago(time.Hour), 0).Text)
	assert.Equal(t, "Wiki was updated", moment(ago(time.Hour), ago(time.Minute)).Text, "the newest of the two")
	assert.Nil(t, moment(ago(momentAge+time.Hour), ago(momentAge+time.Hour)))
	assert.Nil(t, moment(0, 0), "a listing with no date is not news")
}

// An owner image is drawn over the generated visual it hides, which
// gnoweb's script uncovers when the image fails to load; with no gateway
// set, only the generated visual is drawn.
func TestOwnerImagesFallBack(t *testing.T) {
	app := listingJSON("blog", "gnoland", kindApp, "Blog", `"cover":"ipfs://QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG"`)
	h, _ := newTestHandler(t, map[string]string{"api/v1/home": homeBody(app)})
	html := render(t, h, "")
	assert.Contains(t, html, `data-fallback><svg class="cover"`)
	assert.Contains(t, html, `data-fallback><svg class="icon"`)

	h.deps.IPFSGateway = ""
	h.cache = newResponseCache()
	html = render(t, h, "")
	assert.NotContains(t, html, "<img")
	assert.Contains(t, html, `<svg class="cover"`)
}

// Core is the realm's word: an app under the operator's trust list shows
// its owner images, but is not called one of gno.land's core apps.
func TestCoreKeysOffTheRealm(t *testing.T) {
	h, _ := newTestHandler(t, nil)
	trusted := listing{Slug: "x", Kind: kindApp, tier: tierTrusted, Icon: "https://ipfs.example/ipfs/x"}
	c := h.card(&trusted, nil, nil)
	assert.Empty(t, c.Trust)
	assert.False(t, c.Community)
	assert.NotEmpty(t, c.Icon, "trust still lets owner images show")

	trusted.Core = true
	assert.Equal(t, "Core", h.card(&trusted, nil, nil).Trust)

	established := listing{Slug: "y", Kind: kindApp, tier: tierRegistered, earned: true}
	assert.Equal(t, "Established", h.card(&established, nil, nil).Trust)
	seed := listing{Slug: "z", Kind: kindApp, tier: tierRegistered, Core: true}
	assert.Equal(t, "Core", h.card(&seed, nil, nil).Trust, "the realm's seeds, whatever the trust list")
	assert.False(t, h.card(&seed, nil, nil).Community, "a core app is not a community one")
}

func TestEarned(t *testing.T) {
	now := time.Now()
	ago := func(d time.Duration) int64 { return now.Add(-d).Unix() }
	ok := listing{Ranked: earnStars, Created: ago(earnAge), FifthRanked: ago(earnStarsAge)}
	assert.True(t, earned(&ok, now))
	for name, mod := range map[string]func(l *listing){
		"too few ranked stars":    func(l *listing) { l.Ranked = earnStars - 1 },
		"no fifth ranked star":    func(l *listing) { l.FifthRanked = 0 },
		"listing too young":       func(l *listing) { l.Created = ago(earnAge - time.Minute) },
		"fifth star too young":    func(l *listing) { l.FifthRanked = ago(earnStarsAge - time.Minute) },
		"image change in waiting": func(l *listing) { l.Pending = true },
	} {
		l := ok
		mod(&l)
		assert.False(t, earned(&l, now), name)
	}
}

// Eligibility, not the shelf an app sits on, decides the operator's cap
// and the trending hero: a shelf of trusted apps ahead of the earned ones
// fills neither the pool nor the hero.
func TestFeaturedPoolIsDecidedByEligibility(t *testing.T) {
	var ls []listing
	var trusted []string
	for i := 1; i <= spotlightPool; i++ {
		s := fmt.Sprintf("t%02d", i)
		ls = append(ls, listing{Slug: s, Kind: kindApp, tier: tierTrusted})
		trusted = append(trusted, s)
	}
	for _, s := range []string{"e1", "e2", "e3", "e4"} {
		ls = append(ls, listing{Slug: s, Kind: kindApp, tier: tierRegistered, earned: true})
	}
	bySlug := indexListings(ls)
	top := shelf{Slugs: []string{"e1", "e2", "e3", "e4"}}
	for day := range int64(4) {
		res := &homeResponse{Shelves: []shelf{{Slugs: trusted}, top}}
		res.Time = day * secondsPerDay
		hero, _, spot := featured(res, bySlug)
		require.Len(t, spot, spotlightSize, "day %d", day)
		n := 0
		for _, l := range append(spot, hero) {
			if l.tier == tierTrusted {
				n++
			}
		}
		assert.LessOrEqual(t, n, 1, "day %d: the operator's apps take one place at most", day)
	}

	bySlug["e1"].Week = 50
	hot := shelf{Key: trendingKey, Slugs: []string{"e1"}}
	hero, label, _ := featured(&homeResponse{Shelves: []shelf{{Slugs: trusted}, hot}}, bySlug)
	assert.Equal(t, "e1", hero.Slug, "the eligible app with the most momentum")
	assert.Equal(t, labelTrending, label)
}

func TestCacheRecoversAPanickingLoad(t *testing.T) {
	c := newResponseCache()
	_, err := cached(context.Background(), c, "k", func(context.Context) (int, error) { panic("boom") })
	require.ErrorContains(t, err, "boom")
	_, err = cached(context.Background(), c, "k", func(context.Context) (int, error) { return 1, nil })
	assert.ErrorContains(t, err, "boom", "remembered as a failure, briefly")
}

// A failure is remembered for errorTTL, not cacheTTL: a store realm back
// up shows within seconds.
func TestCacheKeepsFailuresBriefly(t *testing.T) {
	c := newResponseCache()
	c.put("ok", 1, nil)
	c.put("failed", nil, errors.New("down"))
	now := time.Now()
	assert.WithinDuration(t, now.Add(cacheTTL), c.entries["ok"].expires, time.Second)
	assert.WithinDuration(t, now.Add(errorTTL), c.entries["failed"].expires, time.Second)
}

// At its cap the cache drops the answer soonest to expire, never every
// answer at once: a crawler walking the pages costs no reader the front page.
func TestCacheEvictsOneAtItsCap(t *testing.T) {
	c := newResponseCache()
	c.put("failed", nil, errors.New("down")) // remembered briefly: the soonest to expire
	for i := range cacheMaxEntries - 2 {
		c.put(strconv.Itoa(i), i, nil)
	}
	c.put("home", 1, nil)
	require.Len(t, c.entries, cacheMaxEntries)

	c.put("one-more", 1, nil)
	assert.Len(t, c.entries, cacheMaxEntries)
	assert.NotContains(t, c.entries, "failed")
	assert.Contains(t, c.entries, "home")
	assert.Contains(t, c.entries, "one-more")

	c.put("home", 2, nil) // a key already held takes no one's place
	assert.Len(t, c.entries, cacheMaxEntries)
	assert.Contains(t, c.entries, "0")
}

// Concurrent misses of one key share a single load.
func TestCacheCoalescesConcurrentMisses(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		c := newResponseCache()
		release := make(chan struct{})
		var loads atomic.Int32
		load := func(context.Context) (int, error) {
			loads.Add(1)
			<-release
			return 1, nil
		}
		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() {
				v, err := cached(context.Background(), c, "k", load)
				assert.NoError(t, err)
				assert.Equal(t, 1, v)
			})
		}
		synctest.Wait() // every reader waits on the key
		close(release)
		wg.Wait()
		assert.Equal(t, int32(1), loads.Load())
	})
}

// A hero the pool did not give, such as a GovDAO pick of an operator app,
// still counts toward the operator's one place once apps have earned theirs.
func TestPickedOperatorHeroKeepsTheCap(t *testing.T) {
	var ls []listing
	var trusted []string
	for i := 1; i <= 3; i++ {
		s := fmt.Sprintf("t%02d", i)
		ls = append(ls, listing{Slug: s, Kind: kindApp, tier: tierTrusted})
		trusted = append(trusted, s)
	}
	for _, s := range []string{"e1", "e2", "e3"} {
		ls = append(ls, listing{Slug: s, Kind: kindApp, tier: tierRegistered, earned: true})
	}
	bySlug := indexListings(ls)
	for day := range int64(4) {
		res := &homeResponse{Pick: "t03", Shelves: []shelf{{Slugs: trusted}, {Slugs: []string{"e1", "e2", "e3"}}}}
		res.Time = day * secondsPerDay
		hero, _, spot := featured(res, bySlug)
		n := 0
		for _, l := range append(spot, hero) {
			if l.tier == tierTrusted {
				n++
			}
		}
		assert.LessOrEqual(t, n, 1, "day %d", day)
	}
}

// Once GovDAO names a successor every store page says so, as every page of
// the realm's markdown does: the lens, a category, a list, and the page that
// does not exist.
func TestMovedNoticeIsOnEveryPage(t *testing.T) {
	const moved = `Explore has moved: <a href="/r/gnoland/store/v1">/r/gnoland/store/v1</a>`
	build := func(to string) string {
		return `{"version":1,"moved_to":"` + to + `","shelves":[],"builders":{"top":[],"new":[]},"listings":[]}`
	}
	page := `{"version":1,"category":{"key":"defi","label":"DeFi","count":4,"confirmed":4},"page":1,"pages":1,"listings":[]}`
	list := `{"version":1,"list":{"key":"latest","count":4},"page":1,"pages":1,"listings":[]}`
	home := strings.Replace(homeBody(trustedApp), `{"key":"new",`, `{"key":"new","more":[{"key":"latest","count":4}],`, 1)
	h, _ := newTestHandler(t, map[string]string{
		"api/v1/home": home, "api/v1/build": build("gno.land/r/gnoland/store/v1"),
		"api/v1/category/defi/1": page, "api/v1/list/latest/1": list,
	})
	for _, target := range []string{"", ":build", ":c/defi", ":latest", ":nope"} {
		assert.Contains(t, render(t, h, target), moved, "on %q", target)
	}

	h, _ = newTestHandler(t, map[string]string{"api/v1/build": build("https://evil.example/x")})
	assert.NotContains(t, render(t, h, ":build"), "Explore has moved", "only a realm of this chain is a successor")
}

// Only a realm path on this chain is shown as the successor.
func TestMovedToIsARealmPath(t *testing.T) {
	h := &Handler{deps: Deps{Domain: "gno.land", RealmPath: "/r/gnoland/store"}}
	for path, want := range map[string]string{
		"gno.land/r/gnoland/store/v1": "/r/gnoland/store/v1",
		"gno.land/r/gnoland/store":    "", // itself
		"gno.land/r/":                 "",
		"gno.land/r/gnoland/":         "",
		"gno.land/r/gnoland//store":   "",
		"gno.land/p/gnoland/store/v1": "",
		"evil.land/r/gnoland/store":   "",
		"https://evil.example/x":      "",
		"":                            "",
	} {
		assert.Equal(t, want, h.realmWebPath(path), path)
	}
}

// The operator's deny list takes a listing, or a whole namespace, off every
// page at once, with no vote and no transaction.
func TestDeniedListingsAreNotShown(t *testing.T) {
	home := homeBody(trustedApp, communityApp, listingJSON("two", "nym-other", kindApp, "Two", confirmed))
	build := `{"version":1,"shelves":[],"builders":{"top":[{"namespace":"nym-acmex123","listings":1,"stars":0},{"namespace":"nym-other","listings":1,"stars":0}],"new":[]},"listings":[]}`
	h, _ := newTestHandler(t, map[string]string{"api/v1/home": home, "api/v1/build": build})
	h.deps.Denied = func(pkg string) bool { return pkg == "nym-acmex123" || strings.HasPrefix(pkg, "nym-acmex123/") }

	front := render(t, h, "")
	assert.NotContains(t, front, "Pixels")
	assert.Contains(t, front, "Two")
	lens := render(t, h, ":build")
	assert.NotContains(t, lens, "/u/nym-acmex123")
	assert.Contains(t, lens, "/u/nym-other")
}

// With an indexer, Explore shows what is used: a shelf of the front page's
// apps by distinct callers, and the count on their cards, both said to be
// the indexer's. The shelf comes after the pinned ones, a free address
// namespace takes no place on it, and the figures earn nothing: no badge,
// no owner image.
func TestUsage(t *testing.T) {
	const anon = "g17zyd8dhz9v7ndecx4wxc4z3ufmc6e43nex9cxg"
	var listings, slugs []string
	callers := map[string]int{"gno.land/r/" + anon + "/poll": 900, "gno.land/r/nym-acmex123/quiet": usedMin - 1}
	for i, slug := range []string{"aaa", "bbb", "ccc", "ddd"} {
		listings = append(listings, listingJSON(slug, "nym-acmex123", kindApp, strings.ToUpper(slug), confirmed))
		slugs = append(slugs, `"`+slug+`"`)
		callers["gno.land/r/nym-acmex123/"+slug] = 10 * (i + 1)
	}
	listings = append(listings,
		listingJSON("poll", anon, kindApp, "Poll", confirmed),
		listingJSON("quiet", "nym-acmex123", kindApp, "Quiet", confirmed))
	// New draws one row of four: Poll and Quiet first, so their cards show.
	home := `{"version":1,"categories":[],"shelves":[{"key":"new","slugs":["poll","quiet",` + strings.Join(slugs, ",") + `]},{"key":"updated","slugs":[]}],"listings":[` + strings.Join(listings, ",") + `]}`

	h, _ := newTestHandler(t, map[string]string{"api/v1/home": home})
	assert.NotContains(t, render(t, h, ""), "Most used", "no indexer, no usage")

	var asked atomic.Int32
	h.deps.Usage = func(context.Context) map[string]int { asked.Add(1); return callers }
	html := render(t, h, "")
	assert.Equal(t, int32(1), asked.Load(), "one read of the indexer's figures a page")

	pinned, shelf, ok := strings.Cut(html, "Most used this week")
	require.True(t, ok)
	assert.Contains(t, pinned, ">New</h2>", "the pinned shelves keep the head of the page")
	assert.Contains(t, shelf, "as this site&#39;s indexer counts them")
	order := []int{strings.Index(shelf, ">DDD<"), strings.Index(shelf, ">CCC<"), strings.Index(shelf, ">BBB<"), strings.Index(shelf, ">AAA<")}
	assert.IsIncreasing(t, order, "most callers first")
	assert.NotContains(t, shelf, "Poll", "a free address namespace is not on the shelf")
	assert.NotContains(t, shelf, "Quiet", "nor an app below usedMin")
	assert.Contains(t, shelf, "40 users this week")
	assert.Contains(t, pinned, "900 users this week", "its card still says how many call it")
	assert.NotContains(t, pinned, strconv.Itoa(usedMin-1)+" users this week", "a count below usedMin is not shown")

	// The figures are shown, and earn nothing.
	assert.NotContains(t, html, "Established")
	assert.NotContains(t, html, `class="icon" src=`, "usage shows no owner image")
	assert.NotContains(t, html, `class="cover" src=`)

	// While the indexer has nothing to say, the page is the chain's own.
	h.deps.Usage = func(context.Context) map[string]int { return nil }
	assert.NotContains(t, render(t, h, ""), "users this week")
}

// A claim stays unverified whatever an indexer counts: the two are said
// side by side, and neither stands for the other.
func TestUsageDoesNotVerifyAClaim(t *testing.T) {
	h, _ := newTestHandler(t, nil)
	claim := listing{Slug: "wugnot", Kind: kindApp, Path: "gno.land/r/nym-acmex123/wugnot", tier: tierRegistered}
	used := h.card(&claim, nil, map[string]int{claim.Path: 845})
	assert.True(t, used.Unverified)
	assert.Equal(t, 845, used.Users)
	claim.Confirmed = true
	assert.False(t, h.card(&claim, nil, nil).Unverified)
}

// The snippet to list an app can be called again: an init runs once.
func TestSnippetCanRelist(t *testing.T) {
	h, _ := newTestHandler(t, map[string]string{"api/v1/home": homeBody(trustedApp)})
	html := render(t, h, "")
	assert.Contains(t, html, "func init(cur realm) { Relist(cross(cur)) }")
	assert.Contains(t, html, "func Relist(cur realm) {")
}
