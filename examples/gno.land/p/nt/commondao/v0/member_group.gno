package commondao

import "strings"

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
	// For security it's recommended that any meta value implements the
	// Sealable interface.
	SetMeta(any)

	// GetMeta returns the group metadata.
	GetMeta() any
}

// IsCanonicalMemberGroup checks wether a member group is the canonical
// one, created using the NewMemberGroup or MustNewMemberGroup functions
// defined within this package.
//
// Use this function at any public entry point that accepts a MemberGroup
// from an external caller before invoking any of its methods.
func IsCanonicalMemberGroup(g MemberGroup) bool {
	_, ok := g.(*memberGroup)
	return ok
}

// NewMemberGroup creates a new group of members.
// It panics if name is empty, or when members is nil or is not canonical.
func NewMemberGroup(name string, members MemberStorage) MemberGroup {
	// Make sure members is canonical so the returned MemberGroup is safe
	if members != nil && !IsCanonicalMemberStorage(members) {
		panic("member storage for group is not canonical")
	}

	// At this point the returned member group is safe. If members is nil it
	// will panic, otherwise members is a canonical type, so it's safe to be use.
	return NewUnsafeMemberGroup(name, members)
}

// NewUnsafeMemberGroup creates a new member group with the checks for canonical
// types disabled. Unsafe member groups allow using a custom non-canonical member
// storage implementation.
//
// The returned unsafe member group is unsafe because if a custom member storage is
// assigned by an untrusted caller it could be used to execute malicious code.
//
// IMPORTANT: Use unsafe member group if you fully understand `commondao` and the
// security implication that disabling these checks could have. If canonical checks
// are disabled it's highly recommended to use the `IsCanonical` functions before
// calling a method.
func NewUnsafeMemberGroup(name string, members MemberStorage) MemberGroup {
	if members == nil {
		panic("member storage is required")
	}

	name = strings.TrimSpace(name)
	if name == "" {
		panic("member group name is required")
	}

	return &memberGroup{
		name:    name,
		members: members,
	}
}

type memberGroup struct {
	name    string
	members MemberStorage
	meta    any
	sealed  bool
}

// Seal returns a sealed copy of the group.
// No further changes are accepted on the returned copy once sealed.
// The original group remains mutable.
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
