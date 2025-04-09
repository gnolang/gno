package commondao

import (
	"gno.land/p/demo/avl"
)

// Groupings stores groups of DAO members.
// Members grouping can be used to store members in different groups
// which can be useful in different situations, for example when
// users are organized in different tiers or grouped by roles.
type Groupings struct {
	groups avl.Tree
}

// Add adds a new member group.
// If a group with the same name exists it's replaced by the one being added.
func (g *Groupings) Add(m MemberGroup) (udpated bool) {
	return g.groups.Set(m.Name(), m)
}

// Remove removes an existing group.
func (g Groupings) Remove(name string) {
	g.groups.Remove(name)
}

// Get returns a group of members.
func (g Groupings) Get(name string) (_ MemberGroup, found bool) {
	v, found := g.groups.Get(name)
	if found {
		return v.(MemberGroup), true
	}
	return MemberGroup{}, false
}

// Exists checks if a group of members exists.
func (g Groupings) Exists(name string) bool {
	return g.groups.Has(name)
}
