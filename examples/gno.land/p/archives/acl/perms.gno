package acl

type perms []perm

func (perms perms) hasPerm(verb, resource string) bool {
	for _, perm := range perms {
		if perm.hasPerm(verb, resource) {
			return true
		}
	}
	return false
}
