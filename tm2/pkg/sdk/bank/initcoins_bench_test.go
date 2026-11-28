package bank

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// Guards the genesis balance-loading regression: SetCoins enumerates an
// address's existing split-tier keys, and cacheStore answers that range query
// by scanning every dirty key written so far, so a loop of n SetCoins calls is
// O(n^2). InitCoins skips the enumeration on addresses that provably hold
// nothing yet.
//
// Run both and compare; the gap grows with n:
//
//	go test ./pkg/sdk/bank/ -run XXX -bench 'CoinsLoad' -benchtime 1x
//
// At n=100_000 the SetCoins form is minutes and the InitCoins form is seconds.
// If they ever converge, either the store was fixed (good -- say so here) or
// the genesis fast path was lost (bad).
func benchCoinsLoad(b *testing.B, n int, init bool) {
	b.Helper()
	amt := std.NewCoins(std.NewCoin(testAccountDenom, 1_000_000))
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		env := setupTestEnv()
		// InitChain runs against a cache-wrapped multistore (BaseApp does this
		// for deliverState), and the quadratic lives in that cache layer --
		// cacheStore.dirtyItems. Benchmarking against the bare IAVL store from
		// setupTestEnv would measure tree cost only and show nothing.
		env.ctx = env.ctx.WithMultiStore(env.ctx.MultiStore().MultiCacheWrap())
		b.StartTimer()
		for a := range n {
			var err error
			if init {
				err = env.bankk.InitCoins(env.ctx, addrN(a), amt)
			} else {
				err = env.bankk.SetCoins(env.ctx, addrN(a), amt)
			}
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkCoinsLoadSetCoins10000(b *testing.B)   { benchCoinsLoad(b, 10_000, false) }
func BenchmarkCoinsLoadInitCoins10000(b *testing.B)  { benchCoinsLoad(b, 10_000, true) }
func BenchmarkCoinsLoadSetCoins20000(b *testing.B)   { benchCoinsLoad(b, 20_000, false) }
func BenchmarkCoinsLoadInitCoins20000(b *testing.B)  { benchCoinsLoad(b, 20_000, true) }
func BenchmarkCoinsLoadSetCoins100000(b *testing.B)  { benchCoinsLoad(b, 100_000, false) }
func BenchmarkCoinsLoadInitCoins100000(b *testing.B) { benchCoinsLoad(b, 100_000, true) }
