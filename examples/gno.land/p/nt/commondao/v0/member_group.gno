package commondao

import (
	"errors"
	"strings"
)

// MemberGroup defines an interface for a group of members.
type MemberGroup interface {
	Sealable

	// Name returns the name of the group.
	Name() string

	// Members returns the members that belong to the group.
	Members() MemberStorage

	// SetMeta sets any metadata relevant to the group.
	// Metadata can be used to store data which is specific to the group.
	// Usually can be used to store parameter values which would be useful
	// during proposal voting or tallying to resolve things like voting
	// weights or rights for example.
	SetMeta(any)

	// GetMeta returns the group metadata.
	GetMeta() any
}

// NewMemberGroup creates a new group of members.
func NewMemberGroup(name string, members MemberStorage) (MemberGroup, error) {
	if members == nil {
		return nil, errors.New("member storage is required")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("member group name is required")
	}

	return &memberGroup{
		name:    name,
		members: members,
	}, nil
}

// MustNewMemberGroup creates a new group of members or panics on error.
func MustNewMemberGroup(name string, members MemberStorage) MemberGroup {
	g, err := NewMemberGroup(name, members)
	if err != nil {
		panic(err)
	}
	return g
}

type memberGroup struct {
	name    string
	members MemberStorage
	meta    any
	sealed  bool
}

// Seal returns a sealed copy of the group.
// No further changes are accepted on the returned copy once sealed.
func (g memberGroup) Seal() Sealable {
	g.sealed = true
	return &g
}

// IsSealed checks if grouping is sealed.
func (g memberGroup) IsSealed() bool {
	return g.sealed
}

// Name returns the name of the group.
func (g memberGroup) Name() string {
	return g.name
}

// Members returns the members that belong to the group.
func (g memberGroup) Members() MemberStorage {
	if g.sealed {
		return g.members.Seal().(MemberStorage)
	}
	return g.members
}

// SetMeta sets any metadata relevant to the group.
// It panics when group is sealed.
func (g *memberGroup) SetMeta(meta any) {
	assertNotSealed(g)
	g.meta = meta
}

// GetMeta returns the group metadata.
// It panics when group is sealed and metadata is not sealable.
// Metadata must implement `Sealable` interface to be exposed when group is sealed.
func (g memberGroup) GetMeta() any {
	if !g.sealed {
		return g.meta
	}

	s, ok := g.meta.(Sealable)
	if !ok {
		// Don't return metadata when group is sealed and metadata
		// is not sealable to avoid exposing data unintentionally.
		panic("failed to get metadata, group is sealed and metadata is not sealable")
	}
	return s.Seal()
}
