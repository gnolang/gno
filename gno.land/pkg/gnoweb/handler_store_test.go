package gnoweb

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/feature/chainmap"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

func TestStoreHeaderEntry(t *testing.T) {
	t.Parallel()

	client := NewMockClient(
		&MockPackage{Domain: "gno.land", Path: "/r/demo/app", Files: map[string]string{"render.gno": `package app; func Render(string) string { return "hi" }`}},
		&MockPackage{Domain: "gno.land", Path: "/r/gnoland/store", Files: map[string]string{"render.gno": `package store; func Render(string) string { return "store" }`}},
	)
	for _, tc := range []struct {
		name, realm, path, want string
	}{
		{"absent", "", "/r/demo/app", ""},
		{"on any page", "/r/gnoland/store", "/r/demo/app", `<a href="/r/gnoland/store" class="store-link">`},
		{"current on the store", "/r/gnoland/store", "/r/gnoland/store", `<a href="/r/gnoland/store" class="store-link" aria-current="page">`},
		{"inside the store on a store page", "/r/gnoland/store", "/r/gnoland/store:c/defi", `<a href="/r/gnoland/store" class="store-link" aria-current="true">`},
		{"not current on another tab", "/r/gnoland/store", "/r/gnoland/store$source", `<a href="/r/gnoland/store" class="store-link">`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h, err := NewHTTPHandler(slog.New(slog.DiscardHandler), &HTTPHandlerConfig{
				ClientAdapter: client,
				Renderer:      NewHTMLRenderer(slog.New(slog.DiscardHandler), NewDefaultRenderConfig(), client),
				Aliases:       map[string]AliasTarget{},
				Meta:          StaticMetadata{Domain: "gno.land"},
				StoreRealm:    tc.realm,
			})
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, tc.path, nil))
			require.Equal(t, http.StatusOK, rr.Code)
			if tc.want == "" {
				assert.NotContains(t, rr.Body.String(), "store-link")
				return
			}
			assert.Contains(t, rr.Body.String(), tc.want)
		})
	}
}

// A store that NewHTTPHandler cannot serve is an error, not a panic or a
// store that never matches.
func TestStoreConfigIsValidated(t *testing.T) {
	t.Parallel()

	client := NewMockClient()
	for _, tc := range []struct {
		name, realm, domain, gateway, want string
	}{
		{"no domain", "/r/gnoland/store", "", "", "Meta.Domain"},
		{"no leading slash", "r/gnoland/store", "gno.land", "", "StoreRealm"},
		{"not a realm", "/p/nt/avl/v0", "gno.land", "", "StoreRealm"},
		{"a directory", "/r/gnoland/store/", "gno.land", "", "StoreRealm"},
		{"gateway with a path", "/r/gnoland/store", "gno.land", "https://dweb.link/ipfs/", "IPFSGateway"},
		{"gateway with no scheme", "/r/gnoland/store", "gno.land", "dweb.link", "IPFSGateway"},
		{"gateway smuggling a CSP directive", "/r/gnoland/store", "gno.land", "https://a;script-src", "IPFSGateway"},
		{"gateway with a quote", "/r/gnoland/store", "gno.land", "https://a'b", "IPFSGateway"},
		{"valid", "/r/gnoland/store", "gno.land", "https://dweb.link", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewHTTPHandler(slog.New(slog.DiscardHandler), &HTTPHandlerConfig{
				ClientAdapter: client,
				Renderer:      NewHTMLRenderer(slog.New(slog.DiscardHandler), NewDefaultRenderConfig(), client),
				Aliases:       map[string]AliasTarget{},
				Meta:          StaticMetadata{Domain: tc.domain},
				StoreRealm:    tc.realm,
				IPFSGateway:   tc.gateway,
			})
			if tc.want == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.want)
		})
	}
}

// storeClient answers the store realm's api/v1 queries with JSON, and every
// other query as the mock does.
type storeClient struct {
	*MockClient
	api map[string]string
}

func (c storeClient) Realm(ctx context.Context, path, args string) ([]byte, error) {
	if body, ok := c.api[args]; ok {
		return []byte(body), nil
	}
	return c.MockClient.Realm(ctx, path, args)
}

