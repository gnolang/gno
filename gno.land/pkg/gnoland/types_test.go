package gnoland

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTxs generates dummy transactions
func generateTxs(t *testing.T, count int) []TxWithMetadata {
	t.Helper()

	txs := make([]TxWithMetadata, count)

	for i := range count {
		txs[i] = TxWithMetadata{
			Tx: std.Tx{
				Msgs: []std.Msg{
					bank.MsgSend{
						FromAddress: crypto.Address{byte(i)},
						ToAddress:   crypto.Address{byte(i)},
						Amount:      std.NewCoins(std.NewCoin(ugnot.Denom, 1)),
					},
				},
				Fee: std.Fee{
					GasWanted: 10,
					GasFee:    std.NewCoin(ugnot.Denom, 1000000),
				},
				Memo: fmt.Sprintf("tx %d", i),
			},
		}
	}

	return txs
}

func TestReadGenesisTxs(t *testing.T) {
	t.Parallel()

	createFile := func(path, data string) {
		file, err := os.Create(path)
		require.NoError(t, err)

		_, err = file.WriteString(data)
		require.NoError(t, err)
	}

	t.Run("invalid path", func(t *testing.T) {
		t.Parallel()

		path := "" // invalid

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		txs, err := ReadGenesisTxs(ctx, path)
		assert.Nil(t, txs)

		assert.Error(t, err)
	})

	t.Run("invalid tx format", func(t *testing.T) {
		t.Parallel()

		var (
			dir  = t.TempDir()
			path = filepath.Join(dir, "txs.jsonl")
		)

		// Create the file
		createFile(
			path,
			"random data",
		)

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		txs, err := ReadGenesisTxs(ctx, path)
		assert.Nil(t, txs)

		assert.Error(t, err)
	})

	t.Run("valid txs", func(t *testing.T) {
		t.Parallel()

		var (
			dir  = t.TempDir()
			path = filepath.Join(dir, "txs.jsonl")
			txs  = generateTxs(t, 1000)
		)

		// Create the file
		file, err := os.Create(path)
		require.NoError(t, err)

		// Write the transactions
		for _, tx := range txs {
			encodedTx, err := amino.MarshalJSON(tx)
			require.NoError(t, err)

			_, err = fmt.Fprintf(file, "%s\n", encodedTx)
			require.NoError(t, err)
		}

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Load the transactions
		readTxs, err := ReadGenesisTxs(ctx, path)
		require.NoError(t, err)

		require.Len(t, readTxs, len(txs))

		for index, readTx := range readTxs {
			assert.Equal(t, txs[index], readTx)
		}
	})
}

func TestGnoAccountRestriction(t *testing.T) {
	testEnv := setupTestEnv()
	ctx, acck, bankk := testEnv.ctx, testEnv.acck, testEnv.bankk

	fromAddress := crypto.AddressFromPreimage([]byte("from"))
	toAddress := crypto.AddressFromPreimage([]byte("to"))
	fromAccount := acck.NewAccountWithAddress(ctx, fromAddress)
	toAccount := acck.NewAccountWithAddress(ctx, toAddress)

	// Default account is not unrestricted
	assert.False(t, fromAccount.(*GnoAccount).IsTokenLockWhitelisted())

	// Send Unrestricted
	fromAccount.SetCoins(std.NewCoins(std.NewCoin("foocoin", 10)))
	acck.SetAccount(ctx, fromAccount)
	acck.SetAccount(ctx, toAccount)

	err := bankk.SendCoins(ctx, fromAddress, toAddress, std.NewCoins(std.NewCoin("foocoin", 3)))
	require.NoError(t, err)
	balance := acck.GetAccount(ctx, toAddress).GetCoins()
	assert.Equal(t, balance.String(), "3foocoin")

	// Send Restricted
	bankk.SetRestrictedDenoms(ctx, []string{"foocoin"})
	err = bankk.SendCoins(ctx, fromAddress, toAddress, std.NewCoins(std.NewCoin("foocoin", 3)))
	require.Error(t, err)
	assert.Equal(t, "restricted token transfer error", err.Error())

	// Set unrestrict Account
	fromAccount.(*GnoAccount).SetTokenLockWhitelisted(true)
	assert.True(t, fromAccount.(*GnoAccount).IsTokenLockWhitelisted())

	// Persisted unrestricted state
	acck.SetAccount(ctx, fromAccount)
	fromAccount = acck.GetAccount(ctx, fromAddress)
	assert.True(t, fromAccount.(*GnoAccount).IsTokenLockWhitelisted())

	// Send Restricted
	bankk.SetRestrictedDenoms(ctx, []string{"foocoin"}) // XXX unnecessary?
	err = bankk.SendCoins(ctx, fromAddress, toAddress, std.NewCoins(std.NewCoin("foocoin", 3)))
	require.NoError(t, err)
	assert.Equal(t, balance.String(), "3foocoin")
}

