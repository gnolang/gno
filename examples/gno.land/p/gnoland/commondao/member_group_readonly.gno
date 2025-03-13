package commondao

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/moul/addrset"
)

// ReadOnlyMemberGroup stores multiple members as a readonly group.
// Member groups can be useful in different situations,
// for example to organize users in different tiers or groups.
type ReadonlyMemberGroup struct {
	name    string
	members *addrset.Set
	values  avl.Tree // string(field name) -> any
}

func (g ReadonlyMemberGroup) Name() string {
	return g.name
}

func (g ReadonlyMemberGroup) Size() int {
	return g.members.Size()
}

func (g ReadonlyMemberGroup) GetValue(name string) any {
	v, _ := g.values.Get(name)
	return v
}

func (g ReadonlyMemberGroup) HasValue(name string) bool {
	return g.values.Has(name)
}

func (g ReadonlyMemberGroup) IsMember(user std.Address) bool {
	return g.members.Has(user)
}

func (g ReadonlyMemberGroup) Iterate(fn func(std.Address) bool) {
	g.members.IterateByOffset(0, g.members.Size(), fn)
}
