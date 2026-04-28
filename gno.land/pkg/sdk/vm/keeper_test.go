package vm

// TODO: move most of the logic in ROOT/gno.land/...

import (
	"encoding/json"
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
	"github.com/gnolang/gno/tm2/pkg/amino"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
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

func TestVMKeeperEvalJSONFormatting2(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(9000000)))

	tests := []struct {
		name      string
		pkgBody   string
		expr      string
		checkStrs []string // substrings that must appear in output
	}{
		{
			name:    "JSON string",
			pkgBody: `func GetString() string { return "hello" }`,
			expr:    "GetString()",
			checkStrs: []string{
				`/gno.PrimitiveType`,
				`/gno.StringValue`,
				`"value":"hello"`,
			},
		},
		{
			name:    "JSON integer",
			pkgBody: `func GetInt() int { return 42 }`,
			expr:    "GetInt()",
			checkStrs: []string{
				`/gno.PrimitiveType`,
				`"N":`, // int value in base64
			},
		},
		{
			name:    "JSON boolean",
			pkgBody: `func GetBool() bool { return true }`,
			expr:    "GetBool()",
			checkStrs: []string{
				`/gno.PrimitiveType`,
				`"N":`, // bool value in base64
			},
		},
		{
			name:    "JSON bytes",
			pkgBody: `func GetBytes() []byte { return []byte("test") }`,
			expr:    "GetBytes()",
			checkStrs: []string{
				`/gno.SliceType`,
				`/gno.SliceValue`,
				`/gno.ArrayValue`,
				`"Data":"dGVzdA=="`,
				`"Length":"4"`,
			},
		},
		{
			name:    "JSON multiple values",
			pkgBody: `func GetMulti() (string, int) { return "hello", 42 }`,
			expr:    "GetMulti()",
			checkStrs: []string{
				`/gno.PrimitiveType`,
				`"value":"hello"`,
				`"N":`, // int value
			},
		},
	}

	for i, tc := range tests {
		pkgPath := fmt.Sprintf("gno.land/r/hello%d", i)
		pkgName := fmt.Sprintf("hello%d", i)
		pkgBody := fmt.Sprintf("package %s\n%s", pkgName, tc.pkgBody)
		t.Run(tc.name, func(t *testing.T) {
			files := []*std.MemFile{
				{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
				{Name: "hello.gno", Body: pkgBody},
			}

			msg1 := NewMsgAddPackage(addr, pkgPath, files)
			err := env.vmk.AddPackage(ctx, msg1)
			assert.NoError(t, err)
			env.vmk.CommitGnoTransactionStore(ctx)

			res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, tc.expr)
			require.NoError(t, err)
			// Verify valid JSON
			var result map[string]json.RawMessage
			require.NoError(t, json.Unmarshal([]byte(res), &result))
			require.Contains(t, string(result["results"]), "@type")
			for _, s := range tc.checkStrs {
				assert.Contains(t, res, s)
			}
		})
	}
}

