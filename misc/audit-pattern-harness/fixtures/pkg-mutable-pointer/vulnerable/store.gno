package store

import "gno.land/p/nt/avl/v0"

var store = avl.NewTree()

// Vulnerable (guide §5.1a): the getter returns a pointer to a /p/ type whose
// exported methods (Set, Remove) mutate the receiver. All of avl.Tree's fields
// are unexported, so a naive reviewer sees no exposed state — but readonly
// taint does not block method dispatch, so GetStore().Set(k, v) commits under
// this realm's authority.
func GetStore() *avl.Tree {
	return store
}

// Vulnerable (guide §5.1b): the same path one level deeper. Container has an
// exported pointer field to the mutable /p/ type, so returning *Container hands
// out an indirect mutator via GetContainer().Items.Set(k, v).
type Container struct {
	Items *avl.Tree
	Label string
}

var container = &Container{Items: avl.NewTree(), Label: "box"}

func GetContainer() *Container {
	return container
}
