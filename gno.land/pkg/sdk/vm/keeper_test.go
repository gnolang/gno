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

func TestVMKeeperEvalJSONFormatting2(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(9000000)))

	tests := []struct {
		name     string
		pkgBody  string
		expr     string
		expected string
	}{
		{
			name:     "JSON string",
			pkgBody:  `func GetString() string { return "hello" }`,
			expr:     "GetString()",
			expected: `{"results":[{"T":"string","V":"hello"}]}`,
		},
		{
			name:     "JSON integer",
			pkgBody:  `func GetInt() int { return 42 }`,
			expr:     "GetInt()",
			expected: `{"results":[{"T":"int","V":42}]}`,
		},
		{
			name:     "JSON boolean",
			pkgBody:  `func GetBool() bool { return true }`,
			expr:     "GetBool()",
			expected: `{"results":[{"T":"bool","V":true}]}`,
		},
		{
			name:     "JSON bytes",
			pkgBody:  `func GetBytes() []byte { return []byte("test") }`,
			expr:     "GetBytes()",
			expected: `{"results":[{"T":"[]uint8","V":[116,101,115,116]}]}`,
		},
		{
			name:     "JSON multiple values",
			pkgBody:  `func GetMulti() (string, int) { return "hello", 42 }`,
			expr:     "GetMulti()",
			expected: `{"results":[{"T":"string","V":"hello"},{"T":"int","V":42}]}`,
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

			cfgEval := QueryMsgEval{Expr: tc.expr, PkgPath: pkgPath, Format: QueryFormatJSON}
			res, err := env.vmk.QueryEval(env.ctx, cfgEval)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, res)
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

		cfgEval := QueryMsgEval{Expr: "GetItem()", PkgPath: pkgPath, Format: QueryFormatJSON}
		res, err := env.vmk.QueryEval(env.ctx, cfgEval)
		require.NoError(t, err)

		// Verify structure
		assert.Contains(t, res, `"T":"*gno.land/r/test/persisted1.Item"`)
		assert.Contains(t, res, `"ID":{"T":"int","V":42}`)
		assert.Contains(t, res, `"Name":{"T":"string","V":"test item"}`)
		// Persisted objects should have objectid
		assert.Contains(t, res, `"objectid"`)
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

		cfgEval := QueryMsgEval{Expr: "GetData()", PkgPath: pkgPath, Format: QueryFormatJSON}
		res, err := env.vmk.QueryEval(env.ctx, cfgEval)
		require.NoError(t, err)

		// Verify map format
		assert.Contains(t, res, `"T":"map[string]int"`)
		assert.Contains(t, res, `"one"`)
		assert.Contains(t, res, `"two"`)
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

		cfgEval := QueryMsgEval{Expr: "GetAmount()", PkgPath: pkgPath, Format: QueryFormatJSON}
		res, err := env.vmk.QueryEval(env.ctx, cfgEval)
		require.NoError(t, err)

		// Verify declared type with base
		assert.Contains(t, res, `"T":"gno.land/r/test/persisted7.Amount"`)
		assert.Contains(t, res, `"V":1000`)
		assert.Contains(t, res, `"base":"int64"`)
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

		cfgEval := QueryMsgEval{Expr: "GetError()", PkgPath: pkgPath, Format: QueryFormatJSON}
		res, err := env.vmk.QueryEval(env.ctx, cfgEval)
		require.NoError(t, err)

		// Verify error type with error field
		assert.Contains(t, res, `"T":"*gno.land/r/test/persisted8.CustomError"`)
		assert.Contains(t, res, `"Code":{"T":"int","V":404}`)
		assert.Contains(t, res, `"Message":{"T":"string","V":"not found"}`)
		assert.Contains(t, res, `"error":"not found"`)
	})

	t.Run("persisted_json_tags", func(t *testing.T) {
		pkgPath := "gno.land/r/test/persisted9"
		pkgBody := "package persisted9\n\ntype Person struct {\n\tFirstName string `json:\"first_name\"`\n\tLastName  string `json:\"last_name\"`\n\tAge       int    `json:\"age,omitempty\"`\n}\n\nvar person *Person\n\nfunc init() {\n\tperson = &Person{FirstName: \"John\", LastName: \"Doe\", Age: 30}\n}\n\nfunc GetPerson() *Person {\n\treturn person\n}"

		files := []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
			{Name: "person.gno", Body: pkgBody},
		}

		msg := NewMsgAddPackage(addr, pkgPath, files)
		err := env.vmk.AddPackage(ctx, msg)
		require.NoError(t, err)
		env.vmk.CommitGnoTransactionStore(ctx)

		cfgEval := QueryMsgEval{Expr: "GetPerson()", PkgPath: pkgPath, Format: QueryFormatJSON}
		res, err := env.vmk.QueryEval(env.ctx, cfgEval)
		require.NoError(t, err)

		// Verify JSON tags are respected
		assert.Contains(t, res, `"first_name":{"T":"string","V":"John"}`)
		assert.Contains(t, res, `"last_name":{"T":"string","V":"Doe"}`)
		assert.Contains(t, res, `"age":{"T":"int","V":30}`)
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

		cfgEval := QueryMsgEval{Expr: "GetPtr()", PkgPath: pkgPath, Format: QueryFormatJSON}
		res, err := env.vmk.QueryEval(env.ctx, cfgEval)
		require.NoError(t, err)

		// Verify nil pointer output
		assert.Contains(t, res, `"T":"*gno.land/r/test/persisted10.Data"`)
		assert.Contains(t, res, `"V":null`)
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

		cfgEval := QueryMsgEval{Expr: "GetParent()", PkgPath: pkgPath, Format: QueryFormatJSON}
		res, err := env.vmk.QueryEval(env.ctx, cfgEval)
		require.NoError(t, err)

		// Verify struct with nil pointer field doesn't panic
		assert.Contains(t, res, `"T":"*gno.land/r/test/persisted11.Parent"`)
		assert.Contains(t, res, `"Name":{"T":"string","V":"test"}`)
		assert.Contains(t, res, `"Child":{"T":"*gno.land/r/test/persisted11.Child","V":null}`)
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

		cfgEval := QueryMsgEval{Expr: "GetNumbers()", PkgPath: pkgPath, Format: QueryFormatJSON}
		res, err := env.vmk.QueryEval(env.ctx, cfgEval)
		require.NoError(t, err)

		// Verify slice is correctly serialized
		assert.Contains(t, res, `"T":"[]int"`)
		assert.Contains(t, res, `[10,20,30,40,50]`)
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

		cfgEval := QueryMsgEval{Expr: "GetItems()", PkgPath: pkgPath, Format: QueryFormatJSON}
		res, err := env.vmk.QueryEval(env.ctx, cfgEval)
		require.NoError(t, err)

		// Verify slice of structs is correctly serialized
		assert.Contains(t, res, `"T":"[]gno.land/r/test/persisted13.Item"`)
		assert.Contains(t, res, `"ID":{"T":"int","V":1}`)
		assert.Contains(t, res, `"Name":{"T":"string","V":"first"}`)
		assert.Contains(t, res, `"ID":{"T":"int","V":2}`)
		assert.Contains(t, res, `"Name":{"T":"string","V":"second"}`)
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

		cfgEval := QueryMsgEval{Expr: "GetRoot()", PkgPath: pkgPath, Format: QueryFormatJSON}
		res, err := env.vmk.QueryEval(env.ctx, cfgEval)
		require.NoError(t, err)

		// Verify nested structure with nil pointers
		assert.Contains(t, res, `"T":"*gno.land/r/test/persisted14.Level1"`)
		assert.Contains(t, res, `"Data":{"T":"string","V":"deep"}`)
		// The nil pointer field should be null
		assert.Contains(t, res, `"Nil":{"T":"*gno.land/r/test/persisted14.Level3","V":null}`)
	})
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
