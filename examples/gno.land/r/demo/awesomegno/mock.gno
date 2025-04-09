package awesomegno

var (
	mockAddr1 = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
	mockAddr2 = "g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn"
	mockAddr3 = "g1w4ek2u3jta047h6lta047h6lta047h6l9huexc"
)

func AddMockData() {
	defiID := AddCategory("DeFi", "Decentralized Finance applications and protocols")
	nftID := AddCategory("NFT", "Non-Fungible Token marketplaces and collections")
	gamesID := AddCategory("Games", "Blockchain games and gaming platforms")
	socialID := AddCategory("Social", "Social networks and communication platforms")
	govID := AddCategory("Governance", "DAOs and governance applications")
	utilityID := AddCategory("Utility", "Utility tools and services")
	dataID := AddCategory("Data & Analytics", "Blockchain data and analytics platforms")
	identityID := AddCategory("Identity", "Identity and authentication solutions")
	marketID := AddCategory("Marketplaces", "Decentralized marketplaces")
	infraID := AddCategory("Infrastructure", "Blockchain infrastructure and developer tools")

	dapp1ID := AddDapp(
		"GnoSwap",
		"Decentralized exchange for Gno.land tokens with automated market making",
		"https://gnoswap.io",
		mockAddr1,
	)
	UpdateDapp(dapp1ID, "", "", "", defiID)

	dapp2ID := AddDapp(
		"GnoArt",
		"NFT marketplace for digital art on Gno.land",
		"https://gnoart.io",
		mockAddr2,
	)
	UpdateDapp(dapp2ID, "", "", "", nftID, marketID)

	dapp3ID := AddDapp(
		"CryptoQuest",
		"RPG game built on Gno.land with player-owned assets",
		"https://cryptoquest.io",
		mockAddr3,
	)
	UpdateDapp(dapp3ID, "", "", "", gamesID, nftID)

	dapp4ID := AddDapp(
		"GnoChat",
		"Decentralized messaging platform with end-to-end encryption",
		"https://gnochat.io",
		mockAddr1,
	)
	UpdateDapp(dapp4ID, "", "", "", socialID)

	dapp5ID := AddDapp(
		"GnoDAO",
		"DAO creation and management platform",
		"https://gnodao.io",
		mockAddr2,
	)
	UpdateDapp(dapp5ID, "", "", "", govID)

	dapp6ID := AddDapp(
		"GnoTools",
		"Developer toolkit for building on Gno.land",
		"https://gnotools.io",
		mockAddr3,
	)
	UpdateDapp(dapp6ID, "", "", "", utilityID, infraID)

	dapp7ID := AddDapp(
		"GnoMetrics",
		"Analytics dashboard for Gno.land blockchain",
		"https://gnometrics.io",
		mockAddr1,
	)
	UpdateDapp(dapp7ID, "", "", "", dataID)

	dapp8ID := AddDapp(
		"GnoID",
		"Decentralized identity verification system",
		"https://gnoid.io",
		mockAddr2,
	)
	UpdateDapp(dapp8ID, "", "", "", identityID)

	dapp9ID := AddDapp(
		"GnoMarket",
		"P2P marketplace for digital and physical goods",
		"https://gnomarket.io",
		mockAddr3,
	)
	UpdateDapp(dapp9ID, "", "", "", marketID)

	dapp10ID := AddDapp(
		"GnoStack",
		"Full-stack development framework for Gno.land",
		"https://gnostack.io",
		mockAddr1,
	)
	UpdateDapp(dapp10ID, "", "", "", infraID, utilityID)
}
