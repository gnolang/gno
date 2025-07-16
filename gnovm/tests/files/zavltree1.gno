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
	tree, updated = tree.Put(types.String("key0"), "value0-new")
	println(updated, tree.Size())
	tree, updated = tree.Put(types.String("key1"), "value1")
	println(updated, tree.Size())
	tree, updated = tree.Put(types.String("key3"), "value3")
	println(updated, tree.Size())
	tree, updated = tree.Put(types.String("key2"), "value2")
	println(updated, tree.Size())
}

// Output:
// false 1
// true 1
// false 2
// false 3
// false 4
