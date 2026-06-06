package main

import (
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/ufmt/v0"
)

func main() {
	// Create a new AVL tree and populate it with some key-value pairs.
	var id seqid.ID
	tree := avl.NewTree()
	for i := 0; i < 42; i++ {
		tree.Set(id.Next().String(), i)
	}

	// Create a new pager.
	pager := pager.NewPager(tree, 7, false)

	for pn := -1; pn < 8; pn++ {
		page := pager.GetPage(pn)

		println(ufmt.Sprintf("## Page %d of %d", page.PageNumber, page.TotalPages))
		for idx, item := range page.Items {
			println(ufmt.Sprintf("- idx=%d key=%s value=%d", idx, item.Key, item.Value))
		}
		println(page.Picker("/"))
		println()
	}
}

// Output:
// ## Page 0 of 6
// _0_ | [1](?page=1) | [2](?page=2) | … | [6](?page=6)
//
// ## Page 0 of 6
// _0_ | [1](?page=1) | [2](?page=2) | … | [6](?page=6)
//
// ## Page 1 of 6
// - idx=0 key=0000001 value=0
// - idx=1 key=0000002 value=1
// - idx=2 key=0000003 value=2
// - idx=3 key=0000004 value=3
// - idx=4 key=0000005 value=4
// - idx=5 key=0000006 value=5
// - idx=6 key=0000007 value=6
// **1** | [2](?page=2) | [3](?page=3) | … | [6](?page=6)
//
// ## Page 2 of 6
// - idx=0 key=0000008 value=7
// - idx=1 key=0000009 value=8
// - idx=2 key=000000a value=9
// - idx=3 key=000000b value=10
// - idx=4 key=000000c value=11
// - idx=5 key=000000d value=12
// - idx=6 key=000000e value=13
// [1](?page=1) | **2** | [3](?page=3) | [4](?page=4) | … | [6](?page=6)
//
// ## Page 3 of 6
// - idx=0 key=000000f value=14
// - idx=1 key=000000g value=15
// - idx=2 key=000000h value=16
// - idx=3 key=000000j value=17
// - idx=4 key=000000k value=18
// - idx=5 key=000000m value=19
// - idx=6 key=000000n value=20
// [1](?page=1) | [2](?page=2) | **3** | [4](?page=4) | [5](?page=5) | [6](?page=6)
//
// ## Page 4 of 6
// - idx=0 key=000000p value=21
// - idx=1 key=000000q value=22
// - idx=2 key=000000r value=23
// - idx=3 key=000000s value=24
// - idx=4 key=000000t value=25
// - idx=5 key=000000v value=26
// - idx=6 key=000000w value=27
// [1](?page=1) | [2](?page=2) | [3](?page=3) | **4** | [5](?page=5) | [6](?page=6)
//
// ## Page 5 of 6
// - idx=0 key=000000x value=28
// - idx=1 key=000000y value=29
// - idx=2 key=000000z value=30
// - idx=3 key=0000010 value=31
// - idx=4 key=0000011 value=32
// - idx=5 key=0000012 value=33
// - idx=6 key=0000013 value=34
// [1](?page=1) | … | [3](?page=3) | [4](?page=4) | **5** | [6](?page=6)
//
// ## Page 6 of 6
// - idx=0 key=0000014 value=35
// - idx=1 key=0000015 value=36
// - idx=2 key=0000016 value=37
// - idx=3 key=0000017 value=38
// - idx=4 key=0000018 value=39
// - idx=5 key=0000019 value=40
// - idx=6 key=000001a value=41
// [1](?page=1) | … | [4](?page=4) | [5](?page=5) | **6**
//
// ## Page 7 of 6
// [1](?page=1) | … | [5](?page=5) | [6](?page=6) | _7_
//
