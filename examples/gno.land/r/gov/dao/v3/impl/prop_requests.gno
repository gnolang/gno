package impl

import (
	"chain/runtime/unsafe"
	"strings"

	"gno.land/p/aeddi/panictoerr"
	"gno.land/p/moul/md"
	"gno.land/p/nt/markdown/sanitize/v0"
	trs_pkg "gno.land/p/nt/treasury/v0"
	"gno.land/p/nt/ufmt/v0"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
	"gno.land/r/gov/dao/v3/treasury"
)

func NewChangeLawRequest(cur realm, newLaw Law) dao.ProposalRequest {
	member, _ := memberstore.Get(0, cur).GetMember(unsafe.OriginCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	cb := func(cur realm) error {
		law = &newLaw
		return nil
	}

	e := dao.NewSimpleExecutor(0, cur, cb, ufmt.Sprintf("A new Law is proposed:\n %v", newLaw))

	return dao.NewProposalRequest("Change Law Proposal", "This proposal is looking to change the actual govDAO Law", e)
}

func NewUpgradeDaoImplRequest(cur realm, newDao dao.DAO, realmPkg, reason string) dao.ProposalRequest {
	// Rejected here as well as in UpdateImpl so the mistake surfaces when the
	// proposal is written, not when it executes. An empty realmPkg would be
	// stored as an allowlist entry matching every user realm's empty
	// PkgPath(), and would render as a blank name in the grant sentence below.
	if strings.TrimSpace(realmPkg) == "" {
		panic("realmPkg must be the realm path being granted govDAO authority")
	}
	// Padding is rejected for the same reason UpdateImpl rejects it: the entry
	// is stored as given and matched whole, so a padded path grants nobody
	// anything. Caught here so the author sees it, not the voters.
	if realmPkg != strings.TrimSpace(realmPkg) {
		panic("realmPkg must not have leading or trailing spaces")
	}

	member, _ := memberstore.Get(0, cur).GetMember(unsafe.OriginCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	cb := func(cur realm) error {
		// dao.UpdateImpl() must be cross-called from v3/impl but
		// what calls this cb function is r/gov/dao.
		// therefore we must cross back into v3/impl and then
		// cross call dao.UpdateRequest().
		dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(newDao, []string{"gno.land/r/gov/dao/v3/impl", realmPkg}))
		return nil
	}

	// State the grant: this executor rewrites AllowedDAOs, and previously
	// rendered nothing at all. realmPkg is caller-supplied and lands in a code
	// span, so it is wrapped with sanitize.InlineCode — which emits its own
	// fence, so do not add backticks around it. (md.EscapeText is wrong here:
	// CommonMark 6.1 does not process backslash escapes inside code spans.)
	// Nothing ties realmPkg to newDao, which is why the text says so.
	e := dao.NewSimpleExecutor(0, cur, cb, ufmt.Sprintf(
		"Replaces the govDAO implementation with the DAO carried by this proposal.\n\n"+
			"After execution, only `gno.land/r/gov/dao/v3/impl` and %s may replace the "+
			"implementation, mutate the member store, or move treasury funds.\n\n"+
			"This grant is stated by the proposer: the incoming implementation is passed "+
			"as a value and is not verified to belong to the realm named above.",
		sanitize.InlineCode(realmPkg)))

	return dao.NewProposalRequest("Change DAO implementation", "This proposal is looking to change the actual govDAO implementation. Reason: "+reason, e)
}

func NewAddMemberRequest(cur realm, addr address, tier string, portfolio string) dao.ProposalRequest {
	// Reject a non-bech32 address at proposal-build time (see AddMember): keeps
	// an unauthenticable / injection-bearing key out of the member store.
	if !addr.IsValid() {
		panic("invalid member address: " + addr.String())
	}
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

	member, _ := memberstore.Get(0, cur).GetMember(unsafe.OriginCaller())
	if member == nil {
		panic("proposer is not a member")
	}

	if member.InvitationPoints <= 0 {
		panic("proposer does not have enough invitation points for inviting new people to the board")
	}

	cb := func(cur realm) error {
		member.RemoveInvitationPoint()
		err := memberstore.Get(0, cur).SetMember(tier, addr, memberByTier(tier))

		return err
	}

	e := dao.NewSimpleExecutor(0, cur, cb, ufmt.Sprintf("A new member with address %v is proposed to be on tier %v. Provided Portfolio information:\n\n%v", addr, tier, portfolio))

	name := tryResolveAddr(addr)
	return dao.NewProposalRequestWithFilter(
		ufmt.Sprintf("New %s Member Proposal", tier),
		ufmt.Sprintf("This is a proposal to add `%s` to **%s**.\n#### `%s`'s Portfolio:\n\n%s\n", name, tier, name, portfolio),
		e,
		FilterByTier{Tier: tier},
	)
}

func NewWithdrawMemberRequest(cur realm, addr address, reason string) dao.ProposalRequest {
	member, tier := memberstore.Get(0, cur).GetMember(addr)
	if member == nil {
		panic("user we want to remove not found")
	}

	reason = strings.TrimSpace(reason)
	if tier == memberstore.T1 && reason == "" {
		panic("T1 user removals must contains a reason.")
	}

	cb := func(cur realm) error {
		memberstore.Get(0, cur).RemoveMember(addr)
		return nil
	}

	e := dao.NewSimpleExecutor(0, cur, cb, ufmt.Sprintf("Member with address %v will be withdrawn.\n\n REASON: %v.", addr, reason))

	return dao.NewProposalRequest(
		"Member Withdrawal Proposal",
		ufmt.Sprintf("This is a proposal to remove %s from the GovDAO", tryResolveAddr(addr)),
		e,
	)
}

func NewPromoteMemberRequest(cur realm, addr address, fromTier string, toTier string) dao.ProposalRequest {
	cb := func(cur realm) error {
		// Fail on an unknown destination tier before touching the member.
		//
		// This does NOT fix a data-loss bug: an unknown tier already panicked a
		// few lines down, because memberByTier's default case panics and it is
		// evaluated as SetMember's third argument. A panic is reverted -- baseapp
		// commits message writes only when the result is OK -- so the earlier
		// RemoveMember never stuck. Checking here only moves the failure earlier
		// and names the actual cause.
		//
		// The real remove-then-add hazard is still open and this guard does not
		// close it. SetMember RETURNS errors rather than panicking, for a member
		// that already exists and for a tier missing from the per-store index,
		// and a returned executor error does not revert: ExecuteOrRejectProposal
		// marks the proposal rejected and the transaction still commits. Those
		// paths would leave the member removed and not re-added. GetTier below
		// reads the global tier table, which is not the index SetMember checks.
		if _, ok := memberstore.GetTier(toTier); !ok {
			panic("unknown destination tier: " + toTier)
		}

		prevTier := memberstore.Get(0, cur).RemoveMember(addr)
		if prevTier == "" {
			panic("member not found, so cannot be promoted")
		}

		if prevTier != fromTier {
			panic("previous tier changed from the one indicated in the proposal")
		}

		err := memberstore.Get(0, cur).SetMember(toTier, addr, memberByTier(toTier))

		return err
	}

	e := dao.NewSimpleExecutor(0, cur, cb, ufmt.Sprintf("A new member with address %v will be promoted from tier %v to tier %v.", addr, fromTier, toTier))

	return dao.NewProposalRequestWithFilter(
		"Member Promotion Proposal",
		ufmt.Sprintf("This is a proposal to promote %s from **%s** to **%s**.", tryResolveAddr(addr), fromTier, toTier),
		e,
		FilterByTier{Tier: toTier},
	)
}

func NewTreasuryPaymentRequest(cur realm, payment trs_pkg.Payment, reason string) dao.ProposalRequest {
	if !treasury.HasBanker(payment.BankerID()) {
		panic("banker not registered in treasury with ID: " + payment.BankerID())
	}

	reason = strings.TrimSpace(reason)
	if reason == "" {
		panic("treasury payment request requires a reason")
	}

	cb := func(cur realm) error {
		return panictoerr.PanicToError(func() {
			treasury.Send(cross(cur), payment)
		})
	}

	e := dao.NewSimpleExecutor(0, cur,
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
func NewTreasuryGRC20TokensUpdate(cur realm, newTokenKeys []string) dao.ProposalRequest {
	if len(newTokenKeys) == 0 {
		panic("the list of new tokens is empty")
	}

	cb := func(cur realm) error {
		return panictoerr.PanicToError(func() {
			// NOTE:: Consider checking if the newTokenKeys are already registered
			// in the grc20reg before updating the treasury tokens keys.
			treasury.SetTokenKeys(cross(cur), newTokenKeys)
		})
	}

	bulletList := md.BulletList(newTokenKeys)

	e := dao.NewSimpleExecutor(0, cur,
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
		return memberstore.NewMember(t.InvitationPoints)
	case memberstore.T2:
		t, _ := memberstore.GetTier(memberstore.T2)
		return memberstore.NewMember(t.InvitationPoints)
	case memberstore.T3:
		t, _ := memberstore.GetTier(memberstore.T3)
		return memberstore.NewMember(t.InvitationPoints)
	default:
		panic("member not found by the specified tier")
	}
}
