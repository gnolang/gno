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
		Sealable

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
	sealed        bool
}

// Seal returns a sealed copy of the grouping.
// No further changes are accepted on the returned copy once sealed.
func (g memberGrouping) Seal() Sealable {
	g.sealed = true
	return &g
}

// IsSealed checks if grouping is sealed.
func (g memberGrouping) IsSealed() bool {
	return g.sealed
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
// It panics when grouping is sealed.
func (g *memberGrouping) Add(name string) (MemberGroup, error) {
	assertNotSealed(g)

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
	return g.maySealGroup(v.(MemberGroup)), true
}

// Delete deletes a member group.
// It panics when grouping is sealed.
func (g *memberGrouping) Delete(name string) error {
	assertNotSealed(g)
	g.groups.Remove(name)
	return nil
}

// IterateByOffset iterates all member groups.
func (g memberGrouping) IterateByOffset(offset, count int, fn MemberGroupIterFn) bool {
	return g.groups.IterateByOffset(offset, count, func(_ string, v any) bool {
		return fn(g.maySealGroup(v.(MemberGroup)))
	})
}

func (g memberGrouping) maySealGroup(mg MemberGroup) MemberGroup {
	if g.sealed {
		return mg.Seal().(MemberGroup)
	}
	return mg
}
