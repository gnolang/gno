package pager

import (
	"testing"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/ufmt"
	"gno.land/p/nt/urequire"
	"gno.land/p/moul/pageable"
)

func TestPager_GetPage(t *testing.T) {
	// Create a new AVL tree and populate it with some key-value pairs.
	tree := avl.NewTree()
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)
	tree.Set("d", 4)
	tree.Set("e", 5)

	t.Run("normal ordering", func(t *testing.T) {
		// Create a new pager.
		pager := NewPager(tree, 10, false)

		// Define test cases.
		tests := []struct {
			pageNumber int
			pageSize   int
			expected   []pageable.Item
		}{
			{1, 2, []pageable.Item{{Index: "a", Value: 1}, {Index: "b", Value: 2}}},
			{2, 2, []pageable.Item{{Index: "c", Value: 3}, {Index: "d", Value: 4}}},
			{3, 2, []pageable.Item{{Index: "e", Value: 5}}},
			{1, 3, []pageable.Item{{Index: "a", Value: 1}, {Index: "b", Value: 2}, {Index: "c", Value: 3}}},
			{2, 3, []pageable.Item{{Index: "d", Value: 4}, {Index: "e", Value: 5}}},
			{1, 5, []pageable.Item{{Index: "a", Value: 1}, {Index: "b", Value: 2}, {Index: "c", Value: 3}, {Index: "d", Value: 4}, {Index: "e", Value: 5}}},
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
		// Create a new pager.
		pager := NewPager(tree, 10, true)

		// Define test cases.
		tests := []struct {
			pageNumber int
			pageSize   int
			expected   []pageable.Item
		}{
			{1, 2, []pageable.Item{{Index: "e", Value: 5}, {Index: "d", Value: 4}}},
			{2, 2, []pageable.Item{{Index: "c", Value: 3}, {Index: "b", Value: 2}}},
			{3, 2, []pageable.Item{{Index: "a", Value: 1}}},
			{1, 3, []pageable.Item{{Index: "e", Value: 5}, {Index: "d", Value: 4}, {Index: "c", Value: 3}}},
			{2, 3, []pageable.Item{{Index: "b", Value: 2}, {Index: "a", Value: 1}}},
			{1, 5, []pageable.Item{{Index: "e", Value: 5}, {Index: "d", Value: 4}, {Index: "c", Value: 3}, {Index: "b", Value: 2}, {Index: "a", Value: 1}}},
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

func TestPager_GetPageByPath(t *testing.T) {
	// Create a new AVL tree and populate it with some key-value pairs.
	tree := avl.NewTree()
	for i := 0; i < 50; i++ {
		tree.Set(ufmt.Sprintf("key%d", i), i)
	}

	// Create a new pager.
	pager := NewPager(tree, 10, false)

	// Define test cases.
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
	// Create a new AVL tree and populate it with some key-value pairs.
	tree := avl.NewTree()
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)
	tree.Set("d", 4)
	tree.Set("e", 5)

	// Create a new pager.
	pager := NewPager(tree, 10, false)

	// Define test cases.
	tests := []struct {
		pageNumber int
		pageSize   int
		path       string
		expected   string
	}{
		{1, 2, "/test", "**1** | [2](?page=2) | [3](?page=3)"},
		{2, 2, "/test", "[1](?page=1) | **2** | [3](?page=3)"},
		{3, 2, "/test", "[1](?page=1) | [2](?page=2) | **3**"},
		{1, 2, "/test?foo=bar", "**1** | [2](?page=2&foo=bar) | [3](?page=3&foo=bar)"},
	}

	for _, tt := range tests {
		page := pager.GetPageWithSize(tt.pageNumber, tt.pageSize)

		ui := page.Picker(tt.path)
		uassert.Equal(t, tt.expected, ui)
	}
}

func TestPager_UI_WithManyPages(t *testing.T) {
	// Create a new AVL tree and populate it with many key-value pairs.
	tree := avl.NewTree()
	for i := 0; i < 100; i++ {
		tree.Set(ufmt.Sprintf("key%d", i), i)
	}

	// Create a new pager.
	pager := NewPager(tree, 10, false)

	// Define test cases for a large number of pages.
	tests := []struct {
		pageNumber int
		pageSize   int
		path       string
		expected   string
	}{
		{1, 10, "/test", "**1** | [2](?page=2) | [3](?page=3) | … | [10](?page=10)"},
		{2, 10, "/test", "[1](?page=1) | **2** | [3](?page=3) | [4](?page=4) | … | [10](?page=10)"},
		{3, 10, "/test", "[1](?page=1) | [2](?page=2) | **3** | [4](?page=4) | [5](?page=5) | … | [10](?page=10)"},
		{4, 10, "/test", "[1](?page=1) | [2](?page=2) | [3](?page=3) | **4** | [5](?page=5) | [6](?page=6) | … | [10](?page=10)"},
		{5, 10, "/test", "[1](?page=1) | … | [3](?page=3) | [4](?page=4) | **5** | [6](?page=6) | [7](?page=7) | … | [10](?page=10)"},
		{6, 10, "/test", "[1](?page=1) | … | [4](?page=4) | [5](?page=5) | **6** | [7](?page=7) | [8](?page=8) | … | [10](?page=10)"},
		{7, 10, "/test", "[1](?page=1) | … | [5](?page=5) | [6](?page=6) | **7** | [8](?page=8) | [9](?page=9) | [10](?page=10)"},
		{8, 10, "/test", "[1](?page=1) | … | [6](?page=6) | [7](?page=7) | **8** | [9](?page=9) | [10](?page=10)"},
		{9, 10, "/test", "[1](?page=1) | … | [7](?page=7) | [8](?page=8) | **9** | [10](?page=10)"},
		{10, 10, "/test", "[1](?page=1) | … | [8](?page=8) | [9](?page=9) | **10**"},
	}

	for _, tt := range tests {
		page := pager.GetPageWithSize(tt.pageNumber, tt.pageSize)

		ui := page.Picker(tt.path)
		uassert.Equal(t, tt.expected, ui)
	}
}

func TestPager_ParseQuery(t *testing.T) {
	// Create a new AVL tree and populate it with some key-value pairs.
	tree := avl.NewTree()
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)
	tree.Set("d", 4)
	tree.Set("e", 5)

	// Create a new pager.
	pager := NewPager(tree, 10, false)

	// Define test cases.
	tests := []struct {
		rawURL        string
		expectedPage  int
		expectedSize  int
		expectedError bool
	}{
		{"/r/foo:bar/baz?size=2&page=1", 1, 2, false},
		{"/r/foo:bar/baz?size=3&page=2", 2, 3, false},
		{"/r/foo:bar/baz?size=5&page=3", 3, 5, false},
		{"/r/foo:bar/baz?page=2", 2, pager.DefaultPageSize, false},
		{"/r/foo:bar/baz?size=3", 1, 3, false},
		{"/r/foo:bar/baz", 1, pager.DefaultPageSize, false},
		{"/r/foo:bar/baz?size=0&page=0", 1, pager.DefaultPageSize, false},
	}

	for _, tt := range tests {
		page, size, err := pager.ParseQuery(tt.rawURL)
		if tt.expectedError {
			uassert.Error(t, err, ufmt.Sprintf("ParseQuery(%s) expected error but got none", tt.rawURL))
		} else {
			urequire.NoError(t, err, ufmt.Sprintf("ParseQuery(%s) returned error: %v", tt.rawURL, err))
			uassert.Equal(t, tt.expectedPage, page, ufmt.Sprintf("ParseQuery(%s) returned page %d, expected %d", tt.rawURL, page, tt.expectedPage))
			uassert.Equal(t, tt.expectedSize, size, ufmt.Sprintf("ParseQuery(%s) returned size %d, expected %d", tt.rawURL, size, tt.expectedSize))
		}
	}
}

func TestWrapper_IterateByOffset(t *testing.T) {
	tree := avl.NewTree()
	tree.Set("a", 1)
	tree.Set("b", 2)
	wrapper := NewWrapper(tree)

	t.Run("empty count", func(t *testing.T) {
		result := wrapper.IterateByOffset(0, 0, func(k, v interface{}) bool { return true })
		uassert.Equal(t, false, result)
	})

	t.Run("empty tree", func(t *testing.T) {
		emptyWrapper := NewWrapper(avl.NewTree())
		result := emptyWrapper.IterateByOffset(0, 1, func(k, v interface{}) bool { return true })
		uassert.Equal(t, false, result)
	})

	t.Run("offset beyond size", func(t *testing.T) {
		result := wrapper.IterateByOffset(3, 1, func(k, v interface{}) bool { return true })
		uassert.Equal(t, false, result)
	})
}

func TestPager_ReversedWithDifferentSizes(t *testing.T) {
	tree := avl.NewTree()
	for i := 0; i < 10; i++ {
		tree.Set(ufmt.Sprintf("key%d", i), i)
	}

	pager := NewPager(tree, 3, true) // Default page size of 3, reversed order

	tests := []struct {
		pageNumber int
		pageSize   int
		wantFirst  string
		wantLast   string
	}{
		{1, 4, "key9", "key6"},
		{2, 4, "key5", "key2"},
		{3, 4, "key1", "key0"},
		{4, 4, "", ""}, // Empty page
	}

	for _, tt := range tests {
		page := pager.GetPageWithSize(tt.pageNumber, tt.pageSize)
		if len(page.Items) > 0 {
			uassert.Equal(t, tt.wantFirst, page.Items[0].Index)
			uassert.Equal(t, tt.wantLast, page.Items[len(page.Items)-1].Index)
		} else {
			uassert.Equal(t, 0, len(page.Items))
		}
	}
}
