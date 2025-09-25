package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/crypto/multisig"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestServer creates a test RPC server.
func createTestServer(
	t *testing.T,
	handler http.Handler,
) *httptest.Server {
	t.Helper()

	s := httptest.NewServer(handler)
	t.Cleanup(s.Close)

	return s
}

// defaultHTTPHandler generates a default HTTP test handler.
func defaultHTTPHandler(
	t *testing.T,
	method string,
	responseResult any,
) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("content-type"))

		// Parse the message
		var req types.RPCRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))

		// Basic request validation
		require.Equal(t, req.JSONRPC, "2.0")
		require.Equal(t, req.Method, method)

		// Marshal the result data to Amino JSON
		result, err := amino.MarshalJSON(responseResult)
		require.NoError(t, err)

		// Send a response back
		response := types.RPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  result,
		}

		// Marshal the response
		marshalledResponse, err := json.Marshal(response)
		require.NoError(t, err)

		_, err = w.Write(marshalledResponse)
		require.NoError(t, err)
	}
}

func Test_execVerify(t *testing.T) {
	t.Parallel()

	const (
		accountNumber   = uint64(10)
		accountSequence = uint64(2)
		fakeKeyName1    = "verifyApp_Key1"
		encPassword     = ""
		chainID         = "dev"
	)

	prepare := func(t *testing.T) (string, std.Tx, func()) {
		t.Helper()

		// Make new test dir.
		kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
		assert.NotNil(t, kbHome)

		// Add test account to keybase.
		kb, err := keys.NewKeyBaseFromDir(kbHome)
		assert.NoError(t, err)
		info, err := kb.CreateAccount(fakeKeyName1, testMnemonic, "", encPassword, 0, 0)
		assert.NoError(t, err)

		// Prepare the signature.
		signOpts := signOpts{
			chainID:         chainID,
			accountSequence: accountSequence,
			accountNumber:   accountNumber,
		}

		keyOpts := keyOpts{
			keyName:     fakeKeyName1,
			decryptPass: "",
		}

		// Construct msg & tx and marshal.
		msg := bank.MsgSend{
			FromAddress: info.GetAddress(),
			ToAddress:   info.GetAddress(),
			Amount: std.Coins{
				std.Coin{
					Denom:  "ugnot",
					Amount: 10,
				},
			},
		}

		tx := std.Tx{
			Msgs: []std.Msg{msg},
			Fee: std.Fee{
				GasWanted: 10,
				GasFee: std.Coin{
					Amount: 10,
					Denom:  "ugnot",
				},
			},
		}

		sig, err := generateSignature(&tx, kb, signOpts, keyOpts)
		assert.NoError(t, err)

		// Add signature to the transaction.
		tx.Signatures = []std.Signature{*sig}

		return kbHome, tx, kbCleanUp
	}

	t.Run("tx path not specified", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			AccountNumber:   commands.Uint64Flag{V: accountNumber},
			AccountSequence: commands.Uint64Flag{V: accountSequence},
			ChainID:         chainID,
			TxPath:          "", // unset
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test: bad key name", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			AccountNumber:   commands.Uint64Flag{V: accountNumber},
			AccountSequence: commands.Uint64Flag{V: accountSequence},
			ChainID:         chainID,
			TxPath:          txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{"bad-key-name"} // Bad key name.

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test: bad transaction", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)
		// Mutate the raw tx to make it bad.
		rawTx[0] = 0xFF

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			AccountNumber:   commands.Uint64Flag{V: accountNumber},
			AccountSequence: commands.Uint64Flag{V: accountSequence},
			ChainID:         chainID,
			TxPath:          txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test: signature ok", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			AccountNumber:   commands.Uint64Flag{V: accountNumber, Defined: true},
			AccountSequence: commands.Uint64Flag{V: accountSequence, Defined: true},
			ChainID:         chainID,
			TxPath:          txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io)
		assert.NoError(t, err)
	})

	t.Run("test: signature ok with Uint64Flag", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		// Initialize the account number flag.
		flagAccountNumber := &commands.Uint64Flag{}
		flagAccountNumber.Set(strconv.FormatUint(accountNumber, 10))
		require.Equal(t, accountNumber, flagAccountNumber.V)

		// Initialize the account sequence flag.
		flagAccountSequence := &commands.Uint64Flag{}
		flagAccountSequence.Set(strconv.FormatUint(accountSequence, 10))
		require.Equal(t, accountSequence, flagAccountSequence.V)

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			AccountNumber:   *flagAccountNumber,
			AccountSequence: *flagAccountSequence,
			ChainID:         chainID,
			TxPath:          txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io)
		assert.NoError(t, err)
	})

	t.Run("test: missing signature", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Remove any signatures from the tx.
		tx.Signatures = nil

		// Marshal the tx.
		rawTxWithoutSig, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTxWithoutSig, 0o644))

		// No signature in tx and no -signature or -sig-path flag.
		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			AccountNumber:   commands.Uint64Flag{V: accountNumber},
			AccountSequence: commands.Uint64Flag{V: accountSequence},
			ChainID:         chainID,
			TxPath:          txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test: -sig-path flag: no signature", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		// Write std.Tx, not std.Signature.
		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			SigPath:         txFile.Name(),
			AccountNumber:   commands.Uint64Flag{V: accountNumber},
			AccountSequence: commands.Uint64Flag{V: accountSequence},
			ChainID:         chainID,
			TxPath:          txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test: -sig-path flag: ok", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		// Marshal the signature.
		rawSig, err := amino.MarshalJSON(tx.Signatures[0])
		assert.NoError(t, err)

		sigFile, err := os.CreateTemp("", "sig-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(sigFile.Name(), rawSig, 0o644))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			SigPath:         sigFile.Name(),
			AccountNumber:   commands.Uint64Flag{V: accountNumber, Defined: true},
			AccountSequence: commands.Uint64Flag{V: accountSequence, Defined: true},
			ChainID:         chainID,
			TxPath:          txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io)
		assert.NoError(t, err)
	})

	t.Run("test: bad -account-sequence flag", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		// Initialize the account number flag.
		flagAccountNumber := &commands.Uint64Flag{}
		flagAccountNumber.Set(strconv.FormatUint(accountNumber, 10))

		// Initialize the account sequence flag.
		flagAccountSequence := &commands.Uint64Flag{}
		flagAccountSequence.Set(strconv.FormatUint(accountSequence+1, 10)) // Bad sequence.

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			AccountNumber:   *flagAccountNumber,
			AccountSequence: *flagAccountSequence,
			ChainID:         chainID,
			TxPath:          txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test: bad -account-number flag", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		// Initialize the account number flag.
		flagAccountNumber := &commands.Uint64Flag{}
		flagAccountNumber.Set(strconv.FormatUint(accountNumber+1, 10)) // Bad account number.

		// Initialize the account sequence flag.
		flagAccountSequence := &commands.Uint64Flag{}
		flagAccountSequence.Set(strconv.FormatUint(accountSequence, 10))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			AccountNumber:   *flagAccountNumber,
			AccountSequence: *flagAccountSequence,
			ChainID:         chainID,
			TxPath:          txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test: bad -chainid flag", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		// Initialize the account number flag.
		flagAccountNumber := &commands.Uint64Flag{}
		flagAccountNumber.Set(strconv.FormatUint(accountNumber+1, 10)) // Bad account number.

		// Initialize the account sequence flag.
		flagAccountSequence := &commands.Uint64Flag{}
		flagAccountSequence.Set(strconv.FormatUint(accountSequence, 10))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			AccountNumber:   *flagAccountNumber,
			AccountSequence: *flagAccountSequence,
			ChainID:         "bad-chainid", // Bad chain ID.
			TxPath:          txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})

	t.Run("test: no -chainid: ok", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		// Initialize the account number flag.
		flagAccountNumber := &commands.Uint64Flag{}
		flagAccountNumber.Set(strconv.FormatUint(accountNumber, 10))

		// Initialize the account sequence flag.
		flagAccountSequence := &commands.Uint64Flag{}
		flagAccountSequence.Set(strconv.FormatUint(accountSequence, 10))

		// Create a test server that will return the account number and sequence.
		handler := defaultHTTPHandler(t, "status", &ctypes.ResultStatus{
			NodeInfo: p2pTypes.NodeInfo{
				Network: chainID,
			},
		})

		server := createTestServer(t, handler)
		defer server.Close()

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
					Remote:                server.URL, // Needs remote to fetch account info.
				},
			},
			AccountNumber:   *flagAccountNumber,
			AccountSequence: *flagAccountSequence,
			TxPath:          txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io) // Account-number and account-sequence wrong.
		assert.NoError(t, err)
	})

	t.Run("test: no -account-number: bad default value", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		// Initialize the account sequence flag.
		flagAccountSequence := &commands.Uint64Flag{}
		flagAccountSequence.Set(strconv.FormatUint(accountSequence, 10))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
					Remote:                "http://localhost:26657", // The node doesn't exist.
				},
			},
			AccountSequence: *flagAccountSequence,
			ChainID:         chainID,
			TxPath:          txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io) // Account-number and account-sequence wrong.
		assert.Error(t, err)
	})

	t.Run("test: no -account-number: ok", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		// Initialize the account sequence flag.
		flagAccountSequence := &commands.Uint64Flag{}
		flagAccountSequence.Set(strconv.FormatUint(accountSequence, 10))

		baseAccount, err := amino.MarshalJSON(
			struct{ BaseAccount std.BaseAccount }{
				std.BaseAccount{
					AccountNumber: accountNumber,
					Sequence:      accountSequence,
				},
			},
		)
		require.NoError(t, err)

		// Create a test server that will return the account number and sequence.
		handler := defaultHTTPHandler(t, "abci_query", &ctypes.ResultABCIQuery{
			Response: abci.ResponseQuery{
				ResponseBase: abci.ResponseBase{
					Data: baseAccount,
				},
			},
		},
		)
		server := createTestServer(t, handler)
		defer server.Close()

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
					Remote:                server.URL, // Needs remote to fetch account info.
				},
			},
			AccountSequence: *flagAccountSequence,
			ChainID:         chainID,
			TxPath:          txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io) // Account-number and account-sequence wrong.
		assert.NoError(t, err)
	})

	t.Run("test: no -account-sequence: bad default value", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		// Initialize the account number flag.
		flagAccountNumber := &commands.Uint64Flag{}
		flagAccountNumber.Set(strconv.FormatUint(accountNumber, 10))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
					Remote:                "http://localhost:26657", // The node doesn't exist.
				},
			},
			AccountNumber: *flagAccountNumber,
			ChainID:       chainID,
			TxPath:        txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io) // Account-number and account-sequence wrong.
		assert.Error(t, err)
	})

	t.Run("test: no -account-sequence: ok", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		// Initialize the account number flag.
		flagAccountNumber := &commands.Uint64Flag{}
		flagAccountNumber.Set(strconv.FormatUint(accountNumber, 10))

		baseAccount, err := amino.MarshalJSON(
			struct{ BaseAccount std.BaseAccount }{
				std.BaseAccount{
					AccountNumber: accountNumber,
					Sequence:      accountSequence,
				},
			},
		)
		require.NoError(t, err)

		// Create a test server that will return the account number and sequence.
		handler := defaultHTTPHandler(t, "abci_query", &ctypes.ResultABCIQuery{
			Response: abci.ResponseQuery{
				ResponseBase: abci.ResponseBase{
					Data: baseAccount,
				},
			},
		},
		)
		server := createTestServer(t, handler)
		defer server.Close()

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
					Remote:                server.URL, // Needs remote to fetch account info.
				},
			},
			AccountNumber: *flagAccountNumber,
			ChainID:       chainID,
			TxPath:        txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io) // Account-number and account-sequence wrong.
		assert.NoError(t, err)
	})

	t.Run("test: no -account-sequence: bad sequence response", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		// Initialize the account number flag.
		flagAccountNumber := &commands.Uint64Flag{}
		flagAccountNumber.Set(strconv.FormatUint(accountNumber, 10))

		baseAccount, err := amino.MarshalJSON(
			struct{ BaseAccount std.BaseAccount }{
				std.BaseAccount{
					AccountNumber: accountNumber,
					Sequence:      accountSequence + 1,
				},
			},
		)
		require.NoError(t, err)

		// Create a test server that will return the account number and sequence.
		handler := defaultHTTPHandler(t, "abci_query", &ctypes.ResultABCIQuery{
			Response: abci.ResponseQuery{
				ResponseBase: abci.ResponseBase{
					Data: baseAccount,
				},
			},
		},
		)
		server := createTestServer(t, handler)
		defer server.Close()

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
					Remote:                server.URL, // Needs remote to fetch account info.
				},
			},
			AccountNumber: *flagAccountNumber,
			ChainID:       chainID,
			TxPath:        txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io) // Account-number and account-sequence wrong.
		assert.Error(t, err)
	})

	t.Run("test: no -account-sequence and -account-number flags: error", func(t *testing.T) {
		t.Parallel()

		kbHome, tx, cleanUp := prepare(t)
		defer cleanUp()

		// Marshal the tx with signature.
		rawTx, err := amino.MarshalJSON(tx)
		assert.NoError(t, err)

		txFile, err := os.CreateTemp("", "tx-*.json")
		require.NoError(t, err)

		require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

		cfg := &VerifyCfg{
			RootCfg: &BaseCfg{
				BaseOptions: BaseOptions{
					Home:                  kbHome,
					InsecurePasswordStdin: true,
				},
			},
			ChainID: chainID,
			TxPath:  txFile.Name(),
		}

		io := commands.NewTestIO()
		args := []string{fakeKeyName1}

		err = execVerify(context.Background(), cfg, args, io)
		assert.Error(t, err)
	})
}

