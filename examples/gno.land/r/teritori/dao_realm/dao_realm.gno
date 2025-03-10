package dao_realm

import (
	"std"
	"time"

	"gno.land/r/teritori/dao_registry"
	"gno.land/r/teritori/tori"

	dao_core "gno.land/p/teritori/dao_core"
	dao_interfaces "gno.land/p/teritori/dao_interfaces"
	proposal_single "gno.land/p/teritori/dao_proposal_single"
	"gno.land/p/teritori/dao_utils"
	voting_group "gno.land/p/teritori/dao_voting_group"
)

// Example DAO realm

var (
	daoCore    dao_interfaces.IDAOCore
	group      *voting_group.VotingGroup
	registered bool
)

func init() {
	votingModuleFactory := func(core dao_interfaces.IDAOCore) dao_interfaces.IVotingModule {
		group = voting_group.NewVotingGroup()
		group.SetMemberPower("g1747t5m2f08plqjlrjk2q0qld7465hxz8gkx59c", 1)
		group.SetMemberPower("g108cszmcvs4r3k67k7h5zuhm4el3qhlrxzhshtv", 1)
		group.SetMemberPower("g14u5eaheavy0ux4dmpykg2gvxpvqvexm9cyg58a", 1)
		group.SetMemberPower("g1ckn395mpttp0vupgtratyufdaakgh8jgkmr3ym", 1)
		group.SetMemberPower(std.GetOrigCaller(), 1)
		return group
	}

	// TODO: consider using factories that return multiple modules and handlers

	proposalModulesFactories := []dao_interfaces.ProposalModuleFactory{
		func(core dao_interfaces.IDAOCore) dao_interfaces.IProposalModule {
			tt := proposal_single.PercentageThresholdPercent(100)
			tq := proposal_single.PercentageThresholdPercent(100)
			return proposal_single.NewDAOProposalSingle(core, &proposal_single.DAOProposalSingleOpts{
				MaxVotingPeriod: dao_utils.DurationTime(time.Hour * 24 * 42),
				Threshold: &proposal_single.ThresholdThresholdQuorum{
					Threshold: &tt, // 1%
					Quorum:    &tq, // 1%
				},
			})
		},
	}

	messageHandlersFactories := []dao_interfaces.MessageHandlerFactory{
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			return group.UpdateMembersHandler()
		},
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			// TODO: add a router to support multiple proposal modules
			propMod := core.ProposalModules()[0]
			return proposal_single.NewUpdateSettingsHandler(propMod.Module.(*proposal_single.DAOProposalSingle))
		},
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			return tori.NewMintToriHandler()
		},
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			return tori.NewBurnToriHandler()
		},
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			return tori.NewChangeAdminHandler()
		},
	}

	daoCore = dao_core.NewDAOCore(votingModuleFactory, proposalModulesFactories, messageHandlersFactories)

	// dao_registry.Register(func() dao_interfaces.IDAOCore { return daoCore }, "DAO Realm", "Default testing DAO", "https://images.unsplash.com/photo-1477959858617-67f85cf4f1df?w=1080&fit=max")
}

// FIXME: the registry is currently broken in 'gno test', see https://github.com/gnolang/gno/issues/1852
// so we're exposing a way to register after DAO instantiation
func RegisterSelf() {
	if registered {
		panic("already registered")
	}

	dao_registry.Register(func() dao_interfaces.IDAOCore { return daoCore }, "DAO Realm", "Default testing DAO", "https://images.unsplash.com/photo-1477959858617-67f85cf4f1df?w=1080&fit=max")
	registered = true
}

func Render(path string) string {
	return daoCore.Render(path)
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
