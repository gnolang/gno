package daokit

import (
	"gno.land/p/nt/avl/v0"
	"gno.land/p/samcrew/daocond"
)

type ResourcesStore struct {
	Tree *avl.Tree // string -> daocond.Condition
}

// XXX: should condition be a pointer?
type Resource struct {
	Handler     ActionHandler
	Condition   daocond.Condition
	DisplayName string
	Description string
}

func NewResourcesStore() *ResourcesStore {
	return &ResourcesStore{
		Tree: avl.NewTree(),
	}
}

func (r *ResourcesStore) Set(resource *Resource) {
	r.Tree.Set(resource.Handler.Type(), resource)
}

func (r *ResourcesStore) Get(name string) *Resource {
	value, ok := r.Tree.Get(name)
	if !ok {
		return nil
	}
	res := value.(*Resource)
	return res
}
