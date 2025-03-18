package awesomegno

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func setupTest() {
	admins = avl.NewTree()
	categories = avl.NewTree()
	dapps = avl.NewTree()
	proposals = avl.NewTree()
	nextDappID = 1
	nextCategoryID = 1

	testAdmin := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	admins.Set(testAdmin.String(), true)

	std.TestSetOriginCaller(testAdmin)
}

func TestAddCategory(t *testing.T) {
	setupTest()

	categoryID, err := AddCategory("Test", "Test category")
	urequire.NoError(t, err)
	uassert.True(t, categoryID > 0)

	category, found := GetCategory("Test")
	uassert.True(t, found)
	uassert.Equal(t, "Test", category.Name)
	uassert.Equal(t, "Test category", category.Description)

	_, err = AddCategory("Test", "Another description")
	uassert.Error(t, err)

	_, err = AddCategory("", "Empty name")
	uassert.Error(t, err)
}

func TestProposeNewDapp(t *testing.T) {
	setupTest()

	_, err := AddCategory("Tools", "Developer tools")
	urequire.NoError(t, err)

	proposalID, err := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", []string{"Tools"})
	urequire.NoError(t, err)
	uassert.True(t, proposalID > 0)

	proposal, found := GetProposal(proposalID)
	uassert.True(t, found)
	uassert.Equal(t, "Test dApp", proposal.Dapp.Title)
	uassert.Equal(t, StatusPending, proposal.Status)

	_, err = ProposeNewDapp("Invalid Category", "Test", "https://example.com", []string{"NonExistent"})
	uassert.Error(t, err)

	_, err = ProposeNewDapp("", "Test", "https://example.com", []string{"Tools"})
	uassert.Error(t, err)
}

func TestApproveProposal(t *testing.T) {
	setupTest()

	_, err := AddCategory("Tools", "Developer tools")
	urequire.NoError(t, err)
	_, err = AddCategory("DeFi", "Decentralized finance")
	urequire.NoError(t, err)

	proposalID, err := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", []string{"Tools", "DeFi"})
	urequire.NoError(t, err)

	std.TestSetOriginCaller(std.Address("g1notadmin"))
	err = ApproveProposal(proposalID)
	uassert.Error(t, err)

	std.TestSetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	err = ApproveProposal(proposalID)
	urequire.NoError(t, err)

	proposal, found := GetProposal(proposalID)
	uassert.True(t, found)
	uassert.Equal(t, StatusApproved, proposal.Status)

	dapp, found := GetDapp(1)
	uassert.True(t, found)
	uassert.Equal(t, "Test dApp", dapp.Title)
	uassert.Equal(t, true, dapp.Approved)

	uassert.Equal(t, 2, len(dapp.Categories))

	toolsCategory, found := GetCategory("Tools")
	uassert.True(t, found)
	uassert.True(t, containsUint64(toolsCategory.DappIDs, 1))

	defiCategory, found := GetCategory("DeFi")
	uassert.True(t, found)
	uassert.True(t, containsUint64(defiCategory.DappIDs, 1))

	err = ApproveProposal(999)
	uassert.Error(t, err)

	err = ApproveProposal(proposalID)
	uassert.Error(t, err)
}

func TestRejectProposal(t *testing.T) {
	setupTest()

	_, err := AddCategory("Tools", "Developer tools")
	urequire.NoError(t, err)

	proposalID, err := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", []string{"Tools"})
	urequire.NoError(t, err)

	std.TestSetOriginCaller(std.Address("g1notadmin"))
	err = RejectProposal(proposalID, "Not suitable")
	uassert.Error(t, err)

	std.TestSetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	err = RejectProposal(proposalID, "Not suitable")
	urequire.NoError(t, err)

	proposal, found := GetProposal(proposalID)
	uassert.True(t, found)
	uassert.Equal(t, StatusRejected, proposal.Status)
	uassert.Equal(t, "Not suitable", proposal.Reason)

	_, found = GetDapp(1)
	uassert.False(t, found)

	err = RejectProposal(999, "Does not exist")
	uassert.Error(t, err)

	err = RejectProposal(proposalID, "Already rejected")
	uassert.Error(t, err)
}

