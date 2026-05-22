package fork

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// signTxAt signs a tx as-if the signer had (accNum, seq) at sign time and
// returns the Signature. The returned Signature embeds the pubkey so
// bruteForceSignerSequence can verify it.
func signTxAt(t *testing.T, priv crypto.PrivKey, tx std.Tx, chainID string, accNum, seq uint64) std.Signature {
	t.Helper()
	payload, err := std.GetSignaturePayload(std.SignDoc{
		ChainID:       chainID,
		AccountNumber: accNum,
		Sequence:      seq,
		Fee:           tx.Fee,
		Msgs:          tx.Msgs,
		Memo:          tx.Memo,
	})
	require.NoError(t, err)

	sig, err := priv.Sign(payload)
	require.NoError(t, err)

	return std.Signature{
		PubKey:    priv.PubKey(),
		Signature: sig,
	}
}

func makeTestTx(t *testing.T, priv crypto.PrivKey) std.Tx {
	t.Helper()
	msg := bank.MsgSend{
		FromAddress: priv.PubKey().Address(),
		ToAddress:   priv.PubKey().Address(), // doesn't matter for sig test
		Amount:      std.NewCoins(std.NewCoin("ugnot", 100)),
	}
	return std.Tx{
		Msgs: []std.Msg{msg},
		Fee:  std.NewFee(50000, std.NewCoin("ugnot", 1000)),
		Memo: "test",
	}
}

