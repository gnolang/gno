package gnoclient

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"

	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallSingle_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	signer := newInMemorySigner(t, "tendermint_test")
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Setup Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         ugnot.ValueString(10000),
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	// Make Msg config
	msg := vm.MsgCall{
		Caller:  caller.GetAddress(),
		PkgPath: "gno.land/r/demo/deep/very/deep",
		Func:    "Render",
		Args:    []string{"test argument"},
		Send:    nil,
	}

	// Execute call
	res, err := client.Call(baseCfg, msg)
	require.NoError(t, err)

	expected := "(\"hi test argument\" string)\n\n"
	got := string(res.DeliverTx.Data)

	assert.Equal(t, expected, got)

	// Test signing separately
	tx, err := NewCallTx(baseCfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, tx)
	signedTx, err := client.SignTx(*tx, 0, 0)
	assert.NoError(t, err)
	require.NotNil(t, signedTx)
	res, err = client.BroadcastTxCommit(signedTx)
	require.NoError(t, err)
	got = string(res.DeliverTx.Data)
	assert.Equal(t, expected, got)
}

func TestCallMultiple_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	signer := newInMemorySigner(t, "tendermint_test")
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Setup Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         ugnot.ValueString(10000),
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	// Make Msg configs
	msg1 := vm.MsgCall{
		Caller:  caller.GetAddress(),
		PkgPath: "gno.land/r/demo/deep/very/deep",
		Func:    "Render",
		Args:    []string{""},
		Send:    nil,
	}

	// Same call, different argument
	msg2 := vm.MsgCall{
		Caller:  caller.GetAddress(),
		PkgPath: "gno.land/r/demo/deep/very/deep",
		Func:    "Render",
		Args:    []string{"test argument"},
		Send:    nil,
	}

	expected := "(\"it works!\" string)\n\n(\"hi test argument\" string)\n\n"

	// Execute call
	res, err := client.Call(baseCfg, msg1, msg2)
	require.NoError(t, err)

	got := string(res.DeliverTx.Data)
	assert.Equal(t, expected, got)

	// Test signing separately
	tx, err := NewCallTx(baseCfg, msg1, msg2)
	assert.NoError(t, err)
	require.NotNil(t, tx)
	signedTx, err := client.SignTx(*tx, 0, 0)
	assert.NoError(t, err)
	require.NotNil(t, signedTx)
	res, err = client.BroadcastTxCommit(signedTx)
	require.NoError(t, err)
	got = string(res.DeliverTx.Data)
	assert.Equal(t, expected, got)
}

func TestSendSingle_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	signer := newInMemorySigner(t, "tendermint_test")
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Setup Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         ugnot.ValueString(10000),
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	// Make Send config for a new address on the blockchain
	toAddress, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")
	amount := 10
	msg := bank.MsgSend{
		FromAddress: caller.GetAddress(),
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

	// Test signing separately
	tx, err := NewSendTx(baseCfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, tx)
	signedTx, err := client.SignTx(*tx, 0, 0)
	assert.NoError(t, err)
	require.NotNil(t, signedTx)
	res, err = client.BroadcastTxCommit(signedTx)
	require.NoError(t, err)
	assert.Equal(t, "", string(res.DeliverTx.Data))

	// Get the new account balance
	account, _, err = client.QueryAccount(toAddress)
	require.NoError(t, err)
	expected2 := std.Coins{{"ugnot", int64(2 * amount)}}
	got = account.GetCoins()
	assert.Equal(t, expected2, got)
}

