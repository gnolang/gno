package vm

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParamsKeeper(t *testing.T) {
	env := setupTestEnv()
	params := NewSDKParams(env.prmk, env.ctx)

	testCases := []struct {
		name        string
		setFunc     func()
		expectedMsg string
	}{
		{
			name: "SetString should panic",
			setFunc: func() {
				params.SetString("foo:name", "foo")
			},
			expectedMsg: `module name <foo> not registered`,
		},
		{
			name: "SetString should panic (with realm)",
			setFunc: func() {
				params.SetString("foo:gno.land/r/user/repo:name", "foo")
			},
			expectedMsg: `module name <foo> not registered`,
		},
		{
			name: "SetBool should panic",
			setFunc: func() {
				params.SetBool("foo:name", true)
			},
			expectedMsg: `module name <foo> not registered`,
		},
		{
			name: "SetInt64 should panic",
			setFunc: func() {
				params.SetInt64("foo:name", -100)
			},
			expectedMsg: `module name <foo> not registered`,
		},
		{
			name: "SetUint64 should panic",
			setFunc: func() {
				params.SetUint64("foo:name", 100)
			},
			expectedMsg: `module name <foo> not registered`,
		},
		{
			name: "SetBytes should panic",
			setFunc: func() {
				params.SetBytes("foo:name", []byte("foo"))
			},
			expectedMsg: `module name <foo> not registered`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.PanicsWithValue(t, tc.expectedMsg, tc.setFunc, "The panic message did not match the expected value")
		})
	}
}

func TestSDKBankerTotalCoin(t *testing.T) {
	env := setupTestEnv()

	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	store := ctx.Store(env.supplyKey)
	banker := &SDKBanker{
		vmk:         env.vmk,
		ctx:         ctx,
		supplyStore: bank.NewSupplyStore(store),
		store:       store,
	}

	// create test accounts and set coins
	addr1 := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))

	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	env.acck.SetAccount(ctx, acc1)
	env.bankk.SetCoins(ctx, addr1, std.MustParseCoins("1000ugnot,500atom"))

	acc2 := env.acck.NewAccountWithAddress(ctx, addr2)
	env.acck.SetAccount(ctx, acc2)
	env.bankk.SetCoins(ctx, addr2, std.MustParseCoins("2000ugnot,1500atom"))

	// initialize supply
	err := banker.supplyStore.SetSupply(banker.store, "ugnot", 3000)
	require.NoError(t, err)
	err = banker.supplyStore.SetSupply(banker.store, "atom", 2000)
	require.NoError(t, err)

	tests := []struct {
		name      string
		denom     string
		expected  int64
		mustPanic bool
	}{
		{"ugnot total", "ugnot", 3000, false},
		{"atom total", "atom", 2000, false},
		{"non-existent denom", "non-existent", 0, false},
		{"zero balance accounts included", "ugnot", 3000, false},
		{"empty string denom", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mustPanic {
				assert.Panics(t, func() {
					banker.TotalCoin(tt.denom)
				})
			} else {
				total := banker.TotalCoin(tt.denom)
				assert.Equal(t, tt.expected, total)
			}
		})
	}
}

func BenchmarkSDKBankerTotal(b *testing.B) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	store := ctx.Store(env.supplyKey)
	banker := &SDKBanker{
		vmk:         env.vmk,
		ctx:         ctx,
		supplyStore: bank.NewSupplyStore(store),
		store:       store,
	}

	denoms := []string{"ugnot", "atom", "foo", "bar"}
	supplies := map[string]int64{
		"ugnot": 1_000_000,
		"atom":  2_000_000,
		"foo":   5_000_000,
		"bar":   100_000,
	}

	if len(denoms) != len(supplies) {
		b.Fatalf("len(denoms) != len(supplies)")
	}

	for _, denom := range denoms {
		err := banker.supplyStore.SetSupply(banker.store, denom, supplies[denom])
		require.NoError(b, err)
	}

	benchs := []struct {
		name  string
		denom string
	}{
		{"ExistingDenom_Small", "bar"},
		{"ExistingDenom_Medium", "atom"},
		{"ExistingDenom_Large", "ugnot"},
		{"NonExistentDenom", "foo"},
		{"NonExistentDenom", "non-existent"},
	}

	for _, bb := range benchs {
		b.Run(bb.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				banker.TotalCoin(bb.denom)
			}
		})
	}
}
