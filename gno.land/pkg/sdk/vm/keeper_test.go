package vm

// TODO: move most of the logic in ROOT/gno.land/...

import (
	"errors"
	"fmt"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

var (
	initialBalance = std.MustParseCoins(ugnot.ValueString(20_000_000))
	coinsToSend    = std.MustParseCoins(ugnot.ValueString(1_000_000))
)

func TestVMKeeperAddPackage(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create test package.
	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{
			Name: "test.gno",
			Body: `package test

func Echo(cur realm) string {
	return "hello world"
}`,
		},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))

	err := env.vmk.AddPackage(ctx, msg1)

	assert.NoError(t, err)
	assert.NotNil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))

	err = env.vmk.AddPackage(ctx, msg1)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, PkgExistError{}))

	// added package is formatted
	store := env.vmk.getGnoTransactionStore(ctx)
	memFile := store.GetMemFile("gno.land/r/test", "test.gno")
	assert.NotNil(t, memFile)
	expected := `package test

func Echo(cur realm) string {
	return "hello world"
}`
	assert.Equal(t, expected, memFile.Body)
}

func TestVMKeeperAddPackage_InvalidDomain(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create test package.
	const pkgPath = "anotherdomain.land/r/test"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{
			Name: "test.gno",
			Body: `package test
func Echo(cur realm) string {
	return "hello world"
}`,
		},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))

	err := env.vmk.AddPackage(ctx, msg1)

	assert.Error(t, err, ErrInvalidPkgPath("invalid domain: anotherdomain.land/r/test"))
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))

	err = env.vmk.AddPackage(ctx, msg1)
	assert.Error(t, err, ErrInvalidPkgPath("invalid domain: anotherdomain.land/r/test"))

	// added package is formatted
	store := env.vmk.getGnoTransactionStore(ctx)
	memFile := store.GetMemFile("gno.land/r/test", "test.gno")
	assert.Nil(t, memFile)
}

func TestVMKeeperAddPackage_DraftPackage(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create test package.
	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{
			Name: "gnomod.toml",
			Body: `module = "gno.land/r/test"
gno = "0.9"
draft = true`,
		},
		{
			Name: "test.gno",
			Body: `package test

func Echo(cur realm) string {
	return "hello world"
}`,
		},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))

	err := env.vmk.AddPackage(ctx, msg1)

	assert.Error(t, err, ErrInvalidPackage("draft packages can only be deployed at genesis time"))
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))
}

func TestVMKeeperAddPackage_ImportDraft(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// create a valid draft pkg at genesis
	ctx = ctx.WithBlockHeader(&bft.Header{ChainID: "test-chain-id", Height: 0})
	// Create test package.
	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{
			Name: "gnomod.toml",
			Body: `module = "gno.land/r/test"
gno = "0.9"
draft = true`,
		},
		{
			Name: "test.gno",
			Body: `package test

func Echo(cur realm) string {
	return "hello world"
}`,
		},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)

	assert.NoError(t, err)
	assert.NotNil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))

	// Try to import the draft package.
	const pkgPath2 = "gno.land/r/test2"
	files2 := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath2)},
		{
			Name: "test2.gno",
			Body: `package test2

import "gno.land/r/test"

func Echo(cur realm) string {
	return test.Echo(cross)
}`,
		},
	}

	ctx = ctx.WithBlockHeader(&bft.Header{ChainID: "test-chain-id", Height: 42})
	msg2 := NewMsgAddPackage(addr, pkgPath2, files2)
	err = env.vmk.AddPackage(ctx, msg2)
	assert.Error(t, err, ErrTypeCheck(gnolang.ImportDraftError{PkgPath: pkgPath}))

	ctx = ctx.WithBlockHeader(&bft.Header{ChainID: "test-chain-id", Height: 0})
	err = env.vmk.AddPackage(ctx, msg2)
	assert.NoError(t, err)
}

func TestVMKeeperAddPackage_PrivatePackage(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create test package.
	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{
			Name: "gnomod.toml",
			Body: `module = "gno.land/r/test"
gno = "0.9"
private = true`,
		},
		{
			Name: "test.gno",
			Body: `package test

func Echo(cur realm) string {
	return "hello world"
}`,
		},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)
}

func TestVMKeeperAddPackage_UpdatePrivatePackage(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create private test package.
	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{
			Name: "gnomod.toml",
			Body: `module = "gno.land/r/test"
gno = "0.9"
private = true`,
		},
		{
			Name: "test.gno",
			Body: `package test

func Echo(cur realm) string {
	return "hello world"
}`,
		},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Re-upload the same private package with updated content.
	files2 := []*std.MemFile{
		{
			Name: "gnomod.toml",
			Body: `module = "gno.land/r/test"
gno = "0.9"
private = true`,
		},
		{
			Name: "test.gno",
			Body: `package test

func Echo(cur realm) string {
	return "hello updated world"
}`,
		},
	}

	msg2 := NewMsgAddPackage(addr, pkgPath, files2)
	err = env.vmk.AddPackage(ctx, msg2)
	assert.NoError(t, err)

	// Verify the package was updated with the new content.
	store := env.vmk.getGnoTransactionStore(ctx)
	memFile := store.GetMemFile(pkgPath, "test.gno")
	assert.NotNil(t, memFile)
	expected := `package test

func Echo(cur realm) string {
	return "hello updated world"
}`
	assert.Equal(t, expected, memFile.Body)
}

