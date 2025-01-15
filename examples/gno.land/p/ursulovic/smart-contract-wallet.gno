package smart_contract_wallet

import (
	"std"

	"gno.land/p/demo/avl"
)

type MultiSigWallet struct {
	owners             []std.Address
	approvalPercentage int
	approvals          int
	transactions       *avl.Tree
}

type Transaction struct {
	to       std.Address
	value    uint64
	executed bool
}

func DepositCoins() {

}

func WithdrawCoins() {

}

func AssertIsApproved() {
	
}