func TestSendMultiple_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	signer := newInMemorySigner(t, "tendermint_test")
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Setup Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         ugnot.ValueString(10000),
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	// Make Msg configs
	toAddress, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")
	amount1 := 10
	msg1 := bank.MsgSend{
		FromAddress: caller.GetAddress(),
		ToAddress:   toAddress,
		Amount:      std.Coins{{Denom: ugnot.Denom, Amount: int64(amount1)}},
	}

	// Same send, different argument
	amount2 := 20
	msg2 := bank.MsgSend{
		FromAddress: caller.GetAddress(),
		ToAddress:   toAddress,
		Amount:      std.Coins{{Denom: ugnot.Denom, Amount: int64(amount2)}},
	}

	// Execute send
	res, err := client.Send(baseCfg, msg1, msg2)
	assert.NoError(t, err)
	assert.Equal(t, "", string(res.DeliverTx.Data))

	// Get the new account balance
	account, _, err := client.QueryAccount(toAddress)
	assert.NoError(t, err)

	expected := std.Coins{{Denom: ugnot.Denom, Amount: int64(amount1 + amount2)}}
	got := account.GetCoins()

	assert.Equal(t, expected, got)

	// Test signing separately
	tx, err := NewSendTx(baseCfg, msg1, msg2)
	assert.NoError(t, err)
	require.NotNil(t, tx)
	signedTx, err := client.SignTx(*tx, 0, 0)
	assert.NoError(t, err)
	require.NotNil(t, signedTx)
	res, err = client.BroadcastTxCommit(signedTx)
	require.NoError(t, err)
	assert.Equal(t, "", string(res.DeliverTx.Data))

	// Get the new account balance
	account, _, err = client.QueryAccount(toAddress)
	require.NoError(t, err)
	expected2 := std.Coins{{"ugnot", int64(2 * (amount1 + amount2))}}
	got = account.GetCoins()
	assert.Equal(t, expected2, got)
}

// Run tests
func TestRunSingle_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	signer := newInMemorySigner(t, "tendermint_test")
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         ugnot.ValueString(10000),
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	fileBody := `package main
import (
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/tests"
)
func main() {
	println(ufmt.Sprintf("- before: %d", tests.Counter()))
	for i := 0; i < 10; i++ {
		tests.IncCounter()
	}
	println(ufmt.Sprintf("- after: %d", tests.Counter()))
}`

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	// Make Msg configs
	msg := vm.MsgRun{
		Caller: caller.GetAddress(),
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

	res, err := client.Run(baseCfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), "- before: 0\n- after: 10\n")

	// Test signing separately
	tx, err := NewRunTx(baseCfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, tx)
	signedTx, err := client.SignTx(*tx, 0, 0)
	assert.NoError(t, err)
	require.NotNil(t, signedTx)
	res, err = client.BroadcastTxCommit(signedTx)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), "- before: 10\n- after: 20\n")
}

// Run tests
func TestRunMultiple_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	signer := newInMemorySigner(t, "tendermint_test")
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         ugnot.ValueString(10000),
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	fileBody1 := `package main
import (
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/tests"
)
func main() {
	println(ufmt.Sprintf("- before: %d", tests.Counter()))
	for i := 0; i < 10; i++ {
		tests.IncCounter()
	}
	println(ufmt.Sprintf("- after: %d", tests.Counter()))
}`

	fileBody2 := `package main
import (
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/deep/very/deep"
)
func main() {
	println(ufmt.Sprintf("%s", deep.Render("gnoclient!")))
}`

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	// Make Msg configs
	msg1 := vm.MsgRun{
		Caller: caller.GetAddress(),
		Package: &std.MemPackage{
			Name: "main",
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody1,
				},
			},
		},
		Send: nil,
	}
	msg2 := vm.MsgRun{
		Caller: caller.GetAddress(),
		Package: &std.MemPackage{
			Name: "main",
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody2,
				},
			},
		},
		Send: nil,
	}

	expected := "- before: 0\n- after: 10\nhi gnoclient!\n"

	res, err := client.Run(baseCfg, msg1, msg2)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))

	// Test signing separately
	tx, err := NewRunTx(baseCfg, msg1, msg2)
	assert.NoError(t, err)
	require.NotNil(t, tx)
	signedTx, err := client.SignTx(*tx, 0, 0)
	assert.NoError(t, err)
	require.NotNil(t, signedTx)
	res, err = client.BroadcastTxCommit(signedTx)
	assert.NoError(t, err)
	require.NotNil(t, res)
	expected2 := "- before: 10\n- after: 20\nhi gnoclient!\n"
	assert.Equal(t, expected2, string(res.DeliverTx.Data))
}

