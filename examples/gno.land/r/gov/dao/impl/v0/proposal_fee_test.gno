package impl

import (
	"chain"
	"chain/banker"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/memberstore/v0"
)

// payer is never a member: TestAbstainVote leaves noMember on T2.
var payer = testutils.TestAddress("fee payer")

// propose calls ProposeAddMember as a msg call from proposer sending sent.
func propose(cur realm, proposer address, sent chain.Coins) dao.ProposalID {
	testing.IssueCoins(chain.PackageAddress("gno.land/r/gov/dao/impl/v0"), sent)
	testing.SetOriginCaller(proposer)
	testing.SetRealm(testing.NewUserRealm(proposer))
	testing.SetOriginSend(sent)
	return ProposeAddMember(cross(cur), payer, memberstore.T2, "portfolio")
}

func TestNonMemberProposesAMemberForTheFee(cur realm, t *testing.T) {
	loadMembers(cur)
	fee := chain.Coins{chain.NewCoin("ugnot", ProposalFeeAmount)}

	before := banker.NewReadonlyBanker().GetCoin(treasuryAddr, "ugnot")
	pid := propose(cur, payer, fee)
	after := banker.NewReadonlyBanker().GetCoin(treasuryAddr, "ugnot")
	uassert.Equal(t, ProposalFeeAmount, after-before)
	p, err := dao.GetProposal(pid)
	urequire.NoError(t, err)
	uassert.Equal(t, payer, p.Author())
}

func TestNonMemberOpensNothingElse(cur realm, t *testing.T) {
	loadMembers(cur)

	testing.SetOriginCaller(payer)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/impl/v0"))
	req := NewWithdrawMemberRequest(cur, m4, "inactive")
	testing.SetRealm(testing.NewUserRealm(payer))
	_, err := dao.CreateProposal(cross(cur), req)
	urequire.ErrorContains(t, err, "only members can create new proposals")
}

func TestProposeAddMemberRefusesAnythingButTheExactFee(cur realm, t *testing.T) {
	loadMembers(cur)

	for _, sent := range []chain.Coins{
		nil,
		{chain.NewCoin("ugnot", ProposalFeeAmount-1)},
		{chain.NewCoin("ugnot", ProposalFeeAmount+1)},
		{chain.NewCoin("foo", ProposalFeeAmount)},
		{chain.NewCoin("foo", 1), chain.NewCoin("ugnot", ProposalFeeAmount)},
	} {
		uassert.AbortsContains(t, cur, "send exactly 1000000ugnot", func(cur realm) {
			propose(cur, payer, sent)
		})
	}
}

func TestProposeAddMemberIsFreeForMembers(cur realm, t *testing.T) {
	loadMembers(cur)

	uassert.NotAborts(t, cur, func(cur realm) {
		propose(cur, m1, nil)
	})
	uassert.AbortsContains(t, cur, "no proposal fee is due", func(cur realm) {
		propose(cur, m1, chain.Coins{chain.NewCoin("ugnot", ProposalFeeAmount)})
	})
}

func TestProposeAddMemberRefusesIndirectCalls(cur realm, t *testing.T) {
	loadMembers(cur)

	// Only a direct msg call lands the coins here; the txtar covers maketx run.
	testing.SetOriginCaller(payer)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/middleman"))
	testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", ProposalFeeAmount)})
	uassert.AbortsContains(t, cur, "must be called directly", func(cur realm) {
		ProposeAddMember(cross(cur), payer, memberstore.T2, "portfolio")
	})
}

func TestZeroFeeLetsNonMembersProposeFree(cur realm, t *testing.T) {
	loadMembers(cur)
	prev := ProposalFeeAmount
	ProposalFeeAmount = 0
	defer func() { ProposalFeeAmount = prev }()

	uassert.NotAborts(t, cur, func(cur realm) {
		propose(cur, payer, nil)
	})
}

func TestProposalFeeAmountRequest(cur realm, t *testing.T) {
	loadMembers(cur)

	testing.SetOriginCaller(payer)
	uassert.AbortsContains(t, cur, "proposer is not a member", func(cur realm) {
		NewProposalFeeAmountRequest(cur, 2_000_000, "too cheap")
	})

	testing.SetOriginCaller(m1)
	uassert.AbortsContains(t, cur, "must not be negative", func(cur realm) {
		NewProposalFeeAmountRequest(cur, -1, "refund")
	})
	uassert.AbortsContains(t, cur, "requires a reason", func(cur realm) {
		NewProposalFeeAmountRequest(cur, 0, " ")
	})
	uassert.NotAborts(t, cur, func(cur realm) {
		NewProposalFeeAmountRequest(cur, 0, "free proposals")
	})
}
