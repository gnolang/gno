package pagination

import (
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
)

func TestPaginator_GetPage(t *testing.T) {
	// Create a new AVL tree and populate it with some key-value pairs.
	tree := avl.NewTree()
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)
	tree.Set("d", 4)
	tree.Set("e", 5)

	// Create a new paginator.
	paginator := Paginator{Tree: tree}

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
		page, err := paginator.GetPage(tt.pageNumber, tt.pageSize)
		urequire.NoError(t, err, ufmt.Sprintf("GetPage(%d, %d) returned error: %v", tt.pageNumber, tt.pageSize, err))

		uassert.Equal(t, len(tt.expected), len(page.Items), ufmt.Sprintf("GetPage(%d, %d) returned %d items, expected %d", tt.pageNumber, tt.pageSize, len(page.Items), len(tt.expected)))

		for i, item := range page.Items {
			uassert.Equal(t, tt.expected[i].Key, item.Key, ufmt.Sprintf("GetPage(%d, %d) returned item %d as %v, expected %v", tt.pageNumber, tt.pageSize, i, item, tt.expected[i]))
			uassert.Equal(t, tt.expected[i].Value, item.Value, ufmt.Sprintf("GetPage(%d, %d) returned item %d as %v, expected %v", tt.pageNumber, tt.pageSize, i, item, tt.expected[i]))
		}
	}
}

func TestPaginator_RenderPageSelector(t *testing.T) {
	// Create a new AVL tree and populate it with some key-value pairs.
	tree := avl.NewTree()
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)
	tree.Set("d", 4)
	tree.Set("e", 5)

	// Create a new paginator.
	paginator := Paginator{Tree: tree}

	// Define test cases.
	tests := []struct {
		pageNumber int
		pageSize   int
		expected   string
	}{
		{1, 2, "First | Previous | **1** | [2](?page=2) | [3](?page=3) | … | [Next](?page=2) | [Last](?page=3)"},
		{2, 2, "[First](?page=1) | [Previous](?page=1) | [1](?page=1) | **2** | [3](?page=3) | … | [Next](?page=3) | [Last](?page=3)"},
		{3, 2, "[First](?page=1) | [Previous](?page=2) | … | [2](?page=2) | **3** | Next | Last"},
	}

	for _, tt := range tests {
		page, err := paginator.GetPage(tt.pageNumber, tt.pageSize)
		urequire.NoError(t, err, ufmt.Sprintf("GetPage(%d, %d) returned error: %v", tt.pageNumber, tt.pageSize, err))

		md := paginator.RenderPageSelector(page)
		uassert.Equal(t, tt.expected, md, ufmt.Sprintf("RenderPageSelector(%d, %d) returned %s, expected %s", tt.pageNumber, tt.pageSize, md, tt.expected))
	}
}

func TestPaginator_RenderPageSelector_WithManyPages(t *testing.T) {
	// Create a new AVL tree and populate it with many key-value pairs.
	tree := avl.NewTree()
	for i := 0; i < 100; i++ {
		tree.Set(ufmt.Sprintf("key%d", i), i)
	}

	// Create a new paginator.
	paginator := Paginator{Tree: tree}

	// Define test cases for a large number of pages.
	tests := []struct {
		pageNumber int
		pageSize   int
		expected   string
	}{
		{1, 10, "First | Previous | **1** | [2](?page=2) | [3](?page=3) | … | [Next](?page=2) | [Last](?page=10)"},
		{5, 10, "[First](?page=1) | [Previous](?page=4) | … | [3](?page=3) | [4](?page=4) | **5** | [6](?page=6) | [7](?page=7) | … | [Next](?page=6) | [Last](?page=10)"},
		{10, 10, "[First](?page=1) | [Previous](?page=9) | … | [8](?page=8) | [9](?page=9) | **10** | Next | Last"},
	}

	for _, tt := range tests {
		page, err := paginator.GetPage(tt.pageNumber, tt.pageSize)
		urequire.NoError(t, err, ufmt.Sprintf("GetPage(%d, %d) returned error: %v", tt.pageNumber, tt.pageSize, err))

		md := paginator.RenderPageSelector(page)
		uassert.Equal(t, tt.expected, md, ufmt.Sprintf("RenderPageSelector(%d, %d) returned %s, expected %s", tt.pageNumber, tt.pageSize, md, tt.expected))
	}
}