func TestBruteForceSignerSequence(t *testing.T) {
	t.Parallel()

	chainID := "test-chain"
	priv := ed25519.GenPrivKey()
	accNum := uint64(42)

	t.Run("finds correct sequence in range", func(t *testing.T) {
		t.Parallel()
		tx := makeTestTx(t, priv)
		actualSeq := uint64(7)
		sig := signTxAt(t, priv, tx, chainID, accNum, actualSeq)

		resolved, err := bruteForceSignerSequence(tx, sig, accNum, 0, 20, chainID)
		require.NoError(t, err)
		assert.Equal(t, actualSeq, resolved)
	})

	t.Run("finds sequence at lo boundary", func(t *testing.T) {
		t.Parallel()
		tx := makeTestTx(t, priv)
		sig := signTxAt(t, priv, tx, chainID, accNum, 5)

		resolved, err := bruteForceSignerSequence(tx, sig, accNum, 5, 10, chainID)
		require.NoError(t, err)
		assert.Equal(t, uint64(5), resolved)
	})

	t.Run("finds sequence at hi boundary", func(t *testing.T) {
		t.Parallel()
		tx := makeTestTx(t, priv)
		sig := signTxAt(t, priv, tx, chainID, accNum, 10)

		resolved, err := bruteForceSignerSequence(tx, sig, accNum, 5, 10, chainID)
		require.NoError(t, err)
		assert.Equal(t, uint64(10), resolved)
	})

	t.Run("lo==hi with correct value", func(t *testing.T) {
		t.Parallel()
		tx := makeTestTx(t, priv)
		sig := signTxAt(t, priv, tx, chainID, accNum, 3)

		resolved, err := bruteForceSignerSequence(tx, sig, accNum, 3, 3, chainID)
		require.NoError(t, err)
		assert.Equal(t, uint64(3), resolved)
	})

	t.Run("sequence outside range returns error", func(t *testing.T) {
		t.Parallel()
		tx := makeTestTx(t, priv)
		sig := signTxAt(t, priv, tx, chainID, accNum, 100)

		_, err := bruteForceSignerSequence(tx, sig, accNum, 0, 20, chainID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no sequence in")
	})

	t.Run("wrong account number returns error", func(t *testing.T) {
		t.Parallel()
		tx := makeTestTx(t, priv)
		sig := signTxAt(t, priv, tx, chainID, accNum, 5)

		// Sign says accNum=42 but we search assuming 99.
		_, err := bruteForceSignerSequence(tx, sig, 99, 0, 20, chainID)
		require.Error(t, err)
	})

	t.Run("wrong chain ID returns error", func(t *testing.T) {
		t.Parallel()
		tx := makeTestTx(t, priv)
		sig := signTxAt(t, priv, tx, chainID, accNum, 5)

		// Sign says chainID="test-chain" but we search with "other-chain".
		_, err := bruteForceSignerSequence(tx, sig, accNum, 0, 20, "other-chain")
		require.Error(t, err)
	})

	t.Run("nil pubkey returns error", func(t *testing.T) {
		t.Parallel()
		tx := makeTestTx(t, priv)
		sig := std.Signature{PubKey: nil, Signature: []byte("dummy")}

		_, err := bruteForceSignerSequence(tx, sig, accNum, 0, 20, chainID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no pubkey")
	})

	t.Run("secp256k1 key also works", func(t *testing.T) {
		t.Parallel()
		sPriv := secp256k1.GenPrivKey()
		tx := makeTestTx(t, sPriv)
		sig := signTxAt(t, sPriv, tx, chainID, accNum, 12)

		resolved, err := bruteForceSignerSequence(tx, sig, accNum, 0, 20, chainID)
		require.NoError(t, err)
		assert.Equal(t, uint64(12), resolved)
	})

	t.Run("tampered tx fee rejects all sequences", func(t *testing.T) {
		t.Parallel()
		tx := makeTestTx(t, priv)
		sig := signTxAt(t, priv, tx, chainID, accNum, 5)

		// Tamper with the tx after signing.
		tampered := tx
		tampered.Fee = std.NewFee(99999, std.NewCoin("ugnot", 9999))

		_, err := bruteForceSignerSequence(tampered, sig, accNum, 0, 20, chainID)
		require.Error(t, err)
	})
}

// makeGnoAccountWire builds the exact wire bytes that gno.land's auth handler
// returns for `auth/accounts/<addr>`: amino.MarshalJSONIndent of the concrete
// *gnoland.GnoAccount that acck.GetAccount yields on a gno.land chain.
// See tm2/pkg/sdk/auth/handler.go queryAccount.
func makeGnoAccountWire(t *testing.T, acc *gnoland.GnoAccount) []byte {
	t.Helper()
	bz, err := amino.MarshalJSONIndent(acc, "", "  ")
	require.NoError(t, err)
	return bz
}

// TestQueryAccountAtHeight_WireFormatContract guards the regression fixed in
// this PR: auth/accounts/<addr> on gno.land returns a GnoAccount (BaseAccount
// + Attributes), and amino's strict-field policy rejects the previous decoders
// (wrapper-with-BaseAccount and bare std.BaseAccount). queryAccountAtHeight
// must decode as gnoland.GnoAccount.
func TestQueryAccountAtHeight_WireFormatContract(t *testing.T) {
	t.Parallel()

	priv := ed25519.GenPrivKey()
	pub := priv.PubKey()
	src := &gnoland.GnoAccount{
		BaseAccount: std.BaseAccount{
			Address:       pub.Address(),
			Coins:         std.NewCoins(std.NewCoin("ugnot", 12345)),
			PubKey:        pub,
			AccountNumber: 42,
			Sequence:      7,
		},
		// Non-zero attributes — the field that broke the old decoders.
		Attributes: gnoland.BitSet(0x3),
	}
	wire := makeGnoAccountWire(t, src)

	t.Run("decodes as GnoAccount and round-trips BaseAccount", func(t *testing.T) {
		t.Parallel()
		var got gnoland.GnoAccount
		require.NoError(t, amino.UnmarshalJSON(wire, &got))

		assert.Equal(t, src.Address, got.Address)
		assert.Equal(t, src.AccountNumber, got.GetAccountNumber())
		assert.Equal(t, src.Sequence, got.GetSequence())
		assert.Equal(t, src.Attributes, got.Attributes)
		assert.True(t, src.Coins.IsEqual(got.Coins), "coins mismatch: want %s, got %s", src.Coins, got.Coins)
	})

	t.Run("regression: old wrapper-with-BaseAccount decoder errors", func(t *testing.T) {
		t.Parallel()
		// First decoder the buggy code tried. amino's strict-field policy
		// rejects the unknown "attributes" key. The buggy production path was
		// gated on (err == nil), so the runtime-relevant property is the error
		// itself — amino does partially populate the BaseAccount field before
		// failing, but the buggy caller never returned that value.
		var wrapper struct {
			BaseAccount std.BaseAccount `json:"BaseAccount"`
		}
		err := amino.UnmarshalJSON(wire, &wrapper)
		require.Error(t, err, "old wrapper decoder must fail on real GnoAccount wire format")
	})

	t.Run("regression: old bare-BaseAccount fallback errors", func(t *testing.T) {
		t.Parallel()
		// Fallback the buggy code tried second. Both "BaseAccount" and
		// "attributes" are unknown to a bare BaseAccount, so amino rejects the
		// payload. The buggy caller was gated on err == nil, so the property
		// we lock in is the error itself; amino's partial-fill behavior on
		// the target isn't part of the contract we depend on.
		var bare std.BaseAccount
		err := amino.UnmarshalJSON(wire, &bare)
		require.Error(t, err, "old bare-BaseAccount decoder must fail on real GnoAccount wire format")
	})

	t.Run("decodes a zero-attributes GnoAccount", func(t *testing.T) {
		t.Parallel()
		// Common on-chain case: a freshly created account with no flags set.
		zero := &gnoland.GnoAccount{
			BaseAccount: std.BaseAccount{
				Address:       pub.Address(),
				PubKey:        pub,
				AccountNumber: 1,
				Sequence:      0,
			},
		}
		zwire := makeGnoAccountWire(t, zero)

		var got gnoland.GnoAccount
		require.NoError(t, amino.UnmarshalJSON(zwire, &got))
		assert.Equal(t, uint64(1), got.GetAccountNumber())
		assert.Equal(t, uint64(0), got.GetSequence())
		assert.Equal(t, gnoland.BitSet(0), got.Attributes)
	})

	t.Run("zero-address GnoAccount decodes with IsZero address", func(t *testing.T) {
		t.Parallel()
		// On gno.land, querying an unknown address returns a zero-valued
		// account rather than an error. queryAccountAtHeight detects this via
		// acc.Address.IsZero() and returns nil. Lock in the wire-format
		// assumption that backs that branch.
		zero := &gnoland.GnoAccount{}
		zwire := makeGnoAccountWire(t, zero)

		var got gnoland.GnoAccount
		require.NoError(t, amino.UnmarshalJSON(zwire, &got))
		assert.True(t, got.Address.IsZero(),
			"zero-valued GnoAccount must round-trip to an IsZero address")
	})
}
