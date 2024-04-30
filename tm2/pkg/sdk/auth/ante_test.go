package auth

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// run the tx through the anteHandler and ensure its valid
func checkValidTx(t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx std.Tx, simulate bool) {
	t.Helper()

	_, result, abort := anteHandler(ctx, tx, simulate)
	require.Equal(t, "", result.Log)
	require.False(t, abort)
	require.Nil(t, result.Error)
	require.True(t, result.IsOK())
}

// run the tx through the anteHandler and ensure it fails with the given code
func checkInvalidTx(t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx std.Tx, simulate bool, err abci.Error) {
	t.Helper()

	newCtx, result, abort := anteHandler(ctx, tx, simulate)
	require.True(t, abort)

	require.Equal(t, reflect.TypeOf(err), reflect.TypeOf(sdk.ABCIError(result.Error)), fmt.Sprintf("Expected %v, got %v", err, result))

	if reflect.TypeOf(err) == reflect.TypeOf(std.OutOfGasError{}) {
		// GasWanted set correctly
		require.Equal(t, tx.Fee.GasWanted, result.GasWanted, "Gas wanted not set correctly")
		require.True(t, result.GasUsed > result.GasWanted, "GasUsed not greated than GasWanted")
		// Check that context is set correctly
		require.Equal(t, result.GasUsed, newCtx.GasMeter().GasConsumed(), "Context not updated correctly")
	}
}

func defaultAnteOptions() AnteOptions {
	return AnteOptions{
		VerifyGenesisSignatures: true,
	}
}

// Test various error cases in the AnteHandler control flow.
func TestAnteHandlerSigErrors(t *testing.T) {
	t.Parallel()

	// setup
	env := setupTestEnv()
	ctx := env.ctx
	anteHandler := NewAnteHandler(env.acck, env.bank, DefaultSigVerificationGasConsumer, defaultAnteOptions())

	// keys and addresses
	priv1, _, addr1 := tu.KeyTestPubAddr()
	priv2, _, addr2 := tu.KeyTestPubAddr()
	priv3, _, addr3 := tu.KeyTestPubAddr()

	// msg and signatures
	var tx std.Tx
	msg1 := tu.NewTestMsg(addr1, addr2)
	msg2 := tu.NewTestMsg(addr1, addr3)
	fee := tu.NewTestFee()

	msgs := []std.Msg{msg1, msg2}

	// test no signatures
	privs, accNums, seqs := []crypto.PrivKey{}, []uint64{}, []uint64{}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accNums, seqs, fee)

	// tx.GetSigners returns addresses in correct order: addr1, addr2, addr3
	expectedSigners := []crypto.Address{addr1, addr2, addr3}
	require.Equal(t, expectedSigners, tx.GetSigners())

	// Check no signatures fails
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.NoSignaturesError{})

	// test num sigs dont match GetSigners
	privs, accNums, seqs = []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accNums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})

	// test an unrecognized account
	privs, accNums, seqs = []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accNums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnknownAddressError{})

	// save the first account, but second is still unrecognized
	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(std.Coins{fee.GasFee})
	env.acck.SetAccount(ctx, acc1)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnknownAddressError{})
}

// Test logic around account number checking with one signer and many signers.
func TestAnteHandlerAccountNumbers(t *testing.T) {
	t.Parallel()

	// setup
	env := setupTestEnv()
	anteHandler := NewAnteHandler(env.acck, env.bank, DefaultSigVerificationGasConsumer, defaultAnteOptions())
	ctx := env.ctx

	// keys and addresses
	priv1, _, addr1 := tu.KeyTestPubAddr()
	priv2, _, addr2 := tu.KeyTestPubAddr()

	// set the accounts
	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	env.acck.SetAccount(ctx, acc1)
	acc2 := env.acck.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	env.acck.SetAccount(ctx, acc2)

	// msg and signatures
	var tx std.Tx
	msg := tu.NewTestMsg(addr1)
	fee := tu.NewTestFee()

	msgs := []std.Msg{msg}

	// test good tx from one signer
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx from wrong account number
	seqs = []uint64{1}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, []uint64{1}, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})

	// from correct account number
	seqs = []uint64{1}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, []uint64{0}, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx with another signer and incorrect account numbers
	msg1 := tu.NewTestMsg(addr1, addr2)
	msg2 := tu.NewTestMsg(addr2, addr1)
	msgs = []std.Msg{msg1, msg2}
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{1, 0}, []uint64{2, 0}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})

	// correct account numbers
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{2, 0}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}

