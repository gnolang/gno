package acl

import (
	"std"

	"gno.land/p/demo/avl"
)

func New() *Directory {
	return &Directory{
		userGroups:  avl.Tree{},
		permBuckets: avl.Tree{},
	}
}

type Directory struct {
	permBuckets avl.Tree // identifier -> perms
	userGroups  avl.Tree // std.Address -> []string
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
	d.addPermsToBucket(bucket, []perm{p})
}

func (d *Directory) AddUserPerms(addr std.Address, verbs []string, resource string) {
	bucket := "u:" + addr.String()
	var ps perms
	for _, verb := range verbs {
		p := perm{
			verbs:     []string{verb},
			resources: []string{resource},
		}
		ps = append(ps, p)
	}
	d.addPermsToBucket(bucket, ps)
}

func (d *Directory) AddGroupPerm(name string, verb, resource string) {
	bucket := "g:" + name
	p := perm{
		verbs:     []string{verb},
		resources: []string{resource},
	}
	d.addPermsToBucket(bucket, []perm{p})
}

func (d *Directory) AddGroupPerms(name string, verbs []string, resource string) {
	bucket := "g:" + name
	var ps perms
	for _, verb := range verbs {
		p := perm{
			verbs:     []string{verb},
			resources: []string{resource},
		}
		ps = append(ps, p)
	}
	d.addPermsToBucket(bucket, ps)
}

func (d *Directory) addPermsToBucket(bucket string, p perms) {
	var ps perms

	existing, ok := d.permBuckets.Get(bucket)
	if ok {
		ps = existing.(perms)
	}
	ps = ps.appendWithoutDuplicates(p)

	d.permBuckets.Set(bucket, ps)
}

func (d *Directory) AddUserToGroup(user std.Address, group string) {
	existing, ok := d.userGroups.Get(user.String())
	var groups []string
	if ok {
		groups = existing.([]string)
	}
	groups = appendGroupsWithoutDuplicates(groups, []string{group})
	d.userGroups.Set(user.String(), groups)
}

func (d *Directory) AddUserToGroups(user std.Address, newGroups []string) {
	existing, ok := d.userGroups.Get(user.String())
	var groups []string
	if ok {
		groups = existing.([]string)
	}
	groups = appendGroupsWithoutDuplicates(groups, newGroups)
	d.userGroups.Set(user.String(), groups)
}

func (d *Directory) RemoveUserFromGroup(user std.Address, group string) {
	existing, ok := d.userGroups.Get(user.String())
	if !ok {
		return
	}
	groups := existing.([]string)
	for i, g := range groups {
		if g == group {
			groups = append(groups[:i], groups[i+1:]...)
			break
		}
	}
	d.userGroups.Set(user.String(), groups)
}

func (d *Directory) RemoveUserFromGroups(user std.Address, groups []string) {
	existing, ok := d.userGroups.Get(user.String())
	if !ok {
		return
	}
	existingGroups := existing.([]string)
	for _, group := range groups {
		for i, g := range existingGroups {
			if g == group {
				existingGroups = append(existingGroups[:i], existingGroups[i+1:]...)
				break
			}
		}
	}
	d.userGroups.Set(user.String(), existingGroups)
}

func (d *Directory) removePermsFromBucket(bucket string, p []perm) {
	existing, ok := d.permBuckets.Get(bucket)
	if !ok {
		return
	}
	ps := existing.(perms)

	for _, perm := range p {
		for i, existingPerm := range ps {
			if perm.Equal(existingPerm) {
				ps = append(ps[:i], ps[i+1:]...)
				break
			}
		}
	}

	d.permBuckets.Set(bucket, ps)
}

func (d *Directory) RemoveUserPerm(addr std.Address, verb, resource string) {
	bucket := "u:" + addr.String()
	p := perm{
		verbs:     []string{verb},
		resources: []string{resource},
	}
	d.removePermsFromBucket(bucket, []perm{p})
}

func (d *Directory) RemoveUserPerms(addr std.Address, verbs []string, resource string) {
	bucket := "u:" + addr.String()
	for _, verb := range verbs {
		p := perm{
			verbs:     []string{verb},
			resources: []string{resource},
		}
		d.removePermsFromBucket(bucket, []perm{p})
	}
}

func (d *Directory) RemoveGroupPerm(name string, verb, resource string) {
	bucket := "g:" + name
	p := perm{
		verbs:     []string{verb},
		resources: []string{resource},
	}
	d.removePermsFromBucket(bucket, []perm{p})
}

func (d *Directory) RemoveGroupPerms(name string, verbs []string, resource string) {
	bucket := "g:" + name
	for _, verb := range verbs {
		p := perm{
			verbs:     []string{verb},
			resources: []string{resource},
		}
		d.removePermsFromBucket(bucket, []perm{p})
	}
}

func (d *Directory) ResetGroupPerms(name string) {
	bucket := "g:" + name
	d.permBuckets.Set(bucket, perms{})
}

func (d *Directory) ResetUserPerms(addr std.Address) {
	bucket := "u:" + addr.String()
	d.permBuckets.Set(bucket, perms{})
}

func (d *Directory) ResetUserGroups(addr std.Address) {
	d.userGroups.Set(addr.String(), []string{})
}
