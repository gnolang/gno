package stake

import (
	"gno.land/p/demo/avl"
	governor "gno.land/p/demo/governance/governor"
	"gno.land/p/demo/testutils"
	"std"
	"testing"
)

// delegate, undelegate
func TestUndelegate(t *testing.T) {
	println("----------TestUndelegate----------")
	// ------------------------prepare work--------------------------
	env := getEnv()
	d0 := env.delegateAddrs[0]
	d1 := env.delegateAddrs[1]
	d2 := env.delegateAddrs[2]
	delegator := env.delegator
	poolAddr := env.pool
	s := env.stakes[0]

	snapshot := std.GetHeight()
	println("delegator: ", delegator)
	println("snapshot: ", snapshot)

	std.TestSetOrigCaller(std.Address(delegator))  
	testOrigSend(std.Address(poolAddr), std.Coins{{"ugnot", 8}}) // simulate a send
	err := s.Delegate(d0, "", 0, false)

	testOrigSend(std.Address(poolAddr), std.Coins{{"ugnot", 27}}) // simulate a send
	err = s.Delegate(d1, "", 0, false)

	testOrigSend(std.Address(poolAddr), std.Coins{{"ugnot", 71}}) // simulate a send
	err = s.Delegate(d2, "", 0, false)

	std.TestSkipHeights(1) //skip current block to
	snapshot = std.GetHeight()

	// check pool balance
	amount := s.GovToken.Balance(std.Address(poolAddr))
	println("pool amount: ", amount)
	assert(t, amount == 106, "pool balance mismatch")
	
	// get delegation and check shares
    indexedDelegation, err := s.GetDelegation(delegator, d0)
	assert(t, err == nil, "get indexedDelegation error")
	delegation, err := indexedDelegation.GetDelegationByIndex(d0)
	shares := delegation.GetSnapshotDelegationShares(snapshot)
	assert(t, shares == 8, "delegate shares not right")

	indexedDelegation, err = s.GetDelegation(delegator, d1)
	assert(t, err == nil, "get indexedDelegation error")
	delegation, err = indexedDelegation.GetDelegationByIndex(d1)
	shares = delegation.GetSnapshotDelegationShares(snapshot)
	assert(t, shares == 27, "delegate shares not right")

	indexedDelegation, err = s.GetDelegation(delegator, d2)
	assert(t, err == nil, "get indexedDelegation error")
	delegation, err = indexedDelegation.GetDelegationByIndex(d2)
	shares = delegation.GetSnapshotDelegationShares(snapshot)
	assert(t, shares == 71, "delegate shares not right")

	// ----------------------------test unUndelegation----------------------------
	// do undelegation
	err = s.Undelegate(delegator, d0, 8)
	assert(t, err == nil, "get delegation error")

	std.TestSkipHeights(1) //skip current block to
	snapshot = std.GetHeight()

	// get delegation and check shares
    indexedDelegation, err = s.GetDelegation(delegator, d0)
	assert(t, err == nil, "get indexedDelegation error")
	delegation, err = indexedDelegation.GetDelegationByIndex(d0)
	shares = delegation.GetSnapshotDelegationShares(snapshot)
	assert(t, shares == 0, "delegate shares not right")
	
	// test complete ubd tasks
	std.TestSkipHeights(21) //skip 1 step, we can only get vote from past block
	// check balance of delegator
	amount = s.GovToken.Balance(std.Address(delegator))
	println("before complete ubd---")
	println("delegator amount: ", amount)
	s.CompleteUnbondDelegation()
	println("after complete ubd---")
	amount = s.GovToken.Balance(std.Address(delegator))
	assert(t, amount == 8, "incorrect amount")
	// TODO: test more mature immature condition
}