func TestVMKeeperAddPackage_ImportPrivate(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create test package 1.
	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{
			Name: "gnomod.toml",
			Body: `module = "gno.land/r/test"
gno = "0.9"
private = true`,
		},
		{
			Name: "test.gno",
			Body: `package test

func Echo(cur realm) string {
	return "hello world"
}`,
		},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	const pkgPath2 = "gno.land/r/test2"
	files2 := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath2)},
		{
			Name: "test2.gno",
			Body: `package test2

import "gno.land/r/test"

func Echo(cur realm) string {
	return test.Echo(cross)
}`,
		},
	}

	msg2 := NewMsgAddPackage(addr, pkgPath2, files2)
	err = env.vmk.AddPackage(ctx, msg2)
	assert.Error(t, err, ErrTypeCheck(gnolang.ImportPrivateError{PkgPath: pkgPath}))
}

func TestVMKeeperAddPackage_ChangePublicToPrivate(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{
			Name: "test.gno",
			Body: `package test

func Echo(cur realm) string {
	return "hello world"
}`,
		},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Try to upload a private version of the same package.
	files2 := []*std.MemFile{
		{
			Name: "gnomod.toml",
			Body: `module = "gno.land/r/test"
gno = "0.9"
private = true`,
		},
		{
			Name: "test.gno",
			Body: `package test

func Echo(cur realm) string {
	return "hello private world"
}`,
		},
	}

	msg2 := NewMsgAddPackage(addr, pkgPath, files2)
	err = env.vmk.AddPackage(ctx, msg2)
	assert.Error(t, err, ErrInvalidPackage(""))
}

func TestVMKeeperAddPackage_ChangePrivateToPublic(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create a private test package first.
	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{
			Name: "gnomod.toml",
			Body: `module = "gno.land/r/test"
gno = "0.9"
private = true`,
		},
		{
			Name: "test.gno",
			Body: `package test

func Echo(cur realm) string {
	return "hello private world"
}`,
		},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Try to upload a public version of the same package.
	files2 := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{
			Name: "test.gno",
			Body: `package test

func Echo(cur realm) string {
	return "hello public world"
}`,
		},
	}

	msg2 := NewMsgAddPackage(addr, pkgPath, files2)
	err = env.vmk.AddPackage(ctx, msg2)
	assert.Error(t, err, ErrInvalidPackage(""))
}

// Sending total send amount succeeds.
func TestVMKeeperOriginSend1(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create test package.
	const pkgPath = "gno.land/r/test"
	pkgAddr := gnolang.DerivePkgCryptoAddr(pkgPath)
	storageDepositAddr := gnolang.DeriveStorageDepositCryptoAddr(pkgPath)
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "init.gno", Body: `
package test

import (
	"chain/runtime"
	"chain/banker"
)

func init() {
}

func Echo(cur realm, msg string) string {
	addr := runtime.OriginCaller()
	pkgAddr := runtime.CurrentRealm().Address()
	send := banker.OriginSend()
	banker := banker.NewBanker(banker.BankerTypeOriginSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}`},
	}
	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Reconcile the account balance
	userAcctBalance := env.bankk.GetCoins(ctx, addr)
	pkgStorageDeposit := env.bankk.GetCoins(ctx, storageDepositAddr)
	assert.True(t, userAcctBalance.Add(pkgStorageDeposit).IsEqual(initialBalance))

	// Run Echo function.
	msg2 := NewMsgCall(addr, coinsToSend, pkgPath, "Echo", []string{"hello world"})
	res, err := env.vmk.Call(ctx, msg2)
	assert.NoError(t, err)
	assert.Equal(t, `("echo:hello world" string)`+"\n\n", res)

	// The Echo() function sends the user back the original sent amount.
	pkgBalance := env.bankk.GetCoins(ctx, pkgAddr)
	assert.True(t, pkgBalance.IsZero())
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(userAcctBalance))
}

// Sending too much fails
func TestVMKeeperOriginSend2(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create test package.
	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "init.gno", Body: `
package test

import (
	"chain/runtime"
	"chain/banker"
)

var admin address

func init() {
     admin = runtime.OriginCaller()
}

func Echo(cur realm, msg string) string {
	addr := runtime.OriginCaller()
	pkgAddr := runtime.CurrentRealm().Address()
	send := banker.OriginSend()
	banker := banker.NewBanker(banker.BankerTypeOriginSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}

func GetAdmin(cur realm) string {
	return admin.String()
}
`},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	coins := std.MustParseCoins(ugnot.ValueString(21000000))
	msg2 := NewMsgCall(addr, coins, pkgPath, "Echo", []string{"hello world"})
	res, err := env.vmk.Call(ctx, msg2)
	assert.Error(t, err)
	assert.Equal(t, "", res)
	assert.True(t, strings.Contains(err.Error(), "insufficient coins error"))
}

// Sending more than tx send fails.
func TestVMKeeperOriginSend3(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create test package.
	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "init.gno", Body: `
package test

import (
	"chain"
	"chain/banker"
	"chain/runtime"
)

func init() {
}

func Echo(cur realm, msg string) string {
	addr := runtime.OriginCaller()
	pkgAddr := runtime.CurrentRealm().Address()
	send := chain.Coins{{"ugnot", 10000000}}
	banker := banker.NewBanker(banker.BankerTypeOriginSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}`},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	coins := std.MustParseCoins(ugnot.ValueString(9000000))
	msg2 := NewMsgCall(addr, coins, pkgPath, "Echo", []string{"hello world"})
	// XXX change this into an error and make sure error message is descriptive.
	_, err = env.vmk.Call(ctx, msg2)
	assert.Error(t, err)
}

