package dao_realm

import (
	"encoding/base64"
	"std"
	"strings"
	"time"

	dao_core "gno.land/p/demo/teritori/daodao/core"
	dao_interfaces "gno.land/p/demo/teritori/daodao/interfaces"
	proposal_single "gno.land/p/demo/teritori/daodao/proposal_single"
	voting_group "gno.land/p/demo/teritori/daodao/voting_group"
	"gno.land/p/demo/ujson"
	registry "gno.land/r/demo/dao_registry"
	"gno.land/r/demo/groups"
	modboards "gno.land/r/demo/modboards"
	"gno.land/r/demo/tori"
)

var (
	daoCore       dao_interfaces.IDAOCore
	mainBoardName = "dao_realm"
	groupName     = mainBoardName + "_voting_group"
	groupID       groups.GroupID
)

func init() {
	modboards.CreateBoard(mainBoardName)

	votingModuleFactory := func(core dao_interfaces.IDAOCore) dao_interfaces.IVotingModule {
		groupID = groups.CreateGroup(groupName)
		groups.AddMember(groupID, "g1747t5m2f08plqjlrjk2q0qld7465hxz8gkx59c", 1, "")
		groups.AddMember(groupID, "g108cszmcvs4r3k67k7h5zuhm4el3qhlrxzhshtv", 1, "")
		groups.AddMember(groupID, "g14u5eaheavy0ux4dmpykg2gvxpvqvexm9cyg58a", 1, "")
		groups.AddMember(groupID, "g1ckn395mpttp0vupgtratyufdaakgh8jgkmr3ym", 1, "")
		groups.AddMember(groupID, std.GetOrigCaller().String(), 1, "")
		return voting_group.NewVotingGroup(groupID)
	}

	// TODO: consider using factories that return multiple modules and handlers

	proposalModulesFactories := []dao_interfaces.ProposalModuleFactory{
		func(core dao_interfaces.IDAOCore) dao_interfaces.IProposalModule {
			tt := proposal_single.PercentageThresholdPercent(100)
			tq := proposal_single.PercentageThresholdPercent(100)
			return proposal_single.NewDAOProposalSingle(core, &proposal_single.DAOProposalSingleOpts{
				MaxVotingPeriod: time.Hour * 24 * 42,
				Threshold: &proposal_single.ThresholdThresholdQuorum{
					Threshold: &tt, // 1%
					Quorum:    &tq, // 1%
				},
			})
		},
	}

	messageHandlersFactories := []dao_interfaces.MessageHandlerFactory{
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			return groups.NewAddMemberHandler()
		},
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			return groups.NewDeleteMemberHandler()
		},
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			// TODO: add a router to support multiple proposal modules
			propMod := core.ProposalModules()[0]
			return proposal_single.NewUpdateSettingsHandler(propMod.Module.(*proposal_single.DAOProposalSingle))
		},
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			return modboards.NewCreateBoardHandler()
		},
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			return modboards.NewDeletePostHandler()
		},
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			return tori.NewMintToriHandler()
		},
	}

	daoCore = dao_core.NewDAOCore(votingModuleFactory, proposalModulesFactories, messageHandlersFactories)

	registry.Register("DAO Realm", "Default testing DAO", "https://images.unsplash.com/photo-1477959858617-67f85cf4f1df?w=1080&fit=max")
}

func Render(path string) string {
	return "[[board](/r/demo/modboards:" + mainBoardName + ")]\n\n" + daoCore.Render(path)
}

func VoteJSON(moduleIndex int, proposalID int, voteJSON string) {
	// move check in dao core
	module := dao_core.GetProposalModule(daoCore, moduleIndex)
	if !module.Enabled {
		panic("proposal module is not enabled")
	}
	module.Module.VoteJSON(proposalID, voteJSON)
}

func Execute(moduleIndex int, proposalID int) {
	// move check in dao core
	module := dao_core.GetProposalModule(daoCore, moduleIndex)
	if !module.Enabled {
		panic("proposal module is not enabled")
	}
	module.Module.Execute(proposalID)
}

func ProposeJSON(moduleIndex int, proposalJSON string) int {
	// move check in dao core
	module := dao_core.GetProposalModule(daoCore, moduleIndex)
	if !module.Enabled {
		panic("proposal module is not enabled")
	}
	return module.Module.ProposeJSON(proposalJSON)
}

func getProposalsJSON(moduleIndex int, limit int, startAfter string, reverse bool) string {
	// move logic in dao core
	module := dao_core.GetProposalModule(daoCore, moduleIndex)
	return module.Module.ProposalsJSON(limit, startAfter, reverse)
}

func getProposalJSON(moduleIndex int, proposalIndex int) string {
	// move logic in dao core
	module := dao_core.GetProposalModule(daoCore, moduleIndex)
	return module.Module.ProposalJSON(proposalIndex)
}