func TestUpdateDappCategories(t *testing.T) {
	setupTest()

	_, err := AddCategory("Tools", "Developer tools")
	urequire.NoError(t, err)
	_, err = AddCategory("DeFi", "Decentralized finance")
	urequire.NoError(t, err)
	_, err = AddCategory("Games", "Gaming dApps")
	urequire.NoError(t, err)

	proposalID, err := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", []string{"Tools"})
	urequire.NoError(t, err)
	err = ApproveProposal(proposalID)
	urequire.NoError(t, err)

	dapp, found := GetDapp(1)
	uassert.True(t, found)
	uassert.Equal(t, 1, len(dapp.Categories))
	uassert.True(t, containsString(dapp.Categories, "Tools"))

	toolsCategory, found := GetCategory("Tools")
	uassert.True(t, found)
	uassert.True(t, containsUint64(toolsCategory.DappIDs, 1))

	std.TestSetOriginCaller(std.Address("g1notadmin"))
	err = UpdateDappCategories(1, []string{"Tools", "DeFi"})
	uassert.Error(t, err)

	std.TestSetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	err = UpdateDappCategories(1, []string{"DeFi", "Games"})
	urequire.NoError(t, err)

	dapp, found = GetDapp(1)
	uassert.True(t, found)
	uassert.Equal(t, 2, len(dapp.Categories))
	uassert.True(t, containsString(dapp.Categories, "DeFi"))
	uassert.True(t, containsString(dapp.Categories, "Games"))
	uassert.False(t, containsString(dapp.Categories, "Tools"))

	toolsCategory, found = GetCategory("Tools")
	uassert.True(t, found)
	uassert.False(t, containsUint64(toolsCategory.DappIDs, 1))

	defiCategory, found := GetCategory("DeFi")
	uassert.True(t, found)
	uassert.True(t, containsUint64(defiCategory.DappIDs, 1))

	gamesCategory, found := GetCategory("Games")
	uassert.True(t, found)
	uassert.True(t, containsUint64(gamesCategory.DappIDs, 1))

	err = UpdateDappCategories(1, []string{"NonExistent"})
	uassert.Error(t, err)

	err = UpdateDappCategories(999, []string{"Tools"})
	uassert.Error(t, err)
}

func TestDeleteDapp(t *testing.T) {
	setupTest()

	_, err := AddCategory("Tools", "Developer tools")
	urequire.NoError(t, err)
	_, err = AddCategory("DeFi", "Decentralized finance")
	urequire.NoError(t, err)

	proposalID, err := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", []string{"Tools", "DeFi"})
	urequire.NoError(t, err)
	err = ApproveProposal(proposalID)
	urequire.NoError(t, err)

	std.TestSetOriginCaller(std.Address("g1notadmin"))
	err = DeleteDapp(1)
	uassert.Error(t, err)

	std.TestSetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	err = DeleteDapp(1)
	urequire.NoError(t, err)

	_, found := GetDapp(1)
	uassert.False(t, found)

	toolsCategory, found := GetCategory("Tools")
	uassert.True(t, found)
	uassert.False(t, containsUint64(toolsCategory.DappIDs, 1))

	defiCategory, found := GetCategory("DeFi")
	uassert.True(t, found)
	uassert.False(t, containsUint64(defiCategory.DappIDs, 1))

	err = DeleteDapp(999)
	uassert.Error(t, err)
}

func TestDeleteCategory(t *testing.T) {
	setupTest()

	_, err := AddCategory("Tools", "Developer tools")
	urequire.NoError(t, err)
	_, err = AddCategory("DeFi", "Decentralized finance")
	urequire.NoError(t, err)

	proposalID, err := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", []string{"Tools", "DeFi"})
	urequire.NoError(t, err)
	err = ApproveProposal(proposalID)
	urequire.NoError(t, err)

	std.TestSetOriginCaller(std.Address("g1notadmin"))
	err = DeleteCategory("Tools")
	uassert.Error(t, err)

	std.TestSetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	err = DeleteCategory("Tools")
	urequire.NoError(t, err)

	_, found := GetCategory("Tools")
	uassert.False(t, found)

	dapp, found := GetDapp(1)
	uassert.True(t, found)
	uassert.Equal(t, 1, len(dapp.Categories))
	uassert.False(t, containsString(dapp.Categories, "Tools"))
	uassert.True(t, containsString(dapp.Categories, "DeFi"))

	err = DeleteCategory("NonExistent")
	uassert.Error(t, err)
}