// TestVMKeeperEvalJSONPersistedObjects tests JSON output for persisted realm objects
// with real ObjectIDs stored in the database.
func TestVMKeeperEvalJSONPersistedObjects(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(20000000)))

	t.Run("persisted_struct_pointer_with_objectid", func(t *testing.T) {
		pkgPath := "gno.land/r/test/persisted1"
		pkgBody := `package persisted1

type Item struct {
	ID   int
	Name string
}

var item *Item

func init() {
	item = &Item{ID: 42, Name: "test item"}
}

func GetItem() *Item {
	return item
}`

		files := []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
			{Name: "item.gno", Body: pkgBody},
		}

		msg := NewMsgAddPackage(addr, pkgPath, files)
		err := env.vmk.AddPackage(ctx, msg)
		require.NoError(t, err)
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetItem()")
		require.NoError(t, err)

		// Verify Amino format: PointerType wrapping RefType
		assert.Contains(t, res, `/gno.PointerType`)
		assert.Contains(t, res, `"ID":"gno.land/r/test/persisted1.Item"`)
		// Pointer shows RefValue to the persisted object
		assert.Contains(t, res, `/gno.PointerValue`)
		assert.Contains(t, res, `/gno.RefValue`)
		// RefValue contains queryable ObjectID
		assert.Contains(t, res, `"ObjectID":"`)
	})

	t.Run("persisted_map", func(t *testing.T) {
		pkgPath := "gno.land/r/test/persisted4"
		pkgBody := `package persisted4

var data map[string]int

func init() {
	data = make(map[string]int)
	data["one"] = 1
	data["two"] = 2
}

func GetData() map[string]int {
	return data
}`

		files := []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
			{Name: "map.gno", Body: pkgBody},
		}

		msg := NewMsgAddPackage(addr, pkgPath, files)
		err := env.vmk.AddPackage(ctx, msg)
		require.NoError(t, err)
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetData()")
		require.NoError(t, err)

		// Verify Amino format - persisted map shows as MapType + RefValue
		assert.Contains(t, res, `/gno.MapType`)
		assert.Contains(t, res, `/gno.RefValue`)
		assert.Contains(t, res, `"ObjectID":"`)
	})

	t.Run("persisted_declared_type", func(t *testing.T) {
		pkgPath := "gno.land/r/test/persisted7"
		pkgBody := `package persisted7

type Amount int64

var amount Amount

func init() {
	amount = 1000
}

func GetAmount() Amount {
	return amount
}`

		files := []*std.MemFile{
			{Name: "a.gno", Body: pkgBody},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		}

		msg := NewMsgAddPackage(addr, pkgPath, files)
		err := env.vmk.AddPackage(ctx, msg)
		require.NoError(t, err)
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetAmount()")
		require.NoError(t, err)

		// Verify Amino format for declared type: RefType with ID
		assert.Contains(t, res, `"ID":"gno.land/r/test/persisted7.Amount"`)
		assert.Contains(t, res, `/gno.RefType`)
		// Primitive value stored in N (base64)
		assert.Contains(t, res, `"N":`)
	})

	t.Run("persisted_error_type", func(t *testing.T) {
		pkgPath := "gno.land/r/test/persisted8"
		pkgBody := `package persisted8

type CustomError struct {
	Code    int
	Message string
}

func (e *CustomError) Error() string {
	return e.Message
}

var lastError *CustomError

func init() {
	lastError = &CustomError{Code: 404, Message: "not found"}
}

func GetError() error {
	return lastError
}`

		files := []*std.MemFile{
			{Name: "a.gno", Body: pkgBody},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		}

		msg := NewMsgAddPackage(addr, pkgPath, files)
		err := env.vmk.AddPackage(ctx, msg)
		require.NoError(t, err)
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetError()")
		require.NoError(t, err)

		// Verify Amino format: PointerType wrapping RefType
		assert.Contains(t, res, `/gno.PointerType`)
		assert.Contains(t, res, `"ID":"gno.land/r/test/persisted8.CustomError"`)
		// Shows as PointerValue with RefValue to the persisted struct
		assert.Contains(t, res, `/gno.PointerValue`)
		assert.Contains(t, res, `/gno.RefValue`)
		// @error is populated from the result's .Error() method.
		assert.Contains(t, res, `"@error":"not found"`)
	})

	t.Run("persisted_json_tags", func(t *testing.T) {
		pkgPath := "gno.land/r/test/persisted9"
		pkgBody := `package persisted9

type Person struct {
	FirstName string ` + "`json:\"first_name\"`" + `
	LastName  string ` + "`json:\"last_name\"`" + `
	Age       int    ` + "`json:\"age,omitempty\"`" + `
}

var person *Person

func init() {
	person = &Person{FirstName: "John", LastName: "Doe", Age: 30}
}

func GetPerson() *Person {
	return person
}`

		files := []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
			{Name: "person.gno", Body: pkgBody},
		}

		msg := NewMsgAddPackage(addr, pkgPath, files)
		err := env.vmk.AddPackage(ctx, msg)
		require.NoError(t, err)
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetPerson()")
		require.NoError(t, err)

		// Verify Amino format: PointerType wrapping RefType
		assert.Contains(t, res, `/gno.PointerType`)
		assert.Contains(t, res, `"ID":"gno.land/r/test/persisted9.Person"`)
		assert.Contains(t, res, `/gno.PointerValue`)
		assert.Contains(t, res, `/gno.RefValue`)
	})

	t.Run("persisted_nil_pointer", func(t *testing.T) {
		pkgPath := "gno.land/r/test/persisted10"
		pkgBody := `package persisted10

type Data struct {
	Value int
}

var ptr *Data

func GetPtr() *Data {
	return ptr
}`

		files := []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
			{Name: "nil.gno", Body: pkgBody},
		}

		msg := NewMsgAddPackage(addr, pkgPath, files)
		err := env.vmk.AddPackage(ctx, msg)
		require.NoError(t, err)
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetPtr()")
		require.NoError(t, err)

		// Verify Amino format for nil pointer: PointerType, no V field (amino omits nil)
		assert.Contains(t, res, `/gno.PointerType`)
		assert.Contains(t, res, `"ID":"gno.land/r/test/persisted10.Data"`)
		// Nil pointer: V is omitted by amino, not "V":null
		assert.NotContains(t, res, `"V":`)
	})

	// Regression test: nil pointer in struct field should not panic
	// This tests the fix for pv.TV == nil check in jsonValueSimple
	t.Run("persisted_struct_with_nil_pointer_field", func(t *testing.T) {
		pkgPath := "gno.land/r/test/persisted11"
		pkgBody := `package persisted11

type Child struct {
	Value int
}

type Parent struct {
	Name  string
	Child *Child
}

var parent *Parent

func init() {
	parent = &Parent{Name: "test", Child: nil}
}

func GetParent() *Parent {
	return parent
}`

		files := []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
			{Name: "parent.gno", Body: pkgBody},
		}

		msg := NewMsgAddPackage(addr, pkgPath, files)
		err := env.vmk.AddPackage(ctx, msg)
		require.NoError(t, err)
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetParent()")
		require.NoError(t, err)

		// Verify Amino format: PointerType wrapping RefType
		assert.Contains(t, res, `/gno.PointerType`)
		assert.Contains(t, res, `"ID":"gno.land/r/test/persisted11.Parent"`)
		assert.Contains(t, res, `/gno.PointerValue`)
		assert.Contains(t, res, `/gno.RefValue`)
	})

	// Regression test: persisted slice with RefValue base should not panic
	// This tests the fix for passing m.Store instead of nil to GetPointerAtIndexInt2
	t.Run("persisted_slice_of_primitives", func(t *testing.T) {
		pkgPath := "gno.land/r/test/persisted12"
		pkgBody := `package persisted12

var numbers []int

func init() {
	numbers = []int{10, 20, 30, 40, 50}
}

func GetNumbers() []int {
	return numbers
}`

		files := []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
			{Name: "numbers.gno", Body: pkgBody},
		}

		msg := NewMsgAddPackage(addr, pkgPath, files)
		err := env.vmk.AddPackage(ctx, msg)
		require.NoError(t, err)
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetNumbers()")
		require.NoError(t, err)

		// Verify Amino format for persisted slice
		assert.Contains(t, res, `/gno.SliceType`)
		assert.Contains(t, res, `/gno.SliceValue`)
		assert.Contains(t, res, `/gno.RefValue`)
	})

	// Regression test: persisted slice of structs with RefValue base
	t.Run("persisted_slice_of_structs", func(t *testing.T) {
		pkgPath := "gno.land/r/test/persisted13"
		pkgBody := `package persisted13

type Item struct {
	ID   int
	Name string
}

var items []Item

func init() {
	items = []Item{
		{ID: 1, Name: "first"},
		{ID: 2, Name: "second"},
	}
}

func GetItems() []Item {
	return items
}`

		files := []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
			{Name: "items.gno", Body: pkgBody},
		}

		msg := NewMsgAddPackage(addr, pkgPath, files)
		err := env.vmk.AddPackage(ctx, msg)
		require.NoError(t, err)
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetItems()")
		require.NoError(t, err)

		// Verify Amino format for slice of structs
		assert.Contains(t, res, `/gno.SliceType`)
		assert.Contains(t, res, `"ID":"gno.land/r/test/persisted13.Item"`)
		assert.Contains(t, res, `/gno.SliceValue`)
		assert.Contains(t, res, `/gno.RefValue`)
	})

	// Regression test: deeply nested persisted struct with nil pointers at various levels
	t.Run("persisted_nested_with_nil_pointers", func(t *testing.T) {
		pkgPath := "gno.land/r/test/persisted14"
		pkgBody := `package persisted14

type Level3 struct {
	Data string
}

type Level2 struct {
	L3  *Level3
	Nil *Level3
}

type Level1 struct {
	L2 *Level2
}

var root *Level1

func init() {
	root = &Level1{
		L2: &Level2{
			L3:  &Level3{Data: "deep"},
			Nil: nil,
		},
	}
}

func GetRoot() *Level1 {
	return root
}`

		files := []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
			{Name: "nested.gno", Body: pkgBody},
		}

		msg := NewMsgAddPackage(addr, pkgPath, files)
		err := env.vmk.AddPackage(ctx, msg)
		require.NoError(t, err)
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetRoot()")
		require.NoError(t, err)

		// Verify Amino format: PointerType wrapping RefType
		assert.Contains(t, res, `/gno.PointerType`)
		assert.Contains(t, res, `"ID":"gno.land/r/test/persisted14.Level1"`)
		assert.Contains(t, res, `/gno.PointerValue`)
		assert.Contains(t, res, `/gno.RefValue`)
		// ObjectID present for fetching the object
		assert.Contains(t, res, `"ObjectID":"`)
	})
}