// The store's Content tab is its own view, with its own title and status;
// a markdown client still gets the realm's markdown.
func TestStoreViewIsServed(t *testing.T) {
	t.Parallel()

	client := storeClient{
		MockClient: NewMockClient(&MockPackage{
			Domain:    "gno.land",
			Path:      "/r/gnoland/store",
			Files:     map[string]string{"render.gno": `package store; func Render(string) string { return "store" }`},
			Functions: []*doc.JSONFunc{{Name: "Render", Params: []*doc.JSONField{{Name: "path", Type: "string"}}, Results: []*doc.JSONField{{Type: "string"}}}},
		}),
		api: map[string]string{
			"api/v1/home":  `{"version":1,"categories":[],"shelves":[],"listings":[]}`,
			"api/v1/build": `{"version":1,"shelves":[],"builders":{"top":[],"new":[]},"listings":[]}`,
		},
	}
	h, err := NewHTTPHandler(slog.New(slog.DiscardHandler), &HTTPHandlerConfig{
		ClientAdapter: client,
		Renderer:      NewHTMLRenderer(slog.New(slog.DiscardHandler), NewDefaultRenderConfig(), client),
		Aliases:       map[string]AliasTarget{},
		Meta:          StaticMetadata{Domain: "gno.land"},
		TrustedPaths:  []string{"gnoland"},
		StoreRealm:    "/r/gnoland/store",
	})
	require.NoError(t, err)

	for _, tc := range []struct {
		name, path, accept string
		status             int
		want, notWant      string
	}{
		{"front page", "/r/gnoland/store", "", http.StatusOK, "<title>Explore · gno.land</title>", "[gno.land]/r/gnoland/store"},
		// gnoweb writes these titles itself: they hold under render args too.
		{"build lens", "/r/gnoland/store:build", "", http.StatusOK, "<title>Build · Explore - gno.land</title>", "store · realm by"},
		{"not found", "/r/gnoland/store:nope", "", http.StatusNotFound, "There is no such page in Explore.", "[gno.land]/r/gnoland/store"},
		{"markdown", "/r/gnoland/store", "text/markdown", http.StatusOK, "[gno.land]/r/gnoland/store", "b-store"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.accept != "" {
				req.Header.Set("Accept", tc.accept)
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			assert.Equal(t, tc.status, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.want)
			assert.NotContains(t, rr.Body.String(), tc.notWant)
		})
	}
}

// Explore reads the real realm, not an example of it: this decodes what
// store/v0 answers from genesis, through a node, into the pages gnoweb
// draws. The api/v1 contract is frozen with the realm, so a field the two
// disagree on must fail here.
func TestStoreReadsTheRealRealm(t *testing.T) {
	cfg := NewDefaultAppConfig()
	cfg.NodeRemote = sharedNodeRemote(t)
	cfg.StoreRealm = "/r/gnoland/store/v0"
	router, err := NewRouter(slog.New(slog.DiscardHandler), cfg)
	require.NoError(t, err)

	for _, tc := range []struct {
		path string
		want []string
	}{
		{"/r/gnoland/store/v0", []string{
			`<div class="b-store u-grid-full">`,              // the store's view, not the realm's markdown
			`>gno.land essentials</h2>`,                      // the realm's "core" shelf, worded by gnoweb
			`<use href="#ico-check"></use></svg>Core</span>`, // "core" is read
			`href="/r/gnoland/store/v0:c/social"`,            // categories and their counts
			`Block `,                                         // height and time
		}},
		{"/r/gnoland/store/v0:c/social", []string{
			"<h1>Social</h1>", "<p>1 app</p>", // "confirmed" is read
			`<h3><a href="/r/gnoland/boards2/v0">Boards</a></h3>`,
			`<span class="u-font-mono" title="/r/gnoland/boards2/v0">`,
		}},
		{"/r/gnoland/store/v0:latest", []string{"<h1>New apps</h1>", `>Valopers</a></h3>`}}, // a paged list and its count
		{"/r/gnoland/store/v0:build", []string{"<h1>Build</h1>", `>AVL Tree</a></h3>`, `href="/u/moul"`}},
	} {
		t.Run(tc.path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, tc.path, nil))
			require.Equal(t, http.StatusOK, rr.Code)
			for _, want := range tc.want {
				assert.Contains(t, rr.Body.String(), want)
			}
			assert.NotContains(t, rr.Body.String(), `>unverified</span>`, "the seed is confirmed")
		})
	}
}

