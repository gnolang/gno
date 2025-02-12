package teams

import (
	"std"

	"gno.land/p/demo/avl"
)

type Permissions struct{ CanAddPackage bool }

// BasicTeam is the default team implementation
type BasicTeam struct {
	members avl.Tree // std.Address.String() -> Perms
}

// Can checks permissions for a specific action
func (t *BasicTeam) Can(member std.Address, action Action) bool {
	val, exists := t.members.Get(member.String())
	if !exists {
		return false
	}

	perms := val.(Permissions)

	switch action.(type) {
	case ActionAddPackage:
		return perms.CanAddPackage
	default:
		return false
	}
}

// Size returns total number of team members
func (t *BasicTeam) Size() int {
	return t.members.Size()
}

// HasMember checks if address belongs to the team
func (t *BasicTeam) HasMember(member std.Address) bool {
	return t.members.Has(member.String())
}

// Page retrieves members with pagination support
func (t *BasicTeam) Page(pageNumber, pageSize int) []std.Address {
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
	t.members.IterateByOffset(start, count, func(key string, perm interface{}) bool {
		result = append(result, std.Address(key))
		return true
	})
	return result
}

// SetMember updates permissions for a team member
// Only team address can set member permission
func OwnerSetMember(member std.Address, perms Permissions) {
	caller := std.GetOrigCaller()
	val, exists := teams.Get(caller.String())
	if !exists {
		panic(ErrInvalidTeam)
	}

	team, ok := val.(*BasicTeam)
	if !ok {
		panic("unable to set permission on not-managed team by `sys/teams`")
	}

	team.members.Set(member.String(), perms)
}