// TestVMKeeperEvalJSONError verifies that QueryEvalJSON populates the
// top-level "@error" field when the evaluated expression returns a non-nil
// value that implements the error interface.
//
// Both an ephemeral error (constructed at query time) and a persisted error
// value must have their .Error() string extracted into the @error field, so
// clients don't have to perform a second round-trip to decode the message.
func TestVMKeeperEvalJSONError(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(20000000)))

	t.Run("ephemeral_error", func(t *testing.T) {
		pkgPath := "gno.land/r/test/ephemerr"
		pkgBody := `package ephemerr

type MyErr struct{ Msg string }

func (e *MyErr) Error() string { return e.Msg }

func GetError() error { return &MyErr{Msg: "boom"} }`

		files := []*std.MemFile{
			{Name: "a.gno", Body: pkgBody},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		}
		msg := NewMsgAddPackage(addr, pkgPath, files)
		require.NoError(t, env.vmk.AddPackage(ctx, msg))
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetError()")
		require.NoError(t, err)

		assert.Contains(t, res, `"@error":"boom"`,
			"ephemeral error should have its .Error() extracted into @error; got: %s", res)
	})

	t.Run("persisted_error", func(t *testing.T) {
		pkgPath := "gno.land/r/test/persisterr"
		pkgBody := `package persisterr

type CustomError struct {
	Code    int
	Message string
}

func (e *CustomError) Error() string { return e.Message }

var lastError *CustomError

func init() {
	lastError = &CustomError{Code: 404, Message: "not found"}
}

func GetError() error { return lastError }`

		files := []*std.MemFile{
			{Name: "a.gno", Body: pkgBody},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		}
		msg := NewMsgAddPackage(addr, pkgPath, files)
		require.NoError(t, env.vmk.AddPackage(ctx, msg))
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetError()")
		require.NoError(t, err)

		assert.Contains(t, res, `"@error":"not found"`,
			"persisted error should have its .Error() extracted into @error; got: %s", res)
	})

	t.Run("nil_error_no_field", func(t *testing.T) {
		pkgPath := "gno.land/r/test/nilerr"
		pkgBody := `package nilerr
func GetError() error { return nil }`

		files := []*std.MemFile{
			{Name: "a.gno", Body: pkgBody},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		}
		msg := NewMsgAddPackage(addr, pkgPath, files)
		require.NoError(t, env.vmk.AddPackage(ctx, msg))
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetError()")
		require.NoError(t, err)

		assert.NotContains(t, res, `"@error"`,
			"nil error should not produce an @error field; got: %s", res)
	})

	t.Run("oog_in_error_method_graceful_degrade", func(t *testing.T) {
		// Contract: if the result's .Error() method exhausts gas, the query
		// must still return the successful Results JSON. The @error field
		// is best-effort — on OOG we drop it rather than destroying results.
		pkgPath := "gno.land/r/test/oogerror"
		pkgBody := `package oogerror

type BigErr struct{}

func (e *BigErr) Error() string {
	// Consume a lot of gas to try to exhaust the per-query meter.
	s := ""
	for i := 0; i < 1000000; i++ {
		s += "x"
	}
	return s
}

func GetError() error { return &BigErr{} }`

		files := []*std.MemFile{
			{Name: "a.gno", Body: pkgBody},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		}
		msg := NewMsgAddPackage(addr, pkgPath, files)
		require.NoError(t, env.vmk.AddPackage(ctx, msg))
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetError()")
		require.NoError(t, err,
			"OOG in .Error() must not fail the query; results should be preserved")
		assert.Contains(t, res, `"results":`,
			"results payload must be present")
		assert.NotContains(t, res, `"@error"`,
			"OOG in .Error() must graceful-degrade — @error omitted")
	})

	t.Run("typed_nil_error_graceful_degrade", func(t *testing.T) {
		// A non-nil error interface wrapping a typed-nil concrete pointer:
		//   var e *MyErr = nil; return e
		// tv.ImplError() is true (static type satisfies error), so
		// tryGetError invokes .Error() — which nil-derefs the receiver.
		// The defer-recover in tryGetError must catch the panic and
		// gracefully degrade: no @error field, no process crash.
		pkgPath := "gno.land/r/test/typednil"
		pkgBody := `package typednil

type MyErr struct{ Msg string }

func (e *MyErr) Error() string { return e.Msg } // panics on nil receiver

func GetError() error {
	var e *MyErr = nil
	return e
}`

		files := []*std.MemFile{
			{Name: "a.gno", Body: pkgBody},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		}
		msg := NewMsgAddPackage(addr, pkgPath, files)
		require.NoError(t, env.vmk.AddPackage(ctx, msg))
		env.vmk.CommitGnoTransactionStore(ctx)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetError()")
		require.NoError(t, err, "typed-nil error must not crash the handler")
		assert.Contains(t, res, `"results":`,
			"results must still be present even when .Error() panics")
		assert.NotContains(t, res, `"@error"`,
			"typed-nil .Error() panic should graceful-degrade, not surface as @error; got: %s", res)
	})
}