// A store realm outside -trusted-paths is a community page like any other:
// its own words, the category labels, stay out of the head. One gnoweb
// vouches for names its pages with them.
func TestStoreRealmLabelsInTheHead(t *testing.T) {
	t.Parallel()

	client := storeClient{
		MockClient: NewMockClient(&MockPackage{
			Domain: "gno.land", Path: "/r/acme/store",
			Files:     map[string]string{"render.gno": `package store; func Render(string) string { return "store" }`},
			Functions: []*doc.JSONFunc{{Name: "Render", Params: []*doc.JSONField{{Name: "path", Type: "string"}}, Results: []*doc.JSONField{{Type: "string"}}}},
		}),
		api: map[string]string{
			"api/v1/home":            `{"version":1,"categories":[{"key":"defi","label":"Official gno.land wallet","count":1,"confirmed":1}],"shelves":[],"listings":[]}`,
			"api/v1/category/defi/1": `{"version":1,"category":{"key":"defi","label":"Official gno.land wallet","count":1,"confirmed":1},"page":1,"pages":1,"listings":[]}`,
		},
	}
	for _, tc := range []struct {
		name    string
		trusted []string
		inHead  bool
	}{
		{"outside the trusted paths", nil, false},
		{"under a trusted path", []string{"acme"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			h, err := NewHTTPHandler(slog.New(slog.DiscardHandler), &HTTPHandlerConfig{
				ClientAdapter: client,
				Renderer:      NewHTMLRenderer(slog.New(slog.DiscardHandler), NewDefaultRenderConfig(), client),
				Aliases:       map[string]AliasTarget{},
				Meta:          StaticMetadata{Domain: "gno.land"},
				TrustedPaths:  tc.trusted,
				StoreRealm:    "/r/acme/store",
			})
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/r/acme/store:c/defi", nil))
			require.Equal(t, http.StatusOK, rr.Code)
			body := rr.Body.String()
			head, _, _ := strings.Cut(body, "</head>")
			assert.Contains(t, body, "<h1>Official gno.land wallet</h1>", "the page is the store's either way")
			assert.Equal(t, tc.inHead, strings.Contains(head, "Official gno.land wallet"), head)
		})
	}
}

// NewHTTPHandler normalizes a copy of its config: a caller may share one
// between handlers built at once.
func TestNewHTTPHandlerLeavesItsConfigAlone(t *testing.T) {
	t.Parallel()

	client := NewMockClient()
	cfg := &HTTPHandlerConfig{
		ClientAdapter: client,
		Renderer:      NewHTMLRenderer(slog.New(slog.DiscardHandler), NewDefaultRenderConfig(), client),
		Aliases:       map[string]AliasTarget{},
		Meta:          StaticMetadata{Domain: "gno.land"},
		IPFSGateway:   "https://DWEB.link/",
	}
	want := *cfg
	h, err := NewHTTPHandler(slog.New(slog.DiscardHandler), cfg)
	require.NoError(t, err)
	require.NotNil(t, h)
	assert.Equal(t, want.IPFSGateway, cfg.IPFSGateway)
}

// Explore shows nothing of an aggregate that left part of the week unread:
// its counts are lower bounds, and would rank apps wrongly.
func TestUsageOf(t *testing.T) {
	t.Parallel()

	callers := map[string]int{"gno.land/r/acme/swap": 3}
	assert.Nil(t, usageOf(nil))
	assert.Nil(t, usageOf(&chainmap.Activity{Callers: callers, Partial: true}))
	assert.Equal(t, callers, usageOf(&chainmap.Activity{Callers: callers}))
}

