package social_feeds_dao

import (
	"encoding/base64"
	"std"
	"strings"
	"time"

	dao_core "gno.land/p/demo/daodao/core"
	dao_interfaces "gno.land/p/demo/daodao/interfaces"
	"gno.land/p/demo/daodao/proposal_single"
	"gno.land/p/demo/daodao/voting_group"
	"gno.land/p/demo/ujson"
	"gno.land/r/demo/dao_registry"
	"gno.land/r/demo/groups"
	social_feeds "gno.land/r/demo/social_feeds"
)

var (
	daoCore      dao_interfaces.IDAOCore
	mainFeedName = "teritori"
	groupName    = "social_feeds_dao_voting_group"
	groupID      groups.GroupID
	feedID       social_feeds.FeedID
)

func init() {
	feedID = social_feeds.CreateFeed(mainFeedName)

	votingModuleFactory := func(core dao_interfaces.IDAOCore) dao_interfaces.IVotingModule {
		groupID = groups.CreateGroup(groupName)
		groups.AddMember(groupID, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", 1, "")
		groups.AddMember(groupID, "g108cszmcvs4r3k67k7h5zuhm4el3qhlrxzhshtv", 1, "")
		return dao_voting_group.NewVotingGroup(groupID)
	}

	proposalModulesFactories := []dao_interfaces.ProposalModuleFactory{
		func(core dao_interfaces.IDAOCore) dao_interfaces.IProposalModule {
			tt := dao_proposal_single.PercentageThresholdPercent(1500)
			tq := dao_proposal_single.PercentageThresholdPercent(5000)
			return dao_proposal_single.NewDAOProposalSingle(core, &dao_proposal_single.DAOProposalSingleOpts{
				MaxVotingPeriod: time.Second * 86400,
				Threshold: &dao_proposal_single.ThresholdThresholdQuorum{
					Threshold: &tt, // 15%
					Quorum:    &tq, // 50%
				},
			})
		},
	}

	messagesHandlersFactories := []dao_interfaces.MessageHandlerFactory{
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			return social_feeds.NewBanPostHandler()
		},
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			return dao_proposal_single.NewUpdateSettingsHandler(core.ProposalModules()[0].Module.(*dao_proposal_single.DAOProposalSingle))
		},
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			return groups.NewAddMemberHandler()
		},
		func(core dao_interfaces.IDAOCore) dao_interfaces.MessageHandler {
			return groups.NewDeleteMemberHandler()
		},
	}

	daoCore = dao_core.NewDAOCore(votingModuleFactory, proposalModulesFactories, messagesHandlersFactories)
	dao_registry.Register("Social Feed Moderation", "DAO responsible for Teritori social feed moderation", "ipfs://bafkreieqcwmjcb64r42ygs6a4dswz63djzgayjn3rhzjber3e42cknawlm")
}

func Render(path string) string {
	return "[[feed](/r/demo/social_feeds:" + mainFeedName + ")]\n\n" + daoCore.Render(path)
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
