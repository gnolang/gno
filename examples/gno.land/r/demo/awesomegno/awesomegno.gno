package awesomegno

import (
	"std"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/seqid"
	"gno.land/p/moul/authz"
	"gno.land/r/leon/hof"
)

var (
	dapps     = avl.NewTree()
	proposals = avl.NewTree()

	nextDappID     seqid.ID
	nextProposalID seqid.ID
	nextCategoryID seqid.ID

	categories = avl.NewTree()

	Auth = authz.NewMemberAuthority(
		std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5"), // leon
		std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5"), // moul
	)
)

func init() {
	hof.Register("AwesomeGno", "Awesome Gno is a registry for gnowesome dApps on gno.land")
}

func AddCategory(name, description string) seqid.ID {
	caller := std.PreviousRealm().Address()
	if !isAdmin(caller) {
		panic("not authorized: only admins can add categories")
	}

	if name == "" {
		panic("category name cannot be empty")
	}

	sanitizedKey := strings.TrimSpace(name)
	if categories.Has(sanitizedKey) {
		panic("category already exists")
	}

	category := &Category{
		ID:          nextCategoryID.Next(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		DappIDs:     []seqid.ID{},
	}

	categories.Set(sanitizedKey, category)

	std.Emit("CategoryAdded",
		"id", category.ID.String(),
		"name", name)

	return category.ID
}

func GetCategory(name string) *Category {
	sanitizedKey := strings.TrimSpace(name)
	value, found := categories.Get(sanitizedKey)
	if !found {
		panic("category not found")
	}

	return value.(*Category)
}

func GetCategoryByID(id seqid.ID) Category {
	value, found := categories.Get(id.String())
	if !found {
		panic("category not found")
	}

	return value.(Category)
}

func RemoveDappFromCategory(dappID seqid.ID, categoryName string) {
	category := GetCategory(categoryName)

	newDappIDs := []seqid.ID{}
	for _, id := range category.DappIDs {
		if id != dappID {
			newDappIDs = append(newDappIDs, id)
		}
	}

	category.DappIDs = newDappIDs
	categories.Set(strings.TrimSpace(categoryName), category)
}

func ListCategories() []Category {
	result := []Category{}

	readOnlyCategories := rotree.Wrap(categories, nil)

	readOnlyCategories.Iterate("", "", func(key string, value any) bool {
		result = append(result, value.(Category))
		return false
	})

	return result
}

func ProposeNewDapp(title, description, url string, categoryNamesStr string) seqid.ID {
	caller := std.PreviousRealm().Address()

	if title == "" {
		panic("title cannot be empty")
	}

	if description == "" {
		panic("description cannot be empty")
	}

	if url == "" {
		panic("url cannot be empty")
	}

	if categoryNamesStr == "" {
		panic("no categories provided")
	}

	categoryNames := []string{}
	for _, name := range strings.Split(categoryNamesStr, ",") {
		trimmedName := strings.TrimSpace(name)
		if trimmedName != "" {
			categoryNames = append(categoryNames, trimmedName)
		}
	}

	for _, catName := range categoryNames {
		GetCategory(catName)
	}

	dapp := &Dapp{
		Title:       title,
		Description: description,
		Author:      caller,
		URL:         url,
		Approved:    false,
		CreatedAt:   time.Now(),
		Categories:  categoryNames,
	}

	proposal := &Proposal{
		ID:        nextProposalID.Next(),
		Dapp:      dapp,
		Status:    StatusPending,
		Proposer:  caller,
		Reason:    "",
		CreatedAt: time.Now(),
	}

	proposals.Set(proposal.ID.String(), proposal)

	std.Emit("ProposalCreated",
		"id", proposal.ID.String(),
		"title", title,
		"proposer", caller.String())

	return proposal.ID
}

func ApproveProposal(proposalID seqid.ID) seqid.ID {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		panic("not authorized: only admins can approve proposals")
	}

	value, found := proposals.Get(proposalID.String())
	if !found {
		panic(ufmt.Sprintf("proposal not found: %s", proposalID.String()))
	}

	proposal := value.(Proposal)
	if proposal.Status != StatusPending {
		panic("proposal is not pending")
	}

	dapp := proposal.Dapp
	dapp.ID = nextDappID.Next()
	dapp.Approved = true

	dapps.Set(dapp.ID.String(), dapp)

	for _, categoryName := range dapp.Categories {
		addDappToCategory(dapp.ID, categoryName)
	}
	proposal.Status = StatusApproved
	proposals.Set(proposalID.String(), proposal)

	std.Emit("ProposalApproved",
		"id", proposalID.String(),
		"dapp_id", dapp.ID.String())

	return dapp.ID
}

