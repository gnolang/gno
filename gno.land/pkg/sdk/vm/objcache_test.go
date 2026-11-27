package vm

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// The decoded-object cache (gnovm/pkg/gnolang/store_objcache.go) keeps
// decoded objects across transactions. It is only safe if the state machine
// cannot tell a hit from a miss. These tests run one workload against a
// cold, a warming and a fully warm cache, and require the gas of every
// transaction, every result and every block's app hash to be identical.

const objcachePkg = "gno.land/r/test/objcache"

const objcacheSrc = `package objcache

import (
	"strconv"
	"strings"
)

type node struct {
	key   int
	val   string
	left  *node
	right *node
}

var root *node

func insert(n *node, k int, v string) *node {
	if n == nil {
		return &node{key: k, val: v}
	}
	switch {
	case k < n.key:
		n.left = insert(n.left, k, v)
	case k > n.key:
		n.right = insert(n.right, k, v)
	default:
		n.val = v
	}
	return n
}

func Set(cur realm, k int, v string) { root = insert(root, k, v) }

// Fill inserts a comma-separated list of keys.
func Fill(cur realm, keys string) {
	for _, k := range strings.Split(keys, ",") {
		n, _ := strconv.Atoi(k)
		root = insert(root, n, "v")
	}
}

// SetThenPanic mutates loaded objects, then fails the transaction.
func SetThenPanic(cur realm, k int, v string) {
	root = insert(root, k, v)
	panic("rolled back")
}

func Get(cur realm, k int) string {
	for n := root; n != nil; {
		switch {
		case k < n.key:
			n = n.left
		case k > n.key:
			n = n.right
		default:
			return n.val
		}
	}
	return "<none>"
}
`

type objcacheTx struct {
	gas int64
	res string
}

type objcacheRun struct {
	txs    []objcacheTx
	hashes []string
}

// runObjcacheWorkload deploys the realm and runs a mixed workload of reads,
// writes, failing writes and out-of-gas writes, committing a block every 10
// transactions.
func runObjcacheWorkload(t *testing.T) objcacheRun {
	t.Helper()
	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("objcache"))
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
	env.bankk.SetCoins(ctx, addr, initialBalance)
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gno.GenGnoModLatest(objcachePkg)},
		{Name: "objcache.gno", Body: objcacheSrc},
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(addr, objcachePkg, files)))
	env.vmk.CommitGnoTransactionStore(ctx)

	var run objcacheRun
	deliver := func(gasWanted int64, fn string, args ...string) {
		tx := objcacheTx{}
		dctx := auth.SetGasMeter(env.ctx.WithMode(sdk.RunTxModeDeliver), gasWanted)
		dctx, write := dctx.CacheContext()
		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(store.OutOfGasError); !ok {
						panic(r)
					}
					tx.res = "out of gas"
				}
			}()
			dctx = env.vmk.MakeGnoTransactionStore(dctx)
			res := env.vmh.Process(dctx, NewMsgCall(addr, nil, objcachePkg, fn, args))
			if res.IsOK() {
				env.vmk.CommitGnoTransactionStore(dctx)
				write()
				tx.res = string(res.Data)
			} else {
				tx.res = "error: " + res.Log
			}
		}()
		tx.gas = dctx.GasMeter().GasConsumed()
		run.txs = append(run.txs, tx)
		if len(run.txs)%10 == 0 {
			run.hashes = append(run.hashes, commitMultiStoreHash(t, env))
		}
	}

	const gas = 50_000_000
	model := map[int]string{} // what committed state must hold
	// Build a tree whose upper levels every later transaction loads.
	for _, k := range []int{50, 25, 75, 12, 37, 62, 87, 6, 18, 31, 43, 56, 68, 81, 93} {
		deliver(gas, "Set", strconv.Itoa(k), "v"+strconv.Itoa(k))
		model[k] = "v" + strconv.Itoa(k)
	}
	for i := range 40 {
		ki := (i * 37) % 100
		k := strconv.Itoa(ki)
		switch i % 5 {
		case 0, 1:
			// H2: later loads of the rewritten objects must see the new value.
			deliver(gas, "Set", k, fmt.Sprintf("w%d", i))
			model[ki] = fmt.Sprintf("w%d", i)
		case 2:
			// H1: mutates cached-then-copied objects, then fails.
			deliver(gas, "SetThenPanic", k, "dirty")
		case 3:
			// H1: runs out of gas half way through the writes.
			deliver(450_000, "Set", k, "oog")
		case 4:
			deliver(gas, "Get", k)
		}
	}
	// Every key, after all of the above: exactly the committed writes, no
	// "dirty", no "oog".
	for k := range 100 {
		deliver(gas, "Get", strconv.Itoa(k))
		want, ok := model[k]
		if !ok {
			want = "<none>"
		}
		require.Equalf(t, fmt.Sprintf("(%q string)\n\n", want), run.txs[len(run.txs)-1].res, "Get(%d)", k)
	}
	return run
}