// Test logic around account number checking with many signers when BlockHeight is 0.
func TestAnteHandlerAccountNumbersAtBlockHeightZero(t *testing.T) {
	t.Parallel()

	// setup
	env := setupTestEnv()
	anteHandler := NewAnteHandler(env.acck, env.bank, DefaultSigVerificationGasConsumer, defaultAnteOptions())
	ctx := env.ctx
	header := ctx.BlockHeader().(*bft.Header)
	header.Height = 0
	ctx = ctx.WithBlockHeader(header)

	// keys and addresses
	priv1, _, addr1 := tu.KeyTestPubAddr()
	priv2, _, addr2 := tu.KeyTestPubAddr()

	// set the accounts, we don't need the acc numbers as it is in the genesis block
	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(tu.NewTestCoins())
	env.acck.SetAccount(ctx, acc1)
	acc2 := env.acck.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	env.acck.SetAccount(ctx, acc2)

	// msg and signatures
	var tx std.Tx
	msg := tu.NewTestMsg(addr1)
	fee := tu.NewTestFee()

	msgs := []std.Msg{msg}

	// test good tx from one signer
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx from wrong account number
	seqs = []uint64{1}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, []uint64{1}, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})

	// from correct account number
	seqs = []uint64{1}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, []uint64{0}, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx with another signer and incorrect account numbers
	msg1 := tu.NewTestMsg(addr1, addr2)
	msg2 := tu.NewTestMsg(addr2, addr1)
	msgs = []std.Msg{msg1, msg2}
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{1, 0}, []uint64{2, 0}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})

	// correct account numbers
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{0, 0}, []uint64{2, 0}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}

// Test logic around sequence checking with one signer and many signers.
func TestAnteHandlerSequences(t *testing.T) {
	t.Parallel()

	// setup
	env := setupTestEnv()
	anteHandler := NewAnteHandler(env.acck, env.bank, DefaultSigVerificationGasConsumer, defaultAnteOptions())
	ctx := env.ctx

	// keys and addresses
	priv1, _, addr1 := tu.KeyTestPubAddr()
	priv2, _, addr2 := tu.KeyTestPubAddr()
	priv3, _, addr3 := tu.KeyTestPubAddr()

	// set the accounts
	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	env.acck.SetAccount(ctx, acc1)
	acc2 := env.acck.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	env.acck.SetAccount(ctx, acc2)
	acc3 := env.acck.NewAccountWithAddress(ctx, addr3)
	acc3.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc3.SetAccountNumber(2))
	env.acck.SetAccount(ctx, acc3)

	// msg and signatures
	var tx std.Tx
	msg := tu.NewTestMsg(addr1)
	fee := tu.NewTestFee()

	msgs := []std.Msg{msg}

	// test good tx from one signer
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// test sending it again fails (replay protection)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})

	// fix sequence, should pass
	seqs = []uint64{1}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// new tx with another signer and correct sequences
	msg1 := tu.NewTestMsg(addr1, addr2)
	msg2 := tu.NewTestMsg(addr3, addr1)
	msgs = []std.Msg{msg1, msg2}

	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{2, 0, 0}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// replay fails
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})

	// tx from just second signer with incorrect sequence fails
	msg = tu.NewTestMsg(addr2)
	msgs = []std.Msg{msg}
	privs, accnums, seqs = []crypto.PrivKey{priv2}, []uint64{1}, []uint64{0}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})

	// fix the sequence and it passes
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, []crypto.PrivKey{priv2}, []uint64{1}, []uint64{1}, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// another tx from both of them that passes
	msg = tu.NewTestMsg(addr1, addr2)
	msgs = []std.Msg{msg}
	privs, accnums, seqs = []crypto.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{3, 2}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}

