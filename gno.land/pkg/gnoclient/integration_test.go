package gnoclient

import (
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"

	"github.com/gnolang/gno/tm2/pkg/std"

	"github.com/gnolang/gno/gno.land/pkg/integration"
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
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	// Make Msg config
	msg := MsgCall{
		PkgPath:  "gno.land/r/demo/deep/very/deep",
		FuncName: "Render",
		Args:     []string{"test argument"},
		Send:     "",
	}

	// Execute call
	res, err := client.Call(baseCfg, msg)

	expected := "(\"hi test argument\" string)"
	got := string(res.DeliverTx.Data)

	assert.Nil(t, err)
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
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	// Make Msg configs
	msg1 := MsgCall{
		PkgPath:  "gno.land/r/demo/deep/very/deep",
		FuncName: "Render",
		Args:     []string{""},
		Send:     "",
	}

	// Same call, different argument
	msg2 := MsgCall{
		PkgPath:  "gno.land/r/demo/deep/very/deep",
		FuncName: "Render",
		Args:     []string{"test argument"},
		Send:     "",
	}

	expected := "(\"it works!\" string)(\"hi test argument\" string)"

	// Execute call
	res, err := client.Call(baseCfg, msg1, msg2)

	got := string(res.DeliverTx.Data)
	assert.Nil(t, err)
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
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	// Make Send config for a new address on the blockchain
	toAddress, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")
	amount := 10
	msg := MsgSend{
		ToAddress: toAddress,
		Send:      std.Coin{"ugnot", int64(amount)}.String(),
	}

	// Execute send
	res, err := client.Send(baseCfg, msg)
	assert.Nil(t, err)
	assert.Equal(t, "", string(res.DeliverTx.Data))

	// Get the new account balance
	account, _, err := client.QueryAccount(toAddress)
	assert.Nil(t, err)

	expected := std.Coins{{"ugnot", int64(amount)}}
	got := account.GetCoins()

	assert.Equal(t, expected, got)
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
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	// Make Msg configs
	toAddress, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")
	amount1 := 10
	msg1 := MsgSend{
		ToAddress: toAddress,
		Send:      std.Coin{"ugnot", int64(amount1)}.String(),
	}

	// Same send, different argument
	amount2 := 20
	msg2 := MsgSend{
		ToAddress: toAddress,
		Send:      std.Coin{"ugnot", int64(amount2)}.String(),
	}

	// Execute send
	res, err := client.Send(baseCfg, msg1, msg2)
	assert.Nil(t, err)
	assert.Equal(t, "", string(res.DeliverTx.Data))

	// Get the new account balance
	account, _, err := client.QueryAccount(toAddress)
	assert.Nil(t, err)

	expected := std.Coins{{"ugnot", int64(amount1 + amount2)}}
	got := account.GetCoins()

	assert.Equal(t, expected, got)
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
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	fileBody := `package main
import (
	"std"
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

	// Make Msg configs
	msg := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody,
				},
			},
		},
		Send: "",
	}

	res, err := client.Run(baseCfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), "- before: 0\n- after: 10\n")
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
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	fileBody1 := `package main
import (
	"std"
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
	"std"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/deep/very/deep"
)
func main() {
	println(ufmt.Sprintf("%s", deep.Render("gnoclient!")))
}`

	// Make Msg configs
	msg1 := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody1,
				},
			},
		},
		Send: "",
	}
	msg2 := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody2,
				},
			},
		},
		Send: "",
	}

	expected := "- before: 0\n- after: 10\nhi gnoclient!\n"

	res, err := client.Run(baseCfg, msg1, msg2)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
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
		GasFee:         "10000ugnot",
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
	deposit := "100ugnot"

	// Make Msg config
	msg := MsgAddPackage{
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
	assert.Nil(t, err)

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
	assert.Equal(t, baseAcc.GetCoins().String(), deposit)
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
		GasFee:         "10000ugnot",
		GasWanted:      8000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	deposit := "100ugnot"
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

	msg1 := MsgAddPackage{
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
		Deposit: "",
	}

	msg2 := MsgAddPackage{
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
	assert.Nil(t, err)

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
	assert.Equal(t, baseAcc.GetCoins().String(), deposit)
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