// Sending realm package coins succeeds.
func TestVMKeeperRealmSend1(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)

	const pkgPath = "gno.land/r/test"
	pkgAddr := gnolang.DerivePkgCryptoAddr(pkgPath)
	storageDepositAddr := gnolang.DeriveStorageDepositCryptoAddr(pkgPath)
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "init.gno", Body: `
package test

import (
	"chain"
	"chain/banker"
	"chain/runtime"
)

func init() {
}

func Echo(cur realm, msg string) string {
	addr := runtime.OriginCaller()
	pkgAddr := runtime.CurrentRealm().Address()
	send := chain.Coins{{"ugnot", 1000000}}
	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)
	banker_.SendCoins(pkgAddr, addr, send) // send back
	return "echo:" + msg
}`},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	msg2 := NewMsgCall(addr, coinsToSend, pkgPath, "Echo", []string{"hello world"})
	res, err := env.vmk.Call(ctx, msg2)
	assert.NoError(t, err)
	assert.Equal(t, `("echo:hello world" string)`+"\n\n", res)
	// Reconcile the account balance
	userAcctBalance := env.bankk.GetCoins(ctx, addr)
	pkgStorageDeposit := env.bankk.GetCoins(ctx, storageDepositAddr)
	pkgBalance := env.bankk.GetCoins(ctx, pkgAddr)
	assert.True(t, pkgBalance.IsZero())
	assert.True(t, initialBalance.Sub(pkgStorageDeposit).IsEqual(userAcctBalance))
}

// Sending too much realm package coins fails.
func TestVMKeeperRealmSend2(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "init.gno", Body: `
package test

import (
	"chain"
	"chain/banker"
	"chain/runtime"
)

func init() {
}

func Echo(cur realm, msg string) string {
	addr := runtime.OriginCaller()
	pkgAddr := runtime.CurrentRealm().Address()
	send := chain.Coins{{"ugnot", 10000000}}
	banker_ := banker.NewBanker(banker.BankerTypeRealmSend)
	banker_.SendCoins(pkgAddr, addr, send) // send back
	return "echo:" + msg
}`},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	coins := std.MustParseCoins(ugnot.ValueString(9000000))
	msg2 := NewMsgCall(addr, coins, pkgPath, "Echo", []string{"hello world"})
	// XXX change this into an error and make sure error message is descriptive.
	_, err = env.vmk.Call(ctx, msg2)
	assert.Error(t, err)
}

// Using x/params from a realm.
func TestVMKeeperParams(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	// env.prmk.
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	const pkgPath = "gno.land/r/myuser/myrealm"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "init.gno", Body: `
package params

import "chain/params"

func init() {
	params.SetString("foo.string", "foo1")
}

func Do(cur realm) string {
	params.SetInt64("bar.int64", int64(1337))
	params.SetString("foo.string", "foo2") // override init

	return "XXX" // return std.GetConfig("gno.land/r/test.foo"), if we want to expose std.GetConfig, maybe as a std.TestGetConfig
}`},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	coins := std.MustParseCoins(ugnot.ValueString(8_000_000))
	msg2 := NewMsgCall(addr, coins, pkgPath, "Do", []string{})

	res, err := env.vmk.Call(ctx, msg2)
	assert.NoError(t, err)
	_ = res
	expected := fmt.Sprintf("(\"%s\" string)\n\n", "XXX") // XXX: return something more useful
	assert.Equal(t, expected, res)

	var foo string
	var bar int64
	env.vmk.prmk.GetString(ctx, "vm:gno.land/r/myuser/myrealm:foo.string", &foo)
	env.vmk.prmk.GetInt64(ctx, "vm:gno.land/r/myuser/myrealm:bar.int64", &bar)
	assert.Equal(t, "foo2", foo)
	assert.Equal(t, int64(1337), bar)
}

// Assign admin as OriginCaller on deploying the package.
func TestVMKeeperOriginCallerInit(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "init.gno", Body: `
package test

import (
	"chain/banker"
	"chain/runtime"
)

var admin address

func init() {
	admin = runtime.OriginCaller()
}

func Echo(cur realm, msg string) string {
	addr := runtime.OriginCaller()
	pkgAddr := runtime.CurrentRealm().Address()
	send := banker.OriginSend()
	banker_ := banker.NewBanker(banker.BankerTypeOriginSend)
	banker_.SendCoins(pkgAddr, addr, send) // send back
	return "echo:" + msg
}

func GetAdmin(cur realm) string { // XXX: remove crossing call ?
	return admin.String()
}

`},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run GetAdmin()
	coins := std.MustParseCoins("")
	msg2 := NewMsgCall(addr, coins, pkgPath, "GetAdmin", []string{})
	res, err := env.vmk.Call(ctx, msg2)
	addrString := fmt.Sprintf("(\"%s\" string)\n\n", addr.String())
	assert.NoError(t, err)
	assert.Equal(t, addrString, res)
}

// Call Run without imports, without variables.
func TestVMKeeperRunSimple(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)

	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "script.gno", Body: `
package main

func main() {
	println("hello world!")
}
`},
	}

	coins := std.MustParseCoins("")
	msg2 := NewMsgRun(addr, coins, files)
	res, err := env.vmk.Run(ctx, msg2)
	assert.NoError(t, err)
	assert.Equal(t, "hello world!\n", res)
}

// Call Run with stdlibs.
func TestVMKeeperRunImportStdlibs(t *testing.T) {
	env := setupTestEnv()
	testVMKeeperRunImportStdlibs(t, env)
}

// Call Run with stdlibs, "cold" loading the standard libraries
func TestVMKeeperRunImportStdlibsColdStdlibLoad(t *testing.T) {
	env := setupTestEnvCold()
	testVMKeeperRunImportStdlibs(t, env)
}

