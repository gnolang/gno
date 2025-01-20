package teams

import (
	"std"

	"gno.land/p/demo/ownable"
)

type OwnableTeam struct {
	*OwnableAccessController
}

func NewOwnableTeam(owner *ownable.Ownable) ITeam {
	return &OwnableTeam{
		OwnableAccessController: NewOwnableAccessController(owner),
	}
}

func (o *OwnableTeam) Init() Task {
	return ChainTasks(
		AddMemberTask(o.Owner()),
		// Burn team address, so only owner can control the team
		BurnTeamAddressTask,
	)
}

func (o *OwnableTeam) ApplyUpdate(cmd Cmd) Task {
	return ApplyDefault(cmd)
}

type OwnableAccessController struct {
	*ownable.Ownable

	EnableAddMember    bool
	EnableRemoveMember bool
	EnableAddPackage   bool
}

func NewOwnableAccessController(ownable *ownable.Ownable) *OwnableAccessController {
	return &OwnableAccessController{Ownable: ownable}
}

func (o *OwnableAccessController) CanRun(member std.Address, cmd Cmd) bool {

	if o.Ownable.Owner() == member { // All mighty owner
		return true
	}

	switch cmd.(type) {
	case AddMemberCmd:
		return o.EnableAddMember
	case RemoveMemberCmd:
		return o.EnableRemoveMember
	case AddPackageCmd:
		return o.EnableAddPackage
	}

	return false
}
