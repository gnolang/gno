package main

import (
	"github.com/gnolang/gno/_test/timtadh/data-structures/tree/avl"
	"github.com/gnolang/gno/_test/timtadh/data-structures/types"
)

func main() {
	var tree *avl.AvlNode
	var updated bool
	tree, updated = tree.Put(types.String("key0"), "value0")
	println(updated, tree.Size())
}

// Output:
// false 1