// Test logic around fee deduction.
func TestAnteHandlerFees(t *testing.T) {
	t.Parallel()

	// setup
	env := setupTestEnv()
	ctx := env.ctx
	anteHandler := NewAnteHandler(env.acck, env.bank, DefaultSigVerificationGasConsumer, defaultAnteOptions())

	// keys and addresses
	priv1, _, addr1 := tu.KeyTestPubAddr()

	// set the accounts
	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	env.acck.SetAccount(ctx, acc1)

	// msg and signatures
	var tx std.Tx
	msg := tu.NewTestMsg(addr1)
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := tu.NewTestFee()
	msgs := []std.Msg{msg}

	// signer does not have enough funds to pay the fee
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.InsufficientFundsError{})

	acc1.SetCoins(std.NewCoins(std.NewCoin("atom", 149)))
	env.acck.SetAccount(ctx, acc1)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.InsufficientFundsError{})

	collector := env.bank.(DummyBankKeeper).acck.GetAccount(ctx, FeeCollectorAddress())
	require.Nil(t, collector)
	require.Equal(t, env.acck.GetAccount(ctx, addr1).GetCoins().AmountOf("atom"), int64(149))

	acc1.SetCoins(std.NewCoins(std.NewCoin("atom", 150)))
	env.acck.SetAccount(ctx, acc1)
	checkValidTx(t, anteHandler, ctx, tx, false)

	require.Equal(t, env.bank.(DummyBankKeeper).acck.GetAccount(ctx, FeeCollectorAddress()).GetCoins().AmountOf("atom"), int64(150))
	require.Equal(t, env.acck.GetAccount(ctx, addr1).GetCoins().AmountOf("atom"), int64(0))
}

// Test logic around memo gas consumption.
func TestAnteHandlerMemoGas(t *testing.T) {
	t.Parallel()

	// setup
	env := setupTestEnv()
	anteHandler := NewAnteHandler(env.acck, env.bank, DefaultSigVerificationGasConsumer, defaultAnteOptions())
	ctx := env.ctx

	// keys and addresses
	priv1, _, addr1 := tu.KeyTestPubAddr()

	// set the accounts
	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	require.NoError(t, acc1.SetAccountNumber(0))
	env.acck.SetAccount(ctx, acc1)

	// msg and signatures
	var tx std.Tx
	msg := tu.NewTestMsg(addr1)
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := std.NewFee(0, std.NewCoin("atom", 0))

	// tx does not have enough gas
	tx = tu.NewTestTx(t, ctx.ChainID(), []std.Msg{msg}, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.OutOfGasError{})

	// tx with memo doesn't have enough gas
	fee = std.NewFee(801, std.NewCoin("atom", 0))
	tx = tu.NewTestTxWithMemo(t, ctx.ChainID(), []std.Msg{msg}, privs, accnums, seqs, fee, "abcininasidniandsinasindiansdiansdinaisndiasndiadninsd")
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.OutOfGasError{})

	// memo too large
	fee = std.NewFee(9000, std.NewCoin("atom", 0))
	tx = tu.NewTestTxWithMemo(t, ctx.ChainID(), []std.Msg{msg}, privs, accnums, seqs, fee, strings.Repeat("01234567890", 99000))
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.MemoTooLargeError{})

	// tx with memo has enough gas
	fee = std.NewFee(9000, std.NewCoin("atom", 0))
	tx = tu.NewTestTxWithMemo(t, ctx.ChainID(), []std.Msg{msg}, privs, accnums, seqs, fee, strings.Repeat("0123456789", 10))
	checkValidTx(t, anteHandler, ctx, tx, false)
}

