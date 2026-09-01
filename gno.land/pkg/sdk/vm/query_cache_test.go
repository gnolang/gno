package vm

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

const countingSrc = `package counting

var N int

func Bump(cur realm) { N++ }

func Render(path string) string {
	s := ""
	for i := 0; i < N; i++ {
		s += "x"
	}
	return s
}
`

// atHeight returns ctx with the chain tip moved, which is what the cache keys
// its whole contents on.
func atHeight(ctx sdk.Context, h int64) sdk.Context {
	return ctx.WithBlockHeader(&bft.Header{ChainID: "test", Height: h})
}

func setupCounting(t *testing.T) (testEnv, vmHandler, crypto.Address, string) {
	t.Helper()
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	addr := crypto.AddressFromPreimage([]byte("qc"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins("100000000000ugnot"))

	pkgPath := "gno.land/r/qc/counting"
	msg := NewMsgAddPackage(addr, pkgPath, []*std.MemFile{
		{Name: "a.gno", Body: countingSrc},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	})
	msg.MaxDeposit = std.MustParseCoins("100000000000ugnot")
	require.NoError(t, env.vmk.AddPackage(ctx, msg))
	env.vmk.CommitGnoTransactionStore(ctx)
	return env, NewHandler(env.vmk), addr, pkgPath
}

func render(t *testing.T, vh vmHandler, ctx sdk.Context, pkgPath string) string {
	t.Helper()
	res := vh.Query(ctx, abci.RequestQuery{Path: "vm/qrender", Data: []byte(pkgPath + ":")})
	require.Nil(t, res.Error, "render failed: %v", res.Error)
	return string(res.Data)
}

// The whole point of the cache is that it must never outlive the state it was
// computed from. A bump changes what Render returns, and the answer has to
// change with it once the chain moves on.
func TestQueryCacheDoesNotOutliveTheState(t *testing.T) {
	env, vh, addr, pkgPath := setupCounting(t)

	before := render(t, vh, atHeight(env.ctx, 10), pkgPath)
	require.Empty(t, before, "no bumps yet")

	// Same tip, same query: served from the cache, same answer.
	require.Equal(t, before, render(t, vh, atHeight(env.ctx, 10), pkgPath))

	// Change the state.
	c := env.vmk.MakeGnoTransactionStore(env.ctx)
	call := NewMsgCall(addr, nil, pkgPath, "Bump", nil)
	call.MaxDeposit = std.MustParseCoins("100000000000ugnot")
	_, err := env.vmk.Call(c, call)
	require.NoError(t, err)
	env.vmk.CommitGnoTransactionStore(c)

	// Still the old tip, so the stale answer is still what the cache holds.
	// This is correct: a query pinned to that tip must see that tip's state.
	require.Equal(t, before, render(t, vh, atHeight(env.ctx, 10), pkgPath))

	// Chain advances: the cache must be dropped and the new state seen.
	after := render(t, vh, atHeight(env.ctx, 11), pkgPath)
	require.Equal(t, "x", after, "a new tip must show the bump")
	require.NotEqual(t, before, after)
}

// req.Height selects which state version is read, so it has to be part of the
// key. Without it a historical query and a current one collide.
func TestQueryCacheKeysOnRequestedHeight(t *testing.T) {
	base := abci.RequestQuery{Path: "vm/qrender", Data: []byte("gno.land/r/qc/counting:")}

	h5, h6 := base, base
	h5.Height, h6.Height = 5, 6
	require.NotEqual(t, queryCacheKey(h5), queryCacheKey(h6),
		"two heights must not share a key")

	// Different endpoint, and different argument, must also differ.
	other := base
	other.Path = "vm/qeval"
	require.NotEqual(t, queryCacheKey(base), queryCacheKey(other))
	otherData := base
	otherData.Data = []byte("gno.land/r/qc/counting:x")
	require.NotEqual(t, queryCacheKey(base), queryCacheKey(otherData))
}

// A path and data pair must not be able to impersonate another by moving the
// separator, which is why the height is fixed-width and leads the key.
func TestQueryCacheKeyIsUnambiguous(t *testing.T) {
	a := abci.RequestQuery{Path: "vm/qrender", Data: []byte("b")}
	b := abci.RequestQuery{Path: "vm/qrende", Data: []byte("rb")}
	require.NotEqual(t, queryCacheKey(a), queryCacheKey(b))

	c := abci.RequestQuery{Path: "vm/q", Data: []byte("\x00x")}
	d := abci.RequestQuery{Path: "vm/q\x00", Data: []byte("x")}
	require.NotEqual(t, queryCacheKey(c), queryCacheKey(d))
}

// An entry over the per-entry cap is not held, so one huge result cannot
// occupy the whole budget.
func TestQueryCacheRefusesOversizedEntries(t *testing.T) {
	qc := newQueryCache()
	req := abci.RequestQuery{Path: "vm/qeval", Data: []byte("big")}
	big := abci.ResponseQuery{}
	big.Data = make([]byte, maxQueryCacheEntry+1)

	qc.put(1, req, big)
	_, ok := qc.get(1, req)
	require.False(t, ok, "an oversized result must not be cached")

	small := abci.ResponseQuery{}
	small.Data = []byte("ok")
	qc.put(1, req, small)
	got, ok := qc.get(1, req)
	require.True(t, ok)
	require.Equal(t, []byte("ok"), got.Data)
}

// A query that panics must leave the cache untouched. Storing in a defer would
// not: a defer runs while panicking too, and the response is still empty at
// that point, so the next identical query would get an empty success in place
// of the panic.
func TestQueryCachePanicIsNotStored(t *testing.T) {
	env, vh, _, pkgPath := setupCounting(t)
	ctx := atHeight(env.ctx, 10)

	// qrender wants <pkgpath>:<path>. Without the colon it panics.
	bad := abci.RequestQuery{Path: "vm/qrender", Data: []byte(pkgPath)}
	mustPanic := func() {
		defer func() {
			require.NotNil(t, recover(), "a malformed render query must panic")
		}()
		vh.Query(ctx, bad)
	}

	mustPanic()
	mustPanic() // the repeat must panic too, not be answered from the cache

	_, ok := vh.vm.queryCache.get(ctx.BlockHeight(), bad)
	require.False(t, ok, "a panicking query must store nothing")
}

// A hit must not hand out the stored array, or a caller could edit what the
// next hit returns.
func TestQueryCacheHitIsACopy(t *testing.T) {
	qc := newQueryCache()
	req := abci.RequestQuery{Path: "vm/qeval", Data: []byte("k")}
	res := abci.ResponseQuery{}
	res.Data = []byte("original")
	qc.put(1, req, res)

	first, ok := qc.get(1, req)
	require.True(t, ok)
	first.Data[0] = 'X'

	second, ok := qc.get(1, req)
	require.True(t, ok)
	require.Equal(t, "original", string(second.Data), "a hit must not see another caller's edit")
}
