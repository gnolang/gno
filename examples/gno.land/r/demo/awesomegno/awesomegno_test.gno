package awesomegno

import (
	"std"
	"testing"

	"gno.land/p/demo/seqid"
	"gno.land/p/demo/urequire"
)

var adminAddr = std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")

func TestAddItem(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	id := AddItem("Test Item", "A test description", "https://test.com", validAddr.String())
	urequire.NotEmpty(t, id.String())

	urequire.PanicsWithMessage(t, "item title cannot be empty", func() {
		AddItem("", "description", "url", validAddr.String())
	})
}

func TestUpdateItem(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	id := AddItem("Original Item", "Original description", "https://original.com", validAddr.String())

	catID := AddCategory("Test Category", "Test Category Description")

	urequire.NotPanics(t, func() {
		UpdateItem(id, "Updated Item", "Updated description", "https://updated.com", catID)
	})

	nonExistentID, err := seqid.FromString("9999999")
	urequire.NoError(t, err)
	urequire.PanicsWithMessage(t, "item not found: "+nonExistentID.String(), func() {
		UpdateItem(nonExistentID, "title", "desc", "url")
	})

	invalidCatID, err := seqid.FromString("9999999")
	urequire.NoError(t, err)
	urequire.PanicsWithMessage(t, "category not found: "+invalidCatID.String(), func() {
		UpdateItem(id, "title", "desc", "url", invalidCatID)
	})
}

func TestUpdateItemsAuthors(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	validAddr2 := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf6")
	id := AddItem("Original Item", "Original description", "https://original.com", validAddr.String())

	authors := []std.Address{validAddr, validAddr2}

	urequire.NotPanics(t, func() {
		UpdateItemsAuthors(id, authors)
	})
}

func TestDeleteItem(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	id := AddItem("Item to Delete", "Will be deleted", "https://delete.com", validAddr.String())

	urequire.NotPanics(t, func() {
		DeleteItem(id)
	})

	urequire.PanicsWithMessage(t, "item not found: "+id.String(), func() {
		DeleteItem(id)
	})
}

func TestAddCategory(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	id := AddCategory("Test Category", "A test category description")
	urequire.NotEmpty(t, id.String())

	urequire.PanicsWithMessage(t, "category name cannot be empty", func() {
		AddCategory("", "description")
	})
}

func TestUpdateCategory(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	id := AddCategory("Original Category", "Original description")

	urequire.NotPanics(t, func() {
		UpdateCategory(id, "Updated Category", "Updated description")
	})

	nonExistentID, err := seqid.FromString("9999999")
	urequire.NoError(t, err)
	urequire.PanicsWithMessage(t, "category not found: "+nonExistentID.String(), func() {
		UpdateCategory(nonExistentID, "name", "desc")
	})
}

func TestDeleteCategory(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	id := AddCategory("Category to Delete", "Will be deleted")

	urequire.NotPanics(t, func() {
		DeleteCategory(id)
	})

	urequire.PanicsWithMessage(t, "category not found: "+id.String(), func() {
		DeleteCategory(id)
	})
}

func TestItemCategoryRelationship(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	itemID := AddItem("Relationship Test Item", "Testing relationships", "https://test.com", validAddr.String())
	catID := AddCategory("Relationship Category", "For testing relationships")

	urequire.NotPanics(t, func() {
		UpdateItem(itemID, "", "", "", catID)
	})

	DeleteCategory(catID)

	urequire.PanicsWithMessage(t, "category not found: "+catID.String(), func() {
		UpdateItem(itemID, "", "", "", catID)
	})
}