func TestAnteHandlerMultiSigner(t *testing.T) {
	t.Parallel()

	// setup
	env := setupTestEnv()
	anteHandler := NewAnteHandler(env.acck, env.bank, DefaultSigVerificationGasConsumer, defaultAnteOptions())
	ctx := env.ctx

	// keys and addresses
	priv1, _, addr1 := tu.KeyTestPubAddr()
	priv2, _, addr2 := tu.KeyTestPubAddr()
	priv3, _, addr3 := tu.KeyTestPubAddr()

	// set the accounts
	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	env.acck.SetAccount(ctx, acc1)
	acc2 := env.acck.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	env.acck.SetAccount(ctx, acc2)
	acc3 := env.acck.NewAccountWithAddress(ctx, addr3)
	acc3.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc3.SetAccountNumber(2))
	env.acck.SetAccount(ctx, acc3)

	// set up msgs and fee
	var tx std.Tx
	msg1 := tu.NewTestMsg(addr1, addr2)
	msg2 := tu.NewTestMsg(addr3, addr1)
	msg3 := tu.NewTestMsg(addr2, addr3)
	msgs := []std.Msg{msg1, msg2, msg3}
	fee := tu.NewTestFee()

	// signers in order
	privs, accnums, seqs := []crypto.PrivKey{priv1, priv2, priv3}, []uint64{0, 1, 2}, []uint64{0, 0, 0}
	tx = tu.NewTestTxWithMemo(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee, "Check signers are in expected order and different account numbers works")

	checkValidTx(t, anteHandler, ctx, tx, false)

	// change sequence numbers
	tx = tu.NewTestTx(t, ctx.ChainID(), []std.Msg{msg1}, []crypto.PrivKey{priv1, priv2}, []uint64{0, 1}, []uint64{1, 1}, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
	tx = tu.NewTestTx(t, ctx.ChainID(), []std.Msg{msg2}, []crypto.PrivKey{priv3, priv1}, []uint64{2, 0}, []uint64{1, 2}, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	// expected seqs = [3, 2, 2]
	tx = tu.NewTestTxWithMemo(t, ctx.ChainID(), msgs, privs, accnums, []uint64{3, 2, 2}, fee, "Check signers are in expected order and different account numbers and sequence numbers works")
	checkValidTx(t, anteHandler, ctx, tx, false)
}

func TestAnteHandlerBadSignBytes(t *testing.T) {
	t.Parallel()

	// setup
	env := setupTestEnv()
	anteHandler := NewAnteHandler(env.acck, env.bank, DefaultSigVerificationGasConsumer, defaultAnteOptions())
	ctx := env.ctx

	// keys and addresses
	priv1, _, addr1 := tu.KeyTestPubAddr()
	priv2, _, addr2 := tu.KeyTestPubAddr()

	// set the accounts
	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	env.acck.SetAccount(ctx, acc1)
	acc2 := env.acck.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	env.acck.SetAccount(ctx, acc2)

	var tx std.Tx
	msg := tu.NewTestMsg(addr1)
	msgs := []std.Msg{msg}
	fee := tu.NewTestFee()
	fee2 := tu.NewTestFee()
	fee2.GasWanted += 100
	fee3 := tu.NewTestFee()
	fee3.GasFee.Amount += 100

	// test good tx and signBytes
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	chainID := ctx.ChainID()
	chainID2 := chainID + "somemorestuff"
	unauthErr := std.UnauthorizedError{}

	cases := []struct {
		chainID string
		accnum  uint64
		seq     uint64
		fee     std.Fee
		msgs    []std.Msg
		err     abci.Error
	}{
		{chainID2, 0, 1, fee, msgs, unauthErr},                           // test wrong chain_id
		{chainID, 0, 2, fee, msgs, unauthErr},                            // test wrong seqs
		{chainID, 1, 1, fee, msgs, unauthErr},                            // test wrong accnum
		{chainID, 0, 1, fee, []std.Msg{tu.NewTestMsg(addr2)}, unauthErr}, // test wrong msg
		{chainID, 0, 1, fee2, msgs, unauthErr},                           // test wrong fee
		{chainID, 0, 1, fee3, msgs, unauthErr},                           // test wrong fee
	}

	privs, seqs = []crypto.PrivKey{priv1}, []uint64{1}
	for _, cs := range cases {
		signPayload, err := std.GetSignaturePayload(std.SignDoc{
			ChainID:       cs.chainID,
			AccountNumber: cs.accnum,
			Sequence:      cs.seq,
			Fee:           cs.fee,
			Msgs:          cs.msgs,
		})
		require.NoError(t, err)

		tx := tu.NewTestTxWithSignBytes(
			msgs, privs, fee,
			signPayload,
			"",
		)
		checkInvalidTx(t, anteHandler, ctx, tx, false, cs.err)
	}

	// test wrong signer if public key exist
	privs, accnums, seqs = []crypto.PrivKey{priv2}, []uint64{0}, []uint64{1}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.UnauthorizedError{})

	// test wrong signer if public doesn't exist
	msg = tu.NewTestMsg(addr2)
	msgs = []std.Msg{msg}
	privs, accnums, seqs = []crypto.PrivKey{priv1}, []uint64{1}, []uint64{0}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.InvalidPubKeyError{})
}

