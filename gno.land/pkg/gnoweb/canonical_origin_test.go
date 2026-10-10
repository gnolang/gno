package gnoweb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

func TestNormalizeCanonicalOrigin(t *testing.T) {
	t.Parallel()

	// Harmless spellings are fixed, not refused: a refusal stops gnoweb.
	for in, want := range map[string]string{
		"":                       "",
		"https://gno.land":       "https://gno.land",
		"http://localhost:8888":  "http://localhost:8888",
		"https://gno.land/":      "https://gno.land",
		"https://gno.land//":     "https://gno.land",
		" https://gno.land \n":   "https://gno.land",
		"HTTPS://GNO.land":       "https://gno.land",
		"https://gno.land:443":   "https://gno.land",
		"http://gno.land:80":     "http://gno.land",
		"https://gno.land:8443":  "https://gno.land:8443",
		"http://gno.land:443":    "http://gno.land:443",
		"http://[::1]:8888":      "http://[::1]:8888",
		"https://gno.land:1":     "https://gno.land:1",
		"https://gno.land:65535": "https://gno.land:65535",
	} {
		got, err := normalizeCanonicalOrigin(in)
		if assert.NoError(t, err, "origin %q", in) {
			assert.Equal(t, want, got, "origin %q", in)
		}
	}

	for _, in := range []string{
		"gno.land",
		"ftp://gno.land",
		"https://",
		"https://gno.land/r/demo",
		"https://gno.land?x=1",
		"https://gno.land#top",
		"https://user@gno.land",
		"javascript:alert(1)",
		"https://gno.land\n/evil",
		"https://gno.land?",
		"https://gno.land#",
		"https:gno.land",
		"https://gno.land:",
		"https://gno.land:x:443",
		"https://gno.land:443:443",
		"https://gno.land:0",
		"https://gno.land:65536",
		"https://gno.land:99999",
		"http://[::1]:99999",
	} {
		_, err := normalizeCanonicalOrigin(in)
		assert.Error(t, err, "origin %q", in)
	}
}

// TestCanonicalURLArgs checks that args stay in the canonical only where they
// reach a realm's Render: a listing shows the same paths whatever args a link
// adds, while a namespace listing reads like a realm and an alias key's args
// are the operator's, so both keep them.
func TestCanonicalURLArgs(t *testing.T) {
	t.Parallel()

	h := &HTTPHandler{
		Static: StaticMetadata{CanonicalOrigin: "https://gno.land"},
		policy: newPagePolicy(nil, DefaultAliases, IndexNoCommunity),
	}
	for in, want := range map[string]string{
		"/r/:Official_GNOT_airdrop_claim_at_evil.example":              "/r/",
		"/u/gnoland:Official_GNOT_airdrop_claim_at_evil.example":       "/u/gnoland",
		"/r/gnoland/blog/:Official_GNOT_airdrop_claim_at_evil.example": "/r/gnoland/blog/",
		"/r/gnoland/blog:t/news":                                       "/r/gnoland/blog:t/news",
		"/r/gnoland:t/news":                                            "/r/gnoland:t/news",
		"/news:latest":                                                 "/news:latest",
		"/u/docs:x":                                                    "/docs",
	} {
		u, err := weburl.Parse(in)
		require.NoError(t, err, in)
		assert.Equal(t, "https://gno.land"+want, h.canonicalURL(u), in)
	}
}
