package bank

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeeper(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))
	addr3 := crypto.AddressFromPreimage([]byte("addr3"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	env.acck.SetAccount(ctx, acc)
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins()))

	env.bankk.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))

	// Test HasCoins
	require.True(t, env.bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, env.bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 5))))
	require.False(t, env.bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15))))
	require.False(t, env.bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 5))))

	// Test AddCoins
	env.bankk.AddCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 25))))

	env.bankk.AddCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 15)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 15), std.NewCoin("foocoin", 25))))

	// Test SubtractCoins
	env.bankk.SubtractCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10)))
	env.bankk.SubtractCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 5)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 15))))

	env.bankk.SubtractCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 11)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 15))))

	env.bankk.SubtractCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 10)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 15))))
	require.False(t, env.bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 1))))

	// Test SendCoins
	env.bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("foocoin", 5)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, env.bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5))))

	_ = env.bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("foocoin", 50)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, env.bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5))))

	env.bankk.AddCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 30)))
	env.bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 5)))
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 20), std.NewCoin("foocoin", 5))))
	require.True(t, env.bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 10))))

	// Test InputOutputCoins
	input1 := NewInput(addr2, std.NewCoins(std.NewCoin("foocoin", 2)))
	output1 := NewOutput(addr, std.NewCoins(std.NewCoin("foocoin", 2)))
	env.bankk.InputOutputCoins(ctx, []Input{input1}, []Output{output1})
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 20), std.NewCoin("foocoin", 7))))
	require.True(t, env.bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 8))))

	inputs := []Input{
		NewInput(addr, std.NewCoins(std.NewCoin("foocoin", 3))),
		NewInput(addr2, std.NewCoins(std.NewCoin("barcoin", 3), std.NewCoin("foocoin", 2))),
	}

	outputs := []Output{
		NewOutput(addr, std.NewCoins(std.NewCoin("barcoin", 1))),
		NewOutput(addr3, std.NewCoins(std.NewCoin("barcoin", 2), std.NewCoin("foocoin", 5))),
	}
	env.bankk.InputOutputCoins(ctx, inputs, outputs)
	require.True(t, env.bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 21), std.NewCoin("foocoin", 4))))
	require.True(t, env.bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 7), std.NewCoin("foocoin", 6))))
	require.True(t, env.bankk.GetCoins(ctx, addr3).IsEqual(std.NewCoins(std.NewCoin("barcoin", 2), std.NewCoin("foocoin", 5))))
}

func TestBankKeeper(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx

	bankk := env.bankk

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	env.acck.SetAccount(ctx, acc)
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins()))

	env.bankk.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10)))
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))

	// Test HasCoins
	require.True(t, bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 5))))
	require.False(t, bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15))))
	require.False(t, bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 5))))

	env.bankk.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15)))

	// Test SendCoins
	bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("foocoin", 5)))
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5))))

	err := bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("foocoin", 50)))
	require.Error(t, err)
	// Balances of addr and addr2 should stay the same.
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 5))))

	env.bankk.AddCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 30)))
	bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 5)))
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 20), std.NewCoin("foocoin", 5))))
	require.True(t, bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 10))))

	// validate coins with invalid denoms or negative values cannot be sent
	// NOTE: We must use the Coin literal as the constructor does not allow
	// negative values.
	err = bankk.SendCoins(ctx, addr, addr2, sdk.Coins{sdk.Coin{Denom: "FOOCOIN", Amount: -5}})
	require.Error(t, err)
}

func TestViewKeeper(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	view := NewViewKeeper(env.acck)

	addr := crypto.AddressFromPreimage([]byte("addr1"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)

	// Test GetCoins/SetCoins
	env.acck.SetAccount(ctx, acc)
	require.True(t, view.GetCoins(ctx, addr).IsEqual(std.NewCoins()))

	env.bankk.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10)))
	require.True(t, view.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))

	// Test HasCoins
	require.True(t, view.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, view.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 5))))
	require.False(t, view.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15))))
	require.False(t, view.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 5))))
}

// Test SetRestrictedDenoms
func TestSetRestrictedDenoms(t *testing.T) {
	env := setupTestEnv()
	ctx := env.ctx
	bankk := env.bankk
	prmk := env.prmk
	// Add a single denom
	prmk.SetStrings(ctx, "bank:p:restricted_denoms", []string{"foo"})
	params := bankk.GetParams(ctx)
	require.Contains(t, params.RestrictedDenoms, "foo")

	// Add multiple denoms
	prmk.SetStrings(ctx, "bank:p:restricted_denoms", []string{"goo", "bar"})
	params = bankk.GetParams(ctx)
	require.NotContains(t, params.RestrictedDenoms, "foo")
	require.Contains(t, params.RestrictedDenoms, "goo")
	require.Contains(t, params.RestrictedDenoms, "bar")

	// Add empty list
	prmk.SetStrings(ctx, "bank:p:restricted_denoms", []string{})
	params = bankk.GetParams(ctx)
	require.Empty(t, params.RestrictedDenoms)
}

