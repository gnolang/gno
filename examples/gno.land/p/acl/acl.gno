package acl

import (
	"std"

	"gno.land/p/avl"
)

type Directory struct {
	permBuckets *avl.MutTree // identifier -> perms
	userGroups  *avl.MutTree // std.Address -> []string
}

func New() *Directory {
	return &Directory{
		userGroups:  avl.NewMutTree(),
		permBuckets: avl.NewMutTree(),
	}
}

func (d *Directory) HasPerm(addr std.Address, verb, resource string) bool {
	userPerms, ok := d.permBuckets.Get("u:" + addr.String())
	if ok {
		if userPerms.(perms).hasPerm(verb, resource) {
			return true
		}
	}
	return false
}

func (d *Directory) HasRole(addr std.Address, role string) bool {
	return d.HasPerm(addr, "role", role)
}

func (d *Directory) AddUserPerm(addr std.Address, verb, resource string) {
	bucket := "u:" + addr.String()
	p := perm{
		verbs:     []string{verb},
		resources: []string{resource},
	}
	d.addPermToBucket(bucket, p)
}

func (d *Directory) AddGroupPerm(name string, verb, resource string) {
	bucket := "g:" + name
	p := perm{
		verbs:     []string{verb},
		resources: []string{resource},
	}
	d.addPermToBucket(bucket, p)
}

func (d *Directory) addPermToBucket(bucket string, p perm) {
	var ps perms

	existing, ok := d.permBuckets.Get(bucket)
	if ok {
		ps = existing.(perms)
	}
	ps = append(ps, p)

	d.permBuckets.Set(bucket, ps)
}

func (d *Directory) AddUserToGroup(user std.Address, group string) {
	existing, ok := d.userGroups.Get(user.String())
	var groups []string
	if ok {
		groups = existing.([]string)
	}
	groups = append(groups, group)
	d.userGroups.Set(user.String(), &groups)
}
