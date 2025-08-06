package vm

// TODO: move most of the logic in ROOT/gno.land/...

import (
	"errors"
	"fmt"
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

import "std"

func init() {
}

func Echo(cur realm, msg string) string {
	addr := std.OriginCaller()
	pkgAddr := std.CurrentRealm().Address()
	send := std.OriginSend()
	banker := std.NewBanker(std.BankerTypeOriginSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}`},
	}
	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	//Reconcile the account balance
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

import "std"

var admin std.Address

func init() {
     admin = std.OriginCaller()
}

func Echo(cur realm, msg string) string {
	addr := std.OriginCaller()
	pkgAddr := std.CurrentRealm().Address()
	send := std.OriginSend()
	banker := std.NewBanker(std.BankerTypeOriginSend)
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

import "std"

func init() {
}

func Echo(cur realm, msg string) string {
	addr := std.OriginCaller()
	pkgAddr := std.CurrentRealm().Address()
	send := std.Coins{{"ugnot", 10000000}}
	banker := std.NewBanker(std.BankerTypeOriginSend)
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

import "std"

func init() {
}

func Echo(cur realm, msg string) string {
	addr := std.OriginCaller()
	pkgAddr := std.CurrentRealm().Address()
	send := std.Coins{{"ugnot", 1000000}}
	banker := std.NewBanker(std.BankerTypeRealmSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
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
	//Reconcile the account balance
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

import "std"

func init() {
}

func Echo(cur realm, msg string) string {
	addr := std.OriginCaller()
	pkgAddr := std.CurrentRealm().Address()
	send := std.Coins{{"ugnot", 10000000}}
	banker := std.NewBanker(std.BankerTypeRealmSend)
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

import "std"

func init() {
	std.SetParamString("foo.string", "foo1")
}

func Do(cur realm) string {
	std.SetParamInt64("bar.int64", int64(1337))
	std.SetParamString("foo.string", "foo2") // override init

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

import "std"

var admin std.Address

func init() {
     admin = std.OriginCaller()
}

func Echo(cur realm, msg string) string {
	addr := std.OriginCaller()
	pkgAddr := std.CurrentRealm().Address()
	send := std.OriginSend()
	banker := std.NewBanker(std.BankerTypeOriginSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
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

import "std"

func main() {
	addr := std.OriginCaller()
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
	//varify the account balance
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
