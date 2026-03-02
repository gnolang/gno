package impl

import (
	"chain/runtime"
	"strings"

	"gno.land/p/aeddi/panictoerr"
	"gno.land/p/moul/md"
	trs_pkg "gno.land/p/nt/treasury/v0"
	"gno.land/p/nt/ufmt/v0"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
	"gno.land/r/gov/dao/v3/treasury"
)

func NewChangeLawRequest(_ realm, newLaw Law) dao.ProposalRequest {
	member, _ := memberstore.Get().GetMember(runtime.OriginCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	cb := func(_ realm) error {
		law = &newLaw
		return nil
	}

	e := dao.NewSimpleExecutor(cb, ufmt.Sprintf("A new Law is proposed:\n %v", newLaw))

	return dao.NewProposalRequest("Change Law Proposal", "This proposal is looking to change the actual govDAO Law", e)
}

func NewUpgradeDaoImplRequest(newDao dao.DAO, realmPkg, reason string) dao.ProposalRequest {
	member, _ := memberstore.Get().GetMember(runtime.OriginCaller())
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

func NewAddMemberRequest(_ realm, addr address, tier string, portfolio string) dao.ProposalRequest {
	_, ok := memberstore.GetTier(tier)
	if !ok {
		panic("provided tier does not exists")
	}

	if tier != memberstore.T1 && tier != memberstore.T2 {
		panic("Only T1 and T2 members can be added by proposal. To add a T3 member use AddMember function directly.")
	}

	if portfolio == "" {
		panic("A portfolio for the proposed member is required")
	}

	member, _ := memberstore.Get().GetMember(runtime.OriginCaller())
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

func NewWithdrawMemberRequest(_ realm, addr address, reason string) dao.ProposalRequest {
	member, tier := memberstore.Get().GetMember(addr)
	if member == nil {
		panic("user we want to remove not found")
	}

	reason = strings.TrimSpace(reason)
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

func NewPromoteMemberRequest(addr address, fromTier string, toTier string) dao.ProposalRequest {
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

func NewTreasuryPaymentRequest(payment trs_pkg.Payment, reason string) dao.ProposalRequest {
	if !treasury.HasBanker(payment.BankerID()) {
		panic("banker not registered in treasury with ID: " + payment.BankerID())
	}

	reason = strings.TrimSpace(reason)
	if reason == "" {
		panic("treasury payment request requires a reason")
	}

	cb := func(_ realm) error {
		return panictoerr.PanicToError(func() {
			treasury.Send(cross, payment)
		})
	}

	e := dao.NewSimpleExecutor(
		cb,
		ufmt.Sprintf(
			"A payment will be sent by the GovDAO treasury.\n\nReason: %s\n\nPayment: %s.",
			reason,
			payment.String(),
		),
	)

	return dao.NewProposalRequest(
		"Treasury Payment",
		ufmt.Sprintf(
			"This proposal is looking to send a payment using the treasury.\n\nReason: %s\n\nPayment: %s",
			reason,
			payment.String(),
		),
		e,
	)
}

// NewTreasuryGRC20TokensUpdate creates a proposal request to update the list of GRC20 tokens registry
// keys used by the treasury. The new list, if voted and accepted, will overwrite the current one.
func NewTreasuryGRC20TokensUpdate(newTokenKeys []string) dao.ProposalRequest {
	if len(newTokenKeys) == 0 {
		panic("the list of new tokens is empty")
	}

	cb := func(_ realm) error {
		return panictoerr.PanicToError(func() {
			// NOTE:: Consider checking if the newTokenKeys are already registered
			// in the grc20reg before updating the treasury tokens keys.
			treasury.SetTokenKeys(cross, newTokenKeys)
		})
	}

	bulletList := md.BulletList(newTokenKeys)

	e := dao.NewSimpleExecutor(
		cb,
		ufmt.Sprintf(
			"The list of GRC20 tokens used by the treasury will be updated.\n\nNew Token Keys:\n%s.\n",
			bulletList,
		),
	)

	return dao.NewProposalRequest(
		"Treasury GRC20 Tokens Update",
		ufmt.Sprintf(
			"This proposal is looking to update the list of GRC20 tokens used by the treasury.\n\nNew Token Keys:\n%s",
			bulletList,
		),
		e,
	)
}

func memberByTier(tier string) *memberstore.Member {
	switch tier {
	case memberstore.T1:
		t, _ := memberstore.GetTier(memberstore.T1)
		return &memberstore.Member{
			InvitationPoints: t.InvitationPoints,
		}
	case memberstore.T2:
		t, _ := memberstore.GetTier(memberstore.T2)
		return &memberstore.Member{
			InvitationPoints: t.InvitationPoints,
		}
	case memberstore.T3:
		t, _ := memberstore.GetTier(memberstore.T3)
		return &memberstore.Member{
			InvitationPoints: t.InvitationPoints,
		}
	default:
		panic("member not found by the specified tier")
	}
}
