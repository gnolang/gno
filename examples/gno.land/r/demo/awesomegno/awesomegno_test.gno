package awesomegno

import (
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/p/moul/authz"
)

func setupTest() {
	categories = avl.NewTree()
	dapps = avl.NewTree()
	proposals = avl.NewTree()
	nextDappID = 1
	nextCategoryID = 1

	auth = authz.New()
	adminAddress := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	auth.Current().(*authz.MemberAuthority).AddMember(adminAddress)

	testing.SetOriginCaller(adminAddress)
}

func TestAddCategory(t *testing.T) {
	setupTest()

	categoryID, err := AddCategory("Test", "Test category")
	urequire.NoError(t, err)
	uassert.True(t, categoryID > 0)

	category, err := GetCategory("Test")
	urequire.NoError(t, err)
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

	proposalID, err := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", "Tools")
	urequire.NoError(t, err)
	uassert.True(t, proposalID > 0)

	proposal, found := GetProposal(proposalID)
	uassert.True(t, found)
	uassert.Equal(t, "Test dApp", proposal.Dapp.Title)
	uassert.Equal(t, StatusPending, proposal.Status)

	_, err = ProposeNewDapp("Invalid Category", "Test", "https://example.com", "NonExistent")
	uassert.Error(t, err)

	_, err = ProposeNewDapp("", "Test", "https://example.com", "Tools")
	uassert.Error(t, err)
}

func TestApproveProposal(t *testing.T) {
	setupTest()

	_, err := AddCategory("Tools", "Developer tools")
	urequire.NoError(t, err)
	_, err = AddCategory("DeFi", "Decentralized finance")
	urequire.NoError(t, err)

	proposalID, err := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", "Tools,DeFi")
	urequire.NoError(t, err)

	testing.SetOriginCaller(std.Address("g1notadmin"))
	approvedID, err := ApproveProposal(proposalID)
	uassert.Error(t, err)
	uassert.Equal(t, uint64(0), approvedID)

	testing.SetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	approvedID, err = ApproveProposal(proposalID)
	urequire.NoError(t, err)

	proposal, found := GetProposal(proposalID)
	uassert.True(t, found, "Proposal should be found")
	uassert.Equal(t, StatusApproved, proposal.Status)

	dapp, found := GetDapp(approvedID)
	uassert.True(t, found, "Dapp should be found")
	uassert.Equal(t, "Test dApp", dapp.Title)
	uassert.Equal(t, true, dapp.Approved)

	uassert.Equal(t, 2, len(dapp.Categories))

	toolsCategory, err := GetCategory("Tools")
	urequire.NoError(t, err)
	uassert.True(t, containsUint64(toolsCategory.DappIDs, 2), "Tools category should contain dapp ID 2")

	defiCategory, err := GetCategory("DeFi")
	urequire.NoError(t, err)
	uassert.True(t, containsUint64(defiCategory.DappIDs, 2), "DeFi category should contain dapp ID 2")

	approvedID, err = ApproveProposal(999)
	uassert.Error(t, err)
	uassert.Equal(t, uint64(0), approvedID)

	approvedID, err = ApproveProposal(proposalID)
	uassert.Error(t, err)
	uassert.Equal(t, uint64(0), approvedID)
}

func TestRejectProposal(t *testing.T) {
	setupTest()

	_, err := AddCategory("Tools", "Developer tools")
	urequire.NoError(t, err)

	proposalID, err := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", "Tools")
	urequire.NoError(t, err)

	testing.SetOriginCaller(std.Address("g1notadmin"))
	err = RejectProposal(proposalID, "Not suitable")
	uassert.Error(t, err)

	testing.SetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
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

	proposalID, err := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", "Tools")
	urequire.NoError(t, err)
	approvedID, err := ApproveProposal(proposalID)
	urequire.NoError(t, err)

	dapp, found := GetDapp(approvedID)
	uassert.True(t, found)
	uassert.Equal(t, 1, len(dapp.Categories), "Dapp should have 1 category")
	uassert.True(t, containsString(dapp.Categories, "Tools"))

	toolsCategory, err := GetCategory("Tools")
	urequire.NoError(t, err)
	uassert.True(t, containsUint64(toolsCategory.DappIDs, approvedID), "Tools category should contain dapp ID"+strconv.FormatUint(approvedID, 10))

	testing.SetOriginCaller(std.Address("g1notadmin"))
	err = UpdateDappCategories(approvedID, "Tools,DeFi")
	uassert.Error(t, err)

	testing.SetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	err = UpdateDappCategories(approvedID, "DeFi,Games")
	urequire.NoError(t, err)

	dapp, found = GetDapp(approvedID)
	uassert.True(t, found, "Dapp should be found")
	uassert.Equal(t, 2, len(dapp.Categories), "Dapp should have 2 categories")
	uassert.True(t, containsString(dapp.Categories, "DeFi"), "Dapp should have DeFi category")
	uassert.True(t, containsString(dapp.Categories, "Games"), "Dapp should have Games category")
	uassert.False(t, containsString(dapp.Categories, "Tools"), "Dapp should not have Tools category")

	toolsCategory, err = GetCategory("Tools")
	urequire.NoError(t, err)
	uassert.False(t, containsUint64(toolsCategory.DappIDs, approvedID), "Tools category should not contain dapp ID"+strconv.FormatUint(approvedID, 10))

	defiCategory, err := GetCategory("DeFi")
	urequire.NoError(t, err)
	uassert.True(t, containsUint64(defiCategory.DappIDs, approvedID), "DeFi category should contain dapp ID"+strconv.FormatUint(approvedID, 10))

	gamesCategory, err := GetCategory("Games")
	urequire.NoError(t, err)
	uassert.True(t, containsUint64(gamesCategory.DappIDs, approvedID), "Games category should contain dapp ID"+strconv.FormatUint(approvedID, 10))

	err = UpdateDappCategories(approvedID, "NonExistent")
	uassert.Error(t, err)
}

