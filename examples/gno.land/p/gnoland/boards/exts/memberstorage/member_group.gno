package memberstorage

import "strings"

// MemberGroup defines an interface for a group of members.
type MemberGroup interface {
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

// IsCanonical checks whether a member group is the canonical one, created
// using the NewMemberGroup() function defined within this package.
//
// Use this function at any public entry point that accepts a MemberGroup
// from an external caller before invoking any of its methods.
func IsCanonical(g MemberGroup) bool {
	_, ok := g.(*memberGroup)
	return ok
}

// NewMemberGroup creates a new group of members.
// It panics if name is empty or when members is nil.
func NewMemberGroup(name string, members MemberStorage) MemberGroup {
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
}

// Name returns the name of the group.
func (g *memberGroup) Name() string {
	return g.name
}

// Members returns the members that belong to the group.
func (g *memberGroup) Members() MemberStorage {
	return g.members
}

// SetMeta sets any metadata relevant to the group.
func (g *memberGroup) SetMeta(meta any) {
	g.meta = meta
}

// GetMeta returns the group metadata.
func (g *memberGroup) GetMeta() any {
	return g.meta
}
