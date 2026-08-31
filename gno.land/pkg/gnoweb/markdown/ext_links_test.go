package markdown

import (
	"bytes"
	"net/url"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

// resolveDestination undoes three things before the scheme is checked:
// backslash-escaped punctuation, numeric references, and named entities.
// Every one of them hides a dangerous scheme from a raw-bytes check, so each
// needs its own case here — dropping any single resolution step must fail a
// test rather than silently reopen the bypass.
func TestExtLinksRejectsObfuscatedDangerousURLs(t *testing.T) {
	cases := []string{
		`[x](&#x6a;avascript:alert(1))`,
		`[x](&#106;avascript:alert(1))`,
		// Leading zeros make goldmark parse the reference as octal
		// (strconv.ParseUint with base 0), so 0000152 is also 'j'.
		`[x](&#0000152;avascript:alert(1))`,
		`[x](javascript&colon;alert(1))`,
		`[x](&#x64;ata:text/html,x)`,
		`[x](&#118;bscript:msgbox(1))`,
		`[x](file&colon;/etc/passwd)`,
		`[x](&#x20;javascript:alert(1))`,
		`[x](&#x09;javascript:alert(1))`,
		`[x](&#10;data:text/html,x)`,
		// Backslash-escaped punctuation. UnescapePunctuations turns `\:`
		// into `:`, so these are dangerous only after resolution — exactly
		// like the entity forms above, and affected by the same
		// check-the-raw-bytes mistake.
		`[x](javascript\:alert(1))`,
		`[x](javascript\:alert\(1\))`,
		`[x](vbscript\:msgbox(1))`,
		`[x](file\:/etc/passwd)`,
		`[x](data\:text/html,x)`,
		// Both mechanisms at once.
		`[x](&#x6a;avascript\:alert(1))`,
		// The three resolvers run over each other's OUTPUT, not over the
		// original bytes, so a numeric reference can manufacture the `&`
		// that the entity-name pass then consumes: `&#38;colon;` resolves
		// to `&colon;`, which resolves to `:`. resolveDestination has to
		// chain them in URLEscape's order to see that; running the passes
		// independently over the raw destination would not.
		`[x](javascript&#38;colon;alert(1))`,
		`[x](javascript&#x26;colon;alert(1))`,
		`[x](data&#38;colon;text/html,x)`,
	}
	for _, src := range cases {
		t.Run(src, func(t *testing.T) {
			html := renderExtLinks(t, src)
			// The destination must never reach an href. Either the anchor
			// renders empty, or the link is dropped as invalid — both are
			// safe, so assert the property rather than one of the shapes.
			require.NotRegexp(t, `<a href="[^"]`, html)
			require.NotContains(t, strings.ToLower(html), "javascript:")
			require.NotContains(t, strings.ToLower(html), "vbscript:")
			require.NotContains(t, strings.ToLower(html), "data:")
			require.NotContains(t, strings.ToLower(html), "file:")
		})
	}
}

func TestExtLinksPreservesSafeURLs(t *testing.T) {
	cases := map[string]string{
		`[x](https://example.com/path)`:                   `href="https://example.com/path"`,
		`[x](?x=1&y=2)`:                                   `href="?x=1&amp;y=2"`,
		`[x](/r/boards$help&func=DeleteThread&boardID=2)`: `href="/r/boards$help&amp;func=DeleteThread&amp;boardID=2"`,
	}
	for src, want := range cases {
		t.Run(src, func(t *testing.T) {
			require.Contains(t, renderExtLinks(t, src), want)
		})
	}
}

// A destination whose resolved form net/url refuses to parse loses its
// anchor rather than being rendered as a broken relative href — but it keeps
// its TEXT. The unparsable class is wider than the attacks that motivate it:
// an entity-spelled `%` in an absolute path (`&percnt;`, `&#x25;`) resolves
// to a bare `%` that net/url rejects, and deleting the anchor text there
// would silently erase author-visible copy from a page.
func TestExtLinksDropsUnparsableResolvedDestinations(t *testing.T) {
	for _, src := range []string{
		`[x](&#x09;https://example.com/path)`,
		`[x](https://exam&#10;ple.com/path)`,
		`[x](https://example.com/50&percnt;)`,
		`[x](/r/x$help&func=F&v=50&percnt;)`,
	} {
		t.Run(src, func(t *testing.T) {
			got := renderExtLinks(t, src)
			require.Contains(t, got, "<!-- invalid link -->")
			require.NotContains(t, got, "<a ")
			require.NotContains(t, got, `href=`)
			// The link text survives as plain inline content.
			require.Contains(t, got, ">x<")
		})
	}
}

// The link classifier drives rel="noopener nofollow ugc" and the
// external-link icon, so it has to resolve the destination too: an
// entity-encoded scheme must not buy a foreign URL the chrome of an
// on-chain one.
func TestExtLinksClassifiesResolvedDestination(t *testing.T) {
	want := renderExtLinks(t, `[x](https://evil.example/p)`)
	require.Contains(t, want, `rel="noopener nofollow ugc"`)
	require.Contains(t, want, tooltipExternalLink)

	for _, src := range []string{
		`[x](&#x68;ttps://evil.example/p)`,
		`[x](&#104;ttps://evil.example/p)`,
		`[x](htt&#x70;s://evil.example/p)`,
		`[x](https&#x3a;//evil.example/p)`,
	} {
		t.Run(src, func(t *testing.T) {
			require.Equal(t, want, renderExtLinks(t, src))
		})
	}
}

// A link between two pages of the same non-realm section (the embedded
// documentation under /docs) is navigation inside one document tree, not a hop
// to another package, so it must not wear the "Cross package link" badge.
// Leaving that section for a realm still must.
func TestExtLinksSameSectionIsNotCrossPackage(t *testing.T) {
	orig := &weburl.GnoURL{Path: "/docs/builders/getting-started", Domain: "gno.land"}

	tests := []struct {
		target string
		want   GnoLinkType
	}{
		{"/docs/builders/local-setup", GnoLinkTypePackage},
		{"/docs/resources/gno-testing", GnoLinkTypePackage},
		{"/docs", GnoLinkTypePackage},
		// Leaving /docs for a realm is a real cross-package link.
		{"/r/gnoland/home", GnoLinkTypeInternal},
		{"/u/moul", GnoLinkTypeUser},
		{"https://github.com/gnolang/gno", GnoLinkTypeExternal},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			dest, err := url.Parse(tt.target)
			require.NoError(t, err)
			_, got := detectLinkType(dest, orig)
			require.Equal(t, tt.want, got, "target %q", tt.target)
		})
	}

	// The rule is scoped to the both-sides-unnamespaced case, so a realm
	// origin is unaffected: /r/a/x -> /r/b/y stays a cross-package link even
	// though both share the "r" first segment.
	realmOrig, err := weburl.Parse("https://gno.land/r/alice/blog")
	require.NoError(t, err)
	dest, err := url.Parse("/r/bob/blog")
	require.NoError(t, err)
	_, got := detectLinkType(dest, realmOrig)
	require.Equal(t, GnoLinkTypeInternal, got)
}

func renderExtLinks(t *testing.T, src string) string {
	t.Helper()
	gnourl, err := weburl.Parse("https://gno.land/r/test")
	require.NoError(t, err)
	m := goldmark.New()
	ExtLinks.Extend(m)
	ctx := parser.WithContext(NewGnoParserContext(GnoContext{GnoURL: gnourl}))
	var out bytes.Buffer
	require.NoError(t, m.Convert([]byte(src), &out, ctx))
	return out.String()
}
