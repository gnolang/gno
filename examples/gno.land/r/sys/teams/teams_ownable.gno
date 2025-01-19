package teams

import (
	"std"

	"gno.land/p/demo/ownable"
)

type OwnableAccessController struct {
	*ownable.Ownable

	EnableAddMember    bool
	EnableRemoveMember bool
	DisableAddPackage  bool
}

func NewOwnableTeam(ownable *ownable.Ownable) *OwnableAccessController {
	return &OwnableAccessController{Ownable: ownable}
}

func (o *OwnableAccessController) CanPerform(member std.Address, msg Msg) bool {
	switch msg.(type) {
	case AddMemberMsg:
		if o.EnableAddMember {
			return true
		}
	case RemoveMemberMsg:
		if o.EnableRemoveMember {
			return true
		}
	case AddPackageMsg:
		if o.DisableAddPackage {
			return false
		}
	}

	return o.Ownable.Owner() == member
}
