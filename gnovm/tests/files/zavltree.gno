package main

import (
	"filetests/extern/timtadh/data_structures/tree/avl"
	"filetests/extern/timtadh/data_structures/types"
)

func main() {
	var tree *avl.AvlNode
	var updated bool
	tree, updated = tree.Put(types.String("key0"), "value0")
	println(updated, tree.Size())
}

// Output:
// false 1