// TestDoRecoverQueryNoMachine exercises the panic-recovery helper added for
// QueryObjectJSON / QueryObjectBinary paths, covering every branch of the
// recover logic: non-error panic, error panic (non-OOG), OOG panic, and the
// two no-panic cases (preserves existing err, preserves nil err).
//
// Note on wrap message visibility: tm2/pkg/errors.Wrapf stores the format
// string as a trace entry on the returned *cmnError. The trace appears in
// fmt.Sprintf("%+v", err) but NOT in err.Error(), which surfaces only the
// inner cause. Tests accordingly check the cause in Error() and the prefix
// in %+v.
func TestDoRecoverQueryNoMachine(t *testing.T) {
	t.Run("string_panic_wraps_as_vm_panic", func(t *testing.T) {
		var err error
		func() {
			defer doRecoverQueryNoMachine(&err)
			panic("synthetic failure")
		}()
		require.Error(t, err)
		// Error() surfaces the cause (fmt.Errorf of the raw panic value).
		assert.Contains(t, err.Error(), "synthetic failure")
		// The full format includes the "VM panic:" trace + stacktrace.
		full := fmt.Sprintf("%+v", err)
		assert.Contains(t, full, "VM panic:",
			"wrap trace should appear in verbose format; got: %s", full)
		assert.Contains(t, full, "Stacktrace:",
			"wrap trace should reference Stacktrace label; got: %s", full)
	})

	t.Run("error_panic_wraps_as_vm_panic", func(t *testing.T) {
		var err error
		func() {
			defer doRecoverQueryNoMachine(&err)
			panic(errors.New("boom"))
		}()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "boom")
		full := fmt.Sprintf("%+v", err)
		assert.Contains(t, full, "VM panic:", "got: %s", full)
	})

	t.Run("oog_surfaces_bare_not_wrapped", func(t *testing.T) {
		var err error
		func() {
			defer doRecoverQueryNoMachine(&err)
			panic(types.OutOfGasError{Descriptor: "test"})
		}()
		require.Error(t, err)
		// OOG is assigned directly (*e = oog), not routed through Wrapf.
		// The concrete type must still be OutOfGasError.
		var oog types.OutOfGasError
		require.True(t, errors.As(err, &oog),
			"OOG must surface as bare OutOfGasError, not wrapped; got: %v", err)
		// And the verbose format must NOT carry the VM panic trace.
		assert.NotContains(t, fmt.Sprintf("%+v", err), "VM panic:",
			"OOG must not be wrapped with the VM panic trace")
	})

	t.Run("no_panic_preserves_existing_err", func(t *testing.T) {
		err := errors.New("original")
		func() {
			defer doRecoverQueryNoMachine(&err)
		}()
		require.Error(t, err)
		assert.Equal(t, "original", err.Error())
	})

	t.Run("no_panic_preserves_nil_err", func(t *testing.T) {
		var err error
		func() {
			defer doRecoverQueryNoMachine(&err)
		}()
		assert.NoError(t, err)
	})
}