func testVMKeeperRunImportStdlibs(t *testing.T, env testEnv) {
	t.Helper()

	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)

	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "script.gno", Body: `
package main

import "chain/runtime"

func main() {
	addr := runtime.OriginCaller()
	println("hello world!", addr)
}
`},
	}

	coins := std.MustParseCoins("")
	msg2 := NewMsgRun(addr, coins, files)
	res, err := env.vmk.Run(ctx, msg2)
	assert.NoError(t, err)
	expectedString := fmt.Sprintf("hello world! %s\n", addr.String())
	assert.Equal(t, expectedString, res)
}

func TestVMKeeperRunImportDraft(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create test package.
	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{
			Name: "gnomod.toml",
			Body: `module = "gno.land/r/test"
gno = "0.9"
draft = true`,
		},
		{
			Name: "test.gno",
			Body: `package test

func Echo(cur realm) string {
	return "hello world"
}`,
		},
	}
	ctx = ctx.WithBlockHeader(&bft.Header{ChainID: "test-chain-id", Height: 0})
	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	ctx = ctx.WithBlockHeader(&bft.Header{ChainID: "test-chain-id", Height: 42})

	// Msg Run Echo function.
	coins := std.MustParseCoins("")
	files = []*std.MemFile{
		{
			Name: "main.gno",
			Body: `
package main

import "gno.land/r/test"

func main() {
	msg := test.Echo(cross)
	println(msg)
}
`,
		},
	}
	msg2 := NewMsgRun(addr, coins, files)
	res, err := env.vmk.Run(ctx, msg2)
	assert.NoError(t, err)
	assert.Equal(t, "hello world\n", res)
}

func TestVMKeeperRunImportPrivate(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create test package 1.
	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{
			Name: "gnomod.toml",
			Body: `module = "gno.land/r/test"
gno = "0.9"
private = true`,
		},
		{
			Name: "test.gno",
			Body: `package test

func Echo(cur realm) string {
	return "hello world"
}`,
		},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	files = []*std.MemFile{
		{
			Name: "main.gno",
			Body: `
package main

import "gno.land/r/test"

func main() {
	msg := test.Echo(cross)
	println(msg)
}
`,
		},
	}

	// Msg Run Echo function.
	coins := std.MustParseCoins("")
	msg2 := NewMsgRun(addr, coins, files)
	_, err = env.vmk.Run(ctx, msg2)
	assert.Error(t, err, ErrTypeCheck(gnolang.ImportPrivateError{PkgPath: pkgPath}))
}

func TestNumberOfArgsError(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create test package.
	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{
			Name: "test.gno",
			Body: `package test

func Echo(cur realm, msg string) string { // XXX remove crossing call ?
	return "echo:"+msg
}`,
		},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Call Echo function with wrong number of arguments
	coins := std.MustParseCoins(ugnot.ValueString(1))
	msg2 := NewMsgCall(addr, coins, pkgPath, "Echo", []string{"hello world", "extra arg"})
	assert.PanicsWithValue(
		t,
		"wrong number of arguments in call to Echo: want 2 got 3",
		func() {
			env.vmk.Call(ctx, msg2)
		},
	)
}

func TestNonCrossingCallError(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create test package.
	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{
			Name: "test.gno",
			Body: `package test
			
func Echo(msg string) string {
	return "echo:"+msg
}
	
func EmptyCall() {
	return
}

`,
		},
	}
	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Call Echo function which is not a crossing call
	coins := std.MustParseCoins(ugnot.ValueString(1))
	msg2 := NewMsgCall(addr, coins, pkgPath, "Echo", []string{"hello world"})
	assert.PanicsWithValue(
		t,
		"function Echo is non-crossing and cannot be called with MsgCall; query with vm/qeval or use MsgRun",
		func() {
			env.vmk.Call(ctx, msg2)
		},
	)

	// Call EmptyCall function which is not a crossing call
	msg3 := NewMsgCall(addr, coins, pkgPath, "EmptyCall", []string{})
	assert.PanicsWithValue(
		t,
		"function EmptyCall is non-crossing and cannot be called with MsgCall; query with vm/qeval or use MsgRun",
		func() {
			env.vmk.Call(ctx, msg3)
		},
	)
}

func TestVMKeeperReinitialize(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)
	assert.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(initialBalance))

	// Create test package.
	const pkgPath = "gno.land/r/test"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "init.gno", Body: `
package test

func Echo(cur realm, msg string) string {
	return "echo:"+msg
}`},
	}

	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	require.NoError(t, err)

	// Run Echo function.
	msg2 := NewMsgCall(addr, nil, pkgPath, "Echo", []string{"hello world"})
	res, err := env.vmk.Call(ctx, msg2)
	require.NoError(t, err)
	assert.Equal(t, `("echo:hello world" string)`+"\n\n", res)

	// Clear out gnovm and reinitialize.
	env.vmk.gnoStore = nil
	mcw := env.ctx.MultiStore().MultiCacheWrap()
	env.vmk.Initialize(log.NewNoopLogger(), mcw)
	mcw.MultiWrite()

	// Run echo again, and it should still work.
	res, err = env.vmk.Call(ctx, msg2)
	require.NoError(t, err)
	assert.Equal(t, `("echo:hello world" string)`+"\n\n", res)
}

