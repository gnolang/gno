package pager

import (
	"testing"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/ufmt"
	"gno.land/p/nt/urequire"
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
			expected   []Item
		}{
			{1, 2, []Item{{Key: "a", Value: 1}, {Key: "b", Value: 2}}},
			{2, 2, []Item{{Key: "c", Value: 3}, {Key: "d", Value: 4}}},
			{3, 2, []Item{{Key: "e", Value: 5}}},
			{1, 3, []Item{{Key: "a", Value: 1}, {Key: "b", Value: 2}, {Key: "c", Value: 3}}},
			{2, 3, []Item{{Key: "d", Value: 4}, {Key: "e", Value: 5}}},
			{1, 5, []Item{{Key: "a", Value: 1}, {Key: "b", Value: 2}, {Key: "c", Value: 3}, {Key: "d", Value: 4}, {Key: "e", Value: 5}}},
			{2, 5, []Item{}},
		}

		for _, tt := range tests {
			page := pager.GetPageWithSize(tt.pageNumber, tt.pageSize)

			uassert.Equal(t, len(tt.expected), len(page.Items))

			for i, item := range page.Items {
				uassert.Equal(t, tt.expected[i].Key, item.Key)
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
			expected   []Item
		}{
			{1, 2, []Item{{Key: "e", Value: 5}, {Key: "d", Value: 4}}},
			{2, 2, []Item{{Key: "c", Value: 3}, {Key: "b", Value: 2}}},
			{3, 2, []Item{{Key: "a", Value: 1}}},
			{1, 3, []Item{{Key: "e", Value: 5}, {Key: "d", Value: 4}, {Key: "c", Value: 3}}},
			{2, 3, []Item{{Key: "b", Value: 2}, {Key: "a", Value: 1}}},
			{1, 5, []Item{{Key: "e", Value: 5}, {Key: "d", Value: 4}, {Key: "c", Value: 3}, {Key: "b", Value: 2}, {Key: "a", Value: 1}}},
			{2, 5, []Item{}},
		}

		for _, tt := range tests {
			page := pager.GetPageWithSize(tt.pageNumber, tt.pageSize)

			uassert.Equal(t, len(tt.expected), len(page.Items))

			for i, item := range page.Items {
				uassert.Equal(t, tt.expected[i].Key, item.Key)
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

func TestPage_PickerQueryParamPreservation(t *testing.T) {
	tree := avl.NewTree()
	for i := 1; i <= 6; i++ {
		tree.Set(ufmt.Sprintf("key%d", i), i)
	}

	pager := NewPager(tree, 2, false)

	tests := []struct {
		name       string
		pageNumber int
		path       string
		expected   string
	}{
		{
			name:       "single query param",
			pageNumber: 1,
			path:       "/test?foo=bar",
			expected:   "**1** | [2](?page=2&foo=bar) | [3](?page=3&foo=bar)",
		},
		{
			name:       "multiple query params",
			pageNumber: 2,
			path:       "/test?foo=bar&baz=qux",
			expected:   "[1](?page=1&baz=qux&foo=bar) | **2** | [3](?page=3&baz=qux&foo=bar)",
		},
		{
			name:       "overwrite existing page param",
			pageNumber: 1,
			path:       "/test?param1=value1&page=999&param2=value2",
			expected:   "**1** | [2](?page=2&param1=value1&param2=value2) | [3](?page=3&param1=value1&param2=value2)",
		},
		{
			name:       "empty query string",
			pageNumber: 2,
			path:       "/test",
			expected:   "[1](?page=1) | **2** | [3](?page=3)",
		},
		{
			name:       "query string with only page param",
			pageNumber: 2,
			path:       "/test?page=2",
			expected:   "[1](?page=1) | **2** | [3](?page=3)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := pager.GetPageWithSize(tt.pageNumber, 2)
			result := page.Picker(tt.path)
			if result != tt.expected {
				t.Errorf("\nwant: %s\ngot:  %s", tt.expected, result)
			}
		})
	}
}
