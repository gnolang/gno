package awesomegno

func MockData() {
	categories := []struct {
		name        string
		description string
	}{
		{"Tools", "Developer tools and utilities for Gno"},
		{"DeFi", "Decentralized finance applications"},
		{"Games", "Games and entertainment applications"},
		{"Social", "Social networks and communication tools"},
		{"NFTs", "Non-fungible tokens and collectibles"},
		{"Infrastructure", "Core infrastructure and protocols"},
		{"Education", "Educational resources and tutorials"},
		{"Governance", "Governance and voting systems"},
		{"Marketplaces", "Decentralized marketplaces and exchanges"},
		{"Utilities", "Utility applications and services"},
	}

	for _, cat := range categories {
		AddCategory(cat.name, cat.description)
	}

	dapps := []struct {
		title       string
		description string
		url         string
		categories  []string
	}{
		{
			"GnoScan",
			"Block explorer for the Gno.land blockchain",
			"https://gno.land/r/demo/gnoscan",
			[]string{"Tools", "Infrastructure"},
		},
		{
			"GnoSwap",
			"Decentralized exchange for Gno tokens",
			"https://gno.land/r/demo/gnoswap",
			[]string{"DeFi", "Marketplaces"},
		},
		{
			"GnoVote",
			"On-chain voting and governance system",
			"https://gno.land/r/demo/gnovote",
			[]string{"Governance", "Infrastructure"},
		},
		{
			"CryptoGnomes",
			"Collectible NFT game on Gno.land",
			"https://gno.land/r/demo/cryptognomes",
			[]string{"Games", "NFTs"},
		},
		{
			"GnoLearn",
			"Interactive tutorials for learning Gno",
			"https://gno.land/r/demo/gnolearn",
			[]string{"Education", "Tools"},
		},
		{
			"GnoChat",
			"Decentralized messaging platform",
			"https://gno.land/r/demo/gnochat",
			[]string{"Social", "Utilities"},
		},
		{
			"GnoLend",
			"Lending and borrowing protocol",
			"https://gno.land/r/demo/gnolend",
			[]string{"DeFi"},
		},
		{
			"GnoDAO",
			"DAO creation and management toolkit",
			"https://gno.land/r/demo/gnodao",
			[]string{"Governance", "Infrastructure"},
		},
		{
			"GnoMarket",
			"P2P marketplace for digital goods",
			"https://gno.land/r/demo/gnomarket",
			[]string{"Marketplaces", "Utilities"},
		},
		{
			"GnoStats",
			"Analytics dashboard for Gno.land metrics",
			"https://gno.land/r/demo/gnostats",
			[]string{"Tools", "Infrastructure"},
		},
	}

	for i, dapp := range dapps {
		proposalID, _ := ProposeNewDapp(dapp.title, dapp.description, dapp.url, dapp.categories)

		if i%2 == 0 {
			ApproveProposal(proposalID)
		}
	}

	pendingProposals := []struct {
		title       string
		description string
		url         string
		categories  []string
	}{
		{
			"GnoWallet",
			"Secure wallet for managing Gno assets",
			"https://gno.land/r/demo/gnowallet",
			[]string{"Tools", "Utilities"},
		},
		{
			"GnoForum",
			"Decentralized discussion forum",
			"https://gno.land/r/demo/gnoforum",
			[]string{"Social"},
		},
		{
			"GnoInsurance",
			"Decentralized insurance protocol",
			"https://gno.land/r/demo/gnoinsurance",
			[]string{"DeFi"},
		},
		{
			"GnoQuest",
			"RPG adventure game on Gno.land",
			"https://gno.land/r/demo/gnoquest",
			[]string{"Games"},
		},
		{
			"GnoIdentity",
			"Self-sovereign identity solution",
			"https://gno.land/r/demo/gnoidentity",
			[]string{"Infrastructure", "Utilities"},
		},
	}

	for _, proposal := range pendingProposals {
		ProposeNewDapp(proposal.title, proposal.description, proposal.url, proposal.categories)
	}

	RejectProposal(8, "URL is not valid or project is not ready for inclusion")
}
