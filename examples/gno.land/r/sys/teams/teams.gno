package teams

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
)

var teams avl.Tree // std.Address -> Team

var (
	ErrNotHomeCaller     = errors.New("must be called from r/<team>/home realm")
	ErrUnauthorized      = errors.New("unauthorized operation")
	ErrAlreadyRegistered = errors.New("team already registered")
	ErrInvalidTeam       = errors.New("invalid or unknown team")
)

type Action interface{}

type (
	ActionAddPackage struct{ Path string }
	// NOTE: More actions to come...
)

type Permissions struct{ CanAddPackage bool }

// AccessController overriding access capabilities for team.
type AccessController interface {
	Can(member std.Address, do Action) bool
}

type Team struct {
	AccessController

	address std.Address
	members avl.Tree // Address -> Permissions // members default permissions
}

// HasMember checks if address belongs to the team
func (t *Team) Has(member std.Address) bool {
	return t.members.Has(member.String())
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
	t.members.IterateByOffset(start, count, func(key string, _ interface{}) bool {
		result = append(result, std.Address(key))
		return false
	})

	return result
}

func (t *Team) Can(member std.Address, do Action) bool {
	if t.IsTeamAddress(member) { // all mighty team address
		return true
	}

	var perms Permissions

	// If member is not known, deny systematically
	val, exists := t.members.Get(member.String())
	if !exists {
		return false
	}

	perms = val.(Permissions)

	if t.AccessController != nil {
		return t.AccessController.Can(member, do)
	}

	// Check default permission
	switch do.(type) {
	case ActionAddPackage:
		return perms.CanAddPackage
	default:
		return false
	}
}

// Admin method

func (t *Team) Tree() *rotree.ReadOnlyTree {
	return rotree.Wrap(&t.members, nil)
}

func (t *Team) Set(member std.Address, perms Permissions) (updated bool) {
	t.assertCallerIsTeamAddress()
	return t.members.Set(member.String(), perms)
}

func (t *Team) Register(member std.Address) {
	t.assertCallerIsTeamAddress()
	if !t.members.Has(member.String()) {
		// we don't want to override actual permission if already set
		t.members.Set(member.String(), Permissions{})
	}
}

func (t *Team) SetAccessController(ac AccessController) {
	t.assertCallerIsTeamAddress()
	t.AccessController = ac
	return
}

func (t *Team) IsTeamAddress(addr std.Address) bool {
	return addr == t.address
}

func (t *Team) assertCallerIsTeamAddress() {
	if !t.IsTeamAddress(std.GetOrigCaller()) {
		panic(ErrUnauthorized)
	}
}

// Register creates a new team from a home realm.
func Register() *Team {
	caller := std.GetOrigCaller()
	if teams.Has(caller.String()) {
		panic("team already register: " + caller.String())
	}

	prevPkg := std.PrevRealm().PkgPath()
	pathParts := strings.Split(prevPkg, "/")
	if len(pathParts) != 4 || !strings.HasPrefix(prevPkg, "gno.land/r/") || pathParts[3] != "home" {
		panic(ErrNotHomeCaller)
	}

	team := &Team{address: caller}
	teams.Set(caller.String(), team)
	return team

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
func Can(team, member std.Address, action Action) bool {
	if t := Get(team); t != nil {
		return t.Can(member, action)
	}

	panic(ErrInvalidTeam)
}

func SetMember(team, member std.Address, perms Permissions) (updated bool) {
	if t := Get(team); t != nil {
		return t.Set(member, perms)
	}

	panic(ErrInvalidTeam)
}