func TestObjectDecodeCacheGasDeterminism(t *testing.T) {
	prev := gno.SetObjectDecodeCacheEnabled(false)
	defer gno.SetObjectDecodeCacheEnabled(prev)
	cold := runObjcacheWorkload(t)

	gno.SetObjectDecodeCacheEnabled(true)
	gno.ClearObjectDecodeCache()
	h0, _ := gno.ObjectDecodeCacheStats()
	warming := runObjcacheWorkload(t) // starts empty, fills as it goes
	h1, _ := gno.ObjectDecodeCacheStats()
	warm := runObjcacheWorkload(t) // every object already decoded once
	h2, _ := gno.ObjectDecodeCacheStats()

	// The workload must exercise what it claims to.
	var failed, oog int
	for _, tx := range cold.txs {
		switch {
		case tx.res == "out of gas":
			oog++
		case strings.HasPrefix(tx.res, "error:"):
			failed++
		}
	}
	require.Equal(t, 8, failed, "SetThenPanic transactions")
	require.Equal(t, 8, oog, "out-of-gas transactions")

	for name, got := range map[string]objcacheRun{"warming": warming, "warm": warm} {
		require.Equal(t, len(cold.txs), len(got.txs))
		for i := range cold.txs {
			require.Equalf(t, cold.txs[i], got.txs[i], "%s run, tx %d: gas or result differs from cold", name, i)
		}
		require.Equalf(t, cold.hashes, got.hashes, "%s run: app hash differs from cold", name)
	}
	require.Greater(t, h1-h0, uint64(0), "warming run never hit the cache")
	require.Greater(t, h2-h1, h1-h0, "warm run should hit more than the warming run")
	t.Logf("cache hits: warming run %d, warm run %d", h1-h0, h2-h1)
}

// BenchmarkObjectDecodeCache runs read-modify-write transactions on a tree of
// size keys with the decoded-object cache off and on. Gas per transaction is
// reported and must be identical between the two.
//
//	go test ./gno.land/pkg/sdk/vm/ -run '^$' -bench ObjectDecodeCache -count 6 | benchstat -col /cache -
func BenchmarkObjectDecodeCache(b *testing.B) {
	for _, size := range []int{1_000, 10_000} {
		for _, enabled := range []bool{false, true} {
			name := fmt.Sprintf("size=%d/cache=%s", size, map[bool]string{false: "off", true: "on"}[enabled])
			b.Run(name, func(b *testing.B) {
				prev := gno.SetObjectDecodeCacheEnabled(enabled)
				defer gno.SetObjectDecodeCacheEnabled(prev)
				gno.ClearObjectDecodeCache()
				benchObjectDecodeCache(b, size)
			})
		}
	}
}

func benchObjectDecodeCache(b *testing.B, size int) {
	b.Helper()
	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("objcache"))
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
	env.bankk.SetCoins(ctx, addr, std.MustParseCoins(ugnot.ValueString(100_000_000_000)))
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gno.GenGnoModLatest(objcachePkg)},
		{Name: "objcache.gno", Body: objcacheSrc},
	}
	require.NoError(b, env.vmk.AddPackage(ctx, NewMsgAddPackage(addr, objcachePkg, files)))
	env.vmk.CommitGnoTransactionStore(ctx)

	call := func(gasWanted int64, fn string, args ...string) int64 {
		dctx := auth.SetGasMeter(env.ctx.WithMode(sdk.RunTxModeDeliver), gasWanted)
		dctx, write := dctx.CacheContext()
		dctx = env.vmk.MakeGnoTransactionStore(dctx)
		res := env.vmh.Process(dctx, NewMsgCall(addr, nil, objcachePkg, fn, args))
		require.True(b, res.IsOK(), res.Log)
		env.vmk.CommitGnoTransactionStore(dctx)
		write()
		return dctx.GasMeter().GasConsumed()
	}
	// Populate in random order, so the tree is balanced on average.
	rng := rand.New(rand.NewPCG(1, 2))
	keys := rng.Perm(size)
	for lo := 0; lo < size; lo += 100 {
		chunk := make([]string, 0, 100)
		for _, k := range keys[lo:min(lo+100, size)] {
			chunk = append(chunk, strconv.Itoa(k))
		}
		call(1_000_000_000, "Fill", strings.Join(chunk, ","))
	}
	committer := env.ctx.MultiStore().(types.Committer)
	committer.Commit()

	var gas int64
	h0, m0 := gno.ObjectDecodeCacheStats()
	b.ResetTimer()
	for i := range b.N {
		gas += call(50_000_000, "Set", strconv.Itoa(rng.IntN(size)), "x"+strconv.Itoa(i))
		if i%100 == 99 {
			committer.Commit()
		}
	}
	b.StopTimer()
	h1, m1 := gno.ObjectDecodeCacheStats()
	b.ReportMetric(float64(gas)/float64(b.N), "gas/op")
	b.ReportMetric(float64(h1-h0)/float64(b.N), "hits/op")
	b.ReportMetric(float64(m1-m0)/float64(b.N), "misses/op")
}
