package administrable

import "std"

type UnprivilegedSet struct {
	Set
}

// Unprivileged returns an unprivileged structure that can be exposed safely.
func (s *Set) Unprivileged() UnprivilegedSet {
	return UnprivilegedSet{Set: *s}
}

func (u *UnprivilegedSet) Add(addr std.Address) {
	u.AssertCurrentHasAccess()
	u.Set.Add(addr)
}

func (u *UnprivilegedSet) Del(addr std.Address) {
	u.AssertCurrentHasAccess()
	u.Set.Del(addr)
}

func (u *UnprivilegedSet) ReplaceAll(addr std.Address) {
	u.AssertCurrentHasAccess()
	u.Set.ReplaceAll(addr)
}
