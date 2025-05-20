package awesomegno

import (
	"std"
	"testing"

	"gno.land/p/demo/seqid"
	"gno.land/p/demo/uassert"
)

var adminAddr = std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")

func TestAddItem(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	uassert.NotPanics(t, func() {
		AddItem("Test Item", "A test description", "https://test.com", validAddr.String())
	})

	uassert.PanicsWithMessage(t, "item title cannot be empty", func() {
		AddItem("", "description", "url", validAddr.String())
	})
}

func TestUpdateItem(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	AddItem("Original Item", "Original description", "https://original.com", validAddr.String())

	AddCategory("Test Category", "Test Category Description")

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)

	uassert.NotPanics(t, func() {
		UpdateItem(item.ID, "Updated Item", "Updated description", "https://updated.com")
	})

	invalidID := seqid.ID(999999)
	uassert.PanicsWithMessage(t, "item not found: "+invalidID.String(), func() {
		UpdateItem(invalidID, "title", "desc", "url")
	})
}

func TestUpdateItemsAuthors(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	validAddr2 := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf6")

	AddItem("Original Item", "Original description", "https://original.com", validAddr.String())

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)

	authors := []std.Address{validAddr, validAddr2}

	uassert.NotPanics(t, func() {
		UpdateItemsAuthors(item.ID, authors)
	})
}

func TestDeleteItem(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	AddItem("Item to Delete", "Will be deleted", "https://delete.com", validAddr.String())

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)

	uassert.NotPanics(t, func() {
		DeleteItem(item.ID)
	})

	uassert.PanicsWithMessage(t, "item not found: "+item.ID.String(), func() {
		DeleteItem(item.ID)
	})
}

func TestAddCategory(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	uassert.NotPanics(t, func() {
		AddCategory("Test Category", "A test category description")
	})

	uassert.PanicsWithMessage(t, "category name cannot be empty", func() {
		AddCategory("", "description")
	})
}

func TestUpdateCategory(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	AddCategory("Original Category", "Original description")

	_, catValue := categories.GetByIndex(0)
	category := catValue.(*Category)

	uassert.NotPanics(t, func() {
		UpdateCategory(category.ID, "Updated Category", "Updated description")
	})

	invalidID := seqid.ID(999999)
	uassert.PanicsWithMessage(t, "category not found: "+invalidID.String(), func() {
		UpdateCategory(invalidID, "name", "desc")
	})
}

func TestDeleteCategory(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	AddCategory("Category to Delete", "Will be deleted")

	_, catValue := categories.GetByIndex(0)
	category := catValue.(*Category)

	uassert.NotPanics(t, func() {
		DeleteCategory(category.ID)
	})

	uassert.PanicsWithMessage(t, "category not found: "+category.ID.String(), func() {
		DeleteCategory(category.ID)
	})
}

func TestItemCategoryRelationship(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	AddItem("Relationship Test Item", "Testing relationships", "https://test.com", validAddr.String())
	AddCategory("Relationship Category", "For testing relationships")

	_, itemValue := items.GetByIndex(0)
	_, catValue := categories.GetByIndex(0)
	item := itemValue.(*Item)
	category := catValue.(*Category)

	uassert.NotPanics(t, func() {
		UpdateItemWithCategories(item.ID, "", "", "", category.ID)
	})

	DeleteCategory(category.ID)

	uassert.PanicsWithMessage(t, "category not found: "+category.ID.String(), func() {
		UpdateItemWithCategories(item.ID, "", "", "", category.ID)
	})
}

func TestUpdateItemWithCategories(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	AddItem("Original Item", "Original description", "https://original.com", validAddr.String())
	AddCategory("Cat1", "Category 1")
	AddCategory("Cat2", "Category 2")

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)
	_, cat1Value := categories.GetByIndex(0)
	cat1 := cat1Value.(*Category)
	_, cat2Value := categories.GetByIndex(1)
	cat2 := cat2Value.(*Category)

	uassert.NotPanics(t, func() {
		UpdateItemWithCategories(item.ID, "Updated Item", "Updated description", "https://updated.com", cat1.ID, cat2.ID)
	})

	invalidID := seqid.ID(999999)
	uassert.PanicsWithMessage(t, "item not found: "+invalidID.String(), func() {
		UpdateItemWithCategories(invalidID, "title", "desc", "url", cat1.ID)
	})
}

func TestUpdateItemNoCategories(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	AddItem("Original Item", "Original description", "https://original.com", validAddr.String())

	_, itemValue := items.GetByIndex(0)
	item := itemValue.(*Item)

	uassert.NotPanics(t, func() {
		UpdateItem(item.ID, "Updated Item", "Updated description", "https://updated.com")
	})

	invalidID := seqid.ID(999999)
	uassert.PanicsWithMessage(t, "item not found: "+invalidID.String(), func() {
		UpdateItem(invalidID, "title", "desc", "url")
	})
}