// A deny entry is taken as an operator types or pastes it, and one that
// could match nothing stops the start: a deny must not fail quietly.
func TestStoreDenyEntries(t *testing.T) {
	t.Parallel()

	for _, entry := range []string{
		"scam", " scam/ ", "scam/app", "/r/scam/app", "r/scam/app", "gno.land/r/scam/app", "gno.land/p/scam/app",
		"https://gno.land/r/scam/app", "https://gno.land/r/scam/app:c/defi?page=2", "/r/scam/app$source",
		"https://gno.land/r/scam/app/render.gno", "/r/scam/app/README.md", // a source page
		"https://gno.land/r/scam/app#readme", "gno.land/r/scam/app?x=1#top", // an anchor
	} {
		got, err := storeDenyEntries([]string{entry, " "}, "gno.land")
		require.NoError(t, err, entry)
		require.Len(t, got, 1, entry)
		assert.True(t, storeDenied(got)("scam/app"), "%q denies scam/app as %q", entry, got[0])
		assert.False(t, storeDenied(got)("scammer/app"), "%q must not deny another namespace", entry)
	}
	for _, entry := range []string{
		"Scam", "*", "scam/*", "sc am", "other.land/r/scam/app", "/r/", "https://gno.land",
		"gno.land", " gno.land ", "gno.land/", "gno.land:443", // the bare domain
		"r", "/r", "gno.land/r", "p", "u", "https://gno.land/r", // a tree, with nothing under it
		"gno.landscam", "localhost:8888/r/scam/app", "//scam/app", "gno.land//scam/app", // would name another namespace
		"scam/app.", "gno.land/r/scam/app.", "scam/App", "scam/app.gno", "scam:app", // would deny all of scam
		"scam$app", "scam?app", "scam#app", "/r/scam$app", "scam%24app\u0455", "u/r", "/p/u",
		"/r/1inch/swap", "gno.land/r/1inch/swap", "/r/-scam/app", "/p/_x", "/u/9lives", // no namespace is named so
		"acme/swap/V2", "acme/swap/x.", "acme/swap/.gno", "c/defi//.&", "c/defi/&.p", "scam/app//", "c/defi//gno.land", "a/0//0.0", "//gno/source/http://", "c/defi/%2Fgno.land", // a last name that is no file, an empty one
	} {
		got, err := storeDenyEntries([]string{entry}, "gno.land")
		assert.ErrorContains(t, err, "StoreDeny", "%q is taken as %q", entry, got)
	}

	client := NewMockClient()
	_, err := NewHTTPHandler(slog.New(slog.DiscardHandler), &HTTPHandlerConfig{
		ClientAdapter: client,
		Renderer:      NewHTMLRenderer(slog.New(slog.DiscardHandler), NewDefaultRenderConfig(), client),
		Aliases:       map[string]AliasTarget{},
		Meta:          StaticMetadata{Domain: "gno.land"},
		StoreDeny:     []string{"Scam"},
	})
	require.ErrorContains(t, err, "StoreDeny")
}

// A -store-deny entry takes its listings off Explore at once.
func TestStoreDenyHidesAListing(t *testing.T) {
	t.Parallel()

	listing := func(ns, title string) string {
		return `{"slug":"` + ns + `-app","kind":"app","path":"gno.land/r/` + ns + `/app","title":"` + title + `","tagline":"x","category":"defi","palette":0,"confirmed":true}`
	}
	client := storeClient{
		MockClient: NewMockClient(&MockPackage{
			Domain: "gno.land", Path: "/r/gnoland/store",
			Files:     map[string]string{"render.gno": `package store; func Render(string) string { return "store" }`},
			Functions: []*doc.JSONFunc{{Name: "Render", Params: []*doc.JSONField{{Name: "path", Type: "string"}}, Results: []*doc.JSONField{{Type: "string"}}}},
		}),
		api: map[string]string{"api/v1/home": `{"version":1,"categories":[],"shelves":[{"key":"new","slugs":["scam-app","fine-app"]}],"listings":[` + listing("scam", "Free Airdrop") + `,` + listing("fine", "Fine App") + `]}`},
	}
	h, err := NewHTTPHandler(slog.New(slog.DiscardHandler), &HTTPHandlerConfig{
		ClientAdapter: client,
		Renderer:      NewHTMLRenderer(slog.New(slog.DiscardHandler), NewDefaultRenderConfig(), client),
		Aliases:       map[string]AliasTarget{},
		Meta:          StaticMetadata{Domain: "gno.land"},
		StoreRealm:    "/r/gnoland/store",
		StoreDeny:     []string{" scam/ "},
	})
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/r/gnoland/store", nil))
	require.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Fine App")
	assert.NotContains(t, rr.Body.String(), "Free Airdrop")
}
