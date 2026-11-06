package auth

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestSigVerifyCacheEvictsOldestWhenFull(t *testing.T) {
	t.Parallel()

	k := func(b byte) [sha256.Size]byte { return [sha256.Size]byte{b} }
	c := NewSigVerifyCache(2)

	c.add(k(1))
	c.add(k(2))
	c.add(k(2)) // already present: no slot used, nothing evicted
	require.Equal(t, 2, c.Len())
	require.True(t, c.has(k(1)))

	c.add(k(3))
	require.Equal(t, 2, c.Len())
	require.False(t, c.has(k(1)), "oldest entry must be evicted")
	require.True(t, c.has(k(2)))
	require.True(t, c.has(k(3)))

	c.add(k(4))
	require.Equal(t, 2, c.Len())
	require.False(t, c.has(k(2)))
	require.True(t, c.has(k(3)))
	require.True(t, c.has(k(4)))

	require.Panics(t, func() { NewSigVerifyCache(0) })
}

func TestSigVerifyKeyCoversEveryInput(t *testing.T) {
	t.Parallel()

	pub1 := secp256k1.GenPrivKey().PubKey()
	pub2 := secp256k1.GenPrivKey().PubKey()
	base := sigVerifyKey(pub1, []byte("sign bytes"), []byte("sig"))

	for name, other := range map[string][sha256.Size]byte{
		"pubkey":     sigVerifyKey(pub2, []byte("sign bytes"), []byte("sig")),
		"sign bytes": sigVerifyKey(pub1, []byte("sign bytez"), []byte("sig")),
		"signature":  sigVerifyKey(pub1, []byte("sign bytes"), []byte("sih")),
		// Same concatenation, different split: the length prefixes must
		// keep these apart.
		"boundary": sigVerifyKey(pub1, []byte("sign byte"), []byte("ssig")),
	} {
		require.NotEqual(t, base, other, name)
	}
	require.Equal(t, base, sigVerifyKey(pub1, []byte("sign bytes"), []byte("sig")))
}

// cachedSignerEnv is singleSignerEnv with a SigVerifyCache wired in.
func cachedSignerEnv(t *testing.T, size int) (singleSignerEnv, *SigVerifyCache) {
	t.Helper()

	cache := NewSigVerifyCache(size)
	return signerEnvWith(t, newSingleSignerEnv(t).priv, cache), cache
}

// signerEnvWith is singleSignerEnv for a given key and cache (nil for none).
func signerEnvWith(t *testing.T, priv crypto.PrivKey, cache *SigVerifyCache) singleSignerEnv {
	t.Helper()

	env := setupTestEnv()
	opts := defaultAnteOptions()
	opts.SigVerifyCache = cache
	addr := priv.PubKey().Address()
	acc := env.acck.NewAccountWithAddress(env.ctx, addr)
	acc.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc.SetAccountNumber(0))
	env.acck.SetAccount(env.ctx, acc)
	fee := tu.NewTestFee()
	fee.GasWanted = checkGasWanted
	return singleSignerEnv{
		anteHandler: NewAnteHandler(env.acck, env.bankk, DefaultSigVerificationGasConsumer, opts),
		ctx:         env.ctx,
		priv:        priv,
		msgs:        []std.Msg{tu.NewTestMsg(addr)},
		fee:         fee,
	}
}

// checkGasWanted covers the flat read depth a CheckTx branch is charged.
const checkGasWanted = 10_000_000

// checkCtx is a CheckTx context on a throwaway branch of e.ctx, so the same
// tx can be checked again, as recheck does, without its sequence moving.
func (e singleSignerEnv) checkCtx() sdk.Context {
	return e.ctx.WithMode(sdk.RunTxModeCheck).
		WithMultiStore(e.ctx.MultiStore().MultiCacheWrap()).
		WithValue(GasPriceContextKey{}, std.GasPrice{})
}

