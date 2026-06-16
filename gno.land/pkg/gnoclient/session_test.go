package gnoclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const (
	TestSessionAccount_Name = "user1"
	TestSessionAccount_Seed = "mention vintage immense fix clerk state magnet embrace meadow buzz captain bar mystery decade mammal rib chunk upset finish athlete maple undo space palace"
)

func TestCallSessionSingle_Integration(t *testing.T) {
	// Set up packages
	rootdir := gnoenv.RootDir()
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	meta := loadpkgs(t, rootdir, "gno.land/r/tests/vm/deep/very/deep")
	state := config.Genesis.AppState.(gnoland.GnoGenesisState)
	state.Txs = append(state.Txs, meta...)
	config.Genesis.AppState = state

	// Init in-memory node and RPCClient
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Make the master and session account
	masterSigner := newInMemorySigner(t, "tendermint_test")
	masterInfo, err := masterSigner.Info()
	require.NoError(t, err)

	signer := newInMemorySessionSigner(t, "tendermint_test", masterInfo.GetAddress())
	signerInfo, err := signer.Info()
	require.NoError(t, err)
	createSession(t, rpcClient, masterSigner, signerInfo.GetPubKey())

	// Set up Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Msg configs
	baseCfg := BaseTxCfg{
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      50000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}
	msg := vm.MsgCall{
		Caller:  masterInfo.GetAddress(),
		PkgPath: "gno.land/r/tests/vm/deep/very/deep",
		Func:    "RenderCrossing",
		Args:    []string{"test argument"},
		Send:    nil,
	}

	// Execute call
	res, err := client.Call(baseCfg, msg)
	require.NoError(t, err)

	expected := "(\"hi test argument\" string)\n\n"
	got := string(res.DeliverTx.Data)

	assert.Equal(t, expected, got)

	res, err = callSigningSeparately(t, client, baseCfg, msg)
	require.NoError(t, err)
	got = string(res.DeliverTx.Data)
	assert.Equal(t, expected, got)
}

