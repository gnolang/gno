package stake

import (
	"gno.land/p/demo/avl"
	governor "gno.land/p/demo/governance/governor"
	stake "gno.land/p/demo/governance/stake"
	"gno.land/p/demo/testutils"
	"std"
	"testing"
)

// delegate, unbound
func TestUnbond(t *testing.T) {
	println("----------TestUnbond----------")
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


	// ----------------------------test unbond----------------------------
	// do unbond
	delegation, tb, err := s.Unbond(delegator, d0, 8)
	assert(t, tb == 8, "token ubond should be 8")

	std.TestSkipHeights(1) //skip current block to
	snapshot = std.GetHeight()

	// check delegation after unbond
	shares = delegation.GetSnapshotDelegationShares(snapshot)
	println("shares: ", shares)
	assert(t, shares == 0, "delegate shares not right")
}