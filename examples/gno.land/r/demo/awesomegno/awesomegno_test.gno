package awesomegno

import (
	"std"
	"testing"

	"gno.land/p/nt/seqid"
	"gno.land/p/nt/uassert"
)

var adminAddr = std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")

func TestAddItem(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	uassert.NotPanics(t, func() {
		AddItem(cross, "Test Item", "A test description", "https://test.com", validAddr.String())
	})

	uassert.AbortsWithMessage(t, ErrItemTitleEmpty, func() {
		AddItem(cross, "", "description", "url", validAddr.String())
	})
}

func TestUpdateItem(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	AddItem(cross, "Original Item", "Original description", "https://original.com", validAddr.String())

	AddCategory(cross, "Test Category", "Test Category Description")

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)

	uassert.NotPanics(t, func() {
		UpdateItem(cross, item.ID, "Updated Item", "Updated description", "https://updated.com")
	})

	invalidID := seqid.ID(999999)
	uassert.AbortsWithMessage(t, ErrItemNotFound+": "+invalidID.String(), func() {
		UpdateItem(cross, invalidID, "title", "desc", "url")
	})
}

func TestUpdateItemsAuthors(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	validAddr2 := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")

	AddItem(cross, "Original Item", "Original description", "https://original.com", validAddr.String())

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)

	authors := []std.Address{validAddr, validAddr2}

	uassert.NotPanics(t, func() {
		UpdateItemsAuthors(cross, item.ID, authors)
	})
}

func TestDeleteItem(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	AddItem(cross, "Item to Delete", "Will be deleted", "https://delete.com", validAddr.String())

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)

	uassert.NotPanics(t, func() {
		DeleteItem(cross, item.ID)
	})

	uassert.AbortsWithMessage(t, ErrItemNotFound+": "+item.ID.String(), func() {
		DeleteItem(cross, item.ID)
	})
}

func TestAddCategory(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	uassert.NotPanics(t, func() {
		AddCategory(cross, "Test Category", "A test category description")
	})

	uassert.AbortsWithMessage(t, ErrCategoryNameEmpty, func() {
		AddCategory(cross, "", "description")
	})
}

func TestUpdateCategory(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	AddCategory(cross, "Original Category", "Original description")

	_, catValue := categories.GetByIndex(0)
	category := catValue.(*Category)

	uassert.NotPanics(t, func() {
		UpdateCategory(cross, category.ID, "Updated Category", "Updated description")
	})

	invalidID := seqid.ID(999999)
	uassert.AbortsWithMessage(t, ErrCategoryNotFound+": "+invalidID.String(), func() {
		UpdateCategory(cross, invalidID, "name", "desc")
	})
}

func TestDeleteCategory(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	AddCategory(cross, "Category to Delete", "Will be deleted")

	_, catValue := categories.GetByIndex(0)
	category := catValue.(*Category)

	uassert.NotPanics(t, func() {
		DeleteCategory(cross, category.ID)
	})

	uassert.AbortsWithMessage(t, ErrCategoryNotFound+": "+category.ID.String(), func() {
		DeleteCategory(cross, category.ID)
	})
}

func TestItemCategoryRelationship(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	AddItem(cross, "Relationship Test Item", "Testing relationships", "https://test.com", validAddr.String())
	AddCategory(cross, "Relationship Category", "For testing relationships")

	_, itemValue := items.GetByIndex(0)
	_, catValue := categories.GetByIndex(0)
	item := itemValue.(*Item)
	category := catValue.(*Category)

	uassert.NotPanics(t, func() {
		UpdateItemWithCategories(cross, item.ID, "", "", "", category.ID)
	})

	DeleteCategory(cross, category.ID)

	uassert.AbortsWithMessage(t, ErrCategoryNotFound+": "+category.ID.String(), func() {
		UpdateItemWithCategories(cross, item.ID, "", "", "", category.ID)
	})
}

func TestUpdateItemWithCategories(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	AddItem(cross, "Original Item", "Original description", "https://original.com", validAddr.String())
	AddCategory(cross, "Cat1", "Category 1")
	AddCategory(cross, "Cat2", "Category 2")

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)
	_, cat1Value := categories.GetByIndex(0)
	cat1 := cat1Value.(*Category)
	_, cat2Value := categories.GetByIndex(1)
	cat2 := cat2Value.(*Category)

	uassert.NotPanics(t, func() {
		UpdateItemWithCategories(cross, item.ID, "Updated Item", "Updated description", "https://updated.com", cat1.ID, cat2.ID)
	})

	invalidID := seqid.ID(999999)
	uassert.AbortsWithMessage(t, ErrItemNotFound+": "+invalidID.String(), func() {
		UpdateItemWithCategories(cross, invalidID, "title", "desc", "url", cat1.ID)
	})
}

func TestUpdateItemNoCategories(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	AddItem(cross, "Original Item", "Original description", "https://original.com", validAddr.String())

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)

	uassert.NotPanics(t, func() {
		UpdateItem(cross, item.ID, "Updated Item", "Updated description", "https://updated.com")
	})

	invalidID := seqid.ID(999999)
	uassert.AbortsWithMessage(t, ErrItemNotFound+": "+invalidID.String(), func() {
		UpdateItem(cross, invalidID, "title", "desc", "url")
	})
}