func TestAnteHandlerSetPubKey(t *testing.T) {
	t.Parallel()

	// setup
	env := setupTestEnv()
	anteHandler := NewAnteHandler(env.acck, env.bank, DefaultSigVerificationGasConsumer, defaultAnteOptions())
	ctx := env.ctx

	// keys and addresses
	priv1, _, addr1 := tu.KeyTestPubAddr()
	_, _, addr2 := tu.KeyTestPubAddr()

	// set the accounts
	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc1.SetAccountNumber(0))
	env.acck.SetAccount(ctx, acc1)
	acc2 := env.acck.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	env.acck.SetAccount(ctx, acc2)

	var tx std.Tx

	// test good tx and set public key
	msg := tu.NewTestMsg(addr1)
	msgs := []std.Msg{msg}
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := tu.NewTestFee()
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)

	acc1 = env.acck.GetAccount(ctx, addr1)
	require.Equal(t, acc1.GetPubKey(), priv1.PubKey())

	// test public key not found
	msg = tu.NewTestMsg(addr2)
	msgs = []std.Msg{msg}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, []uint64{1}, seqs, fee)
	sigs := tx.GetSignatures()
	sigs[0].PubKey = nil
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.InvalidPubKeyError{})

	acc2 = env.acck.GetAccount(ctx, addr2)
	require.Nil(t, acc2.GetPubKey())

	// test invalid signature and public key
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, []uint64{1}, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.InvalidPubKeyError{})

	acc2 = env.acck.GetAccount(ctx, addr2)
	require.Nil(t, acc2.GetPubKey())
}

func TestProcessPubKey(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx

	// keys
	_, _, addr1 := tu.KeyTestPubAddr()
	priv2, _, addr2 := tu.KeyTestPubAddr()
	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	acc2 := env.acck.NewAccountWithAddress(ctx, addr2)

	acc2.SetPubKey(priv2.PubKey())

	type args struct {
		acc      std.Account
		sig      std.Signature
		simulate bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"no sigs, simulate off", args{acc1, std.Signature{}, false}, true},
		{"no sigs, simulate on", args{acc1, std.Signature{}, true}, false},
		{"no sigs, account with pub, simulate on", args{acc2, std.Signature{}, true}, false},
		{"pubkey doesn't match addr, simulate off", args{acc1, std.Signature{PubKey: priv2.PubKey()}, false}, true},
		{"pubkey doesn't match addr, simulate on", args{acc1, std.Signature{PubKey: priv2.PubKey()}, true}, false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ProcessPubKey(tt.args.acc, tt.args.sig, tt.args.simulate)
			require.Equal(t, tt.wantErr, !err.IsOK())
		})
	}
}

func TestConsumeSignatureVerificationGas(t *testing.T) {
	t.Parallel()

	params := DefaultParams()
	msg := []byte{1, 2, 3, 4}

	pkSet1, sigSet1 := generatePubKeysAndSignatures(5, msg, false)
	multisigKey1 := multisig.NewPubKeyMultisigThreshold(2, pkSet1)
	multisignature1 := multisig.NewMultisig(len(pkSet1))
	expectedCost1 := expectedGasCostByKeys(pkSet1)
	for i := 0; i < len(pkSet1); i++ {
		multisignature1.AddSignatureFromPubKey(sigSet1[i], pkSet1[i], pkSet1)
	}

	type args struct {
		meter  store.GasMeter
		sig    []byte
		pubkey crypto.PubKey
		params Params
	}
	tests := []struct {
		name        string
		args        args
		gasConsumed int64
		shouldErr   bool
	}{
		{"PubKeyEd25519", args{store.NewInfiniteGasMeter(), nil, ed25519.GenPrivKey().PubKey(), params}, DefaultSigVerifyCostED25519, true},
		{"PubKeySecp256k1", args{store.NewInfiniteGasMeter(), nil, secp256k1.GenPrivKey().PubKey(), params}, DefaultSigVerifyCostSecp256k1, false},
		{"Multisig", args{store.NewInfiniteGasMeter(), amino.MustMarshal(multisignature1), multisigKey1, params}, expectedCost1, false},
		{"unknown key", args{store.NewInfiniteGasMeter(), nil, nil, params}, 0, true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res := DefaultSigVerificationGasConsumer(tt.args.meter, tt.args.sig, tt.args.pubkey, tt.args.params)

			if tt.shouldErr {
				require.False(t, res.IsOK())
			} else {
				require.True(t, res.IsOK())
				require.Equal(t, tt.gasConsumed, tt.args.meter.GasConsumed(), fmt.Sprintf("%d != %d", tt.gasConsumed, tt.args.meter.GasConsumed()))
			}
		})
	}
}

