// Package team provides decentralized team management with permission control.
// Teams are identified by addresses and manage member permissions.
package teams

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/moul/addrset"
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
type AccessController interface {
	Can(member std.Address, do Action) bool
}

type Team struct {
	AccessController

	address std.Address
	members addrset.Set
}

// HasMember checks if address belongs to the team
func (t *Team) Has(member std.Address) bool {
	return t.members.Has(member)
}

// Members size returns total number of team members
func (t *Team) Size() int {
	return t.members.Size()
}

// Page retrieves members with pagination support
func (t *Team) Page(pageNumber, pageSize int) []std.Address {
	start := pageNumber * pageSize
	end := start + pageSize
	size := t.members.Size()

	if start >= size {
		return nil
	}
	if end > size {
		end = size
	}

	count := end - start
	result := make([]std.Address, 0, count)
	t.members.IterateByOffset(start, count, func(key std.Address) bool {
		result = append(result, std.Address(key))
		return false
	})

	return result
}

func (t *Team) Tree() *rotree.ReadOnlyTree {
	tree := t.members.Tree().(*avl.Tree)
	return rotree.Wrap(tree, nil)
}

// Admin method
func (t *Team) Register(member std.Address) {
	t.assertTeamAddress()
	t.members.Add(member)
	return
}

func (t *Team) SetAccessController(ac AccessController) {
	t.assertTeamAddress()
	t.AccessController = ac
	return
}

func (t *Team) assertTeamAddress() {
	caller := std.GetOrigCaller()
	if caller != t.address {
		panic(ErrUnauthorized)
	}

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

	teams.Set(caller.String(), &BasicAccessController{})
}

func Get(team std.Address) *Team {
	if val, ok := teams.Get(team.String()); ok {
		return val.(*Team)
	}

	return nil
}

// Has checks if a member is part of the specified team.
func Has(team std.Address, member std.Address) bool {
	if t := Get(team); t != nil {
		return t.Has(member)
	}

	panic(ErrInvalidTeam)
}

// Page returns a paginated list of members from the specified team.
func Page(team std.Address, pageNumber, pageSize int) []std.Address {
	if t := Get(team); t != nil {
		return t.Page(pageNumber, pageSize)
	}

	panic(ErrInvalidTeam)
}

// CanTeamMemberDo checks permissions across teams.
func Can(team, member std.Address, action string) bool {
	if t := Get(team); t != nil {
		return t.Can(member, action)
	}

	panic(ErrInvalidTeam)
}
