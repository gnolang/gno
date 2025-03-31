package awesomegno

import (
	"std"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
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

	auth = authz.New()
)

func init() {
	auth.Current().(*authz.MemberAuthority).AddMember(std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")) // matijamarjanovic 
	hof.Register("AwesomeGno", "Awesome Gno is a registry for gnowesome dApps on gno.land")
}

func AddAdmin(addr std.Address) error {
	return auth.Do("add_admin", func() error {
		if !addr.IsValid() {
			return ErrInvalidAddress
		}
		if err := auth.Current().(*authz.MemberAuthority).AddMember(addr); err != nil {
			return ErrAdminAlreadyExists
		}
		std.Emit("AdminAdded", "address", addr.String())
		return nil
	})
}

func RemoveAdmin(addr std.Address) error {
	return auth.Do("remove_admin", func() error {
		if auth.Current().(*authz.MemberAuthority).Tree().Size() <= 1 {
			return ErrCannotRemoveLastAdmin
		}

		if err := auth.Current().(*authz.MemberAuthority).RemoveMember(addr); err != nil {
			return err
		}
		std.Emit("AdminRemoved", "address", addr.String())
		return nil
	})
}

func AddCategory(name, description string) (seqid.ID, error) {
	caller := std.PreviousRealm().Address()
	if !isAdmin(caller) {
		return 0, ErrNotAuthorized
	}

	if name == "" {
		return 0, ErrCategoryExists
	}

	sanitizedKey := strings.TrimSpace(name)
	if categories.Has(sanitizedKey) {
		return 0, ErrCategoryExists
	}

	category := Category{
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

	return category.ID, nil
}

func GetCategory(name string) (Category, error) {
	sanitizedKey := strings.TrimSpace(name)
	value, found := categories.Get(sanitizedKey)
	if !found {
		return Category{}, ErrCategoryNotFound
	}

	return value.(Category), nil
}

func GetCategoryByID(id seqid.ID) (Category, error) {
	value, found := categories.Get(id.String())
	if !found {
		return Category{}, ErrCategoryNotFound
	}

	return value.(Category), nil
}

func RemoveDappFromCategory(dappID seqid.ID, categoryName string) error {
	category, err := GetCategory(categoryName)
	if err != nil {
		return err
	}

	newDappIDs := []seqid.ID{}
	for _, id := range category.DappIDs {
		if id != dappID {
			newDappIDs = append(newDappIDs, id)
		}
	}

	category.DappIDs = newDappIDs
	categories.Set(strings.TrimSpace(categoryName), category)

	return nil
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

func ProposeNewDapp(title, description, url string, categoryNamesStr string) (seqid.ID, error) {
	caller := std.PreviousRealm().Address()

	if title == "" {
		return 0, ErrTitleEmpty
	}

	if description == "" {
		return 0, ErrDescriptionEmpty
	}

	if url == "" {
		return 0, ErrURLCannotBeEmpty
	}

	if categoryNamesStr == "" {
		return 0, ErrNoCategoriesProvided
	}

	categoryNames := []string{}
	for _, name := range strings.Split(categoryNamesStr, ",") {
		trimmedName := strings.TrimSpace(name)
		if trimmedName != "" {
			categoryNames = append(categoryNames, trimmedName)
		}
	}

	for _, catName := range categoryNames {
		if _, err := GetCategory(catName); err != nil {
			return 0, ufmt.Errorf("category '%s' does not exist", catName)
		}
	}

	dapp := Dapp{
		Title:       title,
		Description: description,
		Author:      caller,
		URL:         url,
		Approved:    false,
		CreatedAt:   time.Now(),
		Categories:  categoryNames,
	}

	proposal := Proposal{
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

	return proposal.ID, nil
}

func ApproveProposal(proposalID seqid.ID) (seqid.ID, error) {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		return 0, ufmt.Errorf("not authorized: only admins can approve proposals")
	}

	value, found := proposals.Get(proposalID.String())
	if !found {
		return 0, ufmt.Errorf("proposal not found: %s", proposalID.String())
	}

	proposal := value.(Proposal)
	if proposal.Status != StatusPending {
		return 0, ufmt.Errorf("proposal is not pending")
	}

	dapp := proposal.Dapp
	dapp.ID = nextDappID.Next()
	dapp.Approved = true

	dapps.Set(dapp.ID.String(), dapp)

	for _, categoryName := range dapp.Categories {
		err := addDappToCategory(dapp.ID, categoryName)
		if err != nil {
			return 0, err
		}
	}
	proposal.Status = StatusApproved
	proposals.Set(proposalID.String(), proposal)

	std.Emit("ProposalApproved",
		"id", proposalID.String(),
		"dapp_id", dapp.ID.String())

	return dapp.ID, nil
}

func RejectProposal(proposalID seqid.ID, reason string) error {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		return ufmt.Errorf("not authorized: only admins can reject proposals")
	}

	proposalRaw, exists := proposals.Get(proposalID.String())
	if !exists {
		return ufmt.Errorf("proposal with ID %s does not exist", proposalID.String())
	}

	proposal := proposalRaw.(Proposal)
	if proposal.Status != StatusPending {
		return ufmt.Errorf("proposal is not pending")
	}

	proposal.Status = StatusRejected
	proposal.Reason = reason
	proposals.Set(proposalID.String(), proposal)

	std.Emit("ProposalRejected",
		"id", proposalID.String(),
		"title", proposal.Dapp.Title,
		"reason", reason)

	return nil
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

	category, err := GetCategory(categoryName)
	if err != nil {
		return result
	}

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

func UpdateDappCategories(dappID seqid.ID, newCategories string) error {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		return ErrNotAuthorized
	}

	value, found := dapps.Get(dappID.String())
	if !found {
		return ufmt.Errorf("dApp not found: %s", dappID.String())
	}

	dapp := value.(Dapp)

	newCatTree := avl.NewTree()
	for _, catName := range strings.Split(newCategories, ",") {
		trimmedName := strings.TrimSpace(catName)
		if trimmedName != "" {
			if !categories.Has(trimmedName) {
				return ufmt.Errorf("category not found: %s", trimmedName)
			}
			newCatTree.Set(trimmedName, struct{}{})
		}
	}

	for _, oldCat := range dapp.Categories {
		if !newCatTree.Has(oldCat) {
			if err := RemoveDappFromCategory(dappID, oldCat); err != nil {
				return err
			}
		}
	}

	newCatTree.Iterate("", "", func(key string, _ any) bool {
		if !categories.Has(key) {
			return true
		}
		if err := addDappToCategory(dappID, key); err != nil {
			return true
		}
		return false
	})

	dapp.Categories = []string{}
	newCatTree.Iterate("", "", func(key string, _ any) bool {
		dapp.Categories = append(dapp.Categories, key)
		return false
	})
	dapps.Set(dappID.String(), dapp)

	return nil
}

func DeleteDapp(dappID seqid.ID) error {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		return ErrNotAuthorized
	}

	value, found := dapps.Get(dappID.String())
	if !found {
		return ufmt.Errorf("dApp not found: %s", dappID.String())
	}

	dapp := value.(Dapp)
	for _, categoryName := range dapp.Categories {
		err := RemoveDappFromCategory(dappID, categoryName)
		if err != nil {
			return err
		}
	}

	dapps.Remove(dappID.String())

	std.Emit("DappDeleted", "id", dappID.String())

	return nil
}

func DeleteCategory(categoryName string) error {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		return ErrNotAuthorized
	}

	category, err := GetCategory(categoryName)
	if err != nil {
		return err
	}

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

	return nil
}

func addDappToCategory(dappID seqid.ID, categoryName string) error {
	category, err := GetCategory(categoryName)
	if err != nil {
		return err
	}

	for _, id := range category.DappIDs {
		if id == dappID {
			return nil
		}
	}

	category.DappIDs = append(category.DappIDs, dappID)
	categories.Set(strings.TrimSpace(categoryName), category)

	return nil
}

func isAdmin(addr std.Address) bool {
	return auth.Current().(*authz.MemberAuthority).Has(addr)
}