func generatePubKeysAndSignatures(n int, msg []byte, keyTypeed25519 bool) (pubkeys []crypto.PubKey, signatures [][]byte) {
	pubkeys = make([]crypto.PubKey, n)
	signatures = make([][]byte, n)
	for i := 0; i < n; i++ {
		var privkey crypto.PrivKey
		if rand.Int63()%2 == 0 {
			privkey = ed25519.GenPrivKey()
		} else {
			privkey = secp256k1.GenPrivKey()
		}
		pubkeys[i] = privkey.PubKey()
		signatures[i], _ = privkey.Sign(msg)
	}
	return
}

func expectedGasCostByKeys(pubkeys []crypto.PubKey) int64 {
	cost := int64(0)
	for _, pubkey := range pubkeys {
		pubkeyType := strings.ToLower(fmt.Sprintf("%T", pubkey))
		switch {
		case strings.Contains(pubkeyType, "ed25519"):
			cost += DefaultParams().SigVerifyCostED25519
		case strings.Contains(pubkeyType, "secp256k1"):
			cost += DefaultParams().SigVerifyCostSecp256k1
		default:
			panic("unexpected key type")
		}
	}
	return cost
}

func TestCountSubkeys(t *testing.T) {
	t.Parallel()

	genPubKeys := func(n int) []crypto.PubKey {
		var ret []crypto.PubKey
		for i := 0; i < n; i++ {
			ret = append(ret, secp256k1.GenPrivKey().PubKey())
		}
		return ret
	}
	singleKey := secp256k1.GenPrivKey().PubKey()
	singleLevelMultiKey := multisig.NewPubKeyMultisigThreshold(4, genPubKeys(5))
	multiLevelSubKey1 := multisig.NewPubKeyMultisigThreshold(4, genPubKeys(5))
	multiLevelSubKey2 := multisig.NewPubKeyMultisigThreshold(4, genPubKeys(5))
	multiLevelMultiKey := multisig.NewPubKeyMultisigThreshold(2, []crypto.PubKey{
		multiLevelSubKey1, multiLevelSubKey2, secp256k1.GenPrivKey().PubKey(),
	})
	type args struct {
		pub crypto.PubKey
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"single key", args{singleKey}, 1},
		{"single level multikey", args{singleLevelMultiKey}, 5},
		{"multi level multikey", args{multiLevelMultiKey}, 11},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, std.CountSubKeys(tt.args.pub))
		})
	}
}

func TestAnteHandlerSigLimitExceeded(t *testing.T) {
	t.Parallel()

	// setup
	env := setupTestEnv()
	anteHandler := NewAnteHandler(env.acck, env.bank, DefaultSigVerificationGasConsumer, defaultAnteOptions())
	ctx := env.ctx

	// keys and addresses
	priv1, _, addr1 := tu.KeyTestPubAddr()
	priv2, _, addr2 := tu.KeyTestPubAddr()
	priv3, _, addr3 := tu.KeyTestPubAddr()
	priv4, _, addr4 := tu.KeyTestPubAddr()
	priv5, _, addr5 := tu.KeyTestPubAddr()
	priv6, _, addr6 := tu.KeyTestPubAddr()
	priv7, _, addr7 := tu.KeyTestPubAddr()
	priv8, _, addr8 := tu.KeyTestPubAddr()

	// set the accounts
	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	acc1.SetCoins(tu.NewTestCoins())
	env.acck.SetAccount(ctx, acc1)
	acc2 := env.acck.NewAccountWithAddress(ctx, addr2)
	acc2.SetCoins(tu.NewTestCoins())
	require.NoError(t, acc2.SetAccountNumber(1))
	env.acck.SetAccount(ctx, acc2)

	var tx std.Tx
	msg := tu.NewTestMsg(addr1, addr2, addr3, addr4, addr5, addr6, addr7, addr8)
	msgs := []std.Msg{msg}
	fee := tu.NewTestFee()

	// test rejection logic
	privs, accnums, seqs := []crypto.PrivKey{priv1, priv2, priv3, priv4, priv5, priv6, priv7, priv8},
		[]uint64{0, 0, 0, 0, 0, 0, 0, 0}, []uint64{0, 0, 0, 0, 0, 0, 0, 0}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.TooManySignaturesError{})
}