func Test_VerifyMultisig(t *testing.T) {
	t.Parallel()

	var (
		kbHome      = t.TempDir()
		baseOptions = BaseOptions{
			InsecurePasswordStdin: true,
			Home:                  kbHome,
		}

		encryptPassword = "encrypt"
		multisigName    = "multisig-012"
	)

	// Generate 3 keys, for the multisig.
	privKeys := []secp256k1.PrivKeySecp256k1{
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
		secp256k1.GenPrivKey(),
	}

	kb, err := keys.NewKeyBaseFromDir(kbHome)
	require.NoError(t, err)

	// Import the (public) keys into the keybase.
	require.NoError(t, kb.ImportPrivKey("k0", privKeys[0], encryptPassword))
	require.NoError(t, kb.ImportPrivKey("k1", privKeys[1], encryptPassword))
	require.NoError(t, kb.ImportPrivKey("k2", privKeys[2], encryptPassword))

	// Build the multisig pub-key (2 of 3).
	msPub := multisig.NewPubKeyMultisigThreshold(
		2, // Threshold.
		[]crypto.PubKey{
			privKeys[0].PubKey(),
			privKeys[1].PubKey(),
			privKeys[2].PubKey(),
		},
	)

	msInfo, err := kb.CreateMulti(multisigName, msPub)
	require.NoError(t, err)

	// Generate a minimal tx.
	tx := std.Tx{
		Fee: std.Fee{
			GasWanted: 10,
			GasFee: std.Coin{
				Amount: 10,
				Denom:  "ugnot",
			},
		},
		Msgs: []std.Msg{
			bank.MsgSend{
				FromAddress: msInfo.GetAddress(), // Multisig account is the signer.
			},
		},
	}

	txFile, err := os.CreateTemp("", "tx-*.json")
	require.NoError(t, err)

	rawTx, err := amino.MarshalJSON(tx)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(txFile.Name(), rawTx, 0o644))

	// Have 2 out of 3 key sign the tx, with `gnokey sign`.
	genSignature := func(keyName, sigOut string) {
		// Each invocation needs its own root command.
		io := commands.NewTestIO()
		io.SetIn(
			strings.NewReader(
				fmt.Sprintf(
					"%s\n%s\n",
					encryptPassword,
					encryptPassword,
				),
			),
		)

		signCmd := NewRootCmdWithBaseConfig(io, baseOptions)

		args := []string{
			"sign",
			"--insecure-password-stdin",
			"--home", kbHome,
			"--tx-path", txFile.Name(),
			"--output-document", sigOut,
			keyName,
		}

		require.NoError(t, signCmd.ParseAndRun(context.Background(), args))
	}

	sigs := []string{
		filepath.Join(t.TempDir(), "sig0.json"),
		filepath.Join(t.TempDir(), "sig1.json"),
	}

	genSignature("k0", sigs[0])
	genSignature("k1", sigs[1])

	// Generate the multisig.
	io := commands.NewTestIO()
	multiCmd := NewRootCmdWithBaseConfig(io, baseOptions)

	args := []string{
		"multisign",
		"--insecure-password-stdin",
		"--home", kbHome,
		"--tx-path", txFile.Name(),
		"--signature", sigs[0],
		"--signature", sigs[1],
		multisigName,
	}
	require.NoError(t, multiCmd.ParseAndRun(context.Background(), args))

	// Get the multisig from the transaction file.
	signedRaw, err := os.ReadFile(txFile.Name())
	require.NoError(t, err)

	var signedTx std.Tx
	require.NoError(t, amino.UnmarshalJSON(signedRaw, &signedTx))
	require.Len(t, signedTx.Signatures, 1)

	// Prepare the verify function.
	cfg := &VerifyCfg{
		RootCfg: &BaseCfg{
			BaseOptions: baseOptions,
		},
		ChainID:         "dev",
		AccountNumber:   commands.Uint64Flag{V: 0, Defined: true},
		AccountSequence: commands.Uint64Flag{V: 0, Defined: true},
		TxPath:          txFile.Name(),
	}

	vargs := []string{multisigName}
	err = execVerify(context.Background(), cfg, vargs, io)
	assert.NoError(t, err)
}
