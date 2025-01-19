package pager

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/moul/pageable"
	"gno.land/p/moul/ulist"
)

func TestPager_GetPage(t *testing.T) {
	// Create a new list and populate it with values
	list := ulist.New()
	list.Append("a", "b", "c", "d", "e")

	t.Run("normal ordering", func(t *testing.T) {
		pager := NewPager(list, 10, false)

		tests := []struct {
			pageNumber int
			pageSize   int
			expected   []pageable.Item
		}{
			{1, 2, []pageable.Item{{Index: 0, Value: "a"}, {Index: 1, Value: "b"}}},
			{2, 2, []pageable.Item{{Index: 2, Value: "c"}, {Index: 3, Value: "d"}}},
			{3, 2, []pageable.Item{{Index: 4, Value: "e"}}},
			{1, 3, []pageable.Item{{Index: 0, Value: "a"}, {Index: 1, Value: "b"}, {Index: 2, Value: "c"}}},
			{2, 3, []pageable.Item{{Index: 3, Value: "d"}, {Index: 4, Value: "e"}}},
			{1, 5, []pageable.Item{{Index: 0, Value: "a"}, {Index: 1, Value: "b"}, {Index: 2, Value: "c"}, {Index: 3, Value: "d"}, {Index: 4, Value: "e"}}},
			{2, 5, []pageable.Item{}},
		}

		for _, tt := range tests {
			page := pager.GetPageWithSize(tt.pageNumber, tt.pageSize)
			uassert.Equal(t, len(tt.expected), len(page.Items))
			for i, item := range page.Items {
				uassert.Equal(t, tt.expected[i].Index, item.Index)
				uassert.Equal(t, tt.expected[i].Value, item.Value)
			}
		}
	})

	t.Run("reversed ordering", func(t *testing.T) {
		pager := NewPager(list, 10, true)

		tests := []struct {
			pageNumber int
			pageSize   int
			expected   []pageable.Item
		}{
			{1, 2, []pageable.Item{{Index: 4, Value: "e"}, {Index: 3, Value: "d"}}},
			{2, 2, []pageable.Item{{Index: 2, Value: "c"}, {Index: 1, Value: "b"}}},
			{3, 2, []pageable.Item{{Index: 0, Value: "a"}}},
			{1, 3, []pageable.Item{{Index: 4, Value: "e"}, {Index: 3, Value: "d"}, {Index: 2, Value: "c"}}},
			{2, 3, []pageable.Item{{Index: 1, Value: "b"}, {Index: 0, Value: "a"}}},
			{1, 5, []pageable.Item{{Index: 4, Value: "e"}, {Index: 3, Value: "d"}, {Index: 2, Value: "c"}, {Index: 1, Value: "b"}, {Index: 0, Value: "a"}}},
			{2, 5, []pageable.Item{}},
		}

		for _, tt := range tests {
			page := pager.GetPageWithSize(tt.pageNumber, tt.pageSize)
			uassert.Equal(t, len(tt.expected), len(page.Items))
			for i, item := range page.Items {
				uassert.Equal(t, tt.expected[i].Index, item.Index)
				uassert.Equal(t, tt.expected[i].Value, item.Value)
			}
		}
	})
}

func TestUListWrapperWithDeletedItems(t *testing.T) {
	// Create a list with some deleted items
	list := ulist.New()
	list.Append("a", "b", "c", "d", "e")
	list.Delete(1) // Delete "b"
	list.Delete(3) // Delete "d"

	pager := NewPager(list, 2, false)

	// First page should contain non-deleted items
	page := pager.GetPage(1)
	assertPage(t, page, []string{"a", "c"}, "first page with deleted items")

	// Second page should contain remaining non-deleted item
	page = pager.GetPage(2)
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
			t.Errorf("%s: at index %d expected %q, got %q", desc, i, want[i], got.Items[i].Value)
		}
	}
}
