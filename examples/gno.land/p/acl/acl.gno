package acl

import (
	"std"

	"gno.land/p/avl"
)

func New() *Directory {
	return &Directory{
		userGroups:  avl.NewMutTree(),
		permBuckets: avl.NewMutTree(),
	}
}

type Directory struct {
	permBuckets *avl.MutTree // identifier -> perms
	userGroups  *avl.MutTree // std.Address -> []string
}

func (d *Directory) HasPerm(addr std.Address, verb, resource string) bool {
	// FIXME: consider memoize.

	// user perms
	if d.getBucketPerms("u:"+addr.String()).hasPerm(verb, resource) {
		return true
	}

	// everyone's perms.
	if d.getBucketPerms("g:"+Everyone).hasPerm(verb, resource) {
		return true
	}

	// user groups' perms.
	groups, ok := d.userGroups.Get(addr.String())
	if ok {
		for _, group := range groups.([]string) {
			if d.getBucketPerms("g:"+group).hasPerm(verb, resource) {
				return true
			}
		}
	}

	return false
}

func (d *Directory) getBucketPerms(bucket string) perms {
	res, ok := d.permBuckets.Get(bucket)
	if ok {
		return res.(perms)
	}
	return perms{}
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
	d.userGroups.Set(user.String(), groups)
}
