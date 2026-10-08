package gnoweb

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
