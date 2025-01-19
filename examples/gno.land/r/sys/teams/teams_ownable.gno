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

func NewOwnableAccessController(ownable *ownable.Ownable) *OwnableAccessController {
	return &OwnableAccessController{Ownable: ownable}
}

func (o *OwnableAccessController) CanRun(member std.Address, cmd Cmd) bool {
	switch cmd.(type) {
	case AddMemberCmd:
		if o.EnableAddMember {
			return true
		}
	case RemoveMemberCmd:
		if o.EnableRemoveMember {
			return true
		}
	case AddPackageCmd:
		if o.DisableAddPackage {
			return false
		}
	}

	return o.Ownable.Owner() == member
}
