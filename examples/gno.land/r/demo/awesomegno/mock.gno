package awesomegno

var (
	mockAddr1 = ""
	mockAddr2 = ""
	mockAddr3 = ""
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

	item1ID := AddItem(
		"GnoSwap",
		"Decentralized exchange for Gno.land tokens with automated market making",
		"https://gnoswap.io",
		mockAddr1,
	)
	UpdateItem(item1ID, "", "", "", defiID)

	item2ID := AddItem(
		"GnoArt",
		"NFT marketplace for digital art on Gno.land",
		"https://gnoart.io",
		mockAddr2,
	)
	UpdateItem(item2ID, "", "", "", nftID, marketID)

	item3ID := AddItem(
		"CryptoQuest",
		"RPG game built on Gno.land with player-owned assets",
		"https://cryptoquest.io",
		mockAddr3,
	)
	UpdateItem(item3ID, "", "", "", gamesID, nftID)

	item4ID := AddItem(
		"GnoChat",
		"Decentralized messaging platform with end-to-end encryption",
		"https://gnochat.io",
		mockAddr1,
	)
	UpdateItem(item4ID, "", "", "", socialID)

	item5ID := AddItem(
		"GnoDAO",
		"DAO creation and management platform",
		"https://gnodao.io",
		mockAddr2,
	)
	UpdateItem(item5ID, "", "", "", govID)

	item6ID := AddItem(
		"GnoTools",
		"Developer toolkit for building on Gno.land",
		"https://gnotools.io",
		mockAddr3,
	)
	UpdateItem(item6ID, "", "", "", utilityID, infraID)

	item7ID := AddItem(
		"GnoMetrics",
		"Analytics dashboard for Gno.land blockchain",
		"https://gnometrics.io",
		mockAddr1,
	)
	UpdateItem(item7ID, "", "", "", dataID)

	item8ID := AddItem(
		"GnoID",
		"Decentralized identity verification system",
		"https://gnoid.io",
		mockAddr2,
	)
	UpdateItem(item8ID, "", "", "", identityID)

	item9ID := AddItem(
		"GnoMarket",
		"P2P marketplace for digital and physical goods",
		"https://gnomarket.io",
		mockAddr3,
	)
	UpdateItem(item9ID, "", "", "", marketID)

	item10ID := AddItem(
		"GnoStack",
		"Full-stack development framework for Gno.land",
		"https://gnostack.io",
		mockAddr1,
	)
	UpdateItem(item10ID, "", "", "", infraID, utilityID)
}
