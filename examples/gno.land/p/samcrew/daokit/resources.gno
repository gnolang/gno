package daokit

// Maps action types to their handlers and conditions.
//
// Example: "transfer" action requires 3 member votes for execution.

import (
	"gno.land/p/nt/avl"
	"gno.land/p/samcrew/daocond"
)

type ResourcesStore struct {
	Tree *avl.Tree // string -> daocond.Condition
}

type Resource struct {
	Handler     ActionHandler     // Executes the action when proposal passes
	Condition   daocond.Condition // Voting rules required for approval
	DisplayName string
	Description string
}

func NewResourcesStore() *ResourcesStore {
	return &ResourcesStore{
		Tree: avl.NewTree(),
	}
}

// Retrieves a resource by its handler type name.
func (r *ResourcesStore) Get(name string) *Resource {
	value, ok := r.Tree.Get(name)
	if !ok {
		return nil
	}
	res := value.(*Resource)
	return res
}

// Registers a new resource with its handler type as key.
func (r *ResourcesStore) Set(resource *Resource) {
	r.Tree.Set(resource.Handler.Type(), resource)
}
