package fork

import (
	"testing"

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