func TestEnsureSufficientMempoolFees(t *testing.T) {
	t.Parallel()

	// setup
	env := setupTestEnv()
	ctx := env.ctx.WithMinGasPrices(
		[]std.GasPrice{
			{Gas: 100000, Price: std.Coin{Denom: "photino", Amount: 5}},
			{Gas: 100000, Price: std.Coin{Denom: "stake", Amount: 1}},
		},
	)

	testCases := []struct {
		input      std.Fee
		expectedOK bool
	}{
		{std.NewFee(200000, std.Coin{}), false},
		{std.NewFee(200000, std.NewCoin("photino", 5)), false},
		{std.NewFee(200000, std.NewCoin("stake", 1)), false},
		{std.NewFee(200000, std.NewCoin("stake", 2)), true},
		{std.NewFee(200000, std.NewCoin("photino", 10)), true},
		{std.NewFee(200000, std.NewCoin("stake", 2)), true},
		{std.NewFee(200000, std.NewCoin("atom", 5)), false},
	}

	for i, tc := range testCases {
		res := EnsureSufficientMempoolFees(ctx, tc.input)
		require.Equal(
			t, tc.expectedOK, res.IsOK(),
			"unexpected result; tc #%d, input: %v, log: %v", i, tc.input, res.Log,
		)
	}
}

// Test custom SignatureVerificationGasConsumer
func TestCustomSignatureVerificationGasConsumer(t *testing.T) {
	t.Parallel()

	// setup
	env := setupTestEnv()
	// setup an ante handler that only accepts PubKeyEd25519
	anteHandler := NewAnteHandler(env.acck, env.bank, func(meter store.GasMeter, sig []byte, pubkey crypto.PubKey, params Params) sdk.Result {
		switch pubkey := pubkey.(type) {
		case ed25519.PubKeyEd25519:
			meter.ConsumeGas(params.SigVerifyCostED25519, "ante verify: ed25519")
			return sdk.Result{}
		default:
			return abciResult(std.ErrInvalidPubKey(fmt.Sprintf("unrecognized public key type: %T", pubkey)))
		}
	}, defaultAnteOptions())
	ctx := env.ctx

	// verify that an secp256k1 account gets rejected
	priv1, _, addr1 := tu.KeyTestPubAddr()
	acc1 := env.acck.NewAccountWithAddress(ctx, addr1)
	_ = acc1.SetCoins(std.NewCoins(std.NewCoin("atom", 150)))
	env.acck.SetAccount(ctx, acc1)

	var tx std.Tx
	msg := tu.NewTestMsg(addr1)
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := tu.NewTestFee()
	msgs := []std.Msg{msg}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, std.InvalidPubKeyError{})

	// verify that an ed25519 account gets accepted
	priv2 := ed25519.GenPrivKey()
	pub2 := priv2.PubKey()
	addr2 := pub2.Address()
	acc2 := env.acck.NewAccountWithAddress(ctx, addr2)
	require.NoError(t, acc2.SetCoins(std.NewCoins(std.NewCoin("atom", 150))))
	require.NoError(t, acc2.SetAccountNumber(1))
	env.acck.SetAccount(ctx, acc2)
	msg = tu.NewTestMsg(addr2)
	privs, accnums, seqs = []crypto.PrivKey{priv2}, []uint64{1}, []uint64{0}
	fee = tu.NewTestFee()
	msgs = []std.Msg{msg}
	tx = tu.NewTestTx(t, ctx.ChainID(), msgs, privs, accnums, seqs, fee)
	checkValidTx(t, anteHandler, ctx, tx, false)
}
