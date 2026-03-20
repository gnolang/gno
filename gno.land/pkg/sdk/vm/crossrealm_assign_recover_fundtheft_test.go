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
	caller := runtime.PreviousRealm().Address()
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

// TestCrossRealmAssignRecover_SameTxFundTheft verifies that the
// assign+recover exploit does NOT work: the readonly check fires
// before the mutation, so admin remains unchanged and Withdraw fails.
func TestCrossRealmAssignRecover_SameTxFundTheft(t *testing.T) {
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

	const victimPath = "gno.land/r/test/vault_assign"
	victimAddr := deployAssignVault(t, env, ctx, deployer, victimPath)

	beforeAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	beforeVictim := env.bankk.GetCoins(ctx, victimAddr).AmountOf("ugnot")

	// The exploit script attempts SetAdmin (non-crossing) then Withdraw.
	// With the fix, SetAdmin's assign is blocked before mutation,
	// so admin stays as deployer and Withdraw panics with "unauthorized".
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

	_, err := env.vmk.Run(ctx, NewMsgRun(attacker, std.Coins{}, runFiles))
	// Withdraw panics with "unauthorized" because admin was not corrupted.
	require.Error(t, err, "tx must fail because admin was not corrupted")
	require.Contains(t, err.Error(), "unauthorized",
		"Withdraw must reject the attacker since admin is unchanged")

	// Balances must remain unchanged — no fund theft.
	afterAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	afterVictim := env.bankk.GetCoins(ctx, victimAddr).AmountOf("ugnot")

	require.Equal(t, beforeAttacker, afterAttacker,
		"attacker balance must not change")
	require.Equal(t, beforeVictim, afterVictim,
		"victim balance must not change")
}

// TestCrossRealmAssignRecover_RealmToRealmFundTheft verifies the same
// protection via realm-to-realm composition (MsgCall to attacker realm).
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

	// Deploy attacker realm that attempts the exploit.
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
	// Withdraw panics with "unauthorized" because admin was not corrupted.
	require.Error(t, err, "Attack tx must fail because admin was not corrupted")
	require.Contains(t, err.Error(), "unauthorized")

	// Balances must remain unchanged.
	afterAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	afterVictim := env.bankk.GetCoins(ctx, victimAddr).AmountOf("ugnot")

	require.Equal(t, beforeAttacker, afterAttacker,
		"attacker balance must not change")
	require.Equal(t, beforeVictim, afterVictim,
		"victim balance must not change")
}

// TestCrossRealmAssignRecover_DrainLoop verifies that the drain loop
// exploit does NOT work: each SetAdmin attempt is blocked before mutation.
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

	// Attempt to drain — each iteration fails because SetAdmin is blocked.
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

	_, err := env.vmk.Run(ctx, NewMsgRun(attacker, std.Coins{}, runFiles))
	// First Withdraw panics with "unauthorized", aborting the tx.
	require.Error(t, err, "drain tx must fail")
	require.Contains(t, err.Error(), "unauthorized")

	// Balances must remain unchanged.
	afterAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	afterVictim := env.bankk.GetCoins(ctx, victimAddr).AmountOf("ugnot")

	require.Equal(t, beforeAttacker, afterAttacker,
		"attacker balance must not change")
	require.Equal(t, beforeVictim, afterVictim,
		"vault must retain all funds")
}

// TestCrossRealmAssign_WithoutRecover_Fails is a negative control.
// It verifies that calling SetAdmin without recover causes the tx to fail.
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

	// No recover -- the readonly panic aborts the entire tx.
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
	require.Contains(t, err.Error(), "readonly",
		"error must reference the readonly check")

	// Balances must remain unchanged.
	afterAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	afterVictim := env.bankk.GetCoins(ctx, victimAddr).AmountOf("ugnot")

	require.Equal(t, beforeAttacker, afterAttacker,
		"attacker balance must not change on failed tx")
	require.Equal(t, beforeVictim, afterVictim,
		"victim balance must not change on failed tx")
}