func TestGetDappsByCategory(t *testing.T) {
	setupTest()

	_, err := AddCategory("Tools", "Developer tools")
	urequire.NoError(t, err)
	_, err = AddCategory("DeFi", "Decentralized finance")
	urequire.NoError(t, err)

	proposalID1, err := ProposeNewDapp("Tool dApp", "A tool dApp", "https://example.com/tool", []string{"Tools"})
	urequire.NoError(t, err)
	err = ApproveProposal(proposalID1)
	urequire.NoError(t, err)

	proposalID2, err := ProposeNewDapp("DeFi dApp", "A DeFi dApp", "https://example.com/defi", []string{"DeFi"})
	urequire.NoError(t, err)
	err = ApproveProposal(proposalID2)
	urequire.NoError(t, err)

	proposalID3, err := ProposeNewDapp("Hybrid dApp", "A hybrid dApp", "https://example.com/hybrid", []string{"Tools", "DeFi"})
	urequire.NoError(t, err)
	err = ApproveProposal(proposalID3)
	urequire.NoError(t, err)

	toolsDapps := GetDappsByCategory("Tools")
	uassert.Equal(t, 2, len(toolsDapps))

	defiDapps := GetDappsByCategory("DeFi")
	uassert.Equal(t, 2, len(defiDapps))

	nonExistentDapps := GetDappsByCategory("NonExistent")
	uassert.Equal(t, 0, len(nonExistentDapps))
}

func TestBidirectionalConsistency(t *testing.T) {
	setupTest()

	for _, cat := range []string{"Tools", "DeFi", "Games", "Social"} {
		_, err := AddCategory(cat, "Description for "+cat)
		urequire.NoError(t, err)
	}

	dappConfigs := []struct {
		title       string
		description string
		url         string
		categories  []string
	}{
		{"Tool 1", "A tool", "https://example.com/tool1", []string{"Tools"}},
		{"Tool 2", "Another tool", "https://example.com/tool2", []string{"Tools"}},
		{"DeFi App", "A DeFi app", "https://example.com/defi", []string{"DeFi"}},
		{"Game", "A game", "https://example.com/game", []string{"Games"}},
		{"Social DeFi", "Social DeFi", "https://example.com/social-defi", []string{"Social", "DeFi"}},
		{"Tooled Game", "Game with tools", "https://example.com/tooled-game", []string{"Tools", "Games"}},
	}

	for _, config := range dappConfigs {
		proposalID, err := ProposeNewDapp(config.title, config.description, config.url, config.categories)
		urequire.NoError(t, err)
		err = ApproveProposal(proposalID)
		urequire.NoError(t, err)
	}

	allDapps := ListDapps()
	for _, dapp := range allDapps {
		for _, categoryName := range dapp.Categories {
			category, found := GetCategory(categoryName)
			uassert.True(t, found)
			uassert.True(t, containsUint64(category.DappIDs, dapp.ID))
		}
	}

	allCategories := ListCategories()
	for _, category := range allCategories {
		for _, dappID := range category.DappIDs {
			dapp, found := GetDapp(dappID)
			uassert.True(t, found)
			uassert.True(t, containsString(dapp.Categories, category.Name))
		}
	}

	err := UpdateDappCategories(1, []string{"Tools", "Social"})
	urequire.NoError(t, err)

	dapp, found := GetDapp(1)
	uassert.True(t, found)
	uassert.Equal(t, 2, len(dapp.Categories))
	uassert.True(t, containsString(dapp.Categories, "Tools"))
	uassert.True(t, containsString(dapp.Categories, "Social"))

	toolsCategory, found := GetCategory("Tools")
	uassert.True(t, found)
	uassert.True(t, containsUint64(toolsCategory.DappIDs, 1))

	socialCategory, found := GetCategory("Social")
	uassert.True(t, found)
	uassert.True(t, containsUint64(socialCategory.DappIDs, 1))

	err = DeleteCategory("Games")
	urequire.NoError(t, err)

	_, found = GetCategory("Games")
	uassert.False(t, found)

	allDapps = ListDapps()
	for _, dapp := range allDapps {
		uassert.False(t, containsString(dapp.Categories, "Games"))
	}
}

func containsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func containsUint64(slice []uint64, num uint64) bool {
	for _, item := range slice {
		if item == num {
			return true
		}
	}
	return false
}
