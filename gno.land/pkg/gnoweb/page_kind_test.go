package gnoweb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

func TestPagePolicyKind(t *testing.T) {
	t.Parallel()

	p := pagePolicy{
		trusted: newTrustedPaths([]string{"gnoland"}),
		aliases: map[string]AliasTarget{"/about": {Value: "# About", Kind: StaticMarkdown}},
	}
	cases := map[string]pageKind{
		"/r/gnoland/home":      pageOfficial,
		"/r/gnoland/blog:p/x":  pageOfficial,
		"/p/gnoland/blog/v0":   pageOfficial,
		"/u/gnoland":           pageOfficial,
		"/r/gnoland-evil/home": pageCommunity,
		"/r/nym/app":           pageCommunity,
		"/p/nym/lib":           pageCommunity,
		"/u/nym":               pageCommunity,
		"/u/gnoland/airdrop":   pageCommunity,
		"/r/":                  pageSite,
		"/p/":                  pageSite,
		"/about":               pageOperator,
		"/other":               pageSite,
		"/":                    pageSite,
	}
	for path, want := range cases {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			u, err := weburl.Parse(path)
			require.NoError(t, err)
			assert.Equal(t, want, p.kind(u))
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

func TestPathTitle(t *testing.T) {
	t.Parallel()

	const addr = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	cases := map[string]string{
		"/r/nym/app":             "app · realm by nym",
		"/r/nym/app/":            "app · realm by nym",
		"/r/nym/games/chess":     "games/chess · realm by nym",
		"/r/nym/app:p/hello?x=1": "app · realm by nym",
		"/r/nym/app$source":      "app · realm by nym",
		"/p/nym/lib":             "lib · package by nym",
		"/p/nym/lib/v2":          "lib/v2 · package by nym",
		"/u/nym":                 "nym · user profile",
		"/u/" + addr:             "g1jg...sqf5 · user profile",
		"/r/" + addr + "/app":    "app · realm by g1jg...sqf5",
		"/r/nym":                 "realms by nym",
		"/r/nym/":                "realms by nym",
		"/p/nym":                 "packages by nym",
		"/r/":                    "/r",
		"/about":                 "/about",
		"/":                      "",
	}
	for path, want := range cases {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			u, err := weburl.Parse(path)
			require.NoError(t, err)
			assert.Equal(t, want, pathTitle(u))
		})
	}
}
