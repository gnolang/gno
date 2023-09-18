package social_feeds_dao

import (
	"encoding/base64"
	"std"
	"strings"
	"time"

	dao_core "gno.land/p/demo/daodao/core_v6"
	dao_interfaces "gno.land/p/demo/daodao/interfaces"
	"gno.land/p/demo/daodao/proposal_single_v7"
	"gno.land/p/demo/daodao/voting_group_v8"
	"gno.land/p/demo/ujson"
	"gno.land/r/demo/dao_registry_v5"
	"gno.land/r/demo/groups_v9"
	social_feeds "gno.land/r/demo/social_feeds"
)

var (
	daoCore      dao_core.IDAOCore
	registry     = dao_interfaces.NewMessagesRegistry()
	mainFeedName = "teritori9"
	groupID      groups.GroupID
)

func init() {
	// NOTE: why creator is: g16lcthjkpr9emfj2yemnsfmyc563jw98p4wwvgg ??????
	groupID = groups.CreateGroup(mainFeedName)

	// Create Feed
	currentRealm := std.CurrentRealm()
	feedID := social_feeds.CreateFeed(mainFeedName)

	// For testing purpose we add 2 users
	groups.AddMember(groupID, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", 1, "")

	registry.Register(groups.NewAddMemberHandler())
	registry.Register(groups.NewDeleteMemberHandler())

	daoCore = dao_core.NewDAOCore(dao_voting_group.NewGRC4Voting(groupID), nil)

	tt := dao_interfaces.Percent(1500) // 15%
	tq := dao_interfaces.Percent(5000) // 50%
	proposalMod := dao_proposal_single.NewDAOProposalSingle(daoCore, &dao_proposal_single.DAOProposalSingleOpts{
		MaxVotingPeriod: time.Second * 86400,
		Threshold: dao_interfaces.Threshold{ThresholdQuorum: &dao_interfaces.ThresholdQuorum{
			Threshold: dao_interfaces.PercentageThreshold{Percent: &tt},
			Quorum:    dao_interfaces.PercentageThreshold{Percent: &tq},
		}},
		Registry: registry,
	})
	// TODO: add a router to support multiple proposal modules
	registry.Register(dao_proposal_single.NewUpdateSettingsHandler(proposalMod))
	daoCore.AddProposalModule(proposalMod)

	registry.Register(social_feeds.NewBanPostHandler())
	dao_registry.Register("social_feeds_dao", "Teritori Feed DAO", "https://avatars.githubusercontent.com/u/108656591?s=200&v=4")
}

func Render(path string) string {
	return "[[feed](/r/demo/social_feeds:" + mainFeedName + ")]\n\n" + daoCore.Render(path)
}

func GetCore() dao_core.IDAOCore {
	return daoCore
}

func Vote(moduleIndex int, proposalID int, vote dao_interfaces.Vote, rationale string) {
	dao_core.GetProposalModule(daoCore, moduleIndex).Vote(proposalID, vote, rationale)
}

func Execute(moduleIndex int, proposalID int) {
	dao_core.GetProposalModule(daoCore, moduleIndex).Execute(proposalID)
}

func Propose(moduleIndex int, title string, description string, b64Messages string) {
	mod := dao_core.GetProposalModule(daoCore, moduleIndex)
	var messages []dao_interfaces.ExecutableMessage
	if len(b64Messages) != 0 {
		rawMessages := strings.Split(b64Messages, ",")
		for _, rawMessage := range rawMessages {
			message := registry.FromBase64String(rawMessage)
			messages = append(messages, message)
		}
	}
	mod.Propose(title, description, messages)
}

func GetBinaryMembers() string {
	members := groups.GetMembers(groupID)
	ss := []string{}
	for _, member := range members {
		ss = append(ss, base64.RawURLEncoding.EncodeToString(member.Bytes()))
	}
	return strings.Join(ss, ",")
}

func GetProposalsJSON(moduleIndex int) string {
	return ujson.FormatSlice(dao_core.GetProposalModule(daoCore, moduleIndex).Proposals())
}

func GetGroupID() uint64 {
	return uint64(groupID)
}