func RejectProposal(proposalID seqid.ID, reason string) {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		panic("not authorized: only admins can reject proposals")
	}

	proposalRaw, exists := proposals.Get(proposalID.String())
	if !exists {
		panic(ufmt.Sprintf("proposal with ID %s does not exist", proposalID.String()))
	}

	proposal := proposalRaw.(Proposal)
	if proposal.Status != StatusPending {
		panic("proposal is not pending")
	}

	proposal.Status = StatusRejected
	proposal.Reason = reason
	proposals.Set(proposalID.String(), proposal)

	std.Emit("ProposalRejected",
		"id", proposalID.String(),
		"title", proposal.Dapp.Title,
		"reason", reason)
}

func GetDapp(id seqid.ID) (Dapp, bool) {
	dappRaw, exists := dapps.Get(id.String())
	if !exists {
		return Dapp{}, false
	}

	return dappRaw.(Dapp), true
}

func GetProposal(id seqid.ID) (Proposal, bool) {
	proposalRaw, exists := proposals.Get(id.String())
	if !exists {
		return Proposal{}, false
	}

	return proposalRaw.(Proposal), true
}

func ListDapps() []Dapp {
	result := []Dapp{}

	dapps.Iterate("", "", func(key string, value interface{}) bool {
		result = append(result, value.(Dapp))
		return false
	})

	return result
}

func ListPendingProposals() []Proposal {
	result := []Proposal{}

	proposals.Iterate("", "", func(key string, value interface{}) bool {
		proposal := value.(Proposal)
		if proposal.Status == StatusPending {
			result = append(result, proposal)
		}
		return false
	})

	return result
}

func GetDappsByCategory(categoryName string) []Dapp {
	result := []Dapp{}

	category := GetCategory(categoryName)

	for _, dappID := range category.DappIDs {
		value, found := dapps.Get(dappID.String())
		if found {
			dapp := value.(Dapp)
			if dapp.Approved {
				result = append(result, dapp)
			}
		}
	}

	return result
}

func UpdateDappCategories(dappID seqid.ID, newCategories string) {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		panic("not authorized: only admins can update dapp categories")
	}

	value, found := dapps.Get(dappID.String())
	if !found {
		panic(ufmt.Sprintf("dApp not found: %s", dappID.String()))
	}

	dapp := value.(Dapp)

	newCatTree := avl.NewTree()
	for _, catName := range strings.Split(newCategories, ",") {
		trimmedName := strings.TrimSpace(catName)
		if trimmedName != "" {
			if !categories.Has(trimmedName) {
				panic(ufmt.Sprintf("category not found: %s", trimmedName))
			}
			newCatTree.Set(trimmedName, struct{}{})
		}
	}

	for _, oldCat := range dapp.Categories {
		if !newCatTree.Has(oldCat) {
			RemoveDappFromCategory(dappID, oldCat)
		}
	}

	newCatTree.Iterate("", "", func(key string, _ any) bool {
		if !categories.Has(key) {
			return true
		}
		addDappToCategory(dappID, key)
		return false
	})

	dapp.Categories = []string{}
	newCatTree.Iterate("", "", func(key string, _ any) bool {
		dapp.Categories = append(dapp.Categories, key)
		return false
	})
	dapps.Set(dappID.String(), dapp)
}

func DeleteDapp(dappID seqid.ID) {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		panic("not authorized: only admins can delete dapps")
	}

	value, found := dapps.Get(dappID.String())
	if !found {
		panic(ufmt.Sprintf("dApp not found: %s", dappID.String()))
	}

	dapp := value.(Dapp)
	for _, categoryName := range dapp.Categories {
		RemoveDappFromCategory(dappID, categoryName)
	}

	dapps.Remove(dappID.String())

	std.Emit("DappDeleted", "id", dappID.String())
}

func DeleteCategory(categoryName string) {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		panic("not authorized: only admins can delete categories")
	}

	category := GetCategory(categoryName)

	for _, dappID := range category.DappIDs {
		value, found := dapps.Get(dappID.String())
		if found {
			dapp := value.(Dapp)

			newCategories := []string{}
			for _, cat := range dapp.Categories {
				if cat != categoryName {
					newCategories = append(newCategories, cat)
				}
			}

			dapp.Categories = newCategories
			dapps.Set(dappID.String(), dapp)
		}
	}

	categories.Remove(strings.TrimSpace(categoryName))

	std.Emit("CategoryDeleted", "name", categoryName)
}

func addDappToCategory(dappID seqid.ID, categoryName string) {
	category := GetCategory(categoryName)

	for _, id := range category.DappIDs {
		if id == dappID {
			return
		}
	}

	category.DappIDs = append(category.DappIDs, dappID)
	categories.Set(strings.TrimSpace(categoryName), category)
}

func isAdmin(addr std.Address) bool {
	return Auth.Has(addr)
}