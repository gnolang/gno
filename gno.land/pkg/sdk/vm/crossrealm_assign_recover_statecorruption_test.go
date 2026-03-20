package vm

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	gnolang "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// TestCrossRealmAssignRecover_GovernanceCorruption verifies that the
// assign+recover exploit does NOT corrupt governance state.
// The readonly check fires before the mutation, so votes stay unchanged.
func TestCrossRealmAssignRecover_GovernanceCorruption(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	deployer := crypto.AddressFromPreimage([]byte("deployer"))
	attacker := crypto.AddressFromPreimage([]byte("attacker"))

	deployerAcc := env.acck.NewAccountWithAddress(ctx, deployer)
	env.acck.SetAccount(ctx, deployerAcc)
	env.bankk.SetCoins(ctx, deployer, std.MustParseCoins(ugnot.ValueString(10_000_000)))

	attackerAcc := env.acck.NewAccountWithAddress(ctx, attacker)
	env.acck.SetAccount(ctx, attackerAcc)
	env.bankk.SetCoins(ctx, attacker, std.MustParseCoins(ugnot.ValueString(10_000_000)))

	// Deploy a simple governance realm.
	const govPath = "gno.land/r/test/governance"
	govFiles := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(govPath)},
		{Name: "gov.gno", Body: `package governance

var (
	yesVotes int
	noVotes  int
	finalized bool
)

func init() {
	yesVotes = 10
	noVotes = 90
	finalized = true
}

// SetVotes: non-crossing, should panic if called cross-realm.
func SetVotes(yes, no int) {
	yesVotes = yes
	noVotes = no
}

func GetResult(cur realm) string {
	if !finalized {
		return "pending"
	}
	if yesVotes > noVotes {
		return "approved"
	}
	return "rejected"
}

func GetTally(cur realm) (int, int) {
	return yesVotes, noVotes
}
`},
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(deployer, govPath, govFiles)))

	// Verify initial state: proposal is rejected (10 yes, 90 no).
	res, err := env.vmk.Call(ctx, NewMsgCall(deployer, std.Coins{}, govPath, "GetResult", nil))
	require.NoError(t, err)
	require.Contains(t, res, "rejected", "proposal must start as rejected")

	// Attacker attempts to flip the vote via assign+recover.
	runFiles := []*std.MemFile{
		{Name: "main.gno", Body: `package main

import "gno.land/r/test/governance"

func main() {
	result := governance.GetResult(cross)
	println("before:", result)

	// Attempt to corrupt the tally: flip to 99 yes, 1 no.
	func() {
		defer func() {
			r := recover()
			if r != nil {
				println("panic caught:", r.(string))
			}
		}()
		governance.SetVotes(99, 1)
	}()

	result = governance.GetResult(cross)
	println("after:", result)
}
`},
	}
	res, err = env.vmk.Run(ctx, NewMsgRun(attacker, std.Coins{}, runFiles))
	require.NoError(t, err)

	// The readonly check fired and was recovered.
	require.Contains(t, res, "panic caught:")
	require.Contains(t, res, "readonly")

	// State is NOT corrupted: vote stays rejected.
	require.Contains(t, res, "before: rejected")
	require.Contains(t, res, "after: rejected",
		"vote result must remain 'rejected' — no corruption")

	// Verify state persists correctly after commit.
	env.vmk.CommitGnoTransactionStore(ctx)
	ctx2 := env.vmk.MakeGnoTransactionStore(env.ctx)
	res, err = env.vmk.Call(ctx2, NewMsgCall(deployer, std.Coins{}, govPath, "GetResult", nil))
	require.NoError(t, err)
	require.Contains(t, res, "rejected",
		"vote must remain rejected after commit")
}

// TestCrossRealmAssignRecover_PermanentAdminLockout verifies that
// the assign+recover exploit does NOT corrupt admin state.
// The readonly check prevents the mutation, so admin stays unchanged.
func TestCrossRealmAssignRecover_PermanentAdminLockout(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	deployer := crypto.AddressFromPreimage([]byte("deployer"))
	attacker := crypto.AddressFromPreimage([]byte("attacker"))

	deployerAcc := env.acck.NewAccountWithAddress(ctx, deployer)
	env.acck.SetAccount(ctx, deployerAcc)
	env.bankk.SetCoins(ctx, deployer, std.MustParseCoins(ugnot.ValueString(10_000_000)))

	attackerAcc := env.acck.NewAccountWithAddress(ctx, attacker)
	env.acck.SetAccount(ctx, attackerAcc)
	env.bankk.SetCoins(ctx, attacker, std.MustParseCoins(ugnot.ValueString(10_000_000)))

	// Deploy a realm that uses PreviousRealm (not OriginCaller) for auth.
	const svcPath = "gno.land/r/test/service"
	svcFiles := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(svcPath)},
		{Name: "svc.gno", Body: fmt.Sprintf(`package service

import "chain/runtime"

var admin address

func init() {
	admin = address("%s")
}

// SetAdmin: non-crossing, should panic on cross-realm call.
func SetAdmin(newAdmin address) {
	admin = newAdmin
}

func GetAdmin(cur realm) address {
	return admin
}

// DoSomething: only admin can call.
func DoSomething(cur realm) string {
	caller := runtime.PreviousRealm().Address()
	if caller != admin {
		panic("unauthorized: " + string(caller) + " != " + string(admin))
	}
	return "success"
}
`, deployer.String())},
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(deployer, svcPath, svcFiles)))

	// Deployer can use the service.
	res, err := env.vmk.Call(ctx, NewMsgCall(deployer, std.Coins{}, svcPath, "DoSomething", nil))
	require.NoError(t, err)
	require.Contains(t, res, "success")

	// Attacker attempts to corrupt admin to a dead address.
	deadAddr := crypto.AddressFromPreimage([]byte("dead-address-nobody-controls"))
	runFiles := []*std.MemFile{
		{Name: "main.gno", Body: fmt.Sprintf(`package main

import "gno.land/r/test/service"

func main() {
	adminBefore := service.GetAdmin(cross)
	println("admin before:", adminBefore)

	// Attempt to corrupt admin to a dead address.
	func() {
		defer func() { _ = recover() }()
		service.SetAdmin(address("%s"))
	}()

	adminAfter := service.GetAdmin(cross)
	println("admin after:", adminAfter)
}
`, deadAddr.String())},
	}
	res, err = env.vmk.Run(ctx, NewMsgRun(attacker, std.Coins{}, runFiles))
	require.NoError(t, err)
	require.Contains(t, res, "admin before: "+deployer.String())
	// Admin must be unchanged — the fix prevents corruption.
	require.Contains(t, res, "admin after: "+deployer.String(),
		"admin must remain the deployer — no corruption")

	// Deployer can still use the service after commit.
	env.vmk.CommitGnoTransactionStore(ctx)
	ctx2 := env.vmk.MakeGnoTransactionStore(env.ctx)

	_, err = env.vmk.Call(ctx2, NewMsgCall(deployer, std.Coins{}, svcPath, "DoSomething", nil))
	require.NoError(t, err, "deployer must still be authorized — no lockout")
}
