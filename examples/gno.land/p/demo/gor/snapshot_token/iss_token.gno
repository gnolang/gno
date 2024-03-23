package snapshot_token

import "std"

type IVotes interface {
	// Returns the current amount of votes that `account` has.
	getVotes(account std.Address) uint64

	// Returns the amount of votes that `account` had at the end of a past block (`blockNumber`).
	getPastVotes(account std.Address, blockNumber int64) uint64

	// Returns the total supply of votes available at the end of a past block (`blockNumber`).
	// NOTE: This value is the sum of all available votes, which is not necessarily the sum of all delegated votes.
	// SnapshotToken that have not been delegated are still part of total supply, even though they would not participate in a
	// vote.
	getPastTotalSupply(blockNumber int64) uint64

	// Returns the delegate that `account` has chosen.
	delegates(account std.Address) std.Address

	// Delegates votes from the sender to `delegatee`.
	delegate(delegatee std.Address) bool
}

// Emitted when an account changes their delegate.
type DelegateChangedEvent struct {
	delegator    std.Address
	fromDelegate std.Address
	toDelegate   std.Address
}

// Emitted when a token transfer or delegate change results in changes to a delegate's number of votes.
type DelegateVotesChangedEvent struct {
	delegate        std.Address
	previousBalance uint64
	newBalance      uint64
}
