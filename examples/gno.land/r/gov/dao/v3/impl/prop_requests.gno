package impl

import (
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
)

func NewChangeLawRequest(newLaw *Law) dao.ProposalRequest {
	member, _ := memberstore.Get().GetMember(std.OriginCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	cb := func() error {
		law = newLaw

		return nil
	}

	e := dao.NewSimpleExecutor(cb, ufmt.Sprintf("A new Law is proposed:\n %v", newLaw))

	return dao.NewProposalRequest("Change Law Proposal", "This proposal is looking to change the actual govDAO Law", e)
}

func NewUpgradeDaoImplRequest(newDao dao.DAO, realmPkg, reason string) dao.ProposalRequest {
	member, _ := memberstore.Get().GetMember(std.OriginCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	cb := func() error {
		dao.UpdateImpl(dao.UpdateRequest{
			DAO:         newDao,
			AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl", realmPkg}, // keeping previous realm just in case something went wrong
		})

		return nil
	}

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest("Change DAO implementation", "This proposal is looking to change the actual govDAO implementation. Reason: "+reason, e)
}

func NewAddMemberRequest(addr std.Address, tier string, portfolio string) dao.ProposalRequest {
	_, ok := memberstore.Tiers.GetTier(tier)
	if !ok {
		panic("provided tier does not exists")
	}

	if tier != memberstore.T1 && tier != memberstore.T2 {
		panic("Only T1 and T2 members can be added by proposal. To add a T3 member use AddMember function directly.")
	}

	if portfolio == "" {
		panic("A portfolio for the proposed member is required")
	}

	member, _ := memberstore.Get().GetMember(std.OriginCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	if member.InvitationPoints <= 0 {
		panic("proposer does not have enough invitation points for inviting new people to the board")
	}

	cb := func() error {
		member.RemoveInvitationPoint()

		return memberstore.Get().SetMember(tier, addr, memberByTier(tier))
	}

	e := dao.NewSimpleExecutor(cb, ufmt.Sprintf("A new member with address %v is proposed to be on tier %v. Provided Portfolio information:\n\n%v", addr, tier, portfolio))

	return dao.NewProposalRequestWithFilter(
		"New Member Proposal",
		"This proposal is looking to add a new member to the board.",
		e,
		FilterByTier{Tier: tier},
	)
}

func NewWithdrawMemberRequest(addr std.Address, reason string) dao.ProposalRequest {
	member, tier := memberstore.Get().GetMember(addr)
	if member == nil {
		panic("user we want to remove not found")
	}

	if tier == memberstore.T1 && reason == "" {
		panic("T1 user removals must contains a reason.")
	}

	cb := func() error {
		memberstore.Get().RemoveMember(addr)

		return nil
	}

	e := dao.NewSimpleExecutor(cb, ufmt.Sprintf("Member with address %v will be withdrawn.\n\n REASON: %v.", addr, reason))

	return dao.NewProposalRequest("Member Withdrawn", "This proposal is looking to remove a member from the board.", e)
}

func NewPromoteMemberRequest(addr std.Address, fromTier string, toTier string) dao.ProposalRequest {
	cb := func() error {
		prevTier := memberstore.Get().RemoveMember(addr)
		if prevTier == "" {
			panic("member not found, so cannot be promoted")
		}

		if prevTier != fromTier {
			panic("previous tier changed from the one indicated in the proposal")
		}

		return memberstore.Get().SetMember(toTier, addr, memberByTier(toTier))
	}

	e := dao.NewSimpleExecutor(cb, ufmt.Sprintf("A new member with address %v will be promoted from tier %v to tier %v.", addr, fromTier, toTier))

	return dao.NewProposalRequestWithFilter(
		"Member Promotion",
		"This proposal is looking to promote a member to an upper tier.",
		e,
		FilterByTier{Tier: toTier},
	)
}

func memberByTier(tier string) *memberstore.Member {
	switch tier {
	case memberstore.T1:
		t, _ := memberstore.Tiers.GetTier(memberstore.T1)
		return &memberstore.Member{
			InvitationPoints: t.InvitationPoints,
		}
	case memberstore.T2:
		t, _ := memberstore.Tiers.GetTier(memberstore.T2)
		return &memberstore.Member{
			InvitationPoints: t.InvitationPoints,
		}
	case memberstore.T3:
		t, _ := memberstore.Tiers.GetTier(memberstore.T3)
		return &memberstore.Member{
			InvitationPoints: t.InvitationPoints,
		}
	default:
		panic("member not found by the specified tier")
	}
}
