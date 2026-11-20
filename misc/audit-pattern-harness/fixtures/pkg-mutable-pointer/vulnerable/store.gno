package store

import "gno.land/p/nt/avl/v0"

var store = avl.NewTree()

// Vulnerable (guide §5.1a): the getter returns a pointer to a /p/ type whose
// exported methods (Set, Remove) mutate the receiver. All of avl.Tree's fields
// are unexported, so the pointer still exposes the type's mutator API; the VM
// refuses a foreign caller's write to this realm's stored tree.
func GetStore() *avl.Tree {
	return store
}

// Vulnerable (guide §5.1b): the same path one level deeper. Container has an
// exported pointer field to the mutable /p/ type, so returning *Container
// exposes an indirect mutator via GetContainer().Items.Set(k, v).
type Container struct {
	Items *avl.Tree
	Label string
}

var container = &Container{Items: avl.NewTree(), Label: "box"}

func GetContainer() *Container {
	return container
}
