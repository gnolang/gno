package main

import (
	"fmt"

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
		// tree.Set(fmt.Sprintf("key%d", i), i)
		tree.Set(id.Next().String(), i)
	}

	// Create a new pager.
	pager := pager.Pager{Tree: tree}

	for pn := -1; pn < 8; pn++ {
		page, err := pager.GetPage(pn, 7)
		if err != nil {
			println("Error:", err.Error())
			return
		}

		// Print the items on the first page.
		println(ufmt.Sprintf("Page %d of %d\n", page.PageNumber, page.TotalPages))
		for _, item := range page.Items {
			println(fmt.Sprintf("Key: %s, Value: %v\n", item.Key, item.Value))
		}

		// Render and print the page selector.
		pageSelector := page.RenderSelector()
		println("Page Selector:")
		println(pageSelector)
		println()
	}
}

// Output:
// Page 0 of 6
//
// Page Selector:
// First | Previous | [1](?page=1) | [2](?page=2) | … | [Next](?page=1) | [Last](?page=6)
//
// Page 0 of 6
//
// Page Selector:
// First | Previous | [1](?page=1) | [2](?page=2) | … | [Next](?page=1) | [Last](?page=6)
//
// Page 1 of 6
//
// Key: 0000001, Value: 0
//
// Key: 0000002, Value: 1
//
// Key: 0000003, Value: 2
//
// Key: 0000004, Value: 3
//
// Key: 0000005, Value: 4
//
// Key: 0000006, Value: 5
//
// Key: 0000007, Value: 6
//
// Page Selector:
// First | Previous | **1** | [2](?page=2) | [3](?page=3) | … | [Next](?page=2) | [Last](?page=6)
//
// Page 2 of 6
//
// Key: 0000008, Value: 7
//
// Key: 0000009, Value: 8
//
// Key: 000000a, Value: 9
//
// Key: 000000b, Value: 10
//
// Key: 000000c, Value: 11
//
// Key: 000000d, Value: 12
//
// Key: 000000e, Value: 13
//
// Page Selector:
// [First](?page=1) | [Previous](?page=1) | [1](?page=1) | **2** | [3](?page=3) | [4](?page=4) | … | [Next](?page=3) | [Last](?page=6)
//
// Page 3 of 6
//
// Key: 000000f, Value: 14
//
// Key: 000000g, Value: 15
//
// Key: 000000h, Value: 16
//
// Key: 000000j, Value: 17
//
// Key: 000000k, Value: 18
//
// Key: 000000m, Value: 19
//
// Key: 000000n, Value: 20
//
// Page Selector:
// [First](?page=1) | [Previous](?page=2) | [1](?page=1) | [2](?page=2) | **3** | [4](?page=4) | [5](?page=5) | … | [Next](?page=4) | [Last](?page=6)
//
// Page 4 of 6
//
// Key: 000000p, Value: 21
//
// Key: 000000q, Value: 22
//
// Key: 000000r, Value: 23
//
// Key: 000000s, Value: 24
//
// Key: 000000t, Value: 25
//
// Key: 000000v, Value: 26
//
// Key: 000000w, Value: 27
//
// Page Selector:
// [First](?page=1) | [Previous](?page=3) | … | [2](?page=2) | [3](?page=3) | **4** | [5](?page=5) | [6](?page=6) | [Next](?page=5) | [Last](?page=6)
//
// Page 5 of 6
//
// Key: 000000x, Value: 28
//
// Key: 000000y, Value: 29
//
// Key: 000000z, Value: 30
//
// Key: 0000010, Value: 31
//
// Key: 0000011, Value: 32
//
// Key: 0000012, Value: 33
//
// Key: 0000013, Value: 34
//
// Page Selector:
// [First](?page=1) | [Previous](?page=4) | … | [3](?page=3) | [4](?page=4) | **5** | [6](?page=6) | [Next](?page=6) | [Last](?page=6)
//
// Page 6 of 6
//
// Key: 0000014, Value: 35
//
// Key: 0000015, Value: 36
//
// Key: 0000016, Value: 37
//
// Key: 0000017, Value: 38
//
// Key: 0000018, Value: 39
//
// Key: 0000019, Value: 40
//
// Key: 000001a, Value: 41
//
// Page Selector:
// [First](?page=1) | [Previous](?page=5) | … | [4](?page=4) | [5](?page=5) | **6** | Next | Last
//
// Page 7 of 6
//
// Page Selector:
// [First](?page=1) | [Previous](?page=6) | … | [5](?page=5) | [6](?page=6) | Next | Last