func TestAddPackageSingle_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	signer := newInMemorySigner(t, "tendermint_test")
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Setup Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         ugnot.ValueString(10000),
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	body := `package echo

func Echo(str string) string {
	return str
}`

	fileName := "echo.gno"
	deploymentPath := "gno.land/p/demo/integration/test/echo"
	deposit := std.Coins{{Denom: ugnot.Denom, Amount: int64(100)}}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	// Make Msg config
	msg := vm.MsgAddPackage{
		Creator: caller.GetAddress(),
		Package: &std.MemPackage{
			Name: "echo",
			Path: deploymentPath,
			Files: []*std.MemFile{
				{
					Name: fileName,
					Body: body,
				},
			},
		},
		Deposit: deposit,
	}

	// Execute AddPackage
	_, err = client.AddPackage(baseCfg, msg)
	assert.NoError(t, err)

	// Check for deployed file on the node
	query, err := client.Query(QueryCfg{
		Path: "vm/qfile",
		Data: []byte(deploymentPath),
	})
	require.NoError(t, err)
	assert.Equal(t, string(query.Response.Data), fileName)

	// Query balance to validate deposit
	baseAcc, _, err := client.QueryAccount(gnolang.DerivePkgAddr(deploymentPath))
	require.NoError(t, err)
	assert.Equal(t, baseAcc.GetCoins(), deposit)
}

func TestAddPackageMultiple_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	signer := newInMemorySigner(t, "tendermint_test")
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	// Setup Client
	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	// Make Tx config
	baseCfg := BaseTxCfg{
		GasFee:         ugnot.ValueString(10000),
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	deposit := std.Coins{{Denom: ugnot.Denom, Amount: int64(100)}}
	deploymentPath1 := "gno.land/p/demo/integration/test/echo"

	body1 := `package echo

func Echo(str string) string {
	return str
}`

	deploymentPath2 := "gno.land/p/demo/integration/test/hello"
	body2 := `package hello

func Hello(str string) string {
	return "Hello " + str + "!" 
}`

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg1 := vm.MsgAddPackage{
		Creator: caller.GetAddress(),
		Package: &std.MemPackage{
			Name: "echo",
			Path: deploymentPath1,
			Files: []*std.MemFile{
				{
					Name: "echo.gno",
					Body: body1,
				},
			},
		},
		Deposit: nil,
	}

	msg2 := vm.MsgAddPackage{
		Creator: caller.GetAddress(),
		Package: &std.MemPackage{
			Name: "hello",
			Path: deploymentPath2,
			Files: []*std.MemFile{
				{
					Name: "gno.mod",
					Body: "module gno.land/p/demo/integration/test/hello",
				},
				{
					Name: "hello.gno",
					Body: body2,
				},
			},
		},
		Deposit: deposit,
	}

	// Execute AddPackage
	_, err = client.AddPackage(baseCfg, msg1, msg2)
	assert.NoError(t, err)

	// Check Package #1
	query, err := client.Query(QueryCfg{
		Path: "vm/qfile",
		Data: []byte(deploymentPath1),
	})
	require.NoError(t, err)
	assert.Equal(t, string(query.Response.Data), "echo.gno")

	// Query balance to validate deposit
	baseAcc, _, err := client.QueryAccount(gnolang.DerivePkgAddr(deploymentPath1))
	require.NoError(t, err)
	assert.Equal(t, baseAcc.GetCoins().String(), "")

	// Check Package #2
	query, err = client.Query(QueryCfg{
		Path: "vm/qfile",
		Data: []byte(deploymentPath2),
	})
	require.NoError(t, err)
	assert.Contains(t, string(query.Response.Data), "hello.gno")
	assert.Contains(t, string(query.Response.Data), "gno.mod")

	// Query balance to validate deposit
	baseAcc, _, err = client.QueryAccount(gnolang.DerivePkgAddr(deploymentPath2))
	require.NoError(t, err)
	assert.Equal(t, baseAcc.GetCoins(), deposit)
}

// todo add more integration tests:
// MsgCall with Send field populated (single/multiple)
// MsgRun with Send field populated (single/multiple)

func newInMemorySigner(t *testing.T, chainid string) *SignerFromKeybase {
	t.Helper()

	mnemonic := integration.DefaultAccount_Seed
	name := integration.DefaultAccount_Name

	kb := keys.NewInMemory()
	_, err := kb.CreateAccount(name, mnemonic, "", "", uint32(0), uint32(0))
	require.NoError(t, err)

	return &SignerFromKeybase{
		Keybase:  kb,      // Stores keys in memory or on disk
		Account:  name,    // Account name or bech32 format
		Password: "",      // Password for encryption
		ChainID:  chainid, // Chain ID for transaction signing
	}
}