func Test_loadStdlibPackage(t *testing.T) {
	mdb := memdb.NewMemDB()
	cs := dbadapter.StoreConstructor(mdb, types.StoreOptions{})

	gs := gnolang.NewStore(nil, cs, cs)
	assert.PanicsWithError(t, `failed loading stdlib "notfound": does not exist`, func() {
		loadStdlibPackage("notfound", "./testdata", gs)
	})
	assert.PanicsWithError(t, `failed loading stdlib "emptystdlib": package has no files`, func() {
		loadStdlibPackage("emptystdlib", "./testdata", gs)
	})
}

func TestVMKeeperAddPackage_DevelopmentModeFails(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)

	const pkgPath = "gno.land/r/testdev"
	// gnomod.toml with develop = 1
	gnomodToml := `[module]
path = "gno.land/r/testdev"

[gno]
version = "0.9"

[develop]
[[develop.replace]]
old = "foo"
new = "bar"
`
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnomodToml},
		{Name: "test.gno", Body: `package testdev
func Echo(cur realm) string { return "dev" }`},
	}
	msg := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg)
	assert.Error(t, err, ErrInvalidPackage(""))
}

func TestVMKeeperAddPackage_PatchGnomodToml(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("addr2"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)

	const pkgPath = "gno.land/r/testpatch"
	gnomodToml := `module = "gno.land/r/anothername"
gno = "0.9"
`
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnomodToml},
		{Name: "test.gno", Body: `package testpatch
func Echo(cur realm) string { return "patched" }`},
	}
	msg := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg)
	require.NoError(t, err)

	// Check that gnomod.toml was patched
	store := env.vmk.getGnoTransactionStore(ctx)

	memFile := store.GetMemFile(pkgPath, "gnomod.toml")
	mpkg, err := gnomod.ParseBytes("gnomod.toml", []byte(memFile.Body))
	require.NoError(t, err)
	expected := `module = "gno.land/r/testpatch"
gno = "0.9"

[addpkg]
  creator = "g1cq2j7y4utseeatek2alfy5ttaphjrtdx67mg8v"
  height = 42
`
	// XXX: custom height
	assert.Equal(t, expected, mpkg.WriteString())
}

func TestProcessStorageDeposit(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)

	// Create a test package and it's dependence.
	pkgPathFoo := "gno.land/r/foo"
	files := []*std.MemFile{
		{Name: "foo.gno", Body: `
package foo

var Msg string
func Bar(cur realm, msg string){
	Msg = msg
}`},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPathFoo)},
	}

	msg := NewMsgAddPackage(addr, pkgPathFoo, files)
	err := env.vmk.AddPackage(ctx, msg)
	assert.NoError(t, err)
	// varify the account balance
	depAddrFoo := gnolang.DeriveStorageDepositCryptoAddr(pkgPathFoo)
	userBalance := env.bankk.GetCoins(ctx, addr)
	depFoo := env.bankk.GetCoins(ctx, depAddrFoo)
	assert.True(t, userBalance.Add(depFoo).IsEqual(initialBalance))

	pkgPathTest := "gno.land/r/test"
	files = []*std.MemFile{
		{Name: "foo.gno", Body: `
package test
import "gno.land/r/foo"

var Msg string
func Echo(cur realm, msg string){
	Msg = msg
	foo.Bar(cross, msg)
}`},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPathTest)},
	}
	msg = NewMsgAddPackage(addr, pkgPathTest, files)
	err = env.vmk.AddPackage(ctx, msg)
	assert.NoError(t, err)
	// Varify the account balance
	depAddrTest := gnolang.DeriveStorageDepositCryptoAddr(pkgPathTest)
	userBalance = env.bankk.GetCoins(ctx, addr)
	depTest := env.bankk.GetCoins(ctx, depAddrTest)
	assert.True(t, userBalance.Add(depTest).Add(depFoo).IsEqual(initialBalance))

	// Run Echo function.
	msg2 := NewMsgCall(addr, std.Coins{}, pkgPathTest, "Echo", []string{"hello world"})
	msg2.MaxDeposit = std.MustParseCoins(ugnot.ValueString(8000))
	_, err = env.vmk.Call(ctx, msg2)
	assert.NoError(t, err)

	// Verify that the combined deposit equals msg2.MaxDeposit.
	depDeltaTest := env.bankk.GetCoins(ctx, depAddrTest).Sub(depTest)
	depDeltaFoo := env.bankk.GetCoins(ctx, depAddrFoo).Sub(depFoo)
	assert.True(t, depDeltaTest.Add(depDeltaFoo).IsEqual(msg2.MaxDeposit))
}

