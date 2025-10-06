package acl

import "gno.land/p/nt/avl"

func New() *Directory {
	return &Directory{
		userGroups:  avl.Tree{},
		permBuckets: avl.Tree{},
	}
}

type Directory struct {
	permBuckets avl.Tree // identifier -> perms
	userGroups  avl.Tree // chain.Address -> []string
}

func (d *Directory) HasPerm(addr address, verb, resource string) bool {
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

func (d *Directory) HasRole(addr address, role string) bool {
	return d.HasPerm(addr, "role", role)
}

func (d *Directory) AddUserPerm(addr address, verb, resource string) {
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

func (d *Directory) AddUserToGroup(user address, group string) {
	existing, ok := d.userGroups.Get(user.String())
	var groups []string
	if ok {
		groups = existing.([]string)
	}
	groups = append(groups, group)
	d.userGroups.Set(user.String(), groups)
}

// TODO: helpers to remove permissions.
// TODO: helpers to adds multiple permissions at once -> {verbs: []string{"read","write"}}.
// TODO: helpers to delete users from gorups.
// TODO: helpers to quickly reset states.