func TestSendSessionSingle_Integration(t *testing.T) {
	// Set up in-memory node and RPCClient
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Make the master and session account
	masterSigner := newInMemorySigner(t, "tendermint_test")
	masterInfo, err := masterSigner.Info()
	require.NoError(t, err)

	signer := newInMemorySessionSigner(t, "tendermint_test", masterInfo.GetAddress())
	signerInfo, err := signer.Info()
	require.NoError(t, err)
	createSession(t, rpcClient, masterSigner, signerInfo.GetPubKey())

	// Set up Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Msg configs
	baseCfg := BaseTxCfg{
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      50000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	// Make Send config for a new address on the blockchain
	toAddress, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")
	amount := 10
	msg := bank.MsgSend{
		FromAddress: masterInfo.GetAddress(),
		ToAddress:   toAddress,
		Amount:      std.Coins{{Denom: ugnot.Denom, Amount: int64(amount)}},
	}

	// Execute send
	res, err := client.Send(baseCfg, msg)
	require.NoError(t, err)
	assert.Equal(t, "", string(res.DeliverTx.Data))

	// Get the new account balance
	account, _, err := client.QueryAccount(toAddress)
	require.NoError(t, err)

	expected := std.Coins{{Denom: ugnot.Denom, Amount: int64(amount)}}
	got := account.GetCoins()
	assert.Equal(t, expected, got)

	res, err = sendSigningSeparately(t, client, baseCfg, msg)
	require.NoError(t, err)
	assert.Equal(t, "", string(res.DeliverTx.Data))

	// Get the new account balance
	account, _, err = client.QueryAccount(toAddress)
	require.NoError(t, err)
	expected2 := std.Coins{{Denom: ugnot.Denom, Amount: int64(2 * amount)}}
	got = account.GetCoins()
	assert.Equal(t, expected2, got)
}

func TestRunSessionSingle_Integration(t *testing.T) {
	// Set up packages
	rootdir := gnoenv.RootDir()
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	meta := loadpkgs(t, rootdir, "gno.land/p/nt/ufmt/v0", "gno.land/r/tests/vm")
	state := config.Genesis.AppState.(gnoland.GnoGenesisState)
	state.Txs = append(state.Txs, meta...)
	config.Genesis.AppState = state

	// Init in-memory node and RPCClient
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Make the master and session account
	masterSigner := newInMemorySigner(t, "tendermint_test")
	masterInfo, err := masterSigner.Info()
	require.NoError(t, err)

	signer := newInMemorySessionSigner(t, "tendermint_test", masterInfo.GetAddress())
	signerInfo, err := signer.Info()
	require.NoError(t, err)
	createSession(t, rpcClient, masterSigner, signerInfo.GetPubKey())

	// Set up Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      50000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	fileBody := `package main
import (
	"gno.land/p/nt/ufmt/v0"
	tests "gno.land/r/tests/vm"
)
func main(cur realm) {
	println(ufmt.Sprintf("- before: %d", tests.Counter(cross(cur))))
	for i := 0; i < 10; i++ {
		tests.IncCounter(cross(cur))
	}
	println(ufmt.Sprintf("- after: %d", tests.Counter(cross(cur))))
}`

	// Make Msg configs
	msg := vm.MsgRun{
		Caller: masterInfo.GetAddress(),
		Package: &std.MemPackage{
			Name: "main",
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody,
				},
			},
		},
		Send: nil,
	}

	// Execute run
	res, err := client.Run(baseCfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), "- before: 0\n- after: 10\n")

	res, err = runSigningSeparately(t, client, baseCfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), "- before: 10\n- after: 20\n")
}

// TestValidateSessionKey tests that Signer.Validate works for a session key
func TestValidateSessionKey(t *testing.T) {
	// Set up packages
	rootdir := gnoenv.RootDir()
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	meta := loadpkgs(t, rootdir, "gno.land/r/tests/vm/deep/very/deep")
	state := config.Genesis.AppState.(gnoland.GnoGenesisState)
	state.Txs = append(state.Txs, meta...)
	config.Genesis.AppState = state

	// Make the master and session account
	masterSigner := newInMemorySigner(t, "tendermint_test")
	masterInfo, err := masterSigner.Info()
	require.NoError(t, err)

	// As a baseline, validate the "normal" master account key
	require.NoError(t, masterSigner.Validate())
	// Now, validate the session account key
	signer := newInMemorySessionSigner(t, "tendermint_test", masterInfo.GetAddress())
	require.NoError(t, signer.Validate())
}

func TestQuerySessionAccount_Integration(t *testing.T) {
	// Set up in-memory node and RPCClient
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	masterSigner := newInMemorySigner(t, "tendermint_test")
	masterInfo, err := masterSigner.Info()
	require.NoError(t, err)

	sessionSigner := newInMemorySessionSigner(t, "tendermint_test", masterInfo.GetAddress())
	sessionInfo, err := sessionSigner.Info()
	require.NoError(t, err)
	createSession(t, rpcClient, masterSigner, sessionInfo.GetPubKey())

	client := Client{
		Signer:    masterSigner,
		RPCClient: rpcClient,
	}

	// Query the session account — must succeed and return a valid account
	account, qres, err := client.QuerySessionAccount(masterInfo.GetAddress(), sessionInfo.GetAddress())
	require.NoError(t, err)
	require.NotNil(t, qres)
	require.NotNil(t, account)

	// Query with an unknown session address — must return an error
	unknown, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")
	_, _, err = client.QuerySessionAccount(masterInfo.GetAddress(), unknown)
	require.Error(t, err)
}

func TestRevokeSession_Integration(t *testing.T) {
	// Set up in-memory node and RPCClient
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	masterSigner := newInMemorySigner(t, "tendermint_test")
	masterInfo, err := masterSigner.Info()
	require.NoError(t, err)

	sessionSigner := newInMemorySessionSigner(t, "tendermint_test", masterInfo.GetAddress())
	sessionInfo, err := sessionSigner.Info()
	require.NoError(t, err)
	createSession(t, rpcClient, masterSigner, sessionInfo.GetPubKey())

	masterClient := Client{
		Signer:    masterSigner,
		RPCClient: rpcClient,
	}

	// Revoke the specific session
	baseCfg := BaseTxCfg{
		GasFee:    ugnot.ValueString(2100000),
		GasWanted: 50000000,
	}
	msg := auth.MsgRevokeSession{
		Creator:    masterInfo.GetAddress(),
		SessionKey: sessionInfo.GetPubKey(),
	}
	_, err = masterClient.RevokeSession(baseCfg, msg)
	require.NoError(t, err)

	// The session account should no longer exist
	_, _, err = masterClient.QuerySessionAccount(masterInfo.GetAddress(), sessionInfo.GetAddress())
	require.Error(t, err)
}

func TestRevokeAllSessions_Integration(t *testing.T) {
	// Set up in-memory node and RPCClient
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	masterSigner := newInMemorySigner(t, "tendermint_test")
	masterInfo, err := masterSigner.Info()
	require.NoError(t, err)

	sessionSigner := newInMemorySessionSigner(t, "tendermint_test", masterInfo.GetAddress())
	sessionInfo, err := sessionSigner.Info()
	require.NoError(t, err)
	createSession(t, rpcClient, masterSigner, sessionInfo.GetPubKey())

	masterClient := Client{
		Signer:    masterSigner,
		RPCClient: rpcClient,
	}

	// Revoke all sessions
	baseCfg := BaseTxCfg{
		GasFee:    ugnot.ValueString(2100000),
		GasWanted: 50000000,
	}
	msg := auth.MsgRevokeAllSessions{
		Creator: masterInfo.GetAddress(),
	}
	_, err = masterClient.RevokeAllSessions(baseCfg, msg)
	require.NoError(t, err)

	// The session account should no longer exist
	_, _, err = masterClient.QuerySessionAccount(masterInfo.GetAddress(), sessionInfo.GetAddress())
	require.Error(t, err)
}

func newInMemorySessionSigner(t *testing.T, chainid string, master crypto.Address) *SignerFromKeybase {
	t.Helper()

	mnemonic := TestSessionAccount_Seed
	name := TestSessionAccount_Name

	kb := keys.NewInMemory()
	_, err := kb.CreateAccount(name, mnemonic, "", "", uint32(0), uint32(0))
	require.NoError(t, err)

	return &SignerFromKeybase{
		Keybase:  kb,      // Stores keys in memory or on disk
		Account:  name,    // Account name or bech32 format
		Password: "",      // Password for encryption
		ChainID:  chainid, // Chain ID for transaction signing
		Master:   master,  // The address of the master account
	}
}

func createSession(t *testing.T, rpcClient rpcclient.Client, masterSigner Signer, sessionKey crypto.PubKey) {
	t.Helper()

	client := Client{
		Signer:    masterSigner,
		RPCClient: rpcClient,
	}
	masterInfo, err := masterSigner.Info()
	require.NoError(t, err)

	baseCfg := BaseTxCfg{
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      50000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}
	msg := auth.MsgCreateSession{
		Creator:    masterInfo.GetAddress(),
		SessionKey: sessionKey,
		SpendLimit: std.Coins{std.NewCoin("ugnot", 5000000)},
		AllowPaths: []string{"*"},
	}
	_, err = client.CreateSession(baseCfg, msg)
	require.NoError(t, err)
}
