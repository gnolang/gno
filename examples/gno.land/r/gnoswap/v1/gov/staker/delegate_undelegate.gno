package staker

import (
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var (
	totalDelegated = uint64(0)

	delegatorAmount = avl.NewTree() // caller => amount
	delegatedFromTo = avl.NewTree() // caller => to => amount
	delegatedTo     = avl.NewTree() // to => amount
)
 
func delegate(to std.Address, amount uint64) {
	caller := std.PrevRealm().Addr().String()
	toStr := to.String()

	// initialize the internal tree for callers to `delegatedFromTo`
	innerTree := getOrCreateInnerTree(delegatedFromTo, caller)

	totalDelegated += amount

	// update delegator amount
	updateUint64InTree(delegatorAmount, caller, amount, true)

	// update delegatedFromTo's inner tree
	updateUint64InTree(innerTree, toStr, amount, true)

	// update delegatedTo
	updateUint64InTree(delegatedTo, toStr, amount, true)

	timeStamp := uint64(time.Now().Unix())
	// update delegation history
	delegation := DelegationHistory{
		to:        to,
		amount:    amount,
		timestamp: timeStamp,
		height:    uint64(std.GetHeight()),
		add:       true, // if true, delegation
	}

	history := make([]DelegationHistory, 0)
	if value, exists := delegationHistory.Get(caller); exists {
		history = value.([]DelegationHistory)
	}
	history = append(history, delegation)
	delegationHistory.Set(caller, history)

	// update delegation stat history
	updateAmount := uint64(0)
	snapShotHistory := make([]DelegationSnapShotHistory, 0)
	if value, exists := delegationSnapShotHistory.Get(toStr); exists {
		snapShotHistory = value.([]DelegationSnapShotHistory)
		lastStat := snapShotHistory[len(snapShotHistory)-1]
		updateAmount = lastStat.amount + amount
	} else {
		updateAmount = amount
	}

	snapShotHistory = append(snapShotHistory, DelegationSnapShotHistory{
		to:           to,
		amount:       updateAmount,
		updatedBlock: uint64(std.GetHeight()),
		updatedAt:    timeStamp,
	})
	delegationSnapShotHistory.Set(toStr, snapShotHistory)
}

func undelegate(to std.Address, amount uint64) {
	caller := std.PrevRealm().Addr().String()
	toStr := to.String()

	// check caller's delegatedFromTo
	innerTree, exists := delegatedFromTo.Get(caller)
	if !exists {
		panic(addDetailToError(
			errNoDelegatedAmount,
			ufmt.Sprintf("caller(%s) has no delegated amount", caller),
		))
	}
	// check caller's delegatedFromTo's inner tree
	delegatedAmountValue, exists := innerTree.(*avl.Tree).Get(toStr)
	if !exists {
		panic(addDetailToError(
			errNoDelegatedTarget,
			ufmt.Sprintf("caller(%s) has no delegated amount to %s", caller, to),
		))
	}
	delegatedAmount := delegatedAmountValue.(uint64)
	if delegatedAmount < amount {
		panic(addDetailToError(
			errNotEnoughDelegated,
			ufmt.Sprintf("caller(%s) has only %d delegated amount(request: %d) to %s", caller, delegatedAmount, amount, to),
		))
	}
	innerTree.(*avl.Tree).Set(toStr, delegatedAmount-amount)
	delegatedFromTo.Set(caller, innerTree)

	// update total delegated amount
	totalDelegated -= amount

	currentAmount := uint64(0)
	if value, exists := delegatorAmount.Get(caller); exists {
		currentAmount = value.(uint64)
	} else {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("caller(%s) has no delegated amount", caller),
		))
	}
	if currentAmount < amount {
		panic(addDetailToError(
			errNotEnoughDelegated,
			ufmt.Sprintf("caller(%s) has only %d delegated amount(request: %d)", caller, currentAmount, amount),
		))
	}
	delegatorAmount.Set(caller, currentAmount-amount)

	currentToAmount := uint64(0)
	if value, exists := delegatedTo.Get(toStr); exists {
		currentToAmount = value.(uint64)
	} else {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("to(%s) has no delegated amount", toStr),
		))
	}
	if currentToAmount < amount {
		panic(addDetailToError(
			errNotEnoughDelegated,
			ufmt.Sprintf("to(%s) has only %d delegated amount(request: %d)", toStr, currentToAmount, amount),
		))
	}
	delegatedTo.Set(toStr, currentToAmount-amount)

	// update delegation history
	delegation := DelegationHistory{
		to:        to,
		amount:    amount,
		timestamp: uint64(time.Now().Unix()),
		height:    uint64(std.GetHeight()),
		add:       false,
	}
	var history []DelegationHistory
	if value, exists := delegationHistory.Get(caller); exists {
		history = value.([]DelegationHistory)
	}
	history = append(history, delegation)
	delegationHistory.Set(caller, history)

	// update delegation stat history
	statValue, exists := delegationSnapShotHistory.Get(toStr)
	if !exists {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("caller(%s) has no delegation stat history", caller),
		))
	}

	stat := statValue.([]DelegationSnapShotHistory)
	remainingAmount := amount
	for i := 0; i < len(stat); i++ {
		if stat[i].amount > 0 {
			if stat[i].amount < remainingAmount {
				remainingAmount -= stat[i].amount
				stat = append(stat[:i], stat[i+1:]...)
			} else {
				stat[i].amount -= remainingAmount
				stat[i].updatedAt = uint64(time.Now().Unix())
				break
			}
		}
	}
	delegationSnapShotHistory.Set(to.String(), stat)
}