func (e singleSignerEnv) run(ctx sdk.Context, tx std.Tx) (int64, sdk.Result) {
	newCtx, res, _ := e.anteHandler(ctx, tx, false)
	return newCtx.GasMeter().GasConsumed(), res
}

func TestAnteHandlerSigVerifyCache(t *testing.T) {
	t.Parallel()

	t.Run("check caches a verified signature, gas unchanged on hit", func(t *testing.T) {
		t.Parallel()

		e, cache := cachedSignerEnv(t, 16)
		tx := tu.NewTestTx(t, e.ctx.ChainID(), e.msgs, []crypto.PrivKey{e.priv}, []uint64{0}, []uint64{0}, e.fee)

		missGas, res := e.run(e.checkCtx(), tx)
		require.True(t, res.IsOK(), res.Log)
		require.Equal(t, 1, cache.Len())

		hitGas, res := e.run(e.checkCtx(), tx)
		require.True(t, res.IsOK(), res.Log)
		require.Equal(t, 1, cache.Len())
		require.Equal(t, missGas, hitGas, "a cache hit must not change gas")

		// Same tx through a handler with no cache at all.
		plain := signerEnvWith(t, e.priv, nil)
		plainGas, res := plain.run(plain.checkCtx(), tx)
		require.True(t, res.IsOK(), res.Log)
		require.Equal(t, plainGas, hitGas, "the cache must not change gas")
	})

	t.Run("deliver neither reads nor fills the cache", func(t *testing.T) {
		t.Parallel()

		e, cache := cachedSignerEnv(t, 16)
		tx := tu.NewTestTx(t, e.ctx.ChainID(), e.msgs, []crypto.PrivKey{e.priv}, []uint64{0}, []uint64{0}, e.fee)
		_, res := e.run(e.ctx, tx) // e.ctx is a DeliverTx context
		require.True(t, res.IsOK(), res.Log)
		require.Equal(t, 0, cache.Len())

		// Plant an entry for a signature that does not verify. CheckTx
		// trusts it (which shows the lookup is live); DeliverTx must not.
		bad := std.NewTx(e.msgs, e.fee, []std.Signature{{
			PubKey: e.priv.PubKey(), Signature: []byte("nope"),
		}}, "")
		signBytes, err := bad.GetSignBytes(e.ctx.ChainID(), 0, 1)
		require.NoError(t, err)
		cache.add(sigVerifyKey(e.priv.PubKey(), signBytes, []byte("nope")))

		_, res = e.run(e.checkCtx(), bad)
		require.True(t, res.IsOK(), "planted entry should be a hit in CheckTx: %s", res.Log)
		checkInvalidTx(t, e.anteHandler, e.ctx, bad, false, std.UnauthorizedError{})
	})

	t.Run("a stale sequence misses: no replay", func(t *testing.T) {
		t.Parallel()

		e, cache := cachedSignerEnv(t, 16)
		tx := tu.NewTestTx(t, e.ctx.ChainID(), e.msgs, []crypto.PrivKey{e.priv}, []uint64{0}, []uint64{0}, e.fee)
		_, res := e.run(e.checkCtx(), tx)
		require.True(t, res.IsOK(), res.Log)
		require.Equal(t, 1, cache.Len())

		// The tx is committed: the stored sequence moves to 1.
		_, res = e.run(e.ctx, tx)
		require.True(t, res.IsOK(), res.Log)

		// Resubmitted with its cached signature, it must still fail.
		checkInvalidTx(t, e.anteHandler, e.checkCtx(), tx, false, std.UnauthorizedError{})
	})

	t.Run("a signature moved onto another body misses", func(t *testing.T) {
		t.Parallel()

		e, cache := cachedSignerEnv(t, 16)
		tx := tu.NewTestTx(t, e.ctx.ChainID(), e.msgs, []crypto.PrivKey{e.priv}, []uint64{0}, []uint64{0}, e.fee)
		_, res := e.run(e.checkCtx(), tx)
		require.True(t, res.IsOK(), res.Log)
		require.Equal(t, 1, cache.Len())

		other := e.fee
		other.GasWanted--
		moved := std.NewTx(e.msgs, other, tx.Signatures, "")
		checkInvalidTx(t, e.anteHandler, e.checkCtx(), moved, false, std.UnauthorizedError{})
		moved = std.NewTx(e.msgs, e.fee, tx.Signatures, "memo")
		checkInvalidTx(t, e.anteHandler, e.checkCtx(), moved, false, std.UnauthorizedError{})
	})

	t.Run("a legacy-rendering signature is verified, never cached", func(t *testing.T) {
		t.Parallel()

		e, cache := cachedSignerEnv(t, 16)
		legacyBytes, err := std.GetSignaturePayloadLegacy(e.signDoc())
		require.NoError(t, err)
		tx := tu.NewTestTxWithSignBytes(e.msgs, []crypto.PrivKey{e.priv}, e.fee, legacyBytes, "")
		checkValidTx(t, e.anteHandler, e.checkCtx(), tx, false)
		require.Equal(t, 0, cache.Len())
	})

	t.Run("an evicted entry is verified again", func(t *testing.T) {
		t.Parallel()

		e, cache := cachedSignerEnv(t, 1)
		tx := tu.NewTestTx(t, e.ctx.ChainID(), e.msgs, []crypto.PrivKey{e.priv}, []uint64{0}, []uint64{0}, e.fee)
		_, res := e.run(e.checkCtx(), tx)
		require.True(t, res.IsOK(), res.Log)
		key := sigVerifyKey(e.priv.PubKey(), mustSignBytes(t, tx, e.ctx.ChainID()), tx.Signatures[0].Signature)
		require.True(t, cache.has(key))

		cache.add([sha256.Size]byte{0xff}) // evicts the only entry
		require.False(t, cache.has(key))

		_, res = e.run(e.checkCtx(), tx)
		require.True(t, res.IsOK(), res.Log)
		require.True(t, cache.has(key), "re-verified and stored again")
	})
}

