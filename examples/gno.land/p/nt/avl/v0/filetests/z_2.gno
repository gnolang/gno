// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/nt/avl/v0"
)

var tree avl.Tree

func init() {
	tree.Set("key0", "value0")
	tree.Set("key1", "value1")
}

func main(cur realm) {
	var updated bool
	updated = tree.Set("key2", "value2")
	println(updated, tree.Size())
}

// Output:
// false 3