// TestVMKeeper_RealmDiffIterationDeterminism is a regression test for issue #4580.
// It verifies that the processStorageDeposit function iterates over realms
// in a deterministic order by sorting the realm paths before iteration.
// Without the fix, different runs would produce different error messages
// due to non-deterministic map iteration order.
func TestVMKeeper_RealmDiffIterationDeterminism(t *testing.T) {
	// This test creates multiple realms with different names that would iterate
	// in different orders in a map. It then triggers storage operations that
	// exceed the deposit limit, causing processStorageDeposit to fail.
	// The specific error message depends on which realm is processed first.
	// With proper sorting in processStorageDeposit, the error should be
	// deterministic across multiple runs.
	const numRuns = 5

	runOperations := func() (string, error) {
		env := setupTestEnv()
		ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

		caller := crypto.AddressFromPreimage([]byte("caller"))
		acc := env.acck.NewAccountWithAddress(ctx, caller)
		env.acck.SetAccount(ctx, acc)

		// Give enough coins for package creation
		env.bankk.SetCoins(ctx, caller, std.MustParseCoins(ugnot.ValueString(100_000_000)))

		// Create realms with names designed to have different map iteration orders
		realms := []string{
			"gno.land/r/test/realm_aaa",
			"gno.land/r/test/realm_zzz",
			"gno.land/r/test/realm_mmm",
			"gno.land/r/test/realm_001",
			"gno.land/r/test/realm_999",
			"gno.land/r/test/realm_abc",
			"gno.land/r/test/realm_xyz",
			"gno.land/r/test/realm_123",
			"gno.land/r/test/realm_789",
			"gno.land/r/test/realm_def",
		}

		// Create each realm
		for i, realmPath := range realms {
			files := []*std.MemFile{
				{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(realmPath)},
				{
					Name: "realm.gno",
					Body: fmt.Sprintf(`package %s

var storage []string

func UpdateStorage(cur realm, n int) {
	// Force storage growth based on realm index
	for i := 0; i < n+%d*100; i++ {
		storage = append(storage, "data_data_data_data")
	}
}`, path.Base(realmPath), i),
				},
			}
			msg := NewMsgAddPackage(caller, realmPath, files)
			err := env.vmk.AddPackage(ctx, msg)
			require.NoError(t, err)
		}

		// Create master realm
		masterPath := "gno.land/r/test/master"

		// Build imports and calls dynamically
		imports := ""
		calls := ""
		for _, realmPath := range realms {
			alias := path.Base(realmPath)
			imports += fmt.Sprintf("\t%s \"%s\"\n", alias, realmPath)
			calls += fmt.Sprintf("\t%s.UpdateStorage(cross, 500)\n", alias)
		}

		masterCode := fmt.Sprintf(`package master

import (
%s)

func UpdateAll(cur realm) {
%s}`, imports, calls)

		masterFiles := []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(masterPath)},
			{Name: "master.gno", Body: masterCode},
		}
		msg := NewMsgAddPackage(caller, masterPath, masterFiles)
		err := env.vmk.AddPackage(ctx, msg)
		require.NoError(t, err)

		// Call with limited deposit to force errors in processStorageDeposit
		// The error will depend on which realms get processed first
		callMsg := NewMsgCall(caller, std.Coins{}, masterPath, "UpdateAll", []string{})
		callMsg.MaxDeposit = std.MustParseCoins(ugnot.ValueString(20_000_000))

		// Capture the error - it should vary based on iteration order
		_, err = env.vmk.Call(ctx, callMsg)

		env.vmk.CommitGnoTransactionStore(ctx)

		// Return error string which should vary with iteration order
		if err != nil {
			return err.Error(), err
		}
		return "no_error", nil
	}

	// Track first error message as baseline
	firstMsg, _ := runOperations()

	// Check subsequent runs for differences
	for i := 1; i < numRuns; i++ {
		errMsg, _ := runOperations()

		// If we find a different error message, it indicates non-deterministic behavior.
		// This should NOT happen with the sorting fix in processStorageDeposit.
		if errMsg != firstMsg {
			t.Fatalf("Non-deterministic behavior detected at run %d!\nFirst error: %s\nDifferent error at run %d: %s\n\nThis indicates the determinism fix in processStorageDeposit is not working correctly.",
				i+1, firstMsg, i+1, errMsg)
		}

		// Force GC and allocations to change runtime state
		runtime.GC()
		// Create some allocations to change heap state
		temp := make([]map[string]int, 100)
		for j := range temp {
			temp[j] = make(map[string]int)
			temp[j]["key"] = j
		}
	}

	// All runs produced identical results - this is expected with the fix applied
	t.Logf("SUCCESS: All %d runs produced identical results, confirming deterministic behavior", numRuns)
}

