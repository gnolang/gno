package main

import (
	"gno.land/p/moul/ulist"
	"gno.land/p/moul/ulist/pager"
	"gno.land/p/nt/ufmt"
)

func main() {
	// Create a new list and populate it with some items
	list := ulist.New()
	for i := 0; i < 25; i++ {
		list.Append(ufmt.Sprintf("item_%02d", i))
	}

	// Delete some items to demonstrate gap handling
	list.Delete(3)  // Delete "item_03"
	list.Delete(7)  // Delete "item_07"
	list.Delete(11) // Delete "item_11"

	// Create a new pager with page size 5
	pager := pager.NewPager(list, 5, false)

	println("## Forward Pagination")
	for pn := 1; pn <= 6; pn++ {
		page := pager.GetPage(pn)

		println(ufmt.Sprintf("### Page %d of %d", page.PageNumber, page.TotalPages))
		for idx, item := range page.Items {
			println(ufmt.Sprintf("- [%d] index=%d value=%s", idx, item.Index, item.Value))
		}
		if picker := page.Picker(); picker != "" {
			println("Navigation:", picker)
		}
		println()
	}

	println("---")

	// Create a new pager in reverse mode
	pager = pager.NewPager(list, 5, true)

	println("## Reverse Pagination")
	for pn := 1; pn <= 3; pn++ {
		page := pager.GetPage(pn)

		println(ufmt.Sprintf("### Page %d of %d (reversed)", page.PageNumber, page.TotalPages))
		for idx, item := range page.Items {
			println(ufmt.Sprintf("- [%d] index=%d value=%s", idx, item.Index, item.Value))
		}
		if picker := page.Picker(); picker != "" {
			println("Navigation:", picker)
		}
		println()
	}
}

// Output:
// ## Forward Pagination
// ### Page 1 of 5
// - [0] index=0 value=item_00
// - [1] index=1 value=item_01
// - [2] index=2 value=item_02
// - [3] index=4 value=item_04
// - [4] index=5 value=item_05
// Navigation: **1** | [2](?page=2) | [3](?page=3) | [4](?page=4) | [5](?page=5)
//
// ### Page 2 of 5
// - [0] index=6 value=item_06
// - [1] index=8 value=item_08
// - [2] index=9 value=item_09
// - [3] index=10 value=item_10
// - [4] index=12 value=item_12
// Navigation: [1](?page=1) | **2** | [3](?page=3) | [4](?page=4) | [5](?page=5)
//
// ### Page 3 of 5
// - [0] index=13 value=item_13
// - [1] index=14 value=item_14
// - [2] index=15 value=item_15
// - [3] index=16 value=item_16
// - [4] index=17 value=item_17
// Navigation: [1](?page=1) | [2](?page=2) | **3** | [4](?page=4) | [5](?page=5)
//
// ### Page 4 of 5
// - [0] index=18 value=item_18
// - [1] index=19 value=item_19
// - [2] index=20 value=item_20
// - [3] index=21 value=item_21
// - [4] index=22 value=item_22
// Navigation: [1](?page=1) | [2](?page=2) | [3](?page=3) | **4** | [5](?page=5)
//
// ### Page 5 of 5
// - [0] index=23 value=item_23
// - [1] index=24 value=item_24
// Navigation: [1](?page=1) | & | [3](?page=3) | [4](?page=4) | **5**
//
// ### Page 6 of 5
// Navigation: [1](?page=1) | & | [4](?page=4) | [5](?page=5) | _6_
//
// ---
// ## Reverse Pagination
// ### Page 1 of 5 (reversed)
// - [0] index=24 value=item_24
// - [1] index=23 value=item_23
// - [2] index=22 value=item_22
// - [3] index=21 value=item_21
// - [4] index=20 value=item_20
// Navigation: **1** | [2](?page=2) | [3](?page=3) | [4](?page=4) | [5](?page=5)
//
// ### Page 2 of 5 (reversed)
// - [0] index=19 value=item_19
// - [1] index=18 value=item_18
// - [2] index=17 value=item_17
// - [3] index=16 value=item_16
// - [4] index=15 value=item_15
// Navigation: [1](?page=1) | **2** | [3](?page=3) | [4](?page=4) | [5](?page=5)
//
// ### Page 3 of 5 (reversed)
// - [0] index=14 value=item_14
// - [1] index=13 value=item_13
// - [2] index=12 value=item_12
// - [3] index=10 value=item_10
// - [4] index=9 value=item_09
// Navigation: [1](?page=1) | [2](?page=2) | **3** | [4](?page=4) | [5](?page=5)
//
