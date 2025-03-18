package awesomegno

import (
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var (
	dapps     = avl.NewTree()
	proposals = avl.NewTree()

	nextDappID     uint64 = 1
	nextProposalID uint64 = 1

	admins                = avl.NewTree()
	categories            = avl.NewTree()
	nextCategoryID uint64 = 1
)

func init() {
	admins.Set("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y", true)
	admins.Set(std.CurrentRealm().Address().String(), true)
	admins.Set(std.PreviousRealm().Address().String(), true)
}

func AddAdmin(addr std.Address) error {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		return ufmt.Errorf("not authorized: only admins can add new admins")
	}
	if !addr.IsValid() {
		return ufmt.Errorf("invalid address")
	}

	admins.Set(addr.String(), true)
	std.Emit("AdminAdded", "address", addr.String())
	return nil
}

func RemoveAdmin(addr std.Address) error {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		return ufmt.Errorf("not authorized: only admins can remove admins")
	}
	if !addr.IsValid() {
		return ufmt.Errorf("invalid address")
	}
	if admins.Size() <= 1 {
		return ufmt.Errorf("cannot remove the last admin")
	}

	admins.Remove(addr.String())
	std.Emit("AdminRemoved", "address", addr.String())
	return nil
}

func AddCategory(name, description string) (uint64, error) {
	caller := std.PreviousRealm().Address()
	if !isAdmin(caller) {
		return 0, ufmt.Errorf("not authorized: only admins can add categories")
	}
	if name == "" {
		return 0, ufmt.Errorf("category name cannot be empty")
	}

	exists := false
	categories.Iterate("", "", func(key string, value interface{}) bool {
		cat := value.(Category)
		if cat.Name == name {
			exists = true
			return true
		}
		return false
	})

	if exists {
		return 0, ufmt.Errorf("category already exists")
	}

	category := Category{
		ID:          nextCategoryID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		DappIDs:     []uint64{},
	}

	categories.Set(strconv.FormatUint(nextCategoryID, 10), category)

	std.Emit("CategoryAdded",
		"id", ufmt.Sprintf("%d", nextCategoryID),
		"name", name)

	id := nextCategoryID
	nextCategoryID++
	return id, nil
}

func GetCategory(name string) (Category, bool) {
	var category Category
	found := false

	categories.Iterate("", "", func(key string, value interface{}) bool {
		cat := value.(Category)
		if cat.Name == name {
			category = cat
			found = true
			return true
		}
		return false
	})

	return category, found
}

func GetCategoryByID(id uint64) (Category, bool) {
	value, found := categories.Get(strconv.FormatUint(id, 10))
	if !found {
		return Category{}, false
	}
	return value.(Category), true
}

func RemoveDappFromCategory(dappID uint64, categoryName string) error {
	category, found := GetCategory(categoryName)
	if !found {
		return ufmt.Errorf("category not found: %s", categoryName)
	}

	newDappIDs := []uint64{}
	for _, id := range category.DappIDs {
		if id != dappID {
			newDappIDs = append(newDappIDs, id)
		}
	}

	category.DappIDs = newDappIDs
	categories.Set(strconv.FormatUint(category.ID, 10), category)

	return nil
}

func ListCategories() []Category {
	result := []Category{}

	categories.Iterate("", "", func(key string, value interface{}) bool {
		result = append(result, value.(Category))
		return false
	})

	return result
}

func ProposeNewDapp(title, description, url string, categoryNamesStr string) (uint64, error) {
	caller := std.PreviousRealm().Address()

	if title == "" {
		return 0, ufmt.Errorf("title cannot be empty")
	}
	if description == "" {
		return 0, ufmt.Errorf("description cannot be empty")
	}
	if url == "" {
		return 0, ufmt.Errorf("URL cannot be empty")
	}

	var categoryNames []string
	if categoryNamesStr != "" {
		for _, name := range strings.Split(categoryNamesStr, ",") {
			trimmedName := strings.TrimSpace(name)
			if trimmedName != "" {
				categoryNames = append(categoryNames, trimmedName)
			}
		}
	}

	if len(categoryNames) > 0 {
		for _, catName := range categoryNames {
			categoryExists := false
			categories.Iterate("", "", func(key string, value interface{}) bool {
				cat := value.(Category)
				if cat.Name == catName {
					categoryExists = true
					return true
				}
				return false
			})
			if !categoryExists {
				return 0, ufmt.Errorf("category '%s' does not exist", catName)
			}
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
		ID:        nextProposalID,
		Dapp:      dapp,
		Status:    StatusPending,
		Proposer:  caller,
		Reason:    "",
		CreatedAt: time.Now(),
	}

	proposals.Set(strconv.FormatUint(nextProposalID, 10), proposal)

	std.Emit("ProposalCreated",
		"id", ufmt.Sprintf("%d", nextProposalID),
		"title", title,
		"proposer", caller.String())

	id := nextProposalID
	nextProposalID++
	return id, nil
}

func ApproveProposal(proposalID uint64) error {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		return ufmt.Errorf("not authorized: only admins can approve proposals")
	}

	value, found := proposals.Get(strconv.FormatUint(proposalID, 10))
	if !found {
		return ufmt.Errorf("proposal not found: %d", proposalID)
	}

	proposal := value.(Proposal)
	if proposal.Status != StatusPending {
		return ufmt.Errorf("proposal is not pending")
	}

	dapp := proposal.Dapp
	dapp.ID = nextDappID
	dapp.Approved = true
	dapps.Set(strconv.FormatUint(nextDappID, 10), dapp)

	for _, categoryName := range dapp.Categories {
		err := addDappToCategory(dapp.ID, categoryName)
		if err != nil {
			return err
		}
	}

	proposal.Status = StatusApproved
	proposals.Set(strconv.FormatUint(proposalID, 10), proposal)

	std.Emit("ProposalApproved",
		"id", ufmt.Sprintf("%d", proposalID),
		"dapp_id", ufmt.Sprintf("%d", nextDappID))

	nextDappID++
	return nil
}

