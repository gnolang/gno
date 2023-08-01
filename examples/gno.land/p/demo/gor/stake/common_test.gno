package stake

import (
	"gno.land/p/demo/avl"
	governor "gno.land/p/demo/governance/governor"
	stake "gno.land/p/demo/governance/stake" // shoule be imported, might be a bug
	"gno.land/p/demo/governance/checkpoints"
	"gno.land/p/demo/testutils"
	"std"
	"testing"
	"fmt"
)

func testOrigSend(to std.Address, coins std.Coins){
	// issue coins
	std.TestIssueCoins(to, coins)
	// set context variable
	std.TestSetOrigSend(coins, nil)
}

func MockDelegateSet(addrs []string, denom string) *avl.Tree {
	mds := avl.NewTree()
	for _, a := range addrs {
		d := stake.NewDelegate(0, denom)
		snapshot := std.GetHeight()
		d.AddSnapshotTokens(snapshot, 1, denom)
		d.AddSnapshotShares(snapshot, 1, denom)
		mds.Set(a, d)
	}
	return mds
}

type Env struct {
	delegator     string
	delegateAddrs []string
	delegateSet   *avl.Tree
	stakes        []*stake.Stake
	pool          string
}

func getEnv() Env {
	delegator := std.TestDerivePkgAddr("delegator").String() // delegator is the external user, here a test pkg to simulate
	pool := std.TestDerivePkgAddr("main").String()

	std.TestSetOrigCaller(std.Address(delegator))

	// new a stake with native coin
	nca := stake.NewNativeCoinAdapter("ugnot")
	s1 := stake.NewStake(nca)
	println("s1: ", s1)

	// new stake with grc20 token
	ga := stake.NewGrc20Adapter("Foo", "FOO", 4)
	s2 := stake.NewStake(ga)
	s2.GovToken.Mint(std.Address(delegator), 1000000)

	// reset pool
	amount:= s1.GovToken.Balance(std.Address(pool))
	s1.BurnPool(amount)

	var delegateAddrs []string
	for i:=0;i<3;i++{
		str := fmt.Sprintf("d%s", i)
		addr := testutils.TestAddress(str).String()
		delegateAddrs = append(delegateAddrs, addr)
	}

	dSet1 := MockDelegateSet(delegateAddrs, "ugnot")
	dSet2 := MockDelegateSet(delegateAddrs, "FOO")

	s1.SetDelegates(dSet1)
	s2.SetDelegates(dSet2)

	// mine block to chage state
	std.TestSkipHeights(1) 

	env := Env{
		delegator: delegator,
		delegateAddrs: delegateAddrs,
		pool: pool,
	}

	env.stakes = append(env.stakes, s1)
	env.stakes = append(env.stakes, s2)
	return env
}

func assertE(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func assert(t *testing.T, cond bool, format string, args ...interface{}) {
	if !cond {
		t.Fatalf(format, args...)
	}
}