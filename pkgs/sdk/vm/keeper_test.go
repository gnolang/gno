package vm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jaekwon/testify/assert"

	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/std"
)

// Sending total send amount succeeds.
func TestVMKeeperOrigSend1(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10gnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10gnot")))

	// Create test package.
	files := []*std.MemFile{
		{"init.gno", `
package test

import "std"

func init() {
}

func Echo(msg string) string {
	addr := std.GetOrigCaller()
	pkgAddr := std.GetOrigPkgAddr()
	send := std.GetOrigSend()
	banker := std.GetBanker(std.BankerTypeOrigSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}`},
	}
	pkgPath := "gno.land/r/test"
	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	coins := std.MustParseCoins("10gnot")
	msg2 := NewMsgCall(addr, coins, pkgPath, "Echo", []string{"hello world"})
	res, err := env.vmk.Call(ctx, msg2)
	assert.NoError(t, err)
	assert.Equal(t, res, `("echo:hello world" string)`)
	// t.Log("result:", res)
}

// Sending too much fails
func TestVMKeeperOrigSend2(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10gnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10gnot")))

	// Create test package.
	files := []*std.MemFile{
		{"init.gno", `
package test

import "std"

func init() {
}

func Echo(msg string) string {
	addr := std.GetOrigCaller()
	pkgAddr := std.GetOrigPkgAddr()
	send := std.GetOrigSend()
	banker := std.GetBanker(std.BankerTypeOrigSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}`},
	}
	pkgPath := "gno.land/r/test"
	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	coins := std.MustParseCoins("11gnot")
	msg2 := NewMsgCall(addr, coins, pkgPath, "Echo", []string{"hello world"})
	res, err := env.vmk.Call(ctx, msg2)
	assert.Error(t, err)
	assert.Equal(t, res, "")
	fmt.Println(err.Error())
	assert.True(t, strings.Contains(err.Error(), "insufficient coins error"))
}

// Sending more than tx send fails.
func TestVMKeeperOrigSend3(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10gnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10gnot")))

	// Create test package.
	files := []*std.MemFile{
		{"init.gno", `
package test

import "std"

func init() {
}

func Echo(msg string) string {
	addr := std.GetOrigCaller()
	pkgAddr := std.GetOrigPkgAddr()
	send := std.Coins{{"gnot", 10}}
	banker := std.GetBanker(std.BankerTypeOrigSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}`},
	}
	pkgPath := "gno.land/r/test"
	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	coins := std.MustParseCoins("9gnot")
	msg2 := NewMsgCall(addr, coins, pkgPath, "Echo", []string{"hello world"})
	// XXX change this into an error and make sure error message is descriptive.
	_, err = env.vmk.Call(ctx, msg2)
	assert.Error(t, err)
}

// Sending realm package coins succeeds.
func TestVMKeeperRealmSend1(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10gnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10gnot")))

	// Create test package.
	files := []*std.MemFile{
		{"init.gno", `
package test

import "std"

func init() {
}

func Echo(msg string) string {
	addr := std.GetOrigCaller()
	pkgAddr := std.GetOrigPkgAddr()
	send := std.Coins{{"gnot", 10}}
	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}`},
	}
	pkgPath := "gno.land/r/test"
	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	coins := std.MustParseCoins("10gnot")
	msg2 := NewMsgCall(addr, coins, pkgPath, "Echo", []string{"hello world"})
	res, err := env.vmk.Call(ctx, msg2)
	assert.NoError(t, err)
	assert.Equal(t, res, `("echo:hello world" string)`)
}

// Sending too much realm package coins fails.
func TestVMKeeperRealmSend2(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10gnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10gnot")))

	// Create test package.
	files := []*std.MemFile{
		{"init.gno", `
package test

import "std"

func init() {
}

func Echo(msg string) string {
	addr := std.GetOrigCaller()
	pkgAddr := std.GetOrigPkgAddr()
	send := std.Coins{{"gnot", 10}}
	banker := std.GetBanker(std.BankerTypeRealmSend)
	banker.SendCoins(pkgAddr, addr, send) // send back
	return "echo:"+msg
}`},
	}
	pkgPath := "gno.land/r/test"
	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	// Run Echo function.
	coins := std.MustParseCoins("9gnot")
	msg2 := NewMsgCall(addr, coins, pkgPath, "Echo", []string{"hello world"})
	// XXX change this into an error and make sure error message is descriptive.
	_, err = env.vmk.Call(ctx, msg2)
	assert.Error(t, err)
}
