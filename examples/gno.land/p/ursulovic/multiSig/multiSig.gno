package multi_sig

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type WithdrawProposal struct {
	id        int
	coin      std.Coin
	recipient std.Address
	approvals *avl.Tree
}

var (
	owners            avl.Tree
	newOwnerApprovals avl.Tree
	withdrawProposals avl.Tree
)

func init() {
}

func InitOwners(addresses []string) {
	if owners.Size() != 0 {
		panic("Owners list has already been initiated!")
	}

	for _, v := range addresses {
		addr := std.Address(v)
		if !addr.IsValid() {
			panic(ufmt.Sprintf("Entry %s is not a valid address!", v))
		}
		owners.Set(v, true)
	}
}

func AddNewOwner(address string) {
	assertOwnersInitialized()
	caller := std.GetOrigCaller()
	assertIsOwner(caller.String())

	if owners.Has(address) {
		panic("Proposed address is already an owner!")
	}

	v, exists := newOwnerApprovals.Get(address)
	if !exists {
		v = avl.NewTree()
		newOwnerApprovals.Set(address, v)
	}

	addressApprovals := v.(*avl.Tree)
	if addressApprovals.Has(caller.String()) {
		panic("You have already approved this new owner proposal address!")
	}
	addressApprovals.Set(caller.String(), true)

	if addressApprovals.Size() == owners.Size() {
		newOwnerApprovals.Remove(address)
		owners.Set(address, true)
	}

}

func GiveUpOwnership() {
	caller := std.GetOrigCaller()
	assertOwnersInitialized()
	assertIsOwner(caller.String())

	if owners.Size() == 1 {
		panic("You cannot give up ownership because there must be atleast one owner!")
	}

	owners.Remove(caller.String())

}

func assertOwnersInitialized() {
	if owners.Size() == 0 {
		panic("Owners must be initialised first!")
	}
}

func assertIsOwner(address string) {
	if !owners.Has(address) {
		panic("You are not the owner!")
	}
}

//Coins part

func DepositCoins() {
	assertOwnersInitialized()
	assertIsOwner(std.GetOrigCaller().String())
}

func ProposeWithdraw(amount int64, denom string, recipient string) {
	assertOwnersInitialized()
	assertIsOwner(std.GetOrigCaller().String())
	recipientAddr := std.Address(recipient)

	if denom == "" {
		panic("Denom cannot be empty!")
	}
	if amount <= 0 {
		panic("Amount must greater than 0!")
	}
	if !recipientAddr.IsValid() {
		panic("Recipient address not valid")
	}

	coin := std.Coin{
		Denom:  denom,
		Amount: amount,
	}

	id := withdrawProposals.Size()
	proposal := WithdrawProposal{
		id:        id,
		coin:      coin,
		recipient: recipientAddr,
		approvals: avl.NewTree(),
	}
	proposal.approvals.Set(std.GetOrigCaller().String(), true)
	withdrawProposals.Set(strconv.Itoa(id), proposal)

}
