package staker

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/r/volos/gov/governance"
)

// calculateUnlockTime determines the actual unlock time based on active proposals the user has voted on.
// If the user has voted on active proposals, the unlock time is extended until the latest proposal
// deadline plus the standard cooldown period.
func calculateUnlockTime(delegatee std.Address, baseUnlockAt int64) int64 {
	activeProposals := governance.GetUserActiveProposals(delegatee)
	if len(activeProposals) == 0 {
		return baseUnlockAt
	}

	var latestDeadline int64 = 0
	for _, proposal := range activeProposals {
		deadline := proposal.VotingDeadline().Unix()
		if deadline > latestDeadline {
			latestDeadline = deadline
		}
	}

	extendedUnlockAt := latestDeadline + unstakeLockPeriod

	if extendedUnlockAt > baseUnlockAt {
		return extendedUnlockAt
	}

	return baseUnlockAt
}

// updateDelegation updates the delegation mapping for a staker -> delegatee relationship.
// Positive amount increases delegation, negative amount decreases delegation.
func updateDelegation(staker, delegatee std.Address, amount int64) {
	stakerKey := staker.String()
	delegateeKey := delegatee.String()

	var stakerDelegations *avl.Tree
	if existing, ok := delegations.Get(stakerKey); ok {
		stakerDelegations = existing.(*avl.Tree)
	} else {
		stakerDelegations = avl.NewTree()
	}

	currentAmount := int64(0)
	if existing, ok := stakerDelegations.Get(delegateeKey); ok {
		currentAmount = existing.(int64)
	}

	newAmount := currentAmount + amount
	if newAmount <= 0 {
		stakerDelegations.Remove(delegateeKey)
		if stakerDelegations.Size() == 0 {
			delegations.Remove(stakerKey)
		} else {
			delegations.Set(stakerKey, stakerDelegations)
		}

	} else {
		stakerDelegations.Set(delegateeKey, newAmount)
		delegations.Set(stakerKey, stakerDelegations)
	}
}
