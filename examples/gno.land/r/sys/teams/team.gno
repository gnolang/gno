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

// Team defines the core capabilities for team management.
type Team interface {
	Can(member std.Address, action Action) bool

	// Members listing
	// XXX This should probably not be part of the interface, but should still be
	// something exposed through `sys/team`
	Size() int
	HasMember(member std.Address) bool
	PageMember(pageNumber, pageSize int) []std.Address
}

// Register creates a new team from a home realm.
func Register() {
	caller := std.GetOrigCaller()
	if teams.Has(caller.String()) {
		panic("team already register: " + caller.String())
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

// TeamSize returns the number of members in the specified team.
func TeamSize(team std.Address) int {
	if t := Get(team); t != nil {
		return t.Size()
	}

	panic(ErrInvalidTeam)
}

// TeamHasMember checks if a member is part of the specified team.
func TeamHasMember(team std.Address, member std.Address) bool {
	if t := Get(team); t != nil {
		return t.HasMember(member)
	}

	panic(ErrInvalidTeam)
}

// TeamPageMember returns a paginated list of members from the specified team.
func TeamPageMember(team std.Address, pageNumber, pageSize int) []std.Address {
	if t := Get(team); t != nil {
		return t.PageMember(pageNumber, pageSize)
	}

	panic(ErrInvalidTeam)
}

// CanTeamMemberDo checks permissions across teams.
func MemberCanDo(team, member std.Address, action string) bool {
	if t := Get(team); t != nil {
		return t.Can(member, action)
	}

	panic(ErrInvalidTeam)
}