// TestVMKeeperCLASignature tests CLA enforcement during package deployment.
// Uses a minimal inline CLA realm to test the keeper's CLA check mechanism
// without requiring the full govdao dependency chain.
func TestVMKeeperCLASignature(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Create admin and user addresses
	admin := crypto.AddressFromPreimage([]byte("admin"))
	user := crypto.AddressFromPreimage([]byte("user"))

	// Set up accounts with initial balance
	adminAcc := env.acck.NewAccountWithAddress(ctx, admin)
	env.acck.SetAccount(ctx, adminAcc)
	env.bankk.SetCoins(ctx, admin, initialBalance)

	userAcc := env.acck.NewAccountWithAddress(ctx, user)
	env.acck.SetAccount(ctx, userAcc)
	env.bankk.SetCoins(ctx, user, initialBalance)

	// Deploy a minimal inline CLA realm for testing.
	// This avoids deploying the full govdao dependency chain; the keeper test
	// only needs HasValidSignature, Sign, and a way to set the required hash.
	const claPkgPath = "gno.land/r/sys/cla"
	claFiles := []*std.MemFile{
		{Name: "cla.gno", Body: `package cla

import "chain/runtime"

var (
	requiredHash string
	signatures   map[address]bool
)

func init() { signatures = make(map[address]bool) }

func SetRequiredHash(cur realm, newHash string) {
	requiredHash = newHash
	signatures = make(map[address]bool)
}

func Sign(cur realm, hash string) {
	if hash != requiredHash {
		panic("hash does not match required CLA hash")
	}
	caller := runtime.PreviousRealm().Address()
	signatures[caller] = true
}

func HasValidSignature(addr address) bool {
	if requiredHash == "" {
		return true
	}
	return signatures[addr]
}
`},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(claPkgPath)},
	}
	claMsg := NewMsgAddPackage(admin, claPkgPath, claFiles)
	err := env.vmk.AddPackage(ctx, claMsg)
	require.NoError(t, err, "failed to deploy inline cla realm")

	// Test 1: CLA disabled (empty hash) - user can deploy
	const userPkgPath1 = "gno.land/r/user/pkg1"
	userFiles1 := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(userPkgPath1)},
		{Name: "pkg.gno", Body: `package pkg1
func Hello(cur realm) string { return "hello" }`},
	}
	userMsg1 := NewMsgAddPackage(user, userPkgPath1, userFiles1)
	err = env.vmk.AddPackage(ctx, userMsg1)
	assert.NoError(t, err, "should allow deployment when CLA is disabled")

	// Test 2: Enable CLA - user should be blocked
	setHashMsg := NewMsgCall(admin, nil, claPkgPath, "SetRequiredHash", []string{"testhash123"})
	_, err = env.vmk.Call(ctx, setHashMsg)
	require.NoError(t, err)

	const userPkgPath2 = "gno.land/r/user/pkg2"
	userFiles2 := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(userPkgPath2)},
		{Name: "pkg.gno", Body: `package pkg2
func Hello(cur realm) string { return "hello" }`},
	}
	userMsg2 := NewMsgAddPackage(user, userPkgPath2, userFiles2)
	err = env.vmk.AddPackage(ctx, userMsg2)
	require.Error(t, err, "should block deployment when user hasn't signed CLA")
	assert.True(t, errors.Is(err, UnauthorizedUserError{}), "error should be UnauthorizedUserError, got: %v", err)

	// Test 3: User signs CLA - can deploy
	signMsg := NewMsgCall(user, nil, claPkgPath, "Sign", []string{"testhash123"})
	_, err = env.vmk.Call(ctx, signMsg)
	require.NoError(t, err)

	const userPkgPath3 = "gno.land/r/user/pkg3"
	userFiles3 := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(userPkgPath3)},
		{Name: "pkg.gno", Body: `package pkg3
func Hello(cur realm) string { return "hello" }`},
	}
	userMsg3 := NewMsgAddPackage(user, userPkgPath3, userFiles3)
	err = env.vmk.AddPackage(ctx, userMsg3)
	assert.NoError(t, err, "should allow deployment after signing CLA")

	// Test 4: Admin changes hash - user signature reset, blocked again
	setHashMsg2 := NewMsgCall(admin, nil, claPkgPath, "SetRequiredHash", []string{"newhash456"})
	_, err = env.vmk.Call(ctx, setHashMsg2)
	require.NoError(t, err)

	const userPkgPath4 = "gno.land/r/user/pkg4"
	userFiles4 := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(userPkgPath4)},
		{Name: "pkg.gno", Body: `package pkg4
func Hello(cur realm) string { return "hello" }`},
	}
	userMsg4 := NewMsgAddPackage(user, userPkgPath4, userFiles4)
	err = env.vmk.AddPackage(ctx, userMsg4)
	assert.Error(t, err, "should block after hash change resets signatures")

	// Test 5: Disable CLA - user can deploy again
	setHashMsg3 := NewMsgCall(admin, nil, claPkgPath, "SetRequiredHash", []string{""})
	_, err = env.vmk.Call(ctx, setHashMsg3)
	require.NoError(t, err)

	const userPkgPath5 = "gno.land/r/user/pkg5"
	userFiles5 := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(userPkgPath5)},
		{Name: "pkg.gno", Body: `package pkg5
func Hello(cur realm) string { return "hello" }`},
	}
	userMsg5 := NewMsgAddPackage(user, userPkgPath5, userFiles5)
	err = env.vmk.AddPackage(ctx, userMsg5)
	assert.NoError(t, err, "should allow deployment when CLA is disabled again")
}

// TestVMKeeperCLASignature_RealmNotDeployed tests that CLA check is skipped
// when the CLA realm is not yet deployed (bootstrap scenario).
func TestVMKeeperCLASignature_RealmNotDeployed(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Create user account
	user := crypto.AddressFromPreimage([]byte("user"))
	userAcc := env.acck.NewAccountWithAddress(ctx, user)
	env.acck.SetAccount(ctx, userAcc)
	env.bankk.SetCoins(ctx, user, initialBalance)

	// CLA realm is not deployed, but SysCLAPkgPath is set (default).
	// This must succeed to allow bootstrap (deploying the CLA realm itself).

	const userPkgPath = "gno.land/r/user/pkg1"
	userFiles := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(userPkgPath)},
		{Name: "pkg.gno", Body: `package pkg1
func Hello(cur realm) string { return "hello" }`},
	}
	userMsg := NewMsgAddPackage(user, userPkgPath, userFiles)
	err := env.vmk.AddPackage(ctx, userMsg)
	assert.NoError(t, err, "should allow deployment when CLA realm is not deployed (bootstrap)")
}

