package acl

import (
	"std"

	"gno.land/p/avl"
)

type Directory struct {
	PermBuckets *avl.MutTree // identifier -> []perm
	UserGroups  *avl.MutTree // std.Address -> []string
}

func New() *Directory {
	return &Directory{
		UserGroups:  avl.NewMutTree(),
		PermBuckets: avl.NewMutTree(),
	}
}

func (d *Directory) HasPerm(addr std.Address, verb string, resource string) bool {
	return false
}

func (d *Directory) HasRole(addr std.Address, role string) bool {
	return d.HasPerm(addr, "role", role)
}

type perm struct {
	verbs     []string
	resources []string
}
