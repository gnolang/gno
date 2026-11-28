package store

import "gno.land/p/nt/avl/v0"

var store = avl.NewTree()

// Fixed (guide §5.1a): keep the tree unexported and expose only value results,
// never a pointer to it. Callers cannot reach Tree.Set/Remove.
func Size() int {
	return store.Size()
}

// Fixed (guide §5.1b): the pointer field is unexported and the struct is never
// returned by pointer; expose a read-only accessor instead.
type container struct {
	items *avl.Tree
	label string
}

var box = &container{items: avl.NewTree(), label: "box"}

func Label() string {
	return box.label
}

func Entries() int {
	return box.items.Size()
}
