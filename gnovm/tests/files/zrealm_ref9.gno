// PKGPATH: gno.land/r/test
package test

// Realm filetest: direct pointer-receiver method call on
// package-level value type (no closure). Tests that auto-address
// (&gTree).Set() works after realm persistence.

type Tree struct {
	size int
}

func (t *Tree) Set(k string, v int) {
	t.size += v
}

func (t *Tree) Size() int {
	return t.size
}

var gTree Tree

func init() {
	gTree.Set("a", 1)
	gTree.Set("b", 2)
}

func main(cur realm) {
	gTree.Set("c", 3)
	println(gTree.Size())
}

// Output:
// 6
