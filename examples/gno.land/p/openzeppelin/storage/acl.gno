package storage

// TODO: Figure out ACL permission management later.

func ACLKey() string {
	return "acl"
}

// TODO: Expose a method to specifically manage this after dealing with permission management.
func SetACL(state State, pkgs []string) {
	m := make(map[string]struct{})
	for _, pkg := range pkgs {
		m[pkg] = struct{}{}
	}
	state.Set(ACLKey(), m)
}

func IsInACL(state State, pkg string) bool {
	v, ok := state.Get(ACLKey())
	if !ok {
		return false
	}
	m := v.(map[string]struct{})
	_, ok = m[pkg]
	return ok
}
