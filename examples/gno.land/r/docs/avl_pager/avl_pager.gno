package avl_pager

import (
	"strconv"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
)

// Tree instance for 100 items
var tree *avl.Tree

// Initialize a tree with 100 items.
func init() {
	tree = avl.NewTree()
	for i := 1; i <= 100; i++ {
		key := "Item" + strconv.Itoa(i)
		tree.Set(key, "Value of "+key)
	}
}

// Render paginated content based on the given URL path.
// URL format: `...?page=<page>&size=<size>` (default is page 1 and size 10).
func Render(path string) string {
	p := pager.NewPager(tree, 10, false) // Default page size is 10
	page := p.MustGetPageByPath(path)

	// Header and pagination info
	result := "# Paginated Items\n"
	result += "Page " + strconv.Itoa(page.PageNumber) + " of " + strconv.Itoa(page.TotalPages) + "\n\n"
	result += page.Picker(path) + "\n\n"

	// Display items on the current page
	for _, item := range page.Items {
		result += "- " + item.Key + ": " + item.Value.(string) + "\n"
	}

	result += "\n" + page.Picker(path) // Repeat page picker for ease of navigation
	return result
}
