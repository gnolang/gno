package awesomegno

import (
	"std"
	"testing"

	"gno.land/p/demo/seqid"
	"gno.land/p/demo/urequire"
)

var adminAddr = std.Address("g125em6arxsnj49vx35f0n0z34putv5ty3376fg5")

func TestAddDapp(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	id := AddDapp("Test Dapp", "A test description", "https://test.com", validAddr.String())
	urequire.NotEmpty(t, id.String())

	urequire.PanicsWithMessage(t, "dapp title cannot be empty", func() {
		AddDapp("", "description", "url", validAddr.String())
	})

	urequire.PanicsWithMessage(t, "invalid author address", func() {
		AddDapp("title", "description", "url", "")
	})
}

func TestUpdateDapp(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	id := AddDapp("Original Dapp", "Original description", "https://original.com", validAddr.String())

	catID := AddCategory("Test Category", "Test Category Description")

	urequire.NotPanics(t, func() {
		UpdateDapp(id, "Updated Dapp", "Updated description", "https://updated.com", catID)
	})

	nonExistentID, err := seqid.FromString("9999999")
	urequire.NoError(t, err)
	urequire.PanicsWithMessage(t, "dApp not found: "+nonExistentID.String(), func() {
		UpdateDapp(nonExistentID, "title", "desc", "url")
	})

	invalidCatID, err := seqid.FromString("9999999")
	urequire.NoError(t, err)
	urequire.PanicsWithMessage(t, "category not found: "+invalidCatID.String(), func() {
		UpdateDapp(id, "title", "desc", "url", invalidCatID)
	})
}

func TestUpdateDappsAuthors(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	validAddr2 := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf6")
	id := AddDapp("Original Dapp", "Original description", "https://original.com", validAddr.String())

	authors := []std.Address{validAddr, validAddr2}

	urequire.NotPanics(t, func() {
		UpdateDappsAuthors(id, authors)
	})
}

func TestDeleteDapp(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	id := AddDapp("Dapp to Delete", "Will be deleted", "https://delete.com", validAddr.String())

	urequire.NotPanics(t, func() {
		DeleteDapp(id)
	})

	urequire.PanicsWithMessage(t, "dApp not found: "+id.String(), func() {
		DeleteDapp(id)
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

func TestDappCategoryRelationship(t *testing.T) {
	t.Helper()
	testing.SetOriginCaller(adminAddr)

	validAddr := std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	dappID := AddDapp("Relationship Test Dapp", "Testing relationships", "https://test.com", validAddr.String())
	catID := AddCategory("Relationship Category", "For testing relationships")

	urequire.NotPanics(t, func() {
		UpdateDapp(dappID, "", "", "", catID)
	})

	DeleteCategory(catID)

	urequire.PanicsWithMessage(t, "category not found: "+catID.String(), func() {
		UpdateDapp(dappID, "", "", "", catID)
	})
}