func TestGnoAccountSendRestrictions(t *testing.T) {
	testEnv := setupTestEnv()
	ctx, acck, bankk := testEnv.ctx, testEnv.acck, testEnv.bankk

	bankk.SetRestrictedDenoms(ctx, []string{"foocoin"})
	addr := crypto.AddressFromPreimage([]byte("addr1"))
	addr2 := crypto.AddressFromPreimage([]byte("addr2"))
	acc := acck.NewAccountWithAddress(ctx, addr)

	// All accounts are restricted by default when the transfer restriction is applied.

	// Test GetCoins/SetCoins
	acck.SetAccount(ctx, acc)
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins()))

	bankk.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10)))
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 10))))

	// Test HasCoins
	require.True(t, bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 10))))
	require.True(t, bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 5))))
	require.False(t, bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15))))
	require.False(t, bankk.HasCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 5))))

	bankk.SetCoins(ctx, addr, std.NewCoins(std.NewCoin("foocoin", 15)))

	// Test sending coins restricted to locked accounts.
	err := bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("foocoin", 5)))
	require.ErrorIs(t, err, std.RestrictedTransferError{}, "expected restricted transfer error, got %v", err)
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("foocoin", 15))))
	require.True(t, bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("foocoin", 0))))

	// Test sending coins unrestricted to locked accounts.
	bankk.AddCoins(ctx, addr, std.NewCoins(std.NewCoin("barcoin", 30)))
	err = bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("barcoin", 10)))
	require.NoError(t, err)
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 20), std.NewCoin("foocoin", 15))))
	require.True(t, bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10))))

	// Remove the restrictions
	bankk.SetRestrictedDenoms(ctx, []string{})
	// Test sending coins restricted to locked accounts.
	err = bankk.SendCoins(ctx, addr, addr2, std.NewCoins(std.NewCoin("foocoin", 5)))
	require.NoError(t, err)
	require.True(t, bankk.GetCoins(ctx, addr).IsEqual(std.NewCoins(std.NewCoin("barcoin", 20), std.NewCoin("foocoin", 10))))
	require.True(t, bankk.GetCoins(ctx, addr2).IsEqual(std.NewCoins(std.NewCoin("barcoin", 10), std.NewCoin("foocoin", 5))))
}

func TestSignGenesisTx(t *testing.T) {
	t.Parallel()

	var (
		txs     = generateTxs(t, 100)
		privKey = secp256k1.GenPrivKey()
		pubKey  = privKey.PubKey()
		chainID = "testing"
	)

	// Make sure the transactions are properly signed
	require.NoError(t, SignGenesisTxs(txs, privKey, chainID))

	// Make sure the signatures are valid
	for _, tx := range txs {
		payload, err := tx.Tx.GetSignBytes(chainID, 0, 0)
		require.NoError(t, err)

		sigs := tx.Tx.GetSignatures()
		require.Len(t, sigs, 1)

		assert.True(t, pubKey.Equals(sigs[0].PubKey))
		assert.True(t, pubKey.VerifyBytes(payload, sigs[0].Signature))
	}
}

func TestSetFlag(t *testing.T) {
	account := &GnoAccount{}

	// Test setting a valid flag
	account.setFlag(flagTokenLockWhitelisted)
	assert.True(t, account.hasFlag(flagTokenLockWhitelisted), "Expected flagTokenLockWhitelisted to be set")

	// Test setting an invalid flag
	assert.Panics(t, func() {
		account.setFlag(BitSet(0x1000)) // Invalid flag
	}, "Expected panic for invalid flag")
}

func TestClearFlag(t *testing.T) {
	account := &GnoAccount{}

	// Set and then clear the flag
	account.setFlag(flagTokenLockWhitelisted)
	assert.True(t, account.hasFlag(flagTokenLockWhitelisted), "Expected flagTokenLockWhitelisted to be set before clearing")

	account.clearFlag(flagTokenLockWhitelisted)
	assert.False(t, account.hasFlag(flagTokenLockWhitelisted), "Expected flagTokenLockWhitelisted to be cleared")
}
