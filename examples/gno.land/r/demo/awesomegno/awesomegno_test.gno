package awesomegno

import (
	"std"
	"testing"

	"gno.land/p/demo/seqid"
	"gno.land/p/demo/uassert"
)

var adminAddr = std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")

func TestAddItem(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	uassert.NotPanics(t, func() {
		cross(AddItem)("Test Item", "A test description", "https://test.com", validAddr.String())
	})

	uassert.AbortsWithMessage(t, "item title cannot be empty", func() {
		cross(AddItem)("", "description", "url", validAddr.String())
	})
}

func TestUpdateItem(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	cross(AddItem)("Original Item", "Original description", "https://original.com", validAddr.String())

	cross(AddCategory)("Test Category", "Test Category Description")

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)

	uassert.NotPanics(t, func() {
		cross(UpdateItem)(item.ID, "Updated Item", "Updated description", "https://updated.com")
	})

	invalidID := seqid.ID(999999)
	uassert.AbortsWithMessage(t, "item not found: "+invalidID.String(), func() {
		cross(UpdateItem)(invalidID, "title", "desc", "url")
	})
}

func TestUpdateItemsAuthors(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	validAddr2 := std.Address("g1ej0qca5ptsw9kfr64ey8jvfy9eacga6mpj2z0y")

	cross(AddItem)("Original Item", "Original description", "https://original.com", validAddr.String())

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)

	authors := []std.Address{validAddr, validAddr2}

	uassert.NotPanics(t, func() {
		cross(UpdateItemsAuthors)(item.ID, authors)
	})
}

func TestDeleteItem(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	cross(AddItem)("Item to Delete", "Will be deleted", "https://delete.com", validAddr.String())

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)

	uassert.NotPanics(t, func() {
		cross(DeleteItem)(item.ID)
	})

	uassert.AbortsWithMessage(t, "item not found: "+item.ID.String(), func() {
		cross(DeleteItem)(item.ID)
	})
}

func TestAddCategory(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	uassert.NotPanics(t, func() {
		cross(AddCategory)("Test Category", "A test category description")
	})

	uassert.AbortsWithMessage(t, "category name cannot be empty", func() {
		cross(AddCategory)("", "description")
	})
}

func TestUpdateCategory(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	cross(AddCategory)("Original Category", "Original description")

	_, catValue := categories.GetByIndex(0)
	category := catValue.(*Category)

	uassert.NotPanics(t, func() {
		cross(UpdateCategory)(category.ID, "Updated Category", "Updated description")
	})

	invalidID := seqid.ID(999999)
	uassert.AbortsWithMessage(t, "category not found: "+invalidID.String(), func() {
		cross(UpdateCategory)(invalidID, "name", "desc")
	})
}

func TestDeleteCategory(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	cross(AddCategory)("Category to Delete", "Will be deleted")

	_, catValue := categories.GetByIndex(0)
	category := catValue.(*Category)

	uassert.NotPanics(t, func() {
		cross(DeleteCategory)(category.ID)
	})

	uassert.AbortsWithMessage(t, "category not found: "+category.ID.String(), func() {
		cross(DeleteCategory)(category.ID)
	})
}

func TestItemCategoryRelationship(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	cross(AddItem)("Relationship Test Item", "Testing relationships", "https://test.com", validAddr.String())
	cross(AddCategory)("Relationship Category", "For testing relationships")

	_, itemValue := items.GetByIndex(0)
	_, catValue := categories.GetByIndex(0)
	item := itemValue.(*Item)
	category := catValue.(*Category)

	uassert.NotPanics(t, func() {
		cross(UpdateItemWithCategories)(item.ID, "", "", "", category.ID)
	})

	cross(DeleteCategory)(category.ID)

	uassert.AbortsWithMessage(t, "category not found: "+category.ID.String(), func() {
		cross(UpdateItemWithCategories)(item.ID, "", "", "", category.ID)
	})
}

func TestUpdateItemWithCategories(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	cross(AddItem)("Original Item", "Original description", "https://original.com", validAddr.String())
	cross(AddCategory)("Cat1", "Category 1")
	cross(AddCategory)("Cat2", "Category 2")

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)
	_, cat1Value := categories.GetByIndex(0)
	cat1 := cat1Value.(*Category)
	_, cat2Value := categories.GetByIndex(1)
	cat2 := cat2Value.(*Category)

	uassert.NotPanics(t, func() {
		cross(UpdateItemWithCategories)(item.ID, "Updated Item", "Updated description", "https://updated.com", cat1.ID, cat2.ID)
	})

	invalidID := seqid.ID(999999)
	uassert.AbortsWithMessage(t, "item not found: "+invalidID.String(), func() {
		cross(UpdateItemWithCategories)(invalidID, "title", "desc", "url", cat1.ID)
	})
}

func TestUpdateItemNoCategories(t *testing.T) {
	adminRealm := std.NewUserRealm(adminAddr)
	testing.SetRealm(adminRealm)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	cross(AddItem)("Original Item", "Original description", "https://original.com", validAddr.String())

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)

	uassert.NotPanics(t, func() {
		cross(UpdateItem)(item.ID, "Updated Item", "Updated description", "https://updated.com")
	})

	invalidID := seqid.ID(999999)
	uassert.AbortsWithMessage(t, "item not found: "+invalidID.String(), func() {
		cross(UpdateItem)(invalidID, "title", "desc", "url")
	})
}
