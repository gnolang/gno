package commondao

import (
	"errors"

	"gno.land/p/nt/avl"
)

// MemberGrouping defines an interface for storing multiple member groups.
// Member grouping can be used by implementations that require grouping users
// by roles or by tiers for example.
type MemberGrouping interface {
	// Size returns the number of groups that grouping contains.
	Size() int

	// Has checks if a group exists.
	Has(name string) bool

	// Add adds an new member group if it doesn't exists.
	Add(name string) (MemberGroup, error)

	// Get returns a member group.
	Get(name string) (_ MemberGroup, found bool)

	// Delete deletes a member group.
	Delete(name string) error

	// IterateByOffset iterates all member groups.
	IterateByOffset(offset, count int, fn func(MemberGroup) bool)
}

// NewMemberGrouping creates a new members grouping.
func NewMemberGrouping(options ...MemberGroupingOption) MemberGrouping {
	g := &memberGrouping{
		createStorage: func(string) MemberStorage { return NewMemberStorage() },
	}

	for _, apply := range options {
		apply(g)
	}
	return g
}

type memberGrouping struct {
	groups        avl.Tree // string(name) -> MemberGroup
	createStorage func(group string) MemberStorage
}

// Size returns the number of groups that grouping contains.
func (g memberGrouping) Size() int {
	return g.groups.Size()
}

// Has checks if a group exists.
func (g memberGrouping) Has(name string) bool {
	return g.groups.Has(name)
}

// Add adds an new member group if it doesn't exists.
func (g *memberGrouping) Add(name string) (MemberGroup, error) {
	if g.groups.Has(name) {
		return nil, errors.New("member group already exists: " + name)
	}

	mg, err := NewMemberGroup(name, g.createStorage(name))
	if err != nil {
		return nil, err
	}

	g.groups.Set(name, mg)
	return mg, nil
}

// Get returns a member group.
func (g memberGrouping) Get(name string) (_ MemberGroup, found bool) {
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

// IterateByOffset iterates all member groups.
func (g memberGrouping) IterateByOffset(offset, count int, fn func(MemberGroup) bool) {
	g.groups.IterateByOffset(offset, count, func(_ string, v any) bool {
		return fn(v.(MemberGroup))
	})
}
