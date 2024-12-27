package pager

import (
	"testing"

	"gno.land/p/moul/pageable"
	"gno.land/p/moul/ulist"
)

func TestNewPager(t *testing.T) {
	// Create a list and add some items
	list := ulist.New()
	list.Append("a", "b", "c", "d", "e")

	// Test forward paging
	pager := NewPager(list, 2, false)

	// First page
	page := pager.GetPage(0)
	assertPage(t, page, []string{"a", "b"}, "first page")

	// Second page
	page = pager.GetPage(1)
	assertPage(t, page, []string{"c", "d"}, "second page")

	// Last page (partial)
	page = pager.GetPage(2)
	assertPage(t, page, []string{"e"}, "last page")

	// Test reverse paging
	reversePager := NewPager(list, 2, true)

	// First page (from end)
	page = reversePager.GetPage(0)
	assertPage(t, page, []string{"e", "d"}, "first reverse page")

	// Second page
	page = reversePager.GetPage(1)
	assertPage(t, page, []string{"c", "b"}, "second reverse page")

	// Last page (partial)
	page = reversePager.GetPage(2)
	assertPage(t, page, []string{"a"}, "last reverse page")
}

func TestUListWrapperWithDeletedItems(t *testing.T) {
	// Create a list with some deleted items
	list := ulist.New()
	list.Append("a", "b", "c", "d", "e")
	list.Delete(1, 3) // Delete "b" and "d"

	pager := NewPager(list, 2, false)

	// First page should skip deleted items
	page := pager.GetPage(0)
	assertPage(t, page, []string{"a", "c"}, "first page with deleted items")

	// Second page
	page = pager.GetPage(1)
	assertPage(t, page, []string{"e"}, "second page with deleted items")
}

// Helper function to assert page contents
func assertPage(t *testing.T, got *pageable.Page, want []string, desc string) {
	if len(got.Items) != len(want) {
		t.Errorf("%s: expected length %d, got %d", desc, len(want), len(got.Items))
		return
	}

	for i := 0; i < len(want); i++ {
		if got.Items[i].Value.(string) != want[i] {
			t.Errorf("%s: at index %d expected %q, got %q", desc, i, want[i], got.Items[i])
		}
	}
}