// TestVMKeeperNestedObjectTraversal tests that nested persisted objects
// can be traversed by querying ObjectIDs returned in RefValue fields.
// This verifies the object graph can be explored via qeval + qobject queries.
//
// The object graph structure for nested pointers is:
//
//	HeapItemValue (for *L1) -> StructValue (L1) -> HeapItemValue (for *L2) -> StructValue (L2) -> ...
//
// Each pointer field is wrapped in a HeapItemValue, and the actual struct is a separate object.
// So traversal requires following the RefValue chain through each wrapper.
func TestVMKeeperNestedObjectTraversal(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(20000000)))

	// Create realm with 3-level nested structure: L1 -> L2 -> L3
	pkgPath := "gno.land/r/test/nested"
	pkgBody := `package nested

type L3 struct {
	Value string
}

type L2 struct {
	Name string
	L3   *L3
}

type L1 struct {
	ID int
	L2 *L2
}

var root *L1

func init() {
	root = &L1{
		ID: 1,
		L2: &L2{
			Name: "level2",
			L3: &L3{
				Value: "deepest",
			},
		},
	}
}

func GetRoot() *L1 {
	return root
}`

	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "nested.gno", Body: pkgBody},
	}

	msg := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg)
	require.NoError(t, err)
	env.vmk.CommitGnoTransactionStore(ctx)

	// Step 1: Query via qeval to get root pointer's ObjectID
	// qeval returns Amino JSON format consistent with qobject
	res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetRoot()")
	require.NoError(t, err)
	t.Logf("Step 1 - qeval GetRoot():\n%s\n", res)

	// Extract the ObjectID from the PointerValue's RefValue Base
	oid := extractNestedRefValueObjectID(t, res)
	require.NotEmpty(t, oid, "Should have ObjectID for *L1 HeapItemValue")

	// Step 2: Query the HeapItemValue for *L1 (pure Amino: no auto-unwrapping)
	res, err = env.vmk.QueryObjectJSON(env.ctx, oid)
	require.NoError(t, err)
	t.Logf("Step 2 - HeapItemValue for *L1:\n%s\n", res)
	assert.Contains(t, res, `/gno.HeapItemValue`)

	// HeapItemValue contains a TypedValue with RefValue pointing to the StructValue
	oid = extractNestedRefValueObjectID(t, res)
	require.NotEmpty(t, oid, "Should have ObjectID for L1 StructValue")

	// Step 3: Query L1 StructValue
	res, err = env.vmk.QueryObjectJSON(env.ctx, oid)
	require.NoError(t, err)
	t.Logf("Step 3 - L1 StructValue:\n%s\n", res)
	assert.Contains(t, res, `/gno.StructValue`)
	assert.Contains(t, res, `"Fields"`)

	// Extract ObjectID for *L2 field's HeapItemValue
	oid = extractNestedRefValueObjectID(t, res)
	require.NotEmpty(t, oid, "Should have ObjectID for *L2 HeapItemValue")

	// Step 4: Query *L2 HeapItemValue
	res, err = env.vmk.QueryObjectJSON(env.ctx, oid)
	require.NoError(t, err)
	t.Logf("Step 4 - HeapItemValue for *L2:\n%s\n", res)

	// Follow to L2 StructValue
	oid = extractNestedRefValueObjectID(t, res)
	require.NotEmpty(t, oid, "Should have ObjectID for L2 StructValue")

	// Step 5: Query L2 StructValue
	res, err = env.vmk.QueryObjectJSON(env.ctx, oid)
	require.NoError(t, err)
	t.Logf("Step 5 - L2 StructValue:\n%s\n", res)
	assert.Contains(t, res, `/gno.StructValue`)

	// Follow to *L3 HeapItemValue
	oid = extractNestedRefValueObjectID(t, res)
	require.NotEmpty(t, oid, "Should have ObjectID for *L3 HeapItemValue")

	// Step 6: Query *L3 HeapItemValue
	res, err = env.vmk.QueryObjectJSON(env.ctx, oid)
	require.NoError(t, err)
	t.Logf("Step 6 - HeapItemValue for *L3:\n%s\n", res)

	// Follow to L3 StructValue
	oid = extractNestedRefValueObjectID(t, res)
	require.NotEmpty(t, oid, "Should have ObjectID for L3 StructValue")

	// Step 7: Query L3 StructValue (final)
	res, err = env.vmk.QueryObjectJSON(env.ctx, oid)
	require.NoError(t, err)
	t.Logf("Step 7 - L3 StructValue (final):\n%s\n", res)
	assert.Contains(t, res, `/gno.StructValue`)

	t.Log("Successfully traversed nested object graph from L1 -> L2 -> L3!")
	t.Log("Pure Amino format: 7 steps (HeapItemValue -> StructValue alternating)")
}