// TestSessionMsgCallSendCountsAgainstSpendLimit verifies that when a
// session-signed MsgCall carries a non-zero msg.Send, the VMKeeper's
// vm.bank.SendCoins transfer (caller → pkgAddr) routes through the
// bank.Keeper.SendCoins session hook and debits SpendLimit. This is
// the end-to-end pipeline: VMKeeper.Call -> vm.bank.SendCoins ->
// session hook.
//
// Note on the "in-realm draining" threat model: gno banker types
// restrict realms from draining master's arbitrary coins. Realms can
// only redirect the msg.Send amount (BankerTypeOriginSend) or spend
// their own pkg balance (BankerTypeRealmSend). So the material
// protection is on msg.Send itself, which this test exercises.
func TestSessionMsgCallSendCountsAgainstSpendLimit(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Fund master.
	master := crypto.AddressFromPreimage([]byte("master-mcsend"))
	masterAcc := env.acck.NewAccountWithAddress(ctx, master)
	env.acck.SetAccount(ctx, masterAcc)
	env.bankk.SetCoins(ctx, master, initialBalance)

	// Deploy a simple realm that accepts coins (no-op on them).
	const pkgPath = "gno.land/r/absorb"
	files := []*std.MemFile{
		{Name: "absorb.gno", Body: `
package absorb

func Absorb(cur realm) {}`},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(master, pkgPath, files)))

	// Create a session with 500k ugnot SpendLimit.
	_, sessionPub, sessionPubAddr := tu.KeyTestPubAddr()
	sa := env.acck.NewSessionAccount(ctx, master, sessionPub)
	da := sa.(std.DelegatedAccount)
	da.SetExpiresAt(0)
	da.SetSpendLimit(std.Coins{std.NewCoin("ugnot", 500_000)})
	da.SetSpendReset(ctx.BlockTime().Unix())
	env.acck.SetSessionAccount(ctx, master, sa)

	sessions := map[crypto.Address]std.DelegatedAccount{master: da}
	sessionCtx := ctx.WithValue(std.SessionAccountsContextKey{}, sessions)

	pkgAddr := gnolang.DerivePkgCryptoAddr(pkgPath)
	masterBefore := env.bankk.GetCoins(sessionCtx, master).AmountOf("ugnot")
	pkgBefore := env.bankk.GetCoins(sessionCtx, pkgAddr).AmountOf("ugnot")

	// Call 1: msg.Send = 200k → VMKeeper moves coins via vm.bank.SendCoins
	// → session hook fires → SpendUsed += 200k.
	msg1 := NewMsgCall(master, std.Coins{std.NewCoin("ugnot", 200_000)}, pkgPath, "Absorb", nil)
	_, err := env.vmk.Call(sessionCtx, msg1)
	require.NoError(t, err, "MsgCall with msg.Send within SpendLimit should succeed")

	// SpendUsed reflects the send.
	reloadedSA := env.acck.GetSessionAccount(sessionCtx, master, sessionPubAddr)
	require.NotNil(t, reloadedSA)
	assert.Equal(t, int64(200_000), reloadedSA.(std.DelegatedAccount).GetSpendUsed().AmountOf("ugnot"),
		"msg.Send must debit session SpendUsed via bank-keeper hook")

	// Coins moved master → pkg.
	masterAfter := env.bankk.GetCoins(sessionCtx, master).AmountOf("ugnot")
	pkgAfter := env.bankk.GetCoins(sessionCtx, pkgAddr).AmountOf("ugnot")
	assert.Equal(t, int64(200_000), masterBefore-masterAfter)
	assert.Equal(t, int64(200_000), pkgAfter-pkgBefore)

	// Call 2: msg.Send = 400k — remaining budget is 300k, must reject.
	msg2 := NewMsgCall(master, std.Coins{std.NewCoin("ugnot", 400_000)}, pkgPath, "Absorb", nil)
	_, err = env.vmk.Call(sessionCtx, msg2)
	require.Error(t, err, "msg.Send exceeding remaining SpendLimit must be rejected")

	// SpendUsed unchanged by the rejected call.
	reloadedSA = env.acck.GetSessionAccount(sessionCtx, master, sessionPubAddr)
	assert.Equal(t, int64(200_000), reloadedSA.(std.DelegatedAccount).GetSpendUsed().AmountOf("ugnot"))
}

// TestSessionStorageDepositExceedsLimit verifies that storage deposits
// triggered by session-signed MsgCall count against SpendLimit. A
// tight-budget session cannot grow realm state more than its remaining
// budget allows.
func TestSessionStorageDepositExceedsLimit(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	master := crypto.AddressFromPreimage([]byte("master-storage"))
	masterAcc := env.acck.NewAccountWithAddress(ctx, master)
	env.acck.SetAccount(ctx, masterAcc)
	env.bankk.SetCoins(ctx, master, initialBalance)

	// Deploy a realm that grows state when Grow is called.
	const pkgPath = "gno.land/r/growstate"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "growstate.gno", Body: `
package growstate

var Msg string

func Grow(cur realm, s string) {
	Msg = s
}`},
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(master, pkgPath, files)))

	// Session with a SpendLimit so tight that any meaningful storage
	// growth will exceed it: 1 ugnot.
	_, sessionPub, sessionPubAddr := tu.KeyTestPubAddr()
	sa := env.acck.NewSessionAccount(ctx, master, sessionPub)
	da := sa.(std.DelegatedAccount)
	da.SetExpiresAt(0)
	da.SetSpendLimit(std.Coins{std.NewCoin("ugnot", 1)})
	da.SetSpendReset(ctx.BlockTime().Unix())
	env.acck.SetSessionAccount(ctx, master, sa)

	sessions := map[crypto.Address]std.DelegatedAccount{master: da}
	sessionCtx := ctx.WithValue(std.SessionAccountsContextKey{}, sessions)

	// Call Grow with a non-trivial string → requires storage deposit
	// many orders of magnitude above 1 ugnot.
	longStr := strings.Repeat("x", 256)
	msg := NewMsgCall(master, std.Coins{}, pkgPath, "Grow", []string{longStr})
	msg.MaxDeposit = std.MustParseCoins(ugnot.ValueString(8000))
	_, err := env.vmk.Call(sessionCtx, msg)
	require.Error(t, err, "storage deposit exceeding session SpendLimit must be rejected")

	// Session's SpendUsed unchanged — the check rejected before persisting.
	reloadedSA := env.acck.GetSessionAccount(sessionCtx, master, sessionPubAddr)
	assert.Equal(t, int64(0), reloadedSA.(std.DelegatedAccount).GetSpendUsed().AmountOf("ugnot"))
}

