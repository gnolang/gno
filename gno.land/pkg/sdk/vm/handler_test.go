package vm

import (
	"testing"

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

func TestProcessNoopMsg(t *testing.T) {
	// setup
	env := setupTestEnv()
	ctx := env.ctx
	vmHandler := NewHandler(env.vmk)

	addr := crypto.AddressFromPreimage([]byte("test1"))
	msg := NewMsgNoop(addr)

	res := vmHandler.Process(ctx, msg)
	assert.Empty(t, res)
}

func TestProcessInvalidMsg(t *testing.T) {
	// setup
	env := setupTestEnv()
	ctx := env.ctx
	vmHandler := NewHandler(env.vmk)

	type InvalidMsg struct {
		std.Msg
	}

	msg := InvalidMsg{}

	res := vmHandler.Process(ctx, msg)
	assert.NotEmpty(t, res)
	assert.Equal(t, res.Error, std.UnknownRequestError{})
}
