package awesomegno

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	"gno.land/p/moul/authz"
)

func setupTest() {
	categories = avl.NewTree()
	dapps = avl.NewTree()
	proposals = avl.NewTree()
	nextDappID = seqid.ID(0)
	nextCategoryID = seqid.ID(0)
	nextProposalID = seqid.ID(0)

	adminAddress := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	Auth = authz.NewMemberAuthority(
		adminAddress,
		std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5"), 
	)
	testing.SetOriginCaller(adminAddress)
}

func TestAddCategory(t *testing.T) {
	setupTest()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	categoryID := AddCategory("Test", "Test category")
	uassert.True(t, categoryID > 0)

	category := GetCategory("Test")
	uassert.Equal(t, "Test", category.Name)
	uassert.Equal(t, "Test category", category.Description)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for duplicate category")
		}
	}()
	AddCategory("Test", "Another description")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for empty category name")
		}
	}()
	AddCategory("", "Empty name")
}

func TestProposeNewDapp(t *testing.T) {
	setupTest()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("unexpected panic: %v", r)
		}
	}()

	AddCategory("Tools", "Developer tools")

	proposalID := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", "Tools")
	uassert.True(t, proposalID > 0)

	proposal, found := GetProposal(proposalID)
	uassert.True(t, found)
	uassert.Equal(t, "Test dApp", proposal.Dapp.Title)
	uassert.Equal(t, StatusPending, proposal.Status)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for invalid category")
		}
	}()
	ProposeNewDapp("Invalid Category", "Test", "https://example.com", "NonExistent")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for empty title")
		}
	}()
	ProposeNewDapp("", "Test", "https://example.com", "Tools")
}

func TestApproveProposal(t *testing.T) {
	setupTest()

	AddCategory("Tools", "Developer tools")
	AddCategory("DeFi", "Decentralized finance")

	proposalID := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", "Tools,DeFi")

	testing.SetOriginCaller(std.Address("g1notadmin"))
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for unauthorized approval")
		}
	}()
	ApproveProposal(proposalID)

	testing.SetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	approvedID := ApproveProposal(proposalID)

	proposal, found := GetProposal(proposalID)
	uassert.True(t, found, "Proposal should be found")
	uassert.Equal(t, StatusApproved, proposal.Status)

	dapp, found := GetDapp(approvedID)
	uassert.True(t, found, "Dapp should be found")
	uassert.Equal(t, "Test dApp", dapp.Title)
	uassert.Equal(t, true, dapp.Approved)

	uassert.Equal(t, 2, len(dapp.Categories))

	toolsCategory := GetCategory("Tools")
	uassert.True(t, containsID(toolsCategory.DappIDs, approvedID), "Tools category should contain dapp ID")

	defiCategory := GetCategory("DeFi")
	uassert.True(t, containsID(defiCategory.DappIDs, approvedID), "DeFi category should contain dapp ID")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for invalid proposal ID")
		}
	}()
	ApproveProposal(seqid.ID(999))

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for already approved proposal")
		}
	}()
	ApproveProposal(proposalID)
}

func TestRejectProposal(t *testing.T) {
	setupTest()

	AddCategory("Tools", "Developer tools")

	proposalID := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", "Tools")

	testing.SetOriginCaller(std.Address("g1notadmin"))
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for unauthorized rejection")
		}
	}()
	RejectProposal(proposalID, "Not suitable")

	testing.SetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	RejectProposal(proposalID, "Not suitable")

	proposal, found := GetProposal(proposalID)
	uassert.True(t, found)
	uassert.Equal(t, StatusRejected, proposal.Status)
	uassert.Equal(t, "Not suitable", proposal.Reason)

	_, found = GetDapp(seqid.ID(1))
	uassert.False(t, found)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for invalid proposal ID")
		}
	}()
	RejectProposal(seqid.ID(999), "Does not exist")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for already rejected proposal")
		}
	}()
	RejectProposal(proposalID, "Already rejected")
}

func TestUpdateDappCategories(t *testing.T) {
	setupTest()

	AddCategory("Tools", "Developer tools")
	AddCategory("DeFi", "Decentralized finance")
	AddCategory("Games", "Gaming dApps")

	proposalID := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", "Tools")
	approvedID := ApproveProposal(proposalID)

	dapp, found := GetDapp(approvedID)
	uassert.True(t, found)
	uassert.Equal(t, 1, len(dapp.Categories), "Dapp should have 1 category")
	uassert.True(t, containsString(dapp.Categories, "Tools"))

	toolsCategory := GetCategory("Tools")
	uassert.True(t, containsID(toolsCategory.DappIDs, approvedID), "Tools category should contain dapp ID "+approvedID.String())

	testing.SetOriginCaller(std.Address("g1notadmin"))
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for unauthorized category update")
		}
	}()
	UpdateDappCategories(approvedID, "Tools,DeFi")

	testing.SetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	UpdateDappCategories(approvedID, "DeFi,Games")

	dapp, found = GetDapp(approvedID)
	uassert.True(t, found, "Dapp should be found")
	uassert.Equal(t, 2, len(dapp.Categories), "Dapp should have 2 categories")
	uassert.True(t, containsString(dapp.Categories, "DeFi"), "Dapp should have DeFi category")
	uassert.True(t, containsString(dapp.Categories, "Games"), "Dapp should have Games category")
	uassert.False(t, containsString(dapp.Categories, "Tools"), "Dapp should not have Tools category")

	toolsCategory = GetCategory("Tools")
	uassert.False(t, containsID(toolsCategory.DappIDs, approvedID), "Tools category should not contain dapp ID "+approvedID.String())

	defiCategory := GetCategory("DeFi")
	uassert.True(t, containsID(defiCategory.DappIDs, approvedID), "DeFi category should contain dapp ID "+approvedID.String())

	gamesCategory := GetCategory("Games")
	uassert.True(t, containsID(gamesCategory.DappIDs, approvedID), "Games category should contain dapp ID "+approvedID.String())

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for non-existent category")
		}
	}()
	UpdateDappCategories(approvedID, "NonExistent")
}

