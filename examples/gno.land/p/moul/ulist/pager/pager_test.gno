package pager

import (
	"testing"

	"gno.land/p/moul/pageable"
	"gno.land/p/moul/ulist"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/ufmt"
	"gno.land/p/nt/urequire"
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

func TestPager_WithDeletedItems(t *testing.T) {
	// Create a list with some deleted items
	list := ulist.New()
	list.Append("a", "b", "c", "d", "e", "f", "g", "h")
	list.Delete(1) // Delete "b"
	list.Delete(3) // Delete "d"
	list.Delete(5) // Delete "f"

	t.Run("normal ordering with gaps", func(t *testing.T) {
		pager := NewPager(list, 2, false)

		// First page should skip deleted items
		page1 := pager.GetPage(1)
		// Note: This might be less than 2 items because deleted items are counted in the offset
		// but skipped during iteration
		uassert.Equal(t, 2, len(page1.Items))
		uassert.Equal(t, 0, page1.Items[0].Index)
		uassert.Equal(t, "a", page1.Items[0].Value)
		uassert.Equal(t, 2, page1.Items[1].Index)
		uassert.Equal(t, "c", page1.Items[1].Value)

		// Second page
		page2 := pager.GetPage(2)
		// This page starts at offset 2, which includes indices 2 and 3, but 3 is deleted
		// so we only get one item from indices 2-3, then continue to 4
		uassert.Equal(t, 2, len(page2.Items))
		uassert.Equal(t, 4, page2.Items[0].Index)
		uassert.Equal(t, "e", page2.Items[0].Value)
		uassert.Equal(t, 6, page2.Items[1].Index)
		uassert.Equal(t, "g", page2.Items[1].Value)
	})

	t.Run("reversed ordering with gaps", func(t *testing.T) {
		pager := NewPager(list, 2, true)

		// First page in reverse should start from the end
		page1 := pager.GetPage(1)
		uassert.Equal(t, 2, len(page1.Items))
		uassert.Equal(t, 7, page1.Items[0].Index)
		uassert.Equal(t, "h", page1.Items[0].Value)
		uassert.Equal(t, 6, page1.Items[1].Index)
		uassert.Equal(t, "g", page1.Items[1].Value)
	})
}

func TestPager_GetPageByPath(t *testing.T) {
	// Create a list with many items
	list := ulist.New()
	for i := 0; i < 50; i++ {
		list.Append(ufmt.Sprintf("item%d", i))
	}

	pager := NewPager(list, 10, false)

	// Test query parsing
	tests := []struct {
		rawURL       string
		expectedPage int
		expectedSize int
	}{
		{"/r/foo:bar/baz?size=10&page=1", 1, 10},
		{"/r/foo:bar/baz?size=10&page=2", 2, 10},
		{"/r/foo:bar/baz?page=3", 3, pager.DefaultPageSize},
		{"/r/foo:bar/baz?size=20", 1, 20},
		{"/r/foo:bar/baz", 1, pager.DefaultPageSize},
	}

	for _, tt := range tests {
		page, err := pager.GetPageByPath(tt.rawURL)
		urequire.NoError(t, err, ufmt.Sprintf("GetPageByPath(%s) returned error: %v", tt.rawURL, err))

		uassert.Equal(t, tt.expectedPage, page.PageNumber)
		uassert.Equal(t, tt.expectedSize, page.PageSize)
	}
}

func TestPage_Picker(t *testing.T) {
	// Create a list
	list := ulist.New()
	list.Append("a", "b", "c", "d", "e")

	pager := NewPager(list, 10, false)

	// Test picker UI generation
	tests := []struct {
		pageNumber int
		pageSize   int
		path       string
		expected   string
	}{
		{1, 2, "", "**1** | [2](?page=2) | [3](?page=3)"},
		{2, 2, "", "[1](?page=1) | **2** | [3](?page=3)"},
		{3, 2, "", "[1](?page=1) | [2](?page=2) | **3**"},
		{1, 2, "/test?foo=bar", "**1** | [2](?page=2&foo=bar) | [3](?page=3&foo=bar)"},
	}

	for _, tt := range tests {
		page := pager.GetPageWithSize(tt.pageNumber, tt.pageSize)
		var ui string
		if tt.path != "" {
			ui = page.PickerWithPath(tt.path)
		} else {
			ui = page.Picker()
		}
		uassert.Equal(t, tt.expected, ui)
	}
}

func TestWrapper_EmptyList(t *testing.T) {
	// Test with empty list
	list := ulist.New()
	wrapper := NewWrapper(list)

	uassert.Equal(t, 0, wrapper.Size())

	// IterateByOffset should return false for empty list
	called := false
	result := wrapper.IterateByOffset(0, 5, func(index, value interface{}) bool {
		called = true
		return false
	})
	uassert.False(t, called, "callback should not be called for empty list")
	uassert.False(t, result)
}

func TestWrapper_EdgeCases(t *testing.T) {
	list := ulist.New()
	list.Append("a", "b", "c")
	wrapper := NewWrapper(list)

	t.Run("zero count", func(t *testing.T) {
		result := wrapper.IterateByOffset(0, 0, func(k, v interface{}) bool { return true })
		uassert.False(t, result)
	})

	t.Run("offset beyond size", func(t *testing.T) {
		result := wrapper.IterateByOffset(5, 1, func(k, v interface{}) bool { return true })
		uassert.False(t, result)
	})

	t.Run("negative count for reverse", func(t *testing.T) {
		var items []interface{}
		wrapper.IterateByOffset(2, -2, func(index, value interface{}) bool {
			items = append(items, value)
			return false
		})
		uassert.Equal(t, 2, len(items))
		uassert.Equal(t, "b", items[0])
		uassert.Equal(t, "a", items[1])
	})
}
