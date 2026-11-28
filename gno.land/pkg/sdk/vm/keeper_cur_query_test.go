package vm

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// A query must thread `cur` down a chain of crossing calls the same way a
// transaction does.
//
// doOpCall rebuilds an inherited cur when it looks like the placeholder the
// preprocessor bakes in for a top-level entry: prev is origin-shaped AND its
// address is empty. withQueryEvalMachine deliberately leaves OriginCaller unset
// (a query has no signer), so buildOriginRealm mints a replacement whose
// address is ALSO empty — which matches the same test at the next hop, and the
// next. Under a query the rebuild therefore cascaded the whole way down: every
// callee got a fresh identity of its own pkgPath instead of the caller's, and
// doOpCall's inherit check was skipped at every level, since a rebuilt cur is
// exempt from it.
//
// It failed open, which is the part that matters: the minted identity's
// Previous() reads as a direct user call — a wider answer than the caller's own,
// not a narrower one.
func TestCurIsThreadedIdenticallyUnderQueryAndTx(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(20000000)))

	pkgPath := "gno.land/r/test/curquery"
	const body = `package curquery

func target(cur realm) string { return cur.PkgPath() }

func exec(cur realm, cb func() string) string { return cb() }

// Stale capture: cb closes over Probe's cur but runs inside exec's frame. A
// transaction refuses this. A query must refuse it identically.
func Probe(cur realm) string {
	cb := func() string { return target(cur) }
	return exec(cross(cur), cb)
}
`
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "x.gno", Body: body},
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(addr, pkgPath, files)))
	env.vmk.CommitGnoTransactionStore(ctx)

	_, err := env.vmk.QueryEval(env.ctx, pkgPath, `Probe()`)
	require.Error(t, err,
		"a query must apply the same cur check as a transaction; the rebuild "+
			"exemption cascading through QueryEval skipped it at every hop")
	require.Contains(t, err.Error(), "not the caller's own cur")
}
