package gnoweb

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		api: map[string]string{"api/v1/home": `{"version":1,"categories":[],"shelves":[],"listings":[]}`},
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
