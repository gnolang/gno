// PKGPATH: gno.land/r/main
package main

import (
	"gno.land/p/nt/avl"
)

func main(cur realm) {
	node := avl.NewNode("key0", "value0")
	println(node.Size())
}

// Output:
// 1
