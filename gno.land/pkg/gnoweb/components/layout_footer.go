package components

type FooterData struct {
	Analytics   bool
	AssetsPath  string
	BuildTime   string
	Sections    []FooterSection
	LegalNotice string
	LegalLinks  []FooterLink
}

type FooterLink struct {
	Label string
	URL   string
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
				{Label: "Docs", URL: "https://docs.gno.land/"},
				{Label: "Faucet", URL: "https://faucet.gno.land/"},
				{Label: "Blog", URL: "https://gno.land/r/gnoland/blog"},
				{Label: "Status", URL: "https://status.gnoteam.com/"},
			},
		},
		{
			Title: "Social media",
			Links: []FooterLink{
				{Label: "GitHub", URL: "https://github.com/gnolang/gno"},
				{Label: "Twitter", URL: "https://twitter.com/_gnoland"},
				{Label: "Discord", URL: "https://discord.gg/S8nKUqwkPn"},
				{Label: "YouTube", URL: "https://www.youtube.com/@_gnoland"},
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
		{Label: "Gno GPL License", URL: "https://github.com/gnolang/gno/blob/master/LICENSE.md"},
		{Label: "Gno.land Network Interaction Terms", URL: "https://github.com/gnolang/gno/blob/master/TERMS.md"},
		{Label: "Gno.land Contributor License Agreement", URL: "https://github.com/gnolang/gno/blob/master/CLA.md"},
	}

	return data
}
