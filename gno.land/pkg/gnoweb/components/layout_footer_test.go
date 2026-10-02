package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnrichFooterData_Outbound(t *testing.T) {
	data := EnrichFooterData(FooterData{HasFaucet: true, StatusURL: "https://status.gno.land/"})

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
		"https://docs.gno.land/":                                OutboundDocs,
		"https://faucet.gno.land/":                              OutboundFaucet,
		"https://status.gno.land/":                              OutboundStatus,
		"https://github.com/gnolang/gno":                        OutboundGitHub,
		"https://twitter.com/_gnoland":                          OutboundTwitter,
		"https://discord.com/invite/gnoland":                    OutboundDiscord,
		"https://bubblerumble.net/":                             OutboundBubbleRumble,
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
	assert.Equal(t, OutboundDocs, got["https://docs.gno.land/"])
	assert.Equal(t, OutboundGitHub, got["https://github.com/gnolang"])
	// Same-domain links must not carry data-outbound; SimpleAnalytics already
	// counts them as page views and an outbound tag would double-count them.
	assert.Equal(t, "", got["/about"])
	assert.NotContains(t, got, "https://gno.land/about")
}

// The link used to be unconditional, so a deployment with no faucet (gnodev,
// a local chain) still advertised the hub.
func TestEnrichFooterData_FaucetIsConditional(t *testing.T) {
	labels := func(data FooterData) []string {
		var out []string
		for _, sec := range data.Sections {
			for _, l := range sec.Links {
				out = append(out, l.Label)
			}
		}
		return out
	}

	assert.NotContains(t, labels(EnrichFooterData(FooterData{})), "Faucet",
		"no faucet configured must render no Faucet link")

	withFaucet := EnrichFooterData(FooterData{HasFaucet: true})
	assert.Contains(t, labels(withFaucet), "Faucet")

	var url string
	for _, sec := range withFaucet.Sections {
		for _, l := range sec.Links {
			if l.Label == "Faucet" {
				url = l.URL
			}
		}
	}
	// The hub, not -faucet-url: staging points that flag at a POST-only API.
	assert.Equal(t, faucetHubURL, url)
}

// Each network has its own status page, set per deployment; without one the
// footer shows no Status link rather than another network's.
func TestEnrichFooterData_StatusIsConditional(t *testing.T) {
	statusURL := func(data FooterData) (string, bool) {
		for _, sec := range data.Sections {
			for _, l := range sec.Links {
				if l.Label == "Status" {
					return l.URL, true
				}
			}
		}
		return "", false
	}

	_, found := statusURL(EnrichFooterData(FooterData{}))
	assert.False(t, found, "no status url configured must render no Status link")

	got, found := statusURL(EnrichFooterData(FooterData{StatusURL: "https://status.onyx.testnets.gno.land"}))
	assert.True(t, found)
	assert.Equal(t, "https://status.onyx.testnets.gno.land", got)
}
