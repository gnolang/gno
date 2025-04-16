package awesomegno

import (
	"gno.land/p/demo/seqid"
)

var (
	mockAddr1 = ""
	mockAddr2 = ""
	mockAddr3 = ""
)

func AddMockData() {
	AddCategory("DeFi", "Decentralized Finance applications and protocols")
	AddCategory("NFT", "Non-Fungible Token marketplaces and collections")
	AddCategory("Games", "Blockchain games and gaming platforms")
	AddCategory("Social", "Social networks and communication platforms")
	AddCategory("Governance", "DAOs and governance applications")
	AddCategory("Utility", "Utility tools and services")
	AddCategory("Data & Analytics", "Blockchain data and analytics platforms")
	AddCategory("Identity", "Identity and authentication solutions")
	AddCategory("Marketplaces", "Decentralized marketplaces")
	AddCategory("Infrastructure", "Blockchain infrastructure and developer tools")

	AddItem(
		"GnoSwap",
		"Decentralized exchange for Gno.land tokens with automated market making",
		"https://gnoswap.io",
		mockAddr1,
	)
	updateFirstItemCategories("DeFi")

	AddItem(
		"GnoArt",
		"NFT marketplace for digital art on Gno.land",
		"https://gnoart.io",
		mockAddr2,
	)
	updateFirstItemCategories("NFT", "Marketplaces")

	AddItem(
		"CryptoQuest",
		"RPG game built on Gno.land with player-owned assets",
		"https://cryptoquest.io",
		mockAddr3,
	)
	updateFirstItemCategories("Games", "NFT")

	AddItem(
		"GnoChat",
		"Decentralized messaging platform with end-to-end encryption",
		"https://gnochat.io",
		mockAddr1,
	)
	updateFirstItemCategories("Social")

	AddItem(
		"GnoDAO",
		"DAO creation and management platform",
		"https://gnodao.io",
		mockAddr2,
	)
	updateFirstItemCategories("Governance")

	AddItem(
		"GnoTools",
		"Developer toolkit for building on Gno.land",
		"https://gnotools.io",
		mockAddr3,
	)
	updateFirstItemCategories("Utility", "Infrastructure")

	AddItem(
		"GnoMetrics",
		"Analytics dashboard for Gno.land blockchain",
		"https://gnometrics.io",
		mockAddr1,
	)
	updateFirstItemCategories("Data & Analytics")

	AddItem(
		"GnoID",
		"Decentralized identity verification system",
		"https://gnoid.io",
		mockAddr2,
	)
	updateFirstItemCategories("Identity")

	AddItem(
		"GnoMarket",
		"P2P marketplace for digital and physical goods",
		"https://gnomarket.io",
		mockAddr3,
	)
	updateFirstItemCategories("Marketplaces")

	AddItem(
		"GnoStack",
		"Full-stack development framework for Gno.land",
		"https://gnostack.io",
		mockAddr1,
	)
	updateFirstItemCategories("Infrastructure", "Utility")
}

func updateFirstItemCategories(categoryNames ...string) {
	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)

	var categoryIDs []seqid.ID
	for _, name := range categoryNames {
		categories.Iterate("", "", func(key string, value interface{}) bool {
			category := value.(*Category)
			if category.Name == name {
				categoryIDs = append(categoryIDs, category.ID)
				return true
			}
			return false
		})
	}

	for _, categoryID := range categoryIDs {
		UpdateItem(item.ID, "", "", "", categoryID)
	}
}
