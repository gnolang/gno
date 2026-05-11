// PKGPATH: gno.land/r/test
package test

// Realm filetest: pointer-receiver method call on a package-level
// value type via auto-address (&x).Method(), both direct and
// inside a closure callback. Mirrors the boards2 pattern where
// gFlaggingThresholds.Set() is called inside crossingFn(func(){...}).

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

func withCallback(fn func()) {
	fn()
}

func init() {
	// Direct auto-address: (&gTree).Set(...)
	gTree.Set("a", 1)

	// Auto-address inside closure callback (like crossingFn pattern)
	withCallback(func() {
		gTree.Set("b", 2)
	})
}

func main(cur realm) {
	// Auto-address after realm persistence
	gTree.Set("c", 3)

	// Also inside closure after persistence
	withCallback(func() {
		gTree.Set("d", 4)
	})

	println(gTree.Size())
}

// Output:
// 10
