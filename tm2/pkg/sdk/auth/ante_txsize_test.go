package auth

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// signedTx builds a transaction the ante will accept, funding its signer.
func signedTx(t *testing.T, env testEnv) std.Tx {
	t.Helper()
	priv, _, addr := tu.KeyTestPubAddr()
	acc := env.acck.NewAccountWithAddress(env.ctx, addr)
	require.NoError(t, acc.SetCoins(std.Coins{{Denom: "atom", Amount: 100_000_000}}))
	env.acck.SetAccount(env.ctx, acc)
	return tu.NewTestTx(t, env.ctx.ChainID(), []std.Msg{tu.NewTestMsg(addr)},
		[]crypto.PrivKey{priv}, []uint64{acc.GetAccountNumber()}, []uint64{0}, tu.NewTestFee())
}

// The tx-size charge is TxSizeCostPerByte times the transaction's length. The
// parameter is only validated positive; nothing bounds it above, unlike the
// per-byte and depth parameters in gno.land/pkg/sdk/vm.
//
// So the product has to be taken with overflow.Mulp. A bare multiply wraps, and
// not always into a negative the meter would refuse: at a cost of 2^62+1 a
// 4096-byte transaction wraps to a charge of 4096, paying nothing worth
// mentioning for its size and saying nothing about it.
//
// Both cases set TxBytes. Nothing else in the tree does, so the charge is
// otherwise always taken against a length of zero and proves nothing.
func TestTxSizeChargeRefusesAnOverflowingCost(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	anteHandler := NewAnteHandler(env.acck, env.bankk, DefaultSigVerificationGasConsumer, defaultAnteOptions())
	tx := signedTx(t, env)

	params := DefaultParams()
	params.TxSizeCostPerByte = (1 << 62) + 1 // wraps small and positive, not negative
	ctx := env.ctx.
		WithValue(AuthParamsContextKey{}, params).
		WithTxBytes(make([]byte, 4096))

	require.PanicsWithValue(t, "multiplication overflow", func() {
		anteHandler(ctx, tx, false)
	}, "an overflowing tx-size charge must not be allowed to wrap into a small one")
}

// The ordinary charge still happens, and scales with the transaction. Without
// this the case above would also pass on an ante that never charged at all.
func TestTxSizeChargeGrowsWithTheTransaction(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	anteHandler := NewAnteHandler(env.acck, env.bankk, DefaultSigVerificationGasConsumer, defaultAnteOptions())

	consumed := func(t *testing.T, size int) int64 {
		t.Helper()
		tx := signedTx(t, env)
		ctx := env.ctx.WithTxBytes(make([]byte, size))
		newCtx, res, abort := anteHandler(ctx, tx, false)
		require.False(t, abort, "the ante must accept an ordinary transaction: %v", res.Log)
		return newCtx.GasMeter().GasConsumed()
	}

	short := consumed(t, 100)
	long := consumed(t, 300)

	params := DefaultParams()
	require.Equal(t, params.TxSizeCostPerByte*200, long-short,
		"the extra 200 bytes must be charged at exactly the per-byte rate")
}
