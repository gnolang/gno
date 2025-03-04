// Teams are entities that manage members and their permissions for specific actions.
//
// Basic usage:
//
//	team := teams.Register() // Called from gno.land/r/<team>/home realm
//	team.SetMember("member1", "add_package")
//

package teams

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
)

// Global registry of all teams
var teams avl.Tree // std.Address -> *Team

var (
	ErrNotHomeCaller     = errors.New("must be called from r/<team>/home realm")
	ErrUnauthorized      = errors.New("unauthorized operation")
	ErrAlreadyRegistered = errors.New("team already registered")
	ErrInvalidTeam       = errors.New("invalid or unknown team")
	ErrInvalidMember     = errors.New("invalid or unknown member")
	ErrInvalidPermission = errors.New("invalid permission string")
)

// Action represents an operation that requires authorization
type Action interface{}

type (
	// ActionAddPackage represents permission to add packages to a team's realm
	ActionAddPackage struct{ Path string }
)

// Permissions defines default access rights for team members
type Permissions struct{ CanAddPackage bool }

// AccessController allows custom permission handling for advanced use cases
type AccessController interface {
	// Can determines if member has permission for specific action
	Can(member std.Address, do Action) bool
}

// Team represents a decentralized organization with member access control
type Team struct {
	AccessController // Optional custom permission handler

	address std.Address // Team's blockchain address
	members avl.Tree    // Member permissions storage (Address.String() -> Permissions)
}

// Address returns the team's address
func (t *Team) Address() std.Address {
	return t.address
}

// Has checks if address belongs to the team
func (t *Team) Has(member std.Address) bool {
	return t.members.Has(member.String())
}

// Size returns total number of team members
func (t *Team) Size() int {
	return t.members.Size()
}

// Page retrieves paginated member list
// pageNumber: 0-based page index
// pageSize: number of members per page
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

// Can checks if member has permission for an action
// Team addresses always have full access
func (t *Team) Can(member std.Address, do Action) bool {
	if t.IsTeamAddress(member) {
		return true
	}

	val, exists := t.members.Get(member.String())
	if !exists {
		return false
	}

	perms := val.(Permissions)

	// Delegate to custom access controller if set
	if t.AccessController != nil {
		return t.AccessController.Can(member, do)
	}

	switch do.(type) {
	case ActionAddPackage:
		return perms.CanAddPackage
	default:
		return false
	}
}

// Tree provides read-only access to member permissions
func (t *Team) Tree() *rotree.ReadOnlyTree {
	return rotree.Wrap(&t.members, nil)
}

// IsTeamAddress checks if address matches team's address
func (t *Team) IsTeamAddress(addr std.Address) bool {
	return addr == t.address
}

// Set updates member permissions (team admin only)
func (t *Team) Set(member std.Address, perms Permissions) (updated bool) {
	t.assertCallerIsTeamAddress()
	if !member.IsValid() {
		panic(ErrInvalidMember)
	}
	return t.members.Set(member.String(), perms)
}

// Register adds a member with default permissions (team admin only)
func (t *Team) Register(member std.Address) {
	t.assertCallerIsTeamAddress()
	if !member.IsValid() {
		panic(ErrInvalidMember)
	}
	if !t.members.Has(member.String()) {
		t.members.Set(member.String(), Permissions{})
	}
}

// SetAccessController configures custom permission handler (team admin only)
func (t *Team) SetAccessController(ac AccessController) {
	t.assertCallerIsTeamAddress()
	t.AccessController = ac
}

// assertCallerIsTeamAddress validates transaction authorization
func (t *Team) assertCallerIsTeamAddress() {
	if !t.IsTeamAddress(std.GetOrigCaller()) {
		panic(ErrUnauthorized)
	}
}

var teamSingleton *Team // Cached team reference for Myteam()

// Register creates a new team from a valid home realm:
// Must be called from a package path matching "gno.land/r/<team>/home"
//
// Example:
//
//	// From gno.land/r/myteam/home
//	myteam := teams.Register()
//	myteam.Set(member, Permissions{ CanAddPackage: true })
func Register() *Team {
	caller := std.GetOrigCaller()
	if teams.Has(caller.String()) {
		panic(ErrAlreadyRegistered)
	}

	prevPkg := std.PrevRealm().PkgPath()
	pathParts := strings.Split(prevPkg, "/")
	if len(pathParts) != 4 || !strings.HasPrefix(prevPkg, "gno.land/r/") || pathParts[3] != "home" {
		panic(ErrNotHomeCaller)
	}

	team := &Team{address: caller}
	teams.Set(caller.String(), team)
	teamSingleton = team // Update singleton
	return team
}

// My returns the team associated with current transaction caller
// Primarily used for method chaining after registration:
//
// Example:
//
//	// From gno.land/r/myteam/home
//	teams.Register()
//	teams.MyTeam().SetMember("founder", "add_package")
//	teams.MyTeam().Page(0, 10)
func MyTeam() *Team {
	caller := std.GetOrigCaller()
	if teamSingleton == nil || teamSingleton.Address() != caller {
		teamSingleton = MustGet(caller)
	}
	return teamSingleton
}

func MustGet(team std.Address) *Team {
	if val, ok := teams.Get(team.String()); ok {
		return val.(*Team)
	}

	panic(ErrInvalidMember)
}

// Get retrieves a team by address
func Get(team std.Address) *Team {
	if val, ok := teams.Get(team.String()); ok {
		return val.(*Team)
	}

	return nil
}

// SetMember updates permissions using comma-separated string-based flags
// perm can be:
//   - "add_package"
func SetMember(member std.Address, perms string) (updated bool) {
	var ps Permissions

	for _, perm := range strings.Split(perms, ",") {
		switch perm {
		case "add_package":
			ps.CanAddPackage = true
		case "": // Default permissions
		default:
			panic(ErrInvalidPermission)
		}
	}

	return MyTeam().Set(member, ps)
}

// CheckPermission verifies member access for an action
func HasPermission(team, member std.Address, action Action) bool {
	return MustGet(team).Can(member, action)
}

// HasMember checks membership status
func HasMember(team std.Address, member std.Address) bool {
	return MustGet(team).Has(member)
}

// MembersPage returns paginated member list
func MembersPage(team std.Address, pageNumber, pageSize int) []std.Address {
	return MustGet(team).Page(pageNumber, pageSize)
}

func Size(team std.Address) int {
	if t := Get(team); t != nil {
		return t.Size()
	}

	return 0
}
