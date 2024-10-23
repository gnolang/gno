package acl

import "fmt"

type perms []perm

func (perms perms) hasPerm(verb, resource string) bool {
	for _, perm := range perms {
		if perm.hasPerm(verb, resource) {
			return true
		}
	}
	return false
}

func (p perms) Dedup() perms {
    seen := make(map[string]struct{})
    uniquePerms := make([]perm, 0, len(p))

    for _, perm := range p {
        permStr := perm.String() // Get string representation
        if _, ok := seen[permStr]; ok {
            continue
        }
        seen[permStr] = struct{}{}
        uniquePerms = append(uniquePerms, perm)
    }

    return uniquePerms
}
