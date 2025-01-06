// PKGPATH: gno.land/r/test
package test

import (
	"github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl"
	"github.com/gnolang/gno/_test/timtadh/data_structures/types"
)

var tree *avl.AvlNode

func init() {
	tree, _ = tree.Put(types.String("key0"), "value0")
	tree, _ = tree.Put(types.String("key1"), "value1")
}

func main() {
	var updated bool
	tree, updated = tree.Put(types.String("key3"), "value3")
	println(updated, tree.Size())
}

// Error:
// cannot attach un-real object from external realm: struct{("key0" github.com/gnolang/gno/_test/timtadh/data_structures/types.String),("value0" string),(2 int),(nil *github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode),(&(struct{("key1" github.com/gnolang/gno/_test/timtadh/data_structures/types.String),("value1" string),(1 int),(nil *github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode),(nil *github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode)} github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode) *github.com/gnolang/gno/_test/timtadh/data_structures/tree/avl.AvlNode)}
