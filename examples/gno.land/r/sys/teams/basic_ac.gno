package teams

import (
	"std"

	"gno.land/p/demo/avl"
)

type Permissions struct{ CanAddPackage bool }

// BasicAccessController is the default Access Controller implementation
type BasicAccessController struct {
	members avl.Tree // std.Address -> Perms
}

// Can checks permissions for a specific action
func (ac *BasicAccessController) Can(member std.Address, action Action) bool {
	val, exists := ac.members.Get(member.String())
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

func (ac *BasicAccessController) setPermission(member std.Address, perms Permissions) {
	ac.members.Set(member.String(), perms)
}

// SetMember updates permissions for a team member
// Only team address can set member permission
func OwnerSetMember(member std.Address, perms Permissions) {
	caller := std.GetOrigCaller()
	team := Get(caller)
	if team == nil {
		panic(ErrInvalidTeam)
	}

	ac, ok := team.AccessController.(*BasicAccessController)
	if !ok {
		panic("AccessControler not managed by `sys/teams`")
	}

	team.Register(member)
	ac.setPermission(member, perms)
}