func TestDeleteDapp(t *testing.T) {
	setupTest()

	AddCategory("Tools", "Developer tools")
	AddCategory("DeFi", "Decentralized finance")

	proposalID := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", "Tools,DeFi")
	approvedID := ApproveProposal(proposalID)

	testing.SetOriginCaller(std.Address("g1notadmin"))
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for unauthorized dapp deletion")
		}
	}()
	DeleteDapp(approvedID)

	testing.SetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	DeleteDapp(approvedID)

	_, found1 := GetDapp(approvedID)
	uassert.False(t, found1)

	toolsCategory := GetCategory("Tools")
	uassert.False(t, containsID(toolsCategory.DappIDs, approvedID))

	defiCategory := GetCategory("DeFi")
	uassert.False(t, containsID(defiCategory.DappIDs, approvedID))

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for invalid dapp ID")
		}
	}()
	DeleteDapp(seqid.ID(999))
}

func TestDeleteCategory(t *testing.T) {
	setupTest()

	AddCategory("Tools", "Developer tools")
	AddCategory("DeFi", "Decentralized finance")

	proposalID := ProposeNewDapp("Test dApp", "A test dApp", "https://example.com", "Tools,DeFi")
	approvedID := ApproveProposal(proposalID)

	testing.SetOriginCaller(std.Address("g1notadmin"))
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for unauthorized category deletion")
		}
	}()
	DeleteCategory("Tools")

	testing.SetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	DeleteCategory("Tools")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for getting deleted category")
		}
	}()
	GetCategory("Tools")

	dapp, found := GetDapp(approvedID)
	uassert.True(t, found)
	uassert.Equal(t, 1, len(dapp.Categories))
	uassert.False(t, containsString(dapp.Categories, "Tools"))
	uassert.True(t, containsString(dapp.Categories, "DeFi"))

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for non-existent category")
		}
	}()
	DeleteCategory("NonExistent")
}

func TestGetDappsByCategory(t *testing.T) {
	setupTest()

	AddCategory("Tools", "Developer tools")
	AddCategory("DeFi", "Decentralized finance")

	proposalID1 := ProposeNewDapp("Tool dApp", "A tool dApp", "https://example.com/tool", "Tools")
	approvedID1 := ApproveProposal(proposalID1)

	proposalID2 := ProposeNewDapp("DeFi dApp", "A DeFi dApp", "https://example.com/defi", "DeFi")
	approvedID2 := ApproveProposal(proposalID2)

	proposalID3 := ProposeNewDapp("Hybrid dApp", "A hybrid dApp", "https://example.com/hybrid", "Tools,DeFi")
	approvedID3 := ApproveProposal(proposalID3)

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
		AddCategory(cat, "Description for "+cat)
	}

	var firstDappID seqid.ID

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
		proposalID := ProposeNewDapp(config.title, config.description, config.url, config.categories)
		dappID := ApproveProposal(proposalID)
		if i == 0 {
			firstDappID = dappID
		}
	}

	allDapps := ListDapps()
	for _, dapp := range allDapps {
		for _, categoryName := range dapp.Categories {
			category := GetCategory(categoryName)
			uassert.True(t, containsID(category.DappIDs, dapp.ID))
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

	testing.SetOriginCaller(std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))
	UpdateDappCategories(firstDappID, "Tools,Social")

	dapp, found := GetDapp(firstDappID)
	uassert.True(t, found)
	uassert.Equal(t, 2, len(dapp.Categories))
	uassert.True(t, containsString(dapp.Categories, "Tools"))
	uassert.True(t, containsString(dapp.Categories, "Social"))

	toolsCategory := GetCategory("Tools")
	uassert.True(t, containsID(toolsCategory.DappIDs, firstDappID))

	socialCategory := GetCategory("Social")
	uassert.True(t, containsID(socialCategory.DappIDs, firstDappID))

	DeleteCategory("Games")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for getting deleted category")
		}
	}()
	GetCategory("Games")

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

func containsID(slice []seqid.ID, id seqid.ID) bool {
	for _, item := range slice {
		if item == id {
			return true
		}
	}
	return false
}
