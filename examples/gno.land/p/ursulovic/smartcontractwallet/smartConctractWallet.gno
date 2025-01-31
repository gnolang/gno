package smartcontractwallet

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/ursulovic/functionmodifiers"
)

type Wallet struct {
	modifiers         functionmodifiers.ModifierRegistry
	withdrawProposals *avl.Tree
}

type WithdrawProposal struct {
	id        int64
	recipient std.Address
	coin      std.Coin
}

func init() {
}

func (wallet *Wallet) InitOwners(owners []string, newOwnerPercentage float64) {
	wallet.withdrawProposals = avl.NewTree()
	wallet.modifiers.InitializeOwners(owners, newOwnerPercentage)
}

func (wallet *Wallet) DepositCoins() {
	caller := std.GetOrigCaller()

	wallet.modifiers.AssertIsOwner(caller)
	wallet.modifiers.AssertOwnersInitialized()
}

func (wallet *Wallet) ProposeWithdraw(recipient string, amount int64, denom string, approvalPercentage float64) int64 {
	caller := std.GetOrigCaller()

	wallet.modifiers.AssertIsOwner(caller)
	wallet.modifiers.AssertOwnersInitialized()

	coin := std.NewCoin(denom, amount)

	id := int64(wallet.withdrawProposals.Size())
	withdrawProposal := &WithdrawProposal{
		id:        id,
		recipient: caller,
		coin:      coin,
	}

	wallet.modifiers.RegisterModifier(formatProposalKey(id), approvalPercentage)

	wallet.withdrawProposals.Set(strconv.FormatInt(id, 10), withdrawProposal)
	return id
}

func (wallet *Wallet) ExecuteWithdraw(id int64) {
	wallet.modifiers.AssertIsApproved(formatProposalKey(id))

	v, exists := wallet.withdrawProposals.Get(strconv.FormatInt(id, 10))

	if !exists {
		panic("Withdraw proposal with provided id doesn't exist!")
	}

	withdrawProposal := v.(*WithdrawProposal)

	coins := std.NewCoins(withdrawProposal.coin)

	banker := std.GetBanker(std.BankerTypeRealmSend)

	banker.SendCoins(std.CurrentRealm().Addr(), withdrawProposal.recipient, coins)

	wallet.withdrawProposals.Remove(formatProposalKey(id))
}

func (wallet *Wallet) ApproveWithdraw(id int64) {
	wallet.modifiers.ApproveModifier(formatProposalKey(id))
}

func formatProposalKey(id int64) string {
	return "WITHDRAW-COIN" + strconv.FormatInt(id, 10)
}