//nolint:tparallel // subtests share keeper/state; must run serially
func TestTrackBalanceChange(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx

	addr1 := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))
	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	acc2 := env.acck.NewAccountWithAddress(ctx, addr2)
	balance1 := std.NewCoins(std.NewCoin("foo", 10), std.NewCoin("bar", 5), std.NewCoin("baz", 2), std.NewCoin("qux", 6))
	balance2 := std.NewCoins(std.NewCoin("bar", 3), std.NewCoin("baz", 2))

	// Track only foo and bar in TotalSupply (qux excluded on purpose)
	params := env.bankk.GetParams(ctx)
	params.TotalSupply = []std.Coin{
		std.NewCoin("foo", 10),
		std.NewCoin("bar", 10),
		std.NewCoin("baz", 10),
	}
	env.bankk.SetParams(ctx, params)
	// Test GetCoins/SetCoins
	env.acck.SetAccount(ctx, acc1)
	env.bankk.SetCoins(ctx, addr1, balance1)
	env.acck.SetAccount(ctx, acc2)
	env.bankk.SetCoins(ctx, addr2, balance2)

	sendCoinsCases := []struct {
		name    string
		amt     std.Coins
		wantInc std.Coins
		wantDec std.Coins
	}{
		{
			name:    "send: from addr1 to addr2 2foo, 1bar",
			amt:     std.NewCoins(std.NewCoin("foo", 2), std.NewCoin("bar", 1)),
			wantInc: std.NewCoins(std.NewCoin("foo", 2), std.NewCoin("bar", 1)),
			wantDec: std.NewCoins(std.NewCoin("foo", 2), std.NewCoin("bar", 1)),
		},
		{
			name:    "send: from addr1 to addr2 1qux and 1bar",
			amt:     std.NewCoins(std.NewCoin("qux", 1), std.NewCoin("bar", 1)),
			wantInc: std.NewCoins(std.NewCoin("foo", 2), std.NewCoin("bar", 2)),
			wantDec: std.NewCoins(std.NewCoin("foo", 2), std.NewCoin("bar", 2)),
		},
		{
			name:    "send: from addr1 to addr2 1foo, 1bar, 1baz 1qux",
			amt:     std.NewCoins(std.NewCoin("foo", 1), std.NewCoin("bar", 1), std.NewCoin("baz", 1), std.NewCoin("qux", 1)),
			wantInc: std.NewCoins(std.NewCoin("foo", 3), std.NewCoin("bar", 3), std.NewCoin("baz", 1)),
			wantDec: std.NewCoins(std.NewCoin("foo", 3), std.NewCoin("bar", 3), std.NewCoin("baz", 1)),
		},
	}

	// Run all steps and check cumulative counters
	for _, tc := range sendCoinsCases {
		t.Run(tc.name, func(t *testing.T) {
			env.bankk.SetCoins(ctx, addr1, balance1)
			env.bankk.SetCoins(ctx, addr2, balance2)

			env.bankk.SendCoins(ctx, addr1, addr2, tc.amt)
			inc, dec := readCounters(t, ctx, env.bankk)
			assert.True(t, inc.IsEqual(tc.wantInc),
				"inc mismatch: got=%v want=%v", inc.Sort(), tc.wantInc.Sort())
			assert.True(t, dec.IsEqual(tc.wantDec),
				"inc mismatch: got=%v want=%v", dec.Sort(), tc.wantDec.Sort())
		})
	}

	t.Run("CheckTx: tracking skipped", func(t *testing.T) {
		checkCtx := ctx.WithMode(sdk.RunTxModeCheck)
		beforeInc, beforeDec := readCounters(t, ctx, env.bankk)

		old1 := env.bankk.GetCoins(ctx, addr1)
		new1 := old1.Add(std.NewCoins(std.NewCoin("foo", 1)))
		// In CheckTx: should be NO-OP
		env.bankk.trackBalanceChange(checkCtx, old1, new1)

		afterInc, afterDec := readCounters(t, ctx, env.bankk)
		if !afterInc.IsEqual(beforeInc) || !afterDec.IsEqual(beforeDec) {
			t.Fatalf("counters changed during CheckTx; inc %v->%v dec %v->%v",
				beforeInc, afterInc, beforeDec, afterDec)
		}
	})

	inc, dec := readCounters(t, ctx, env.bankk)
	assert.True(t, inc.IsEqual(dec),
		"inc and dec mismatch: inc=%v dec=%v", inc.Sort(), dec.Sort())
}
func readCounters(t *testing.T, ctx sdk.Context, bank BankKeeper) (std.Coins, std.Coins) {
	t.Helper()
	store := ctx.Store(bank.key)

	var inc, dec std.Coins
	if bz := store.Get([]byte(balanceIncKey)); bz != nil {
		amino.MustUnmarshal(bz, &inc)
	}
	if bz := store.Get([]byte(balanceDecKey)); bz != nil {
		amino.MustUnmarshal(bz, &dec)
	}
	return inc, dec
}

