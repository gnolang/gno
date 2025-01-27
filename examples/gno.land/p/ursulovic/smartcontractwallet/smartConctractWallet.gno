package smartcontractwallet

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/ursulovic/functionmodifiers"
)

type WithdrawProposal struct {
	id        int64
	recipient std.Address
	coin      std.Coin
}

var (
	modifiers         functionmodifiers.ModiferRegistry
	withdrawProposals *avl.Tree
)

func init() {
	withdrawProposals = avl.NewTree()
}

func initOwners(owners []string, newOwnerPercentage float64) {
	modifiers.InitializeOwners(owners, newOwnerPercentage)
}

func DepositCoins() {
	caller := std.GetOrigCaller()

	modifiers.AssertIsOwner(caller)
	modifiers.AssertOwnersInitialised()
}

func ProposeWithdraw(recipient string, amount int64, denom string, approvalPercentage float64) {
	caller := std.GetOrigCaller()

	modifiers.AssertIsOwner(caller)
	modifiers.AssertOwnersInitialised()

	coin := std.NewCoin(denom, amount)

	id := int64(withdrawProposals.Size())
	withdrawProposal := WithdrawProposal{
		id:        id,
		recipient: caller,
		coin:      coin,
	}

	modifiers.RegisterModifier(formatProposalKey(id), approvalPercentage)

	withdrawProposals.Set(strconv.FormatInt(id, 10), withdrawProposal)
}

func ExecuteWithdraw(id int64) {
	modifiers.AssertIsApproved(formatProposalKey(id))

	v, exists := withdrawProposals.Get(strconv.FormatInt(id, 10))

	if !exists {
		panic("Withdraw proposal with provided id doesn't exist!")
	}

	withdrawProposal := v.(*WithdrawProposal)

	coins := std.NewCoins(withdrawProposal.coin)

	banker := std.GetBanker(std.BankerTypeRealmSend)

	banker.SendCoins(std.CurrentRealm().Addr(), withdrawProposal.recipient, coins)
}

func formatProposalKey(id int64) string {
	return "WITHDRAW-" + strconv.FormatInt(id, 10)
}
