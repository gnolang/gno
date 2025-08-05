package xvls

import (
	"std"
	"strconv"
)

const (
	EventMint         = "Mint"
	EventBurn         = "Burn"
	EventVotingSupply = "VotingSupply"
)

const (
	EventToKey        = "to"
	EventFromKey      = "from"
	EventAmountKey    = "amount"
	EventTotalSupplyKey = "total_supply"
	EventVotingSupplyKey = "voting_supply"
	EventExcludedKey  = "excluded_balance"
	EventCallerKey    = "caller"
	EventTimestampKey = "timestamp"
	EventBalanceKey   = "balance"
	EventLaunchpadKey = "launchpad_address"
)

// emitMint emits an event when xVLS tokens are minted
func emitMint(caller, to std.Address, amount int64) {
	std.Emit(
		EventMint,
		EventCallerKey, caller.String(),
		EventToKey, to.String(),
		EventAmountKey, strconv.FormatInt(amount, 10),
		EventTotalSupplyKey, strconv.FormatInt(TotalSupply(), 10),
		EventVotingSupplyKey, strconv.FormatInt(VotingSupply(), 10),
	)
}

// emitBurn emits an event when xVLS tokens are burned
func emitBurn(caller, from std.Address, amount int64) {
	std.Emit(
		EventBurn,
		EventCallerKey, caller.String(),
		EventFromKey, from.String(),
		EventAmountKey, strconv.FormatInt(amount, 10),
		EventTotalSupplyKey, strconv.FormatInt(TotalSupply(), 10),
		EventVotingSupplyKey, strconv.FormatInt(VotingSupply(), 10),
		EventBalanceKey, strconv.FormatInt(BalanceOf(from), 10),
	)
}

// emitVotingSupply emits an event when voting supply changes
func emitVotingSupply(caller std.Address, totalSupply, votingSupply int64, excludedBalance int64) {
	std.Emit(
		EventVotingSupply,
		EventCallerKey, caller.String(),
		EventTotalSupplyKey, strconv.FormatInt(totalSupply, 10),
		EventVotingSupplyKey, strconv.FormatInt(votingSupply, 10),
		EventExcludedKey, strconv.FormatInt(excludedBalance, 10),
		EventLaunchpadKey, Launchpad.String(),
	)
} 
