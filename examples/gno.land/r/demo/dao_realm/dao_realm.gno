package dao_realm

import (
	"encoding/base64"
	"std"
	"strings"
	"time"

	dao_core "gno.land/p/demo/daodao/core"
	dao_interfaces "gno.land/p/demo/daodao/interfaces"
	"gno.land/p/demo/daodao/proposal_single"
	"gno.land/p/demo/daodao/voting_group"
	"gno.land/r/demo/groups"
	modboards "gno.land/r/demo/modboards"
)

var (
	daoCore       dao_core.IDAOCore
	registry      = dao_interfaces.NewMessagesRegistry()
	mainBoardName = "foo_dao_3"
	groupID       groups.GroupID
)

func init() {
	groupID = groups.CreateGroup(mainBoardName)
	groups.AddMember(groupID, "g1747t5m2f08plqjlrjk2q0qld7465hxz8gkx59c", 1, "")
	groups.AddMember(groupID, "g108cszmcvs4r3k67k7h5zuhm4el3qhlrxzhshtv", 1, "")
	groups.AddMember(groupID, "g14u5eaheavy0ux4dmpykg2gvxpvqvexm9cyg58a", 1, "")
	groups.AddMember(groupID, "g1ckn395mpttp0vupgtratyufdaakgh8jgkmr3ym", 1, "")
	registry.Register(groups.NewAddMemberHandler())
	registry.Register(groups.NewDeleteMemberHandler())

	daoCore = dao_core.NewDAOCore(dao_voting_group.NewGRC4Voting(groupID), nil)

	tt := dao_interfaces.Percent(100) // 1%
	tq := dao_interfaces.Percent(100) // 1%
	proposalMod := dao_proposal_single.NewDAOProposalSingle(daoCore, &dao_proposal_single.DAOProposalSingleOpts{
		MaxVotingPeriod: time.Hour * 24 * 42,
		Threshold: dao_interfaces.Threshold{ThresholdQuorum: &dao_interfaces.ThresholdQuorum{
			Threshold: dao_interfaces.PercentageThreshold{Percent: &tt},
			Quorum:    dao_interfaces.PercentageThreshold{Percent: &tq},
		}},
		Registry: registry,
	})
	// TODO: add a router to support multiple proposal modules
	registry.Register(dao_proposal_single.NewUpdateSettingsHandler(proposalMod))
	daoCore.AddProposalModule(proposalMod)

	registry.Register(modboards.NewCreateBoardHandler())
	registry.Register(modboards.NewDeletePostHandler())
	modboards.CreateBoard(mainBoardName)
}

func Render(path string) string {
	return "[[board](/r/demo/modboards:" + mainBoardName + ")]\n\n" + daoCore.Render(path)
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
