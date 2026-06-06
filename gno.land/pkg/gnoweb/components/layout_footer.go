package components

type FooterData struct {
	Analytics   AnalyticsData
	Sections    []FooterSection
	LegalNotice string
	LegalLinks  []FooterLink
}

type FooterLink struct {
	Label string
	URL   string
	// Outbound, when set to one of the Outbound* constants, is rendered as
	// data-outbound on the link so SimpleAnalytics fires a named
	// outbound_<label> event instead of an anonymous outbound click.
	Outbound string
}

type FooterSection struct {
	Title string
	Links []FooterLink
}

func EnrichFooterData(data FooterData) FooterData {
	data.Sections = []FooterSection{
		{
			Title: "Footer navigation",
			Links: []FooterLink{
				{Label: "About", URL: "/about"},
				{Label: "Docs", URL: "https://docs.gno.land/", Outbound: OutboundDocs},
				{Label: "Faucet", URL: "https://faucet.gno.land/", Outbound: OutboundFaucet},
				{Label: "Blog", URL: "https://gno.land/r/gnoland/blog"},
				{Label: "Status", URL: "https://status.gnoteam.com/", Outbound: OutboundStatus},
			},
		},
		{
			Title: "Social media",
			Links: []FooterLink{
				{Label: "GitHub", URL: "https://github.com/gnolang/gno", Outbound: OutboundGitHub},
				{Label: "Twitter", URL: "https://twitter.com/_gnoland", Outbound: OutboundTwitter},
				{Label: "Discord", URL: "https://discord.gg/S8nKUqwkPn", Outbound: OutboundDiscord},
				{Label: "YouTube", URL: "https://www.youtube.com/@_gnoland", Outbound: OutboundYouTube},
			},
		},
	}

	data.LegalNotice = "\u00a9 2026 NewTendermint, LLC. NewTendermint provides software and user " +
		"interfaces for interacting with the Gno.land blockchain network. The Network is " +
		"decentralized and not controlled by NewTendermint. By using this site or any official " +
		"interface, you agree to the Network Interaction Terms (and any linked policies). Code " +
		"and other copyrightable works published through Gno.land are made available under the " +
		"Network License (GNO Network General Public License v6 or later), including the Strong " +
		"Attribution additional terms, which require an Attribution Notice with an active " +
		"hyperlink to the Attribution URL designated under the Attribution Policy maintained by " +
		"NewTendermint for the Network. Do not submit code, content, or data unless you have " +
		"sufficient rights to do so. On-chain activity is public and may be copied, indexed, " +
		"and displayed by others. THE SITE AND INTERFACES ARE PROVIDED \u201cAS IS\u201d " +
		"WITHOUT WARRANTY."

	data.LegalLinks = []FooterLink{
		{Label: "Gno GPL License", URL: "https://github.com/gnolang/gno/blob/master/LICENSE.md", Outbound: OutboundGitHub},
		{Label: "Gno.land Network Interaction Terms", URL: "https://github.com/gnolang/gno/blob/master/TERMS.md", Outbound: OutboundGitHub},
		{Label: "Gno.land Contributor License Agreement", URL: "https://github.com/gnolang/gno/blob/master/CLA.md", Outbound: OutboundGitHub},
	}

	return data
}
