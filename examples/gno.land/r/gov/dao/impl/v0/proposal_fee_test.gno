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

const feeErr = "non-members must first pay the 1000000ugnot proposal fee with PayProposalFee"

// payer is never a member: TestAbstainVote leaves noMember on T2.
var payer = testutils.TestAddress("fee payer")

// payFee issues sent to the realm as a msg call would, then pays with it.
func payFee(cur realm, payer address, sent chain.Coins) {
	testing.IssueCoins(chain.PackageAddress("gno.land/r/gov/dao/impl/v0"), sent)
	testing.SetOriginCaller(payer)
	testing.SetRealm(testing.NewUserRealm(payer))
	testing.SetOriginSend(sent)
	PayProposalFee(cross(cur))
}

func createAsUser(cur realm, author address) error {
	testing.SetOriginCaller(author)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/gov/dao/impl/v0"))
	req := NewAddMemberRequest(cur, author, memberstore.T2, "portfolio")
	testing.SetRealm(testing.NewUserRealm(author))
	_, err := dao.CreateProposal(cross(cur), req)
	return err
}

func TestNonMemberSpendsOnePaidFeePerProposal(cur realm, t *testing.T) {
	loadMembers(cur)
	proposalCredits = map[address]int{}
	fee := chain.Coins{chain.NewCoin("ugnot", ProposalFeeAmount)}

	urequire.ErrorContains(t, createAsUser(cur, payer), feeErr)

	before := banker.NewReadonlyBanker().GetCoin(treasuryAddr, "ugnot")
	payFee(cur, payer, fee)
	payFee(cur, payer, fee)
	after := banker.NewReadonlyBanker().GetCoin(treasuryAddr, "ugnot")
	uassert.Equal(t, 2*ProposalFeeAmount, after-before)
	urequire.NoError(t, createAsUser(cur, payer))
	urequire.NoError(t, createAsUser(cur, payer))
	urequire.ErrorContains(t, createAsUser(cur, payer), feeErr)
}

func TestPayProposalFeeRefusesAnythingButTheExactFee(cur realm, t *testing.T) {
	loadMembers(cur)
	proposalCredits = map[address]int{}

	for _, sent := range []chain.Coins{
		nil,
		{chain.NewCoin("ugnot", ProposalFeeAmount-1)},
		{chain.NewCoin("ugnot", ProposalFeeAmount+1)},
		{chain.NewCoin("foo", ProposalFeeAmount)},
		{chain.NewCoin("foo", 1), chain.NewCoin("ugnot", ProposalFeeAmount)},
	} {
		uassert.AbortsContains(t, cur, "send exactly 1000000ugnot", func(cur realm) {
			payFee(cur, payer, sent)
		})
	}
	uassert.Equal(t, 0, proposalCredits[payer])
}

func TestPayProposalFeeRefusesMembersAndIndirectCalls(cur realm, t *testing.T) {
	loadMembers(cur)
	fee := chain.Coins{chain.NewCoin("ugnot", ProposalFeeAmount)}

	uassert.AbortsContains(t, cur, "members propose for free", func(cur realm) {
		payFee(cur, m1, fee)
	})

	// Only a direct msg call lands the coins here; the txtar covers maketx run.
	testing.SetOriginCaller(payer)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/middleman"))
	testing.SetOriginSend(fee)
	uassert.AbortsContains(t, cur, "paid with a direct msg call", func(cur realm) {
		PayProposalFee(cross(cur))
	})
}

func TestZeroFeeLetsNonMembersProposeFree(cur realm, t *testing.T) {
	loadMembers(cur)
	proposalCredits = map[address]int{}
	prev := ProposalFeeAmount
	ProposalFeeAmount = 0
	defer func() { ProposalFeeAmount = prev }()

	urequire.NoError(t, createAsUser(cur, payer))
	uassert.AbortsContains(t, cur, "no fee to pay", func(cur realm) {
		payFee(cur, payer, nil)
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
