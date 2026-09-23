package markdown

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

const (
	testGateway = "https://gw.test"
	testCIDv1   = "bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi"
	testCIDv0   = "QmbWqxBEKC3P8tqsKc98xmWNzrzDtRLMiMPL8wBuTGsMnR"
	testIPNSKey = "k51qzi5uqu5dlvj2baxnqndepeb86cbk3ng7n3i46uzyxzyqj2xjonzllnv0v8"
)

func TestRewriteIPFSURL(t *testing.T) {
	t.Parallel()

	gwIPFS := testGateway + "/ipfs/"
	rewritten := []struct{ in, want string }{
		{"ipfs://" + testCIDv1, gwIPFS + testCIDv1},
		{"ipfs://" + testCIDv1 + "/dir/a%20b.png?filename=x.png#top", gwIPFS + testCIDv1 + "/dir/a%20b.png?filename=x.png#top"},
		{"IPFS://" + testCIDv0 + "/", gwIPFS + testCIDv0 + "/"},
		{"ipns://en.wikipedia-on-ipfs.org/wiki/", testGateway + "/ipns/en.wikipedia-on-ipfs.org/wiki/"},
		{"ipns://" + testIPNSKey, testGateway + "/ipns/" + testIPNSKey},

		// Retired path gateways.
		{"https://ipfs.io/ipfs/" + testCIDv0 + "/img.png", gwIPFS + testCIDv0 + "/img.png"},
		{"http://IPFS.IO/ipfs/" + testCIDv1, gwIPFS + testCIDv1},
		{"https://gateway.ipfs.io/ipns/docs.ipfs.tech/x", testGateway + "/ipns/docs.ipfs.tech/x"},
		{"https://dweb.link/ipfs/" + testCIDv1 + "?a=1", gwIPFS + testCIDv1 + "?a=1"},
		{"https://cloudflare-ipfs.com/ipfs/" + testCIDv1, gwIPFS + testCIDv1},
		{"https://cf-ipfs.com/ipfs/" + testCIDv1, gwIPFS + testCIDv1},
		{"https://nftstorage.link/ipfs/" + testCIDv1, gwIPFS + testCIDv1},
		{"https://w3s.link/ipfs/" + testCIDv1, gwIPFS + testCIDv1},

		// Retired subdomain gateways.
		{"https://" + testCIDv1 + ".ipfs.dweb.link/img.png", gwIPFS + testCIDv1 + "/img.png"},
		{"https://" + testCIDv1 + ".ipfs.w3s.link", gwIPFS + testCIDv1},
		{"https://" + testCIDv1 + ".ipfs.nftstorage.link/", gwIPFS + testCIDv1 + "/"},
		{"https://" + testCIDv1 + ".ipfs.cf-ipfs.com/", gwIPFS + testCIDv1 + "/"},
		{"https://en-wikipedia--on--ipfs-org.ipns.dweb.link/wiki/", testGateway + "/ipns/en.wikipedia-on-ipfs.org/wiki/"},
		{"https://" + testIPNSKey + ".ipns.dweb.link/", testGateway + "/ipns/" + testIPNSKey + "/"},
	}
	for _, tc := range rewritten {
		got, ok := rewriteIPFSURL(testGateway, tc.in)
		assert.True(t, ok, "expected %q to be rewritten", tc.in)
		assert.Equal(t, tc.want, got, "input %q", tc.in)
		assert.True(t, mayBeIPFSURL([]byte(tc.in)), "the pre-check must let %q through", tc.in)
	}

	untouched := []string{
		"https://ipfs.filebase.io/ipfs/" + testCIDv1, // a gateway that still works
		"https://example.com/ipfs/" + testCIDv1,
		"https://ipfs.io/",
		"https://ipfs.io/docs",
		"https://ipfs.io/ipfs/",
		"https://ipfs.io/ipfs//x",
		"https://ipfs.io:8443/ipfs/" + testCIDv1,
		"https://user@ipfs.io/ipfs/" + testCIDv1,
		"https://ipfs.io/ipfs/%41" + testCIDv1,
		"https://ipfs.io/ipfs/" + testCIDv1 + "%2F..%2Fx",
		"https://" + testCIDv1 + ".ipfs.example.com/",
		"https://" + testCIDv1 + ".ipfs.dweb.link.example.com/",
		"https://x." + testCIDv1 + ".ipfs.dweb.link/",
		"https://" + testCIDv1 + ".ipfz.dweb.link/",
		"ipfs:///x",
		"ipfs:" + testCIDv1,
		"ipfs://" + testCIDv1 + ":80/x",
		"ipfs://user@" + testCIDv1,
		"ipfs://bafy.example",
		"ipfs://bafy-x",
		"ipns://.example.com",
		"ipns://-example.com",
		"ipns://example..com",
		"ipns://example.com.",
		"ipfs://" + strings.Repeat("a", 254),
		"javascript:alert(1)",
		"mailto:a@b.c",
		"/r/demo/foo",
		"",
	}
	for _, in := range untouched {
		got, ok := rewriteIPFSURL(testGateway, in)
		assert.False(t, ok, "expected %q to be left alone, got %q", in, got)
	}
}