func mustSignBytes(t *testing.T, tx std.Tx, chainID string) []byte {
	t.Helper()
	bz, err := tx.GetSignBytes(chainID, 0, 0)
	require.NoError(t, err)
	return bz
}

func BenchmarkAnteHandlerRecheck(b *testing.B) {
	for _, bc := range []struct {
		name  string
		cache *SigVerifyCache
	}{
		{"cache=off", nil},
		{"cache=on", NewSigVerifyCache(16)},
	} {
		b.Run(bc.name, func(b *testing.B) {
			env := setupTestEnv()
			opts := defaultAnteOptions()
			opts.SigVerifyCache = bc.cache
			ante := NewAnteHandler(env.acck, env.bankk, DefaultSigVerificationGasConsumer, opts)
			priv, _, addr := tu.KeyTestPubAddr()
			acc := env.acck.NewAccountWithAddress(env.ctx, addr)
			acc.SetCoins(tu.NewTestCoins())
			if err := acc.SetAccountNumber(0); err != nil {
				b.Fatal(err)
			}
			env.acck.SetAccount(env.ctx, acc)
			msgs, fee := []std.Msg{tu.NewTestMsg(addr)}, tu.NewTestFee()
			fee.GasWanted = checkGasWanted
			signBytes, err := std.GetSignaturePayload(std.SignDoc{
				ChainID: env.ctx.ChainID(), Fee: fee, Msgs: msgs,
			})
			if err != nil {
				b.Fatal(err)
			}
			tx := tu.NewTestTxWithSignBytes(msgs, []crypto.PrivKey{priv}, fee, signBytes, "")
			check := env.ctx.WithMode(sdk.RunTxModeCheck).WithValue(GasPriceContextKey{}, std.GasPrice{})

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				// Each iteration is one recheck of the same pending tx
				// against the same committed state.
				ctx := check.WithMultiStore(check.MultiStore().MultiCacheWrap())
				if _, res, abort := ante(ctx, tx, false); abort {
					b.Fatal(res.Log)
				}
			}
		})
	}
}
