package gnoweb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

func TestPackageKind(t *testing.T) {
	t.Parallel()

	h := &HTTPHandler{trusted: newTrustedPaths([]string{"gnoland"})}
	cases := map[string]pageKind{
		"/r/gnoland/home":      pageOfficial,
		"/r/gnoland/blog:p/x":  pageOfficial,
		"/p/gnoland/blog/v0":   pageOfficial,
		"/u/gnoland":           pageOfficial,
		"/r/gnoland-evil/home": pageCommunity,
		"/r/nym/app":           pageCommunity,
		"/p/nym/lib":           pageCommunity,
		"/u/nym":               pageCommunity,
		"/r/":                  pageSite,
		"/p/":                  pageSite,
		"/about":               pageSite,
		"/":                    pageSite,
	}
	for path, want := range cases {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			u, err := weburl.Parse(path)
			require.NoError(t, err)
			assert.Equal(t, want, h.packageKind(u))
		})
	}
}

func TestPageKindMayRepeat(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		kind pageKind
		url  string
		want bool
	}{
		{"official", pageOfficial, "/r/gnoland/blog", true},
		{"official view", pageOfficial, "/r/gnoland/blog$source", true},
		{"community", pageCommunity, "/r/nym/app", false},
		{"site", pageSite, "/r/", false},
		{"args reach Render", pageOfficial, "/r/gnoland/blog:t/Claim_At_Evil", false},
		{"a query reaches Render", pageOfficial, "/r/gnoland/blog?page=2", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			u, err := weburl.Parse(tc.url)
			require.NoError(t, err)
			assert.Equal(t, tc.want, tc.kind.mayRepeat(u))
		})
	}
}