func renderIPFS(t *testing.T, src, gateway string) string {
	t.Helper()
	gnourl, err := weburl.Parse("https://gno.land/r/test")
	require.NoError(t, err)
	m := goldmark.New()
	NewGnoExtension(WithImageValidator(AllowSvgDataImage)).Extend(m)
	ctx := parser.WithContext(NewGnoParserContext(GnoContext{GnoURL: gnourl, IPFSGateway: gateway}))
	var out bytes.Buffer
	require.NoError(t, m.Convert([]byte(src), &out, ctx))
	return out.String()
}

var reURLAttr = regexp.MustCompile(`(?:href|src)="[^"]*"`)

// TestIPFSRenderMatchesGatewayURL checks that a rewritten destination
// renders exactly like the same gateway URL written by hand. Destinations
// are resolved again at render time, so this is what catches escapes and
// character references being decoded twice.
func TestIPFSRenderMatchesGatewayURL(t *testing.T) {
	t.Parallel()

	suffixes := []string{
		"",
		"/a.png",
		"/?a=1&amp;amp;b=2",
		`/?q=a\\b&c=d`,
		`/?q=a\\!b`, // a backslash before punctuation is unescaped at render time
		"/a%20b?x=%26#frag",
		"/&#x61;.png",
	}
	sources := []string{
		"ipfs://" + testCIDv1,
		"&#x69;pfs://" + testCIDv1,
		"https://ipfs.io/ipfs/" + testCIDv1,
		"https://" + testCIDv1 + ".ipfs.dweb.link",
	}
	forms := []string{"[x](%s)", "![x](%s)", "<%s>", "[x][r]\n\n[r]: %s"}

	for _, form := range forms {
		for _, suffix := range suffixes {
			want := renderIPFS(t, fmt.Sprintf(form, testGateway+"/ipfs/"+testCIDv1+suffix), testGateway)
			for _, source := range sources {
				got := renderIPFS(t, fmt.Sprintf(form, source+suffix), testGateway)
				switch {
				case form != "<%s>":
					assert.Equal(t, want, got, "form %q, source %q, suffix %q", form, source, suffix)
				case !strings.HasPrefix(source, "&"): // autolinks keep their own label
					assert.Equal(t, reURLAttr.FindAllString(want, -1), reURLAttr.FindAllString(got, -1),
						"form %q, source %q, suffix %q", form, source, suffix)
				}
			}
		}
	}
}

// The golden files under golden/ext_ipfs pin the rewritten output; these
// cover what that harness cannot, since it always sets a gateway.
func TestIPFSRenderUntouched(t *testing.T) {
	t.Parallel()

	t.Run("no gateway", func(t *testing.T) {
		t.Parallel()
		out := renderIPFS(t, "[x](ipfs://"+testCIDv1+") ![y](https://ipfs.io/ipfs/"+testCIDv1+")", "")
		assert.Contains(t, out, `href="ipfs://`+testCIDv1+`"`)
		assert.Contains(t, out, `src="https://ipfs.io/ipfs/`+testCIDv1+`"`)
	})

	t.Run("other urls", func(t *testing.T) {
		t.Parallel()
		src := "[a](https://example.com/ipfs/" + testCIDv1 + ") [b](/r/demo/foo) [c](javascript:alert(1))"
		assert.Equal(t, renderIPFS(t, src, ""), renderIPFS(t, src, testGateway))
	})
}
