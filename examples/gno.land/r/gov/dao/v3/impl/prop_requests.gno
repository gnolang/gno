package impl

import (
	"std"

	"gno.land/p/aeddi/panictoerr"
	"gno.land/p/demo/ufmt"
	trs_pkg "gno.land/p/nt/treasury"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
	"gno.land/r/gov/dao/v3/treasury"
)

func NewChangeLawRequest(_ realm, newLaw *Law) dao.ProposalRequest {
	member, _ := memberstore.Get().GetMember(std.OriginCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	cb := func(_ realm) error {
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

	cb := func(_ realm) error {
		// dao.UpdateImpl() must be cross-called from v3/impl but
		// what calls this cb function is r/gov/dao.
		// therefore we must cross back into v3/impl and then
		// cross call dao.UpdateRequest().
		dao.UpdateImpl(cross, dao.UpdateRequest{
			DAO:         newDao,
			AllowedDAOs: []string{"gno.land/r/gov/dao/v3/impl", realmPkg}, // keeping previous realm just in case something went wrong
		})
		return nil
	}

	e := dao.NewSimpleExecutor(cb, "")

	return dao.NewProposalRequest("Change DAO implementation", "This proposal is looking to change the actual govDAO implementation. Reason: "+reason, e)
}

func NewAddMemberRequest(_ realm, addr std.Address, tier string, portfolio string) dao.ProposalRequest {
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

	cb := func(_ realm) error {
		member.RemoveInvitationPoint()
		err := memberstore.Get().SetMember(tier, addr, memberByTier(tier))

		return err
	}

	e := dao.NewSimpleExecutor(cb, ufmt.Sprintf("A new member with address %v is proposed to be on tier %v. Provided Portfolio information:\n\n%v", addr, tier, portfolio))

	name := tryResolveAddr(addr)
	return dao.NewProposalRequestWithFilter(
		ufmt.Sprintf("New %s Member Proposal", tier),
		ufmt.Sprintf("This is a proposal to add `%s` to **%s**.\n#### `%s`'s Portfolio:\n\n%s\n", name, tier, name, portfolio),
		e,
		FilterByTier{Tier: tier},
	)
}

func NewWithdrawMemberRequest(_ realm, addr std.Address, reason string) dao.ProposalRequest {
	member, tier := memberstore.Get().GetMember(addr)
	if member == nil {
		panic("user we want to remove not found")
	}

	if tier == memberstore.T1 && reason == "" {
		panic("T1 user removals must contains a reason.")
	}

	cb := func(_ realm) error {
		memberstore.Get().RemoveMember(addr)
		return nil
	}

	e := dao.NewSimpleExecutor(cb, ufmt.Sprintf("Member with address %v will be withdrawn.\n\n REASON: %v.", addr, reason))

	return dao.NewProposalRequest(
		"Member Withdrawal Proposal",
		ufmt.Sprintf("This is a proposal to remove %s from the GovDAO", tryResolveAddr(addr)),
		e,
	)
}

func NewPromoteMemberRequest(addr std.Address, fromTier string, toTier string) dao.ProposalRequest {
	cb := func(_ realm) error {
		prevTier := memberstore.Get().RemoveMember(addr)
		if prevTier == "" {
			panic("member not found, so cannot be promoted")
		}

		if prevTier != fromTier {
			panic("previous tier changed from the one indicated in the proposal")
		}

		err := memberstore.Get().SetMember(toTier, addr, memberByTier(toTier))

		return err
	}

	e := dao.NewSimpleExecutor(cb, ufmt.Sprintf("A new member with address %v will be promoted from tier %v to tier %v.", addr, fromTier, toTier))

	return dao.NewProposalRequestWithFilter(
		"Member Promotion Proposal",
		ufmt.Sprintf("This is a proposal to promote %s from **%s** to **%s**.", tryResolveAddr(addr), fromTier, toTier),
		e,
		FilterByTier{Tier: toTier},
	)
}

func NewTreasuryPaymentRequest(_ realm, payment trs_pkg.Payment, reason string) dao.ProposalRequest {
	if !treasury.HasBanker(payment.BankerID()) {
		panic("banker not registered in treasury with ID: " + payment.BankerID())
	}

	if reason == "" {
		panic("treasury payment request requires a reason")
	}

	cb := func(_ realm) error {
    return panictoerr.PanicToError(func() {
      treasury.Send(cross, payment)
    })
	}

	// NOTE: Should this be a NewSafeExecutor instead?
	e := dao.NewSimpleExecutor(
		cb,
		ufmt.Sprintf(
			"A payment will be sent by the GovDAO treasury.\n\nReason: %s\n\nPayment: %s.",
			reason,
			payment.String(),
		),
	)

	// NOTE: Should this be filtered by tier?
	return dao.NewProposalRequest(
		"Treasury Payment",
		"This proposal is looking to send a payment using the treasury.",
		e,
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