func TestDiffCoins(t *testing.T) {
	tests := []struct {
		name    string
		old     std.Coins
		new     std.Coins
		denoms  []string
		wantInc std.Coins
		wantDec std.Coins
	}{
		{
			name:    "no changes (exact equal)",
			old:     std.NewCoins(std.NewCoin("acoin", 10), std.NewCoin("bcoin", 5)),
			new:     std.NewCoins(std.NewCoin("acoin", 10), std.NewCoin("bcoin", 5)),
			denoms:  []string{"acoin", "bcoin"},
			wantInc: std.NewCoins(),
			wantDec: std.NewCoins(),
		},
		{
			name:    "increase same denom",
			old:     std.NewCoins(std.NewCoin("acoin", 10)),
			new:     std.NewCoins(std.NewCoin("acoin", 15)),
			denoms:  []string{"acoin"},
			wantInc: std.NewCoins(std.NewCoin("acoin", 5)),
			wantDec: std.NewCoins(),
		},
		{
			name:    "decrease same denom",
			old:     std.NewCoins(std.NewCoin("acoin", 10)),
			new:     std.NewCoins(std.NewCoin("acoin", 6)),
			denoms:  []string{"acoin"},
			wantInc: std.NewCoins(),
			wantDec: std.NewCoins(std.NewCoin("acoin", 4)),
		},
		{
			name:    "denom only in old -> full decrease",
			old:     std.NewCoins(std.NewCoin("ccoin", 7)),
			new:     std.NewCoins(),
			denoms:  []string{"ccoin"},
			wantInc: std.NewCoins(),
			wantDec: std.NewCoins(std.NewCoin("ccoin", 7)),
		},
		{
			name:    "denom only in new -> full increase",
			old:     std.NewCoins(),
			new:     std.NewCoins(std.NewCoin("dcoin", 9)),
			denoms:  []string{"dcoin"},
			wantInc: std.NewCoins(std.NewCoin("dcoin", 9)),
			wantDec: std.NewCoins(),
		},
		{
			name:   "mixed increases/decreases + ignore non-total-supply denoms",
			old:    std.NewCoins(std.NewCoin("acoin", 5), std.NewCoin("ccoin", 2), std.NewCoin("ecoin", 3)),
			new:    std.NewCoins(std.NewCoin("acoin", 7), std.NewCoin("bcoin", 4), std.NewCoin("ecoin", 1), std.NewCoin("fcoin", 10)),
			denoms: []string{"acoin", "ccoin", "ecoin", "fcoin"}, // "bcoin" is excluded -> ignored
			// a: +2, c: -2, e: -2, f: +10
			wantInc: std.NewCoins(std.NewCoin("acoin", 2), std.NewCoin("fcoin", 10)),
			wantDec: std.NewCoins(std.NewCoin("ccoin", 2), std.NewCoin("ecoin", 2)),
		},
		{
			name:   "unsorted inputs handled via Sort()",
			old:    std.NewCoins(std.NewCoin("bcoin", 1), std.NewCoin("acoin", 2)),
			new:    std.NewCoins(std.NewCoin("acoin", 3), std.NewCoin("bcoin", 1)),
			denoms: []string{"acoin", "bcoin"},
			// a: +1, b: 0
			wantInc: std.NewCoins(std.NewCoin("acoin", 1)),
			wantDec: std.NewCoins(),
		},
		{
			name:   "tails: leftovers in old and new",
			old:    std.NewCoins(std.NewCoin("acoin", 1), std.NewCoin("bcoin", 1), std.NewCoin("ccoin", 1)),
			new:    std.NewCoins(std.NewCoin("acoin", 1), std.NewCoin("dcoin", 2)),
			denoms: []string{"acoin", "bcoin", "ccoin", "dcoin"},
			// dec: b:1,c:1 ; inc: d:2
			wantInc: std.NewCoins(std.NewCoin("dcoin", 2)),
			wantDec: std.NewCoins(std.NewCoin("bcoin", 1), std.NewCoin("ccoin", 1)),
		},

		{
			name:    "ignored denoms not in denomSet",
			old:     std.NewCoins(std.NewCoin("zcoin", 100)),
			new:     std.NewCoins(std.NewCoin("zcoin", 1)),
			denoms:  []string{"acoin", "bcoin"}, // z not tracked
			wantInc: std.NewCoins(),
			wantDec: std.NewCoins(),
		},

		{
			name:    "redundant zero change (equal amounts) produces no inc/dec",
			old:     std.NewCoins(std.NewCoin("ycoin", 5)),
			new:     std.NewCoins(std.NewCoin("ycoin", 5)),
			denoms:  []string{"ycoin"},
			wantInc: std.NewCoins(),
			wantDec: std.NewCoins(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inc, dec := diffCoins(tc.old, tc.new, toSet(tc.denoms))

			assert.True(t, inc.IsEqual(tc.wantInc),
				"inc mismatch: got=%v want=%v", inc.Sort(), tc.wantInc.Sort())

			assert.True(t, dec.IsEqual(tc.wantDec),
				"dec mismatch: got=%v want=%v", dec.Sort(), tc.wantDec.Sort())
		})
	}
}
