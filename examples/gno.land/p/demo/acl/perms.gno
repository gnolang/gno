package acl

type perms []perm

func (p perms) Contains(p2 perm) bool {
	for _, pd := range p {
		if pd.Equal(p2) {
			return true
		}
	}
	return false
}

func (perms perms) hasPerm(verb, resource string) bool {
	for _, perm := range perms {
		if perm.hasPerm(verb, resource) {
			return true
		}
	}
	return false
}

func (p perms) appendWithoutDuplicates(newPerms perms) perms {
	for _, pd := range newPerms {
		if !p.Contains(pd) {
			p = append(p, pd)
		}
	}
	return p
}
