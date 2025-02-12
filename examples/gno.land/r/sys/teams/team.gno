// Package team provides decentralized team management with permission control.
// Teams are identified by addresses and manage member permissions.
package teams

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/avl"
)

var teams avl.Tree // std.Address -> Team

var (
	ErrNotHomeCaller = errors.New("must be called from r/<team>/home realm")
	ErrUnauthorized  = errors.New("unauthorized operation")
	ErrInvalidTeam   = errors.New("invalid or unknown team")
)

type Action interface{}

type (
	ActionAddPackage struct{ Path string }
	// NOTE: More actions to come...
)

// Team defines the core capabilities for team management
type Team interface {
	Can(member std.Address, action Action) bool

	// Members listing
	Size() int
	HasMember(member std.Address) bool
	PageMember(pageNumber, pageSize int) []std.Address
}

// Register creates a new team from a home realm
func Register() {
	caller := std.GetOrigCaller()
	if teams.Has(caller.String()) {
		panic("team already exists: " + caller.String())
	}

	prev := std.PrevRealm()
	pathParts := strings.Split(prev.PkgPath(), "/")
	if len(pathParts) != 4 || pathParts[2] != caller.String() || pathParts[3] != "home" {
		panic(ErrNotHomeCaller)
	}

	teams.Set(caller.String(), &BasicTeam{})
}

func Get(team std.Address) Team {
	if val, ok := teams.Get(team.String()); ok {
		return val.(Team)
	}

	return nil
}

// SetMember updates permissions for a team member
func HasMember(team, member std.Address) bool {
	caller := std.GetOrigCaller()
	val, exists := teams.Get(caller.String())
	if !exists {
		panic(ErrInvalidTeam)
	}

	return val.(Team).HasMember(member)
}

// CanTeamMemberDo checks permissions across teams
func MemberCanDo(team, member std.Address, action string) bool {
	val, exists := teams.Get(team.String())
	if !exists {
		return false
	}
	return val.(Team).Can(member, action)
}