func TestDeleteDapp(t *testing.T) {
	setupTest()

	_, err := AddCategory("Tools", "Developer tools")
	urequire.NoError(t, err)
	_, err = AddCategory("DeFi", "Decentralized finance")
	urequire.NoError(t, err)

	proposalID, err := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", "Tools,DeFi")
	urequire.NoError(t, err)
	approvedID, err := ApproveProposal(proposalID)
	urequire.NoError(t, err)

	testing.SetOriginCaller(std.Address("g1notadmin"))
	err = DeleteDapp(approvedID)
	uassert.Error(t, err)

	testing.SetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	err = DeleteDapp(approvedID)
	urequire.NoError(t, err)

	_, found1 := GetDapp(approvedID)
	uassert.False(t, found1)

	toolsCategory, err := GetCategory("Tools")
	urequire.NoError(t, err)
	uassert.False(t, containsUint64(toolsCategory.DappIDs, 1))

	defiCategory, err := GetCategory("DeFi")
	urequire.NoError(t, err)
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

	proposalID, err := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", "Tools,DeFi")
	urequire.NoError(t, err)
	approvedID, err := ApproveProposal(proposalID)
	urequire.NoError(t, err)

	testing.SetOriginCaller(std.Address("g1notadmin"))
	err = DeleteCategory("Tools")
	uassert.Error(t, err, "Should not be able to delete category with dapps")

	testing.SetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	err = DeleteCategory("Tools")
	urequire.NoError(t, err)

	_, err = GetCategory("Tools")
	uassert.Error(t, err, "Should not be able to get category that was deleted")

	dapp, found := GetDapp(approvedID)
	uassert.True(t, found)
	uassert.Equal(t, 1, len(dapp.Categories))
	uassert.False(t, containsString(dapp.Categories, "Tools"))
	uassert.True(t, containsString(dapp.Categories, "DeFi"))

	err = DeleteCategory("NonExistent")
	uassert.Error(t, err, "Should not be able to delete non-existent category")
}

func TestGetDappsByCategory(t *testing.T) {
	setupTest()

	_, err := AddCategory("Tools", "Developer tools")
	urequire.NoError(t, err)
	_, err = AddCategory("DeFi", "Decentralized finance")
	urequire.NoError(t, err)

	proposalID1, err := ProposeNewDapp("Tool dApp", "A tool dApp", "https://example.com/tool", "Tools")
	urequire.NoError(t, err)
	approvedID1, err := ApproveProposal(proposalID1)
	urequire.NoError(t, err)

	proposalID2, err := ProposeNewDapp("DeFi dApp", "A DeFi dApp", "https://example.com/defi", "DeFi")
	urequire.NoError(t, err)
	approvedID2, err := ApproveProposal(proposalID2)
	urequire.NoError(t, err)

	proposalID3, err := ProposeNewDapp("Hybrid dApp", "A hybrid dApp", "https://example.com/hybrid", "Tools,DeFi")
	urequire.NoError(t, err)
	approvedID3, err := ApproveProposal(proposalID3)
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

	var firstDappID uint64

	dappConfigs := []struct {
		title       string
		description string
		url         string
		categories  string
	}{
		{"Tool 1", "A tool", "https://example.com/tool1", "Tools"},
		{"Tool 2", "Another tool", "https://example.com/tool2", "Tools"},
		{"DeFi App", "A DeFi app", "https://example.com/defi", "DeFi"},
		{"Game", "A game", "https://example.com/game", "Games"},
		{"Social DeFi", "Social DeFi", "https://example.com/social-defi", "Social,DeFi"},
		{"Tooled Game", "Game with tools", "https://example.com/tooled-game", "Tools,Games"},
	}

	for i, config := range dappConfigs {
		proposalID, err := ProposeNewDapp(config.title, config.description, config.url, config.categories)
		urequire.NoError(t, err)
		dappID, err := ApproveProposal(proposalID)
		urequire.NoError(t, err)
		if i == 0 {
			firstDappID = dappID
		}
	}

	allDapps := ListDapps()
	for _, dapp := range allDapps {
		for _, categoryName := range dapp.Categories {
			category, err := GetCategory(categoryName)
			urequire.NoError(t, err)
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

	err := UpdateDappCategories(firstDappID, "Tools,Social")
	urequire.NoError(t, err)

	dapp, found := GetDapp(firstDappID)
	uassert.True(t, found)
	uassert.Equal(t, 2, len(dapp.Categories))
	uassert.True(t, containsString(dapp.Categories, "Tools"))
	uassert.True(t, containsString(dapp.Categories, "Social"))

	toolsCategory, err := GetCategory("Tools")
	urequire.NoError(t, err)
	uassert.True(t, containsUint64(toolsCategory.DappIDs, firstDappID))

	socialCategory, err := GetCategory("Social")
	urequire.NoError(t, err)
	uassert.True(t, containsUint64(socialCategory.DappIDs, firstDappID))

	err = DeleteCategory("Games")
	urequire.NoError(t, err)

	_, err = GetCategory("Games")
	uassert.Error(t, err, "Should not be able to get category that was deleted")

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