func RejectProposal(proposalID uint64, reason string) error {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		return ufmt.Errorf("not authorized: only admins can reject proposals")
	}

	proposalRaw, exists := proposals.Get(strconv.FormatUint(proposalID, 10))
	if !exists {
		return ufmt.Errorf("proposal with ID %d does not exist", proposalID)
	}

	proposal := proposalRaw.(Proposal)
	if proposal.Status != StatusPending {
		return ufmt.Errorf("proposal is not pending")
	}

	proposal.Status = StatusRejected
	proposal.Reason = reason
	proposals.Set(strconv.FormatUint(proposalID, 10), proposal)

	std.Emit("ProposalRejected",
		"id", ufmt.Sprintf("%d", proposalID),
		"title", proposal.Dapp.Title,
		"reason", reason)

	return nil
}

func GetDapp(id uint64) (Dapp, bool) {
	dappRaw, exists := dapps.Get(strconv.FormatUint(id, 10))
	if !exists {
		return Dapp{}, false
	}
	return dappRaw.(Dapp), true
}

func GetProposal(id uint64) (Proposal, bool) {
	proposalRaw, exists := proposals.Get(strconv.FormatUint(id, 10))
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

	category, found := GetCategory(categoryName)
	if !found {
		return result
	}

	for _, dappID := range category.DappIDs {
		value, found := dapps.Get(strconv.FormatUint(dappID, 10))
		if found {
			dapp := value.(Dapp)
			if dapp.Approved {
				result = append(result, dapp)
			}
		}
	}

	return result
}

func UpdateDappCategories(dappID uint64, newCategories string) error {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		return ufmt.Errorf("not authorized: only admins can update dApp categories")
	}

	value, found := dapps.Get(strconv.FormatUint(dappID, 10))
	if !found {
		return ufmt.Errorf("dApp not found: %d", dappID)
	}

	dapp := value.(Dapp)
	for _, catName := range strings.Split(newCategories, ",") {
		_, found := GetCategory(catName)
		if !found {
			return ufmt.Errorf("category not found: %s", catName)
		}
	}

	for _, oldCat := range dapp.Categories {
		found := false
		for _, newCat := range strings.Split(newCategories, ",") {
			if oldCat == newCat {
				found = true
				break
			}
		}
		if !found {
			err := RemoveDappFromCategory(dappID, oldCat)
			if err != nil {
				return err
			}
		}
	}

	for _, newCat := range strings.Split(newCategories, ",") {
		found := false
		for _, oldCat := range dapp.Categories {
			if newCat == oldCat {
				found = true
				break
			}
		}
		if !found {
			err := addDappToCategory(dappID, newCat)
			if err != nil {
				return err
			}
		}
	}

	dapp.Categories = strings.Split(newCategories, ",")
	dapps.Set(strconv.FormatUint(dappID, 10), dapp)

	return nil
}

func DeleteDapp(dappID uint64) error {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		return ufmt.Errorf("not authorized: only admins can delete dApps")
	}

	value, found := dapps.Get(strconv.FormatUint(dappID, 10))
	if !found {
		return ufmt.Errorf("dApp not found: %d", dappID)
	}

	dapp := value.(Dapp)
	for _, categoryName := range dapp.Categories {
		err := RemoveDappFromCategory(dappID, categoryName)
		if err != nil {
			return err
		}
	}

	dapps.Remove(strconv.FormatUint(dappID, 10))

	std.Emit("DappDeleted", "id", ufmt.Sprintf("%d", dappID))

	return nil
}

func DeleteCategory(categoryName string) error {
	caller := std.PreviousRealm().Address()

	if !isAdmin(caller) {
		return ufmt.Errorf("not authorized: only admins can delete categories")
	}

	category, found := GetCategory(categoryName)
	if !found {
		return ufmt.Errorf("category not found: %s", categoryName)
	}

	for _, dappID := range category.DappIDs {
		value, found := dapps.Get(strconv.FormatUint(dappID, 10))
		if found {
			dapp := value.(Dapp)

			newCategories := []string{}
			for _, cat := range dapp.Categories {
				if cat != categoryName {
					newCategories = append(newCategories, cat)
				}
			}

			dapp.Categories = newCategories
			dapps.Set(strconv.FormatUint(dappID, 10), dapp)
		}
	}

	categories.Remove(strconv.FormatUint(category.ID, 10))

	std.Emit("CategoryDeleted", "name", categoryName)

	return nil
}

func addDappToCategory(dappID uint64, categoryName string) error {
	category, found := GetCategory(categoryName)
	if !found {
		return ufmt.Errorf("category not found: %s", categoryName)
	}

	for _, id := range category.DappIDs {
		if id == dappID {
			return nil
		}
	}

	category.DappIDs = append(category.DappIDs, dappID)
	categories.Set(strconv.FormatUint(category.ID, 10), category)

	return nil
}

func isAdmin(addr std.Address) bool {
	isAdminRaw, exists := admins.Get(addr.String())
	return exists && isAdminRaw.(bool)
}
