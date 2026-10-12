package impl

import (
	"chain"
	"chain/banker"
	"chain/runtime/unsafe"

	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/gov/dao/memberstore/v0"
)

var (
	law    *Law
	govDAO *GovDAO = NewGovDAO()
)

func init() {
	law = &Law{
		Supermajority: 66.66, // Two thirds
	}
}

func Render(cur realm, in string) string {
	// Same-realm: pass cur to govDAO.Render (also crossing, but same realm
	// so use the literal cur form rather than cross(cur)).
	return govDAO.Render(cur, cur.PkgPath(), in)
}

// AddMember allows T1 and T2 members to freely add T3 members using their invitation points.
func AddMember(cur realm, addr address) {
	// AGENTS.md: check IsCurrent() before cur.Previous() in a crossing function.
	if !cur.IsCurrent() {
		panic("AddMember: realm value is not the caller's live cur")
	}
	// address args are not VM-validated (the raw MsgCall string is stored), so
	// reject a non-bech32 addr here — matching InitWithUsers. This keeps the
	// member store free of unauthenticable keys and of markdown/pipe/HTML
	// metachars that would otherwise inject into the members render table.
	if !addr.IsValid() {
		panic("invalid member address: " + addr.String())
	}
	caller := cur.Previous()
	if !caller.IsUser() {
		panic("this function must be called by an EOA through msg call or msg run")
	}
	m, t := memberstore.Get(0, cur).GetMember(caller.Address())
	if m == nil {
		panic("caller is not a member")
	}

	if t != memberstore.T1 && t != memberstore.T2 {
		panic("caller is not on T1 or T2. To add members, propose them through proposals")
	}

	m.RemoveInvitationPoint()

	if err := memberstore.Get(0, cur).SetMember(memberstore.T3, addr, memberByTier(memberstore.T3)); err != nil {
		panic(err.Error())
	}
}

var treasuryAddr = chain.PackageAddress("gno.land/r/gov/dao/treasury/v0")

// PayProposalFee takes ProposalFeeAmount from a non-member into the govDAO
// treasury, and the payer spends it on their next proposal. It must be a
// direct MsgCall: a maketx run sends its coins back to the signer, so its
// OriginSend proves no payment.
func PayProposalFee(cur realm) {
	caller := cur.Previous()
	if !caller.IsUserCall() {
		panic("the proposal fee must be paid with a direct msg call")
	}
	if ProposalFeeAmount == 0 {
		panic("proposals are free, there is no fee to pay")
	}
	if m, _ := memberstore.Get(0, cur).GetMember(caller.Address()); m != nil {
		panic("members propose for free")
	}
	sent := unsafe.OriginSend()
	if len(sent) != 1 || sent[0].Denom != "ugnot" || sent[0].Amount != ProposalFeeAmount {
		panic(ufmt.Sprintf("send exactly %dugnot to pay the proposal fee", ProposalFeeAmount))
	}
	banker.NewBanker(banker.BankerTypeRealmSend, cur).SendCoins(cur.Address(), treasuryAddr, sent)
	proposalCredits[caller.Address()]++
}

// GetInstance returns the singleton *GovDAO. Only the loader realm may
// call it (used during the bootstrap UpdateImpl handoff). The
// IsCurrent() check rejects stale or stashed realm values; PkgPath()
// after the check is the authentic immediate caller.
func GetInstance(_ int, rlm realm) *GovDAO {
	if !rlm.IsCurrent() {
		panic("GetInstance: rlm is not the caller's live cur (stale capture or sibling frame)")
	}
	if rlm.PkgPath() != "gno.land/r/gov/dao/loader/v0" {
		panic("not allowed")
	}

	return govDAO
}
