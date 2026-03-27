package vm

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	gnolang "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// deployVault deploys a vault realm with a banker-based Withdraw.
// The vault has an exported SetOwner (non-crossing) — simulating an
// internal helper that was accidentally exported (should be setOwner).
// owner is initialized to a DAO realm address.
func deployVault(t *testing.T, env testEnv, ctx sdk.Context, deployer crypto.Address, pkgPath string) crypto.Address {
	t.Helper()

	vaultAddr := gnolang.DerivePkgCryptoAddr(pkgPath)
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "vault.gno", Body: fmt.Sprintf(`package %s

import (
	"chain"
	"chain/banker"
	"chain/runtime"
)

var owner address

func init() {
	// Set initial owner to a DAO realm via internal helper.
	// This is why SetOwner exists — to set the owner during init.
	SetOwner(address("g1dao0000000000000000000000000000000000"))
}

// SetOwner is an internal helper that was exported by mistake
// (should be setOwner). Without the pre-mutation readonly check,
// a cross-realm caller could call SetOwner + recover to silently
// hijack ownership in memory, then call Withdraw to steal funds.
func SetOwner(o address) {
	owner = o
}

func GetOwner(cur realm) address {
	return owner
}

func Withdraw(cur realm, amount int64) {
	caller := runtime.PreviousRealm().Address()
	if caller != owner {
		panic("unauthorized")
	}
	b := banker.NewBanker(banker.BankerTypeRealmSend)
	pkgAddr := runtime.CurrentRealm().Address()
	b.SendCoins(pkgAddr, caller, chain.Coins{{"ugnot", amount}})
}
`, lastPkgSegment(pkgPath))},
	}

	msg := NewMsgAddPackage(deployer, pkgPath, files)
	require.NoError(t, env.vmk.AddPackage(ctx, msg))

	// Fund the vault.
	env.bankk.SetCoins(ctx, vaultAddr, std.MustParseCoins(ugnot.ValueString(5_000_000)))
	return vaultAddr
}

func lastPkgSegment(pkgPath string) string {
	for i := len(pkgPath) - 1; i >= 0; i-- {
		if pkgPath[i] == '/' {
			return pkgPath[i+1:]
		}
	}
	return pkgPath
}

// TestCrossRealmAssignRecover_FundTheft verifies that the assign+recover
// pattern cannot be used to hijack ownership and steal funds.
//
// Scenario: a vault realm has an exported SetOwner (non-crossing) that
// was meant to be unexported. An attacker calls SetOwner(myAddr) +
// recover, then Withdraw(cross). With the pre-mutation fix, the
// ownership change is blocked before it lands in memory, so Withdraw
// panics with "unauthorized" and no funds move.
func TestCrossRealmAssignRecover_FundTheft(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	deployer := crypto.AddressFromPreimage([]byte("deployer"))
	attacker := crypto.AddressFromPreimage([]byte("attacker"))

	// Set up accounts.
	deployerAcc := env.acck.NewAccountWithAddress(ctx, deployer)
	env.acck.SetAccount(ctx, deployerAcc)
	env.bankk.SetCoins(ctx, deployer, std.MustParseCoins(ugnot.ValueString(30_000_000)))

	attackerAcc := env.acck.NewAccountWithAddress(ctx, attacker)
	env.acck.SetAccount(ctx, attackerAcc)
	env.bankk.SetCoins(ctx, attacker, std.MustParseCoins(ugnot.ValueString(20_000_000)))

	// Deploy vault with deployer as owner.
	const vaultPath = "gno.land/r/test/vault"
	vaultAddr := deployVault(t, env, ctx, deployer, vaultPath)

	beforeAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	beforeVault := env.bankk.GetCoins(ctx, vaultAddr).AmountOf("ugnot")
	require.Equal(t, int64(5_000_000), beforeVault, "vault should start with 5M ugnot")

	// Attacker script: SetOwner(myAddr) + recover, then Withdraw.
	runFiles := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest("gno.land/r/test/attack")},
		{Name: "main.gno", Body: fmt.Sprintf(`package main

import "%s"

func main() {
	// Attempt to hijack ownership via assign+recover.
	func() {
		defer func() { _ = recover() }()
		%s.SetOwner(address("%s"))
	}()

	// Try to withdraw — should fail because owner is unchanged.
	%s.Withdraw(cross, 1000000)
}
`, vaultPath, lastPkgSegment(vaultPath), attacker.String(),
			lastPkgSegment(vaultPath))},
	}

	_, err := env.vmk.Run(ctx, NewMsgRun(attacker, std.Coins{}, runFiles))
	require.Error(t, err, "tx must fail because admin was not corrupted")
	require.Contains(t, err.Error(), "unauthorized",
		"Withdraw must reject the attacker since owner is unchanged")

	// Owner must still be the DAO address — not the attacker.
	const daoAddr = "g1dao0000000000000000000000000000000000"
	ownerRes, err := env.vmk.Call(ctx, NewMsgCall(attacker, std.Coins{}, vaultPath, "GetOwner", nil))
	require.NoError(t, err)
	require.Contains(t, ownerRes, daoAddr,
		"owner must still be the DAO address")

	// Balances must remain unchanged — no fund theft.
	afterAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	afterVault := env.bankk.GetCoins(ctx, vaultAddr).AmountOf("ugnot")

	require.Equal(t, beforeAttacker, afterAttacker,
		"attacker balance must not change")
	require.Equal(t, beforeVault, afterVault,
		"vault must retain all funds")
}

// TestCrossRealmAssignRecover_NoRecover is a negative control.
// Without recover, the readonly panic aborts the entire tx.
func TestCrossRealmAssignRecover_NoRecover(t *testing.T) {
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

	const vaultPath = "gno.land/r/test/vault_neg"
	vaultAddr := deployVault(t, env, ctx, deployer, vaultPath)

	beforeAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	beforeVault := env.bankk.GetCoins(ctx, vaultAddr).AmountOf("ugnot")

	// No recover — the readonly panic aborts the entire tx.
	runFiles := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest("gno.land/r/test/attack_neg")},
		{Name: "main.gno", Body: fmt.Sprintf(`package main

import "%s"

func main() {
	%s.SetOwner(address("%s"))
	%s.Withdraw(cross, 1000000)
}
`, vaultPath, lastPkgSegment(vaultPath), attacker.String(),
			lastPkgSegment(vaultPath))},
	}

	_, err := env.vmk.Run(ctx, NewMsgRun(attacker, std.Coins{}, runFiles))
	require.Error(t, err, "tx must fail")
	require.Contains(t, err.Error(), "readonly",
		"error must reference the readonly check")

	// Owner must still be the DAO address.
	const daoAddr = "g1dao0000000000000000000000000000000000"
	ownerRes, err := env.vmk.Call(ctx, NewMsgCall(attacker, std.Coins{}, vaultPath, "GetOwner", nil))
	require.NoError(t, err)
	require.Contains(t, ownerRes, daoAddr,
		"owner must still be the DAO address")

	afterAttacker := env.bankk.GetCoins(ctx, attacker).AmountOf("ugnot")
	afterVault := env.bankk.GetCoins(ctx, vaultAddr).AmountOf("ugnot")

	require.Equal(t, beforeAttacker, afterAttacker,
		"attacker balance must not change")
	require.Equal(t, beforeVault, afterVault,
		"vault must retain all funds")
}
