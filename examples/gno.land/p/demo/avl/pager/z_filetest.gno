package main

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

func main() {
	// Create a new AVL tree and populate it with some key-value pairs.
	var id seqid.ID
	tree := avl.NewTree()
	for i := 0; i < 42; i++ {
		tree.Set(id.Next().String(), i)
	}

	// Create a new pager.
	pager := pager.NewPager(tree)

	for pn := -1; pn < 8; pn++ {
		page, err := pager.GetPage(pn, 7)
		if err != nil {
			println("Error:", err.Error())
			return
		}

		// Print the items on the page.
		println(ufmt.Sprintf("Page %d of %d", page.PageNumber, page.TotalPages))
		for _, item := range page.Items {
			println(ufmt.Sprintf("Key: %s, Value: %v", item.Key, item.Value))
		}

		// Render and print the page selector.
		pageSelector := page.RenderSelector(pager.PageQueryParam)
		println("Page Selector:")
		println(pageSelector)
		println()
	}
}

// Output:
// Page 0 of 6
// Page Selector:
// **0** | [1](?page=1) | [2](?page=2) | … | [6](?page=6)
//
// Page 0 of 6
// Page Selector:
// **0** | [1](?page=1) | [2](?page=2) | … | [6](?page=6)
//
// Page 1 of 6
// Key: 0000001, Value: (unhandled)
// Key: 0000002, Value: (unhandled)
// Key: 0000003, Value: (unhandled)
// Key: 0000004, Value: (unhandled)
// Key: 0000005, Value: (unhandled)
// Key: 0000006, Value: (unhandled)
// Key: 0000007, Value: (unhandled)
// Page Selector:
// **1** | [2](?page=2) | [3](?page=3) | … | [6](?page=6)
//
// Page 2 of 6
// Key: 0000008, Value: (unhandled)
// Key: 0000009, Value: (unhandled)
// Key: 000000a, Value: (unhandled)
// Key: 000000b, Value: (unhandled)
// Key: 000000c, Value: (unhandled)
// Key: 000000d, Value: (unhandled)
// Key: 000000e, Value: (unhandled)
// Page Selector:
// [1](?page=1) | **2** | [3](?page=3) | [4](?page=4) | [5](?page=5) | [6](?page=6)
//
// Page 3 of 6
// Key: 000000f, Value: (unhandled)
// Key: 000000g, Value: (unhandled)
// Key: 000000h, Value: (unhandled)
// Key: 000000j, Value: (unhandled)
// Key: 000000k, Value: (unhandled)
// Key: 000000m, Value: (unhandled)
// Key: 000000n, Value: (unhandled)
// Page Selector:
// [1](?page=1) | [2](?page=2) | **3** | [4](?page=4) | [5](?page=5) | [6](?page=6)
//
// Page 4 of 6
// Key: 000000p, Value: (unhandled)
// Key: 000000q, Value: (unhandled)
// Key: 000000r, Value: (unhandled)
// Key: 000000s, Value: (unhandled)
// Key: 000000t, Value: (unhandled)
// Key: 000000v, Value: (unhandled)
// Key: 000000w, Value: (unhandled)
// Page Selector:
// [1](?page=1) | [2](?page=2) | [3](?page=3) | **4** | [5](?page=5) | [6](?page=6)
//
// Page 5 of 6
// Key: 000000x, Value: (unhandled)
// Key: 000000y, Value: (unhandled)
// Key: 000000z, Value: (unhandled)
// Key: 0000010, Value: (unhandled)
// Key: 0000011, Value: (unhandled)
// Key: 0000012, Value: (unhandled)
// Key: 0000013, Value: (unhandled)
// Page Selector:
// [1](?page=1) | [2](?page=2) | [3](?page=3) | [4](?page=4) | **5** | [6](?page=6)
//
// Page 6 of 6
// Key: 0000014, Value: (unhandled)
// Key: 0000015, Value: (unhandled)
// Key: 0000016, Value: (unhandled)
// Key: 0000017, Value: (unhandled)
// Key: 0000018, Value: (unhandled)
// Key: 0000019, Value: (unhandled)
// Key: 000001a, Value: (unhandled)
// Page Selector:
// [1](?page=1) | … | [4](?page=4) | [5](?page=5) | **6**
//
// Page 7 of 6
// Page Selector:
// [1](?page=1) | … | [5](?page=5) | [6](?page=6) | **7**
