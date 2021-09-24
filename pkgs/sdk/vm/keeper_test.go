package vm

import (
	"testing"

	"github.com/jaekwon/testify/assert"

	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/std"
)

func TestVMKeeper(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10gnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10gnot")))

	// Create test package.
	files := []NamedFile{
		{"init.go", `
package test

func init() {
}

func Echo(msg string) string {
	return "echo:"+msg
}`},
	}
	pkgPath := "gno.land/r/test"
	err := env.vmk.AddPackage(ctx, addr, pkgPath, files)

	// Run Echo function.
	res, err := env.vmk.Eval(ctx, addr, pkgPath, `Echo("hello world")`)
	assert.NoError(t, err)
	assert.Equal(t, res, `("echo:hello world" string)`)
	t.Log("result:", res)
}
