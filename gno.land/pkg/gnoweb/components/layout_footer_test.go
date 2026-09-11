package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrichFooterData_Outbound(t *testing.T) {
	data := EnrichFooterData(FooterData{})

	// Flatten Sections and LegalLinks into a single URL→Outbound map; URLs
	// across the two groups are distinct, so a last-write collision would
	// itself indicate a regression.
	got := map[string]string{}
	for _, sec := range data.Sections {
		for _, l := range sec.Links {
			got[l.URL] = l.Outbound
		}
	}
	for _, l := range data.LegalLinks {
		got[l.URL] = l.Outbound
	}

	want := map[string]string{
		// /docs is served by DocsHandler from the embedded docs/ tree, so it is
		// a first-party path and must NOT carry data-outbound: SimpleAnalytics
		// already counts it as a page view and an outbound tag double-counts it.
		"/docs":                                                 "",
		"https://faucet.gno.land/":                              OutboundFaucet,
		"https://status.gnoteam.com/":                           OutboundStatus,
		"https://github.com/gnolang/gno":                        OutboundGitHub,
		"https://twitter.com/_gnoland":                          OutboundTwitter,
		"https://discord.gg/S8nKUqwkPn":                         OutboundDiscord,
		"https://www.youtube.com/@_gnoland":                     OutboundYouTube,
		"https://github.com/gnolang/gno/blob/master/LICENSE.md": OutboundGitHub,
		"https://github.com/gnolang/gno/blob/master/TERMS.md":   OutboundGitHub,
		"https://github.com/gnolang/gno/blob/master/CLA.md":     OutboundGitHub,
	}
	for url, outbound := range want {
		assert.Equal(t, outbound, got[url], "URL %q must carry outbound %q", url, outbound)
	}
}

func TestStaticHeaderGeneralLinks_Outbound(t *testing.T) {
	links := StaticHeaderGeneralLinks()
	got := map[string]string{}
	for _, l := range links {
		got[l.URL] = l.Outbound
	}
	assert.Equal(t, OutboundGitHub, got["https://github.com/gnolang"])
	// Same-domain links must not carry data-outbound; SimpleAnalytics already
	// counts them as page views and an outbound tag would double-count them.
	assert.Equal(t, "", got["https://gno.land/about"])
	// The Docs entry must point at the locally served /docs — this is the only
	// link into DocsHandler, so pointing it back at docs.gno.land leaves the
	// embedded pages unreachable from the UI.
	require.Contains(t, got, "/docs")
	assert.Equal(t, "", got["/docs"])
}
