package pager

import (
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

func TestPager_GetPage(t *testing.T) {
	// Create a new AVL tree and populate it with some key-value pairs.
	tree := avl.NewTree()
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)
	tree.Set("d", 4)
	tree.Set("e", 5)

	// Create a new pager.
	pager := NewPager(tree)

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
		page, err := pager.GetPage(tt.pageNumber, tt.pageSize)
		urequire.NoError(t, err, ufmt.Sprintf("GetPage(%d, %d) returned error: %v", tt.pageNumber, tt.pageSize, err))

		uassert.Equal(t, len(tt.expected), len(page.Items), ufmt.Sprintf("GetPage(%d, %d) returned %d items, expected %d", tt.pageNumber, tt.pageSize, len(page.Items), len(tt.expected)))

		for i, item := range page.Items {
			uassert.Equal(t, tt.expected[i].Key, item.Key, ufmt.Sprintf("GetPage(%d, %d) returned item %d as %v, expected %v", tt.pageNumber, tt.pageSize, i, item, tt.expected[i]))
			uassert.Equal(t, tt.expected[i].Value, item.Value, ufmt.Sprintf("GetPage(%d, %d) returned item %d as %v, expected %v", tt.pageNumber, tt.pageSize, i, item, tt.expected[i]))
		}
	}
}

func TestPage_RenderSelector(t *testing.T) {
	// Create a new AVL tree and populate it with some key-value pairs.
	tree := avl.NewTree()
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)
	tree.Set("d", 4)
	tree.Set("e", 5)

	// Create a new pager.
	pager := NewPager(tree)

	// Define test cases.
	tests := []struct {
		pageNumber int
		pageSize   int
		expected   string
	}{
		{1, 2, "**1** | [2](?page=2) | [3](?page=3) | … | [Next](?page=2) | [Last](?page=3)"},
		{2, 2, "[1](?page=1) | **2** | [3](?page=3) | … | [Next](?page=3) | [Last](?page=3)"},
		{3, 2, "[1](?page=1) | … | [2](?page=2) | **3**"},
	}

	for _, tt := range tests {
		page, err := pager.GetPage(tt.pageNumber, tt.pageSize)
		urequire.NoError(t, err, ufmt.Sprintf("GetPage(%d, %d) returned error: %v", tt.pageNumber, tt.pageSize, err))

		md := page.RenderSelector(pager.PageQueryParam)
		uassert.Equal(t, tt.expected, md, ufmt.Sprintf("RenderSelector(%d, %d) returned %s, expected %s", tt.pageNumber, tt.pageSize, md, tt.expected))
	}
}

func TestPager_RenderSelector_WithManyPages(t *testing.T) {
	// Create a new AVL tree and populate it with many key-value pairs.
	tree := avl.NewTree()
	for i := 0; i < 100; i++ {
		tree.Set(ufmt.Sprintf("key%d", i), i)
	}

	// Create a new pager.
	pager := NewPager(tree)

	// Define test cases for a large number of pages.
	tests := []struct {
		pageNumber int
		pageSize   int
		expected   string
	}{
		{1, 10, "**1** | [2](?page=2) | [3](?page=3) | … | [Next](?page=2) | [Last](?page=10)"},
		{5, 10, "[1](?page=1) | … | [4](?page=4) | **5** | [6](?page=6) | [7](?page=7) | … | [Next](?page=6) | [Last](?page=10)"},
		{10, 10, "[1](?page=1) | … | [9](?page=9) | **10**"},
	}

	for _, tt := range tests {
		page, err := pager.GetPage(tt.pageNumber, tt.pageSize)
		urequire.NoError(t, err, ufmt.Sprintf("GetPage(%d, %d) returned error: %v", tt.pageNumber, tt.pageSize, err))

		md := page.RenderSelector(pager.PageQueryParam)
		uassert.Equal(t, tt.expected, md, ufmt.Sprintf("RenderSelector(%d, %d) returned %s, expected %s", tt.pageNumber, tt.pageSize, md, tt.expected))
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
	pager := NewPager(tree)

	// Define test cases.
	tests := []struct {
		rawURL        string
		expectedPage  int
		expectedSize  int
		expectedError bool
	}{
		{"https://example.com?size=2&page=1", 1, 2, false},
		{"https://example.com?size=3&page=2", 2, 3, false},
		{"https://example.com?size=5&page=3", 3, 5, false},
		{"https://example.com?page=2", 2, pager.DefaultPageSize, false},
		{"https://example.com?size=3", 1, 3, false},
		{"https://example.com", 1, pager.DefaultPageSize, false},
		{"https://example.com?size=0&page=0", 1, pager.DefaultPageSize, false},
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
