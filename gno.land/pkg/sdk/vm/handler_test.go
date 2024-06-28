package vm

import (
	"testing"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
)

func Test_parseQueryEvalData(t *testing.T) {
	t.Parallel()
	tt := []struct {
		input   string
		pkgpath string
		expr    string
	}{
		{
			"gno.land/r/realm.Expression()",
			"gno.land/r/realm",
			"Expression()",
		},
		{
			"a.b/c/d.e",
			"a.b/c/d",
			"e",
		},
		{
			"a.b.c.d.e/c/d.e",
			"a.b.c.d.e/c/d",
			"e",
		},
		{
			"abcde/c/d.e",
			"abcde/c/d",
			"e",
		},
	}
	for _, tc := range tt {
		path, expr := parseQueryEvalData(tc.input)
		assert.Equal(t, tc.pkgpath, path)
		assert.Equal(t, tc.expr, expr)
	}
}

func Test_parseQueryEval_panic(t *testing.T) {
	t.Parallel()

	assert.PanicsWithValue(t, panicInvalidQueryEvalData, func() {
		parseQueryEvalData("gno.land/r/demo/users")
	})
}

// Call Run with stdlibs.
func TestQuery_Eval(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx
	vmHandler := env.vmh

	// Give "addr1" some gnots.
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bank.SetCoins(ctx, addr, std.MustParseCoins("10000000ugnot"))
	assert.True(t, env.bank.GetCoins(ctx, addr).IsEqual(std.MustParseCoins("10000000ugnot")))

	// Create test package.
	files := []*std.MemFile{
		{"echo.gno", `
package echo

func Echo(msg string) string {
	return "echo:"+msg
}`},
	}
	pkgPath := "gno.land/r/echo"
	msg1 := NewMsgAddPackage(addr, pkgPath, files)
	err := env.vmk.AddPackage(ctx, msg1)
	assert.NoError(t, err)

	req := abci.RequestQuery{
		Path: "vm/qeval",
		Data: []byte(`gno.land/r/echo.Echo("hello")`),
	}
	res := vmHandler.Query(env.ctx, req)
	assert.True(t, res.IsOK())
	assert.Equal(t, string(res.Data), `("echo:hello" string)`)
}
