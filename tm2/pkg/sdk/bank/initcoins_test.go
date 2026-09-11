package bank

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// InitCoins is only sound because it produces byte-identical store state to
// SetCoins whenever the address holds nothing yet. These tests pin that
// equivalence rather than trusting the argument.

// everyStoreKV dumps the whole bank store, so two runs can be compared exactly
// (values and ordering both).
func everyStoreKV(t *testing.T, env testEnv) []std.KVPair {
	t.Helper()
	var out []std.KVPair
	stor := env.ctx.Store(env.bankk.key)
	it := stor.Iterator(nil, nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		k := append([]byte(nil), it.Key()...)
		v := append([]byte(nil), it.Value()...)
		out = append(out, std.KVPair{Key: k, Value: v})
	}
	return out
}

func addrN(i int) crypto.Address {
	var a crypto.Address
	copy(a[:], fmt.Sprintf("addr-%014d", i))
	return a
}

// The core claim: on a fresh address the two calls are indistinguishable.
func TestInitCoinsMatchesSetCoinsOnFreshAddress(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		amt  std.Coins
	}{
		{"single ugnot", std.NewCoins(std.NewCoin("ugnot", 1_000_000))},
		{"split tier only", std.NewCoins(std.NewCoin("zzzcoin", 5))},
		{"both tiers", std.NewCoins(std.NewCoin("ugnot", 7), std.NewCoin("zzzcoin", 9))},
		{"many denoms", std.NewCoins(
			std.NewCoin("aaa", 1), std.NewCoin("bbb", 2), std.NewCoin("ccc", 3),
			std.NewCoin("ddd", 4), std.NewCoin("ugnot", 5),
		)},
		{"empty", std.NewCoins()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			addr := addrN(1)

			envSet := setupTestEnv()
			errSet := envSet.bankk.SetCoins(envSet.ctx, addr, tc.amt)

			envInit := setupTestEnv()
			errInit := envInit.bankk.InitCoins(envInit.ctx, addr, tc.amt)

			require.Equal(t, errSet == nil, errInit == nil, "error agreement")
			require.Equal(t, everyStoreKV(t, envSet), everyStoreKV(t, envInit),
				"store state must be byte-identical")
			require.Equal(t,
				envSet.bankk.GetCoins(envSet.ctx, addr),
				envInit.bankk.GetCoins(envInit.ctx, addr),
				"GetCoins must agree")
		})
	}
}

// Invalid input must be rejected identically, before anything is written.
func TestInitCoinsRejectsInvalidCoinsLikeSetCoins(t *testing.T) {
	t.Parallel()
	addr := addrN(2)
	bad := std.Coins{std.Coin{Denom: "ugnot", Amount: -1}}

	envSet := setupTestEnv()
	errSet := envSet.bankk.SetCoins(envSet.ctx, addr, bad)

	envInit := setupTestEnv()
	errInit := envInit.bankk.InitCoins(envInit.ctx, addr, bad)

	require.Error(t, errSet)
	require.Error(t, errInit)
	require.Equal(t, errSet.Error(), errInit.Error())
	require.Equal(t, everyStoreKV(t, envSet), everyStoreKV(t, envInit))
}

// A whole genesis-shaped load, both ways, compared as a single store image.
// This is the case the genesis loader actually runs.
func TestInitCoinsWholeLoadMatchesSetCoins(t *testing.T) {
	t.Parallel()

	amounts := func(i int) std.Coins {
		switch i % 4 {
		case 0:
			return std.NewCoins(std.NewCoin("ugnot", int64(i+1)))
		case 1:
			return std.NewCoins(std.NewCoin("zzzcoin", int64(i+1)))
		case 2:
			return std.NewCoins(std.NewCoin("ugnot", int64(i+1)), std.NewCoin("zzzcoin", int64(i+2)))
		default:
			return std.NewCoins(std.NewCoin("aaa", int64(i+1)), std.NewCoin("mmm", int64(i+3)))
		}
	}

	envSet := setupTestEnv()
	envInit := setupTestEnv()
	for i := range 200 {
		require.NoError(t, envSet.bankk.SetCoins(envSet.ctx, addrN(i), amounts(i)))
		require.NoError(t, envInit.bankk.InitCoins(envInit.ctx, addrN(i), amounts(i)))
	}
	require.Equal(t, everyStoreKV(t, envSet), everyStoreKV(t, envInit))
}

// The precondition has teeth: on an address that already holds a split-tier
// denom the new amount does not cover, InitCoins leaves the stale key behind
// where SetCoins removes it. Pinned so nobody "optimises" SetCoins into
// InitCoins at a call site that cannot guarantee freshness.
func TestInitCoinsDoesNotDrainStaleKeys(t *testing.T) {
	t.Parallel()
	addr := addrN(3)
	first := std.NewCoins(std.NewCoin("aaa", 10), std.NewCoin("bbb", 20))
	second := std.NewCoins(std.NewCoin("aaa", 1))

	envSet := setupTestEnv()
	require.NoError(t, envSet.bankk.SetCoins(envSet.ctx, addr, first))
	require.NoError(t, envSet.bankk.SetCoins(envSet.ctx, addr, second))

	envInit := setupTestEnv()
	require.NoError(t, envInit.bankk.SetCoins(envInit.ctx, addr, first))
	require.NoError(t, envInit.bankk.InitCoins(envInit.ctx, addr, second))

	require.Equal(t, second, envSet.bankk.GetCoins(envSet.ctx, addr),
		"SetCoins drains bbb")
	require.NotEqual(t, second, envInit.bankk.GetCoins(envInit.ctx, addr),
		"InitCoins is documented not to drain; if this ever passes, the "+
			"precondition has been silently relaxed and the genesis fast path "+
			"needs rechecking")
}
