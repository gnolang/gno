package commondao

import (
	"errors"

	"gno.land/p/nt/avl/v0"
)

type (
	// MemberGroupIterFn defines a callback to iterate DAO members groups.
	MemberGroupIterFn func(MemberGroup) bool

	// MemberGrouping defines an interface for storing multiple member groups.
	// Member grouping can be used by implementations that require grouping users
	// by roles or by tiers for example.
	MemberGrouping interface {
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
		// The callback can return true to stop iteration.
		IterateByOffset(offset, count int, fn MemberGroupIterFn) (stopped bool)
	}
)

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
func (g memberGrouping) IterateByOffset(offset, count int, fn MemberGroupIterFn) bool {
	return g.groups.IterateByOffset(offset, count, func(_ string, v any) bool {
		return fn(v.(MemberGroup))
	})
}

// NewReadonlyMemberGrouping creates a new grouping if member.
func NewReadonlyMemberGrouping(g MemberGrouping) (*ReadonlyMemberGrouping, error) {
	if g == nil {
		return nil, errors.New("member grouping is required")
	}
	return &ReadonlyMemberGrouping{g}, nil
}

// MustNewReadonlyMemberGrouping creates a new grouping if member or panics on error.
func MustNewReadonlyMemberGrouping(g MemberGrouping) *ReadonlyMemberGrouping {
	grouping, err := NewReadonlyMemberGrouping(g)
	if err != nil {
		panic(err)
	}
	return grouping
}

// ReadonlyMemberGrouping defines a type for storing multiple readonly member groups.
type ReadonlyMemberGrouping struct {
	grouping MemberGrouping
}

// Size returns the number of groups that grouping contains.
func (g ReadonlyMemberGrouping) Size() int {
	if g.grouping == nil {
		return 0
	}
	return g.grouping.Size()
}

// Has checks if a group exists.
func (g ReadonlyMemberGrouping) Has(name string) bool {
	if g.grouping == nil {
		return false
	}
	return g.grouping.Has(name)
}

// Get returns a member group.
func (g ReadonlyMemberGrouping) Get(name string) (_ *ReadonlyMemberGroup, found bool) {
	if g.grouping == nil {
		return nil, false
	}

	group, found := g.grouping.Get(name)
	if !found {
		return nil, false
	}
	return MustNewReadonlyMemberGroup(group), true
}

// IterateByOffset iterates all member groups.
func (g ReadonlyMemberGrouping) IterateByOffset(offset, count int, fn func(*ReadonlyMemberGroup) bool) bool {
	if g.grouping == nil {
		return false
	}

	return g.grouping.IterateByOffset(offset, count, func(group MemberGroup) bool {
		return fn(MustNewReadonlyMemberGroup(group))
	})
}
