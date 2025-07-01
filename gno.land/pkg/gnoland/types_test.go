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

			_, err = file.WriteString(fmt.Sprintf("%s\n", encodedTx))
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
	assert.False(t, fromAccount.(*GnoAccount).IsUnrestricted())

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
	fromAccount.(*GnoAccount).SetUnrestricted()
	assert.True(t, fromAccount.(*GnoAccount).IsUnrestricted())

	// Persisted unrestricted state
	acck.SetAccount(ctx, fromAccount)
	fromAccount = acck.GetAccount(ctx, fromAddress)
	assert.True(t, fromAccount.(*GnoAccount).IsUnrestricted())

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

func TestSetAccountFlag(t *testing.T) {
	account := &GnoAccount{}

	// Test setting a valid flag
	account.setFlag(flagUnrestrictedAccount)
	assert.True(t, account.hasFlag(flagUnrestrictedAccount), "Expected unrestricted flag to be set")

	// Test setting an invalid flag
	assert.Panics(t, func() {
		account.setFlag(BitSet(0x1000)) // Invalid flag
	}, "Expected panic for invalid flag")
}

func TestClearAccountFlag(t *testing.T) {
	account := &GnoAccount{}

	// Set and then clear the flag
	account.setFlag(flagUnrestrictedAccount)
	assert.True(t, account.hasFlag(flagUnrestrictedAccount), "Expected unrestricted flag to be set before clearing")

	account.clearFlag(flagUnrestrictedAccount)
	assert.False(t, account.hasFlag(flagUnrestrictedAccount), "Expected unrestricted flag to be cleared")
}

func TestSetSessionFlag(t *testing.T) {
	session := &GnoSession{}

	// Test setting a valid flag
	session.setFlag(flagSessionCanManageSessions)
	assert.True(t, session.hasFlag(flagSessionCanManageSessions), "Expected session manager flag to be set")

	// Test setting an invalid flag
	assert.Panics(t, func() {
		session.setFlag(BitSet(0x1000)) // Invalid flag
	}, "Expected panic for invalid flag")
}

func TestClearSessionFlag(t *testing.T) {
	session := &GnoSession{}

	// Test clearing a valid flag
	session.setFlag(flagSessionCanManageSessions)
	assert.True(t, session.hasFlag(flagSessionCanManageSessions), "Expected session manager flag to be set before clearing")

	session.clearFlag(flagSessionCanManageSessions)
	assert.False(t, session.hasFlag(flagSessionCanManageSessions), "Expected session manager flag to be cleared")
}


func TestSessionValidationOnly(t *testing.T) {
	session := &GnoSession{}

	// Initially false
	assert.False(t, session.IsValidationOnly(), "Expected IsValidationOnly to be false initially")

	// Set flag and test
	session.SetValidationOnly()
	assert.True(t, session.IsValidationOnly(), "Expected IsValidationOnly to be true after setting")
}

func TestSessionCanManagePackages(t *testing.T) {
	session := &GnoSession{}

	// Initially false
	assert.False(t, session.CanManagePackages(), "Expected CanManagePackages to be false initially")

	// Set flag and test
	session.SetCanManagePackages()
	assert.True(t, session.CanManagePackages(), "Expected CanManagePackages to be true after setting")
}

func TestSessionUnlimitedTransferCapacity(t *testing.T) {
	session := &GnoSession{}

	// Initially false
	assert.False(t, session.HasUnlimitedTransferCapacity(), "Expected HasUnlimitedTransferCapacity to be false initially")

	// Set flag and test
	session.SetUnlimitedTransferCapacity()
	assert.True(t, session.HasUnlimitedTransferCapacity(), "Expected HasUnlimitedTransferCapacity to be true after setting")
}

func TestSessionTransferCapacityLogic(t *testing.T) {
	session := &GnoSession{}

	// Test 1: No capacity means no transfers
	amount := std.Coins{std.NewCoin("ugnot", 100)}
	assert.False(t, session.CanTransferAmount(amount), "Should not be able to transfer without capacity or unlimited flag")

	// Test 2: With unlimited flag
	session.SetUnlimitedTransferCapacity()
	assert.True(t, session.CanTransferAmount(amount), "Should be able to transfer unlimited amounts")

	// Test 3: Clear unlimited flag and set specific capacity
	session.clearFlag(flagSessionUnlimitedTransferCapacity)
	capacity := std.Coins{std.NewCoin("ugnot", 1000)}
	session.SetCoinsTransferCapacity(capacity)
	
	smallAmount := std.Coins{std.NewCoin("ugnot", 100)}
	largeAmount := std.Coins{std.NewCoin("ugnot", 2000)}
	
	assert.True(t, session.CanTransferAmount(smallAmount), "Should be able to transfer within capacity")
	assert.False(t, session.CanTransferAmount(largeAmount), "Should not be able to transfer beyond capacity")
}

func TestSessionConsumeTransferCapacity(t *testing.T) {
	session := &GnoSession{}

	// Test 1: Unlimited capacity - should not consume anything
	session.SetUnlimitedTransferCapacity()
	amount := std.Coins{std.NewCoin("ugnot", 100)}
	err := session.ConsumeTransferCapacity(amount)
	assert.NoError(t, err, "Should not error when consuming from unlimited capacity")

	// Test 2: Limited capacity - successful consumption
	session.clearFlag(flagSessionUnlimitedTransferCapacity)
	initialCapacity := std.Coins{std.NewCoin("ugnot", 1000)}
	session.SetCoinsTransferCapacity(initialCapacity)
	
	consumeAmount := std.Coins{std.NewCoin("ugnot", 300)}
	err = session.ConsumeTransferCapacity(consumeAmount)
	assert.NoError(t, err, "Should successfully consume within capacity")
	
	// Check remaining capacity
	remainingCapacity := session.CoinsTransferCapacity
	expectedRemaining := std.Coins{std.NewCoin("ugnot", 700)}
	assert.True(t, remainingCapacity.IsEqual(expectedRemaining), "Remaining capacity should be correct")

	// Test 3: Insufficient capacity
	largeAmount := std.Coins{std.NewCoin("ugnot", 800)}
	err = session.ConsumeTransferCapacity(largeAmount)
	assert.Error(t, err, "Should error when trying to consume more than available capacity")
	assert.Contains(t, err.Error(), "insufficient transfer capacity", "Error should mention insufficient capacity")

	// Test 4: Zero capacity
	session.SetCoinsTransferCapacity(std.Coins{})
	err = session.ConsumeTransferCapacity(std.Coins{std.NewCoin("ugnot", 1)})
	assert.Error(t, err, "Should error when trying to consume from zero capacity")
	assert.Contains(t, err.Error(), "no transfer capacity available", "Error should mention no capacity available")
}

// XXX: test session coin restriction
// XXX: test session validity
