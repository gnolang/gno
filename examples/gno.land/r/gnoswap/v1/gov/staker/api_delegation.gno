package staker

import (
	"std"

	"gno.land/p/demo/avl"

	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/gnoswap/v1/gov/xgns"
)

// GetTotalxGnsSupply returns the total amount of xGNS supply.
func GetTotalxGnsSupply() uint64 {
	return xgns.TotalSupply()
}

// GetTotalVoteWeight returns the total amount of xGNS used for voting.
func GetTotalVoteWeight() uint64 {
	return xgns.VotingSupply()
}

// GetTotalDelegated returns the total amount of xGNS delegated.
func GetTotalDelegated() uint64 {
	return totalDelegated
}

// GetTotalLockedAmount returns the total amount of locked GNS.
func GetTotalLockedAmount() uint64 {
	return lockedAmount
}

// GetTotalDelegatedFrom returns the total amount of xGNS delegated by given address.
func GetTotalDelegatedFrom(from std.Address) uint64 {
	amount, exist := delegatorAmount.Get(from.String())
	if !exist {
		return 0
	}
	return amount.(uint64)
}

// GetTotalDelegatedTo returns the total amount of xGNS delegated to given address.
func GetTotalDelegatedTo(to std.Address) uint64 {
	amount, exist := delegatedTo.Get(to.String())
	if !exist {
		return 0
	}
	return amount.(uint64)
}

// GetDelegationAmountFromTo returns the amount of xGNS delegated by given address to given address.
func GetDelegationAmountFromTo(from, to std.Address) uint64 {
	toAmount, exist := delegatedFromTo.Get(from.String())
	if !exist {
		return 0
	}

	amount, exist := toAmount.(*avl.Tree).Get(to.String())
	if !exist {
		return 0
	}

	return amount.(uint64)
}

// GetRealmGnsBalance returns the amount of GNS in the current realm.
func GetRealmGnsBalance() uint64 {
	return gns.BalanceOf(std.CurrentRealm().Address())
}
