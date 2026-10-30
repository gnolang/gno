package vm

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/stretchr/testify/require"
)

// Two realms' folded copies of one imported string const must never share
// a backing: not on a long-running node, where each addpkg tx loaded pp
// separately, and not on a restarted node, where one pp load preprocessed
// both. The GC recount dedups by backing, so sharing would make it differ.
func TestConstBackingPartitionSurvivesRestart(t *testing.T) {
	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)

	add := func(path, body string) { // one tx store per addpkg
		deployJSONTestPkg(t, env, env.vmk.MakeGnoTransactionStore(env.ctx), addr, path, body)
	}
	add("gno.land/r/pp", "package pp\n\nconst A = \""+strings.Repeat("a", 1000)+"\"\n")
	add("gno.land/r/qq", "package qq\n\nimport \"gno.land/r/pp\"\n\nconst B = pp.A\n")
	add("gno.land/r/rr", "package rr\n\nimport \"gno.land/r/pp\"\n\nconst C = pp.A\n")

	backing := func(path string, name gnolang.Name) any {
		pn := env.vmk.gnoStore.GetBlockNode(gnolang.PackageNodeLocation(path)).(*gnolang.PackageNode)
		idx, ok := pn.GetLocalIndex(name)
		require.True(t, ok)
		return pn.StaticBlock.Block.Values[idx].V.(gnolang.StringValue).B
	}
	notShared := func() {
		t.Helper()
		require.NotSame(t, backing("gno.land/r/qq", "B"), backing("gno.land/r/rr", "C"),
			"folded consts must own their backing")
	}
	notShared()
	require.Nil(t, restartBootPreprocess(env, ctx))
	notShared()
}
