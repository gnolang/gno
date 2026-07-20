package memberstorage

import (
	"errors"

	"gno.land/p/nt/bptree/v0"
)

type (
	// MemberGroupIterFn defines a callback to iterate member groups.
	MemberGroupIterFn func(MemberGroup) bool

	// MemberGrouping defines an interface for storing multiple member groups.
	// Member grouping can be used by implementations that require grouping users
	// by roles or by tiers for example.
	MemberGrouping interface {
		// Size returns the number of groups that grouping contains.
		Size() int

		// Has checks if a group exists.
		Has(name string) bool

		// Add adds a new member group if it doesn't exist.
		Add(name string) (MemberGroup, error)

		// Get returns a member group.
		Get(name string) (_ MemberGroup, found bool)

		// Delete deletes a member group.
		Delete(name string) error

		// IterateByOffset iterates member groups starting at the given offset.
		// The callback can return true to stop iteration.
		IterateByOffset(offset, count int, fn MemberGroupIterFn) (stopped bool)
	}
)

// IsCanonicalMemberGrouping checks whether a member grouping is the canonical
// one, created using the NewMemberGrouping function defined within this package.
//
// Use this function at any public entry point that accepts a MemberGrouping
// from an external caller before invoking any of its methods.
func IsCanonicalMemberGrouping(g MemberGrouping) bool {
	_, ok := g.(*memberGrouping)
	return ok
}

// NewMemberGrouping creates a new member grouping.
func NewMemberGrouping() MemberGrouping {
	return &memberGrouping{}
}

type memberGrouping struct {
	groups bptree.BPTree // string(name) -> MemberGroup
}

// Size returns the number of groups that grouping contains.
func (g *memberGrouping) Size() int {
	return g.groups.Size()
}

// Has checks if a group exists.
func (g *memberGrouping) Has(name string) bool {
	return g.groups.Has(name)
}

// Add adds a new member group if it doesn't exist.
func (g *memberGrouping) Add(name string) (MemberGroup, error) {
	if g.groups.Has(name) {
		return nil, errors.New("member group already exists: " + name)
	}

	group := NewMemberGroup(name, New())
	g.groups.Set(name, group)
	return group, nil
}

// Get returns a member group.
func (g *memberGrouping) Get(name string) (_ MemberGroup, found bool) {
	v, found := g.groups.Get(name)
	if !found {
		return nil, false
	}
	return v.(MemberGroup), true
}

// Delete deletes a member group.
func (g *memberGrouping) Delete(name string) error {
	g.groups.Remove(name)
	return nil
}

// IterateByOffset iterates member groups starting at the given offset.
// The callback can return true to stop iteration.
func (g *memberGrouping) IterateByOffset(offset, count int, fn MemberGroupIterFn) bool {
	return g.groups.IterateByOffset(offset, count, func(_ string, v any) bool {
		return fn(v.(MemberGroup))
	})
}

// GetMemberGroups returns the groups that a member belongs to.
func GetMemberGroups(s MemberStorage, member address) []string {
	grouping := s.Grouping()
	if grouping == nil {
		return nil
	}

	var groups []string
	grouping.IterateByOffset(0, grouping.Size(), func(g MemberGroup) bool {
		if g.Members().Has(member) {
			groups = append(groups, g.Name())
		}
		return false
	})
	return groups
}