// deployJSONTestPkg deploys a single-file package for contract/edge-case tests.
func deployJSONTestPkg(t *testing.T, env testEnv, ctx sdk.Context, addr crypto.Address, pkgPath, body string) {
	t.Helper()
	files := []*std.MemFile{
		{Name: "a.gno", Body: body},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}
	msg := NewMsgAddPackage(addr, pkgPath, files)
	require.NoError(t, env.vmk.AddPackage(ctx, msg))
	env.vmk.CommitGnoTransactionStore(ctx)
}

// TestVMKeeperJSONContract anchors the ADR-002 design decisions in regression
// tests. Each subtest asserts a specific contract the ADR documents, so future
// refactors cannot silently change the wire shape.
func TestVMKeeperJSONContract(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)

	t.Run("map_nonstring_keys_tuple_shape", func(t *testing.T) {
		// Contract: maps serialize as an ordered MapList of {Key, Value}
		// tuples, never as a JSON object (ADR §"Map Encoding"). Use an
		// ephemeral (query-time-constructed) map so the full structure
		// appears inline instead of as a RefValue.
		pkgPath := "gno.land/r/contract/mapint"
		deployJSONTestPkg(t, env, ctx, addr, pkgPath, `package mapint
func GetMap() map[int]string {
	return map[int]string{1: "one", 2: "two"}
}`)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetMap()")
		require.NoError(t, err)

		// MapValue wrapper present with a nested List of {Key, Value} pairs.
		assert.Contains(t, res, `"@type":"/gno.MapType"`)
		assert.Contains(t, res, `"@type":"/gno.MapValue"`)
		// {Key, Value} tuple shape — each entry has both keys.
		assert.Contains(t, res, `"Key":{`)
		assert.Contains(t, res, `"Value":{`)
		// The keys are int (PrimitiveType/32), encoded in N — not as JSON
		// object keys. One amino/base64 int-key form:
		assert.Contains(t, res, `"Key":{"T":{"@type":"/gno.PrimitiveType","value":"32"}`,
			"int keys must appear as typed Key fields, not as JSON object keys")
		// Values "one" and "two" are present as StringValues.
		assert.Contains(t, res, `"one"`)
		assert.Contains(t, res, `"two"`)
		// Must NOT be shaped as a JSON object with stringified numeric keys.
		assert.NotContains(t, res, `"1":"one"`)
		assert.NotContains(t, res, `"2":"two"`)
	})

	t.Run("unexported_fields_are_emitted", func(t *testing.T) {
		// Contract: all struct fields, including unexported, appear in the
		// output (ADR §"Visibility of Unexported Fields"). Chain state is
		// already public; hiding unexported fields would be misleading.
		pkgPath := "gno.land/r/contract/unexported"
		deployJSONTestPkg(t, env, ctx, addr, pkgPath, `package unexported
type Person struct {
	Name   string
	secret string
}
func New() Person { return Person{Name: "alice", secret: "shh"} }`)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "New()")
		require.NoError(t, err)

		// Both the exported and unexported field values must appear.
		assert.Contains(t, res, `"alice"`)
		assert.Contains(t, res, `"shh"`)
	})

	t.Run("qobject_single_hop_resolution", func(t *testing.T) {
		// Contract: qobject_json returns the target object inline but any
		// nested persisted Object stays as a RefValue — it is never
		// recursively expanded (ADR §"Single-Hop Object Resolution"). This
		// keeps per-query cost proportional to a single persisted blob.
		//
		// gno persists a `*Outer` as a chain of HeapItemValues, each its
		// own persisted object. Qeval returns the outermost HeapItemValue's
		// ObjectID. Fetching it must yield the HeapItemValue wrapper with a
		// RefValue pointing onward — the wrapped StructValue must NOT be
		// inlined in the same response.
		pkgPath := "gno.land/r/contract/singlehop"
		deployJSONTestPkg(t, env, ctx, addr, pkgPath, `package singlehop
type Inner struct { V int }
type Outer struct { Next *Inner }

var inner *Inner
var outer *Outer

func init() {
	inner = &Inner{V: 42}
	outer = &Outer{Next: inner}
}
func GetOuter() *Outer { return outer }`)

		evalRes, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetOuter()")
		require.NoError(t, err)
		hivOID := extractNestedRefValueObjectID(t, evalRes)
		require.NotEmpty(t, hivOID)

		objRes, err := env.vmk.QueryObjectJSON(env.ctx, hivOID)
		require.NoError(t, err)

		// The HeapItemValue's inner TypedValue.V must be a RefValue —
		// proving the nested Outer struct was NOT recursively inlined.
		assert.Contains(t, objRes, `"@type":"/gno.HeapItemValue"`)
		assert.Contains(t, objRes, `"@type":"/gno.RefValue"`,
			"single-hop: inner value must be a RefValue, not the expanded Outer struct")
		assert.NotContains(t, objRes, `"@type":"/gno.StructValue"`,
			"single-hop: Outer's StructValue body must NOT appear in this response. Got: %s", objRes)
		// And definitely not Inner's "V":42 content (which is two hops away).
		assert.NotContains(t, objRes, `"Fields":[{"T":{"@type":"/gno.PrimitiveType"`,
			"single-hop: Inner's primitive field data must NOT appear")
	})

	t.Run("qobject_binary_amino_roundtrip", func(t *testing.T) {
		// Contract: qobject_binary returns Amino-encoded bytes that a client
		// sharing the node's type registry can decode back into a gno Value
		// (ADR §"Amino Type Registry for qobject_binary"). The roundtrip
		// proves: (a) type registration is complete for the kinds produced
		// by ExportObject, and (b) single-hop semantics hold in binary too.
		pkgPath := "gno.land/r/contract/binary"
		deployJSONTestPkg(t, env, ctx, addr, pkgPath, `package binary
type Item struct {
	Name  string
	Count int
}
var item = &Item{Name: "widget", Count: 7}
func GetItem() *Item { return item }`)

		evalRes, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "GetItem()")
		require.NoError(t, err)
		itemOID := extractNestedRefValueObjectID(t, evalRes)
		require.NotEmpty(t, itemOID)

		binRes, err := env.vmk.QueryObjectBinary(env.ctx, itemOID)
		require.NoError(t, err)
		require.NotEmpty(t, binRes)

		// Decode through the shared amino registry.
		var decoded gnolang.Value
		require.NoError(t, amino.UnmarshalAny(binRes, &decoded),
			"binary output must decode through the shared amino registry")

		// The wrapping HeapItemValue must carry the queried ObjectID and
		// its inner TypedValue must hold a RefValue (single-hop in binary).
		hiv, ok := decoded.(*gnolang.HeapItemValue)
		require.True(t, ok, "expected *HeapItemValue, got %T", decoded)
		assert.Equal(t, itemOID, hiv.GetObjectID().String(),
			"decoded HeapItemValue must carry the queried ObjectID")
		_, isRef := hiv.Value.V.(gnolang.RefValue)
		assert.True(t, isRef,
			"single-hop: decoded inner value must be RefValue, got %T", hiv.Value.V)
	})
}

// TestVMKeeperJSONEdgeCases exercises corners the reviewer flagged as
// undertested: OOG, numeric extremes, nil-vs-empty collections.
func TestVMKeeperJSONEdgeCases(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)

	t.Run("qeval_json_oog_infinite_loop", func(t *testing.T) {
		// An infinite-loop expression must exhaust maxGasQuery and surface
		// as an OutOfGasError, not panic through the ABCI handler.
		pkgPath := "gno.land/r/edge/oogloop"
		deployJSONTestPkg(t, env, ctx, addr, pkgPath, `package oogloop
func Loop() int { for {} }`)

		_, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "Loop()")
		require.Error(t, err, "infinite loop must return an error, not hang or crash")
		assert.Contains(t, err.Error(), "out of gas",
			"expected out-of-gas error, got: %v", err)
	})

	t.Run("qobject_json_malformed_oid", func(t *testing.T) {
		// Malformed ObjectID must surface as a structured error, not panic.
		_, err := env.vmk.QueryObjectJSON(env.ctx, "not-a-valid-oid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid")
	})

	t.Run("qobject_json_unknown_oid", func(t *testing.T) {
		// Well-formed but nonexistent ObjectID → structured not-found error.
		_, err := env.vmk.QueryObjectJSON(env.ctx,
			"0000000000000000000000000000000000000000:999")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("int64_extremes", func(t *testing.T) {
		// int64 min/max survive the spec's base64-in-N encoding.
		pkgPath := "gno.land/r/edge/int64ext"
		deployJSONTestPkg(t, env, ctx, addr, pkgPath, `package int64ext
func Max() int64 { return 9223372036854775807 }
func Min() int64 { return -9223372036854775808 }`)

		resMax, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "Max()")
		require.NoError(t, err)
		// int64 max = 0x7FFFFFFFFFFFFFFF, little-endian = FF FF FF FF FF FF FF 7F
		// base64 of that = /////////38=
		assert.Contains(t, resMax, `"N":"/////////38="`,
			"int64 max must encode as little-endian base64; got: %s", resMax)

		resMin, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "Min()")
		require.NoError(t, err)
		// int64 min = 0x8000000000000000, little-endian = 00 00 00 00 00 00 00 80
		// base64 of that = AAAAAAAAAIA=
		assert.Contains(t, resMin, `"N":"AAAAAAAAAIA="`,
			"int64 min must encode as little-endian base64; got: %s", resMin)
	})

	t.Run("nil_vs_empty_slice", func(t *testing.T) {
		// nil slice and empty slice must be distinguishable in output.
		pkgPath := "gno.land/r/edge/nilslice"
		deployJSONTestPkg(t, env, ctx, addr, pkgPath, `package nilslice
func Nil() []int { return nil }
func Empty() []int { return []int{} }`)

		resNil, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "Nil()")
		require.NoError(t, err)
		resEmpty, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "Empty()")
		require.NoError(t, err)

		// A nil slice has no V field (amino omitempty); empty has a
		// SliceValue with Length "0".
		assert.NotContains(t, resNil, `"@type":"/gno.SliceValue"`,
			"nil slice should not carry a SliceValue; got: %s", resNil)
		assert.Contains(t, resEmpty, `"@type":"/gno.SliceValue"`,
			"empty slice must carry a SliceValue; got: %s", resEmpty)
		assert.Contains(t, resEmpty, `"Length":"0"`)
	})

	t.Run("empty_map", func(t *testing.T) {
		// Empty map round-trips as a MapValue with an empty list.
		pkgPath := "gno.land/r/edge/emptymap"
		deployJSONTestPkg(t, env, ctx, addr, pkgPath, `package emptymap
func M() map[string]int { return map[string]int{} }`)

		res, err := env.vmk.QueryEvalJSON(env.ctx, pkgPath, "M()")
		require.NoError(t, err)
		assert.Contains(t, res, `"@type":"/gno.MapValue"`)
		// No MapListItem entries.
		assert.NotContains(t, res, `"@type":"/gno.MapListItem"`,
			"empty map should have no MapListItem entries; got: %s", res)
	})
}

// extractNestedRefValueObjectID extracts the first nested ObjectID from a JSON response.
// Works with both qeval format ({"results":[...]}) and qobject format ({"objectid":"...","value":{...}}).
// This finds the first RefValue.ObjectID that appears in the content (Amino format).
func extractNestedRefValueObjectID(t *testing.T, jsonStr string) string {
	t.Helper()

	// For qobject responses, skip past the wrapper "objectid" and "value" fields
	// For qeval responses, we just search in "results"
	searchFrom := jsonStr

	// If it's a qobject response with "value" wrapper, skip to that section
	valuePattern := `"value":`
	if valueIdx := strings.Index(jsonStr, valuePattern); valueIdx != -1 {
		searchFrom = jsonStr[valueIdx:]
	}

	// Look for RefValue's ObjectID (Amino format uses uppercase ObjectID)
	pattern := `"ObjectID":"`
	idx := strings.Index(searchFrom, pattern)
	if idx == -1 {
		return ""
	}

	start := idx + len(pattern)
	end := strings.Index(searchFrom[start:], `"`)
	if end == -1 {
		return ""
	}

	return searchFrom[start : start+end]
}
