package gnoclient

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestCallSingle_Integration(t *testing.T) {
	// Setup packages
	rootdir := gnoenv.RootDir()
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	meta := loadpkgs(t, rootdir, "gno.land/r/tests/vm/deep/very/deep")
	state := config.Genesis.AppState.(gnoland.GnoGenesisState)
	state.Txs = append(state.Txs, meta...)
	config.Genesis.AppState = state

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
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      21000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	// Make Msg config
	msg := vm.MsgCall{
		Caller:  caller.GetAddress(),
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

func TestCallMultiple_Integration(t *testing.T) {
	// Setup packages
	rootdir := gnoenv.RootDir()
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	meta := loadpkgs(t, rootdir, "gno.land/r/tests/vm/deep/very/deep")
	state := config.Genesis.AppState.(gnoland.GnoGenesisState)
	state.Txs = append(state.Txs, meta...)
	config.Genesis.AppState = state

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
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      21000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	// Make Msg configs
	msg1 := vm.MsgCall{
		Caller:  caller.GetAddress(),
		PkgPath: "gno.land/r/tests/vm/deep/very/deep",
		Func:    "RenderCrossing",
		Args:    []string{""},
		Send:    nil,
	}

	// Same call, different argument
	msg2 := vm.MsgCall{
		Caller:  caller.GetAddress(),
		PkgPath: "gno.land/r/tests/vm/deep/very/deep",
		Func:    "RenderCrossing",
		Args:    []string{"test argument"},
		Send:    nil,
	}

	expected := "(\"it works!\" string)\n\n(\"hi test argument\" string)\n\n"

	// Execute call
	res, err := client.Call(baseCfg, msg1, msg2)
	require.NoError(t, err)

	got := string(res.DeliverTx.Data)
	assert.Equal(t, expected, got)

	res, err = callSigningSeparately(t, client, baseCfg, msg1, msg2)
	require.NoError(t, err)
	got = string(res.DeliverTx.Data)
	assert.Equal(t, expected, got)
}

func TestSendSingle_Integration(t *testing.T) {
	// Set up in-memory node
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
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
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      21000000,
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

func TestSendMultiple_Integration(t *testing.T) {
	// Set up in-memory node
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
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
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      21000000,
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

	res, err = sendSigningSeparately(t, client, baseCfg, msg1, msg2)
	require.NoError(t, err)
	assert.Equal(t, "", string(res.DeliverTx.Data))

	// Get the new account balance
	account, _, err = client.QueryAccount(toAddress)
	require.NoError(t, err)
	expected2 := std.Coins{{Denom: ugnot.Denom, Amount: int64(2 * (amount1 + amount2))}}
	got = account.GetCoins()
	assert.Equal(t, expected2, got)
}

// Run tests
func TestRunSingle_Integration(t *testing.T) {
	// Setup packages
	rootdir := gnoenv.RootDir()
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	meta := loadpkgs(t, rootdir, "gno.land/p/nt/ufmt", "gno.land/r/tests/vm")
	state := config.Genesis.AppState.(gnoland.GnoGenesisState)
	state.Txs = append(state.Txs, meta...)
	config.Genesis.AppState = state

	// Set up in-memory node
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
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      21000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	fileBody := `package main
import (
	"gno.land/p/nt/ufmt"
	tests "gno.land/r/tests/vm"
)
func main() {
	println(ufmt.Sprintf("- before: %d", tests.Counter(cross)))
	for i := 0; i < 10; i++ {
		tests.IncCounter(cross)
	}
	println(ufmt.Sprintf("- after: %d", tests.Counter(cross)))
}`

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	// Make Msg configs
	msg := vm.MsgRun{
		Caller: caller.GetAddress(),
		Package: &std.MemPackage{
			Name: "main",
			// Path: fmt.Sprintf("gno.land/e/%s/run", caller.GetAddress().String()),
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

	res, err = runSigningSeparately(t, client, baseCfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), "- before: 10\n- after: 20\n")
}

// Run tests
func TestRunMultiple_Integration(t *testing.T) {
	// Set up in-memory node
	rootdir := gnoenv.RootDir()
	config := integration.TestingMinimalNodeConfig(rootdir)
	meta := loadpkgs(t, rootdir,
		"gno.land/p/nt/ufmt",
		"gno.land/r/tests/vm",
		"gno.land/r/tests/vm/deep/very/deep",
	)
	state := config.Genesis.AppState.(gnoland.GnoGenesisState)
	state.Txs = append(state.Txs, meta...)
	config.Genesis.AppState = state

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
		GasFee:         ugnot.ValueString(2300000),
		GasWanted:      23000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	fileBody1 := `package main
import (
	"gno.land/p/nt/ufmt"
	tests "gno.land/r/tests/vm"
)
func main() {
	println(ufmt.Sprintf("- before: %d", tests.Counter(cross)))
	for i := 0; i < 10; i++ {
		tests.IncCounter(cross)
	}
	println(ufmt.Sprintf("- after: %d", tests.Counter(cross)))
}`

	fileBody2 := `package main
import (
	"gno.land/p/nt/ufmt"
	"gno.land/r/tests/vm/deep/very/deep"
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
			// Path: fmt.Sprintf("gno.land/e/%s/run", caller.GetAddress().String()),
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
			// Path: fmt.Sprintf("gno.land/e/%s/run", caller.GetAddress().String()),
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

	res, err = runSigningSeparately(t, client, baseCfg, msg1, msg2)
	require.NoError(t, err)
	require.NotNil(t, res)
	expected2 := "- before: 10\n- after: 20\nhi gnoclient!\n"
	assert.Equal(t, expected2, string(res.DeliverTx.Data))
}

func TestAddPackageSingle_Integration(t *testing.T) {
	// Set up in-memory node
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
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
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      21000000,
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
	deposit := std.Coins{{Denom: ugnot.Denom, Amount: int64(10000000)}}

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
				{
					Name: "gnomod.toml",
					Body: gnolang.GenGnoModLatest(deploymentPath),
				},
			},
		},
		MaxDeposit: deposit,
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
	assert.Equal(t, fileName+"\ngnomod.toml", string(query.Response.Data))

	// Query balance to validate deposit
	baseAcc, _, err := client.QueryAccount(gnolang.DeriveStorageDepositCryptoAddr(deploymentPath))
	require.NoError(t, err)
	assert.Equal(t, std.Coins{std.Coin{Denom: "ugnot", Amount: 177600}}, baseAcc.GetCoins())

	// Test signing separately (using a different deployment path)
	deploymentPathB := "gno.land/p/demo/integration/test/echo2"
	msg.Package.Path = deploymentPathB
	_, err = addPackageSigningSeparately(t, client, baseCfg, msg)
	assert.NoError(t, err)
	query, err = client.Query(QueryCfg{
		Path: "vm/qfile",
		Data: []byte(deploymentPathB),
	})
	require.NoError(t, err)
	assert.Equal(t, fileName+"\ngnomod.toml", string(query.Response.Data))
}

func TestAddPackageMultiple_Integration(t *testing.T) {
	// Set up in-memory node
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
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
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      21000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	deposit := std.Coins{{Denom: ugnot.Denom, Amount: int64(10000000)}}
	send := std.Coins{{Denom: ugnot.Denom, Amount: int64(1000000)}}
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
				{
					Name: "gnomod.toml",
					Body: gnolang.GenGnoModLatest(deploymentPath1),
				},
			},
		},
		MaxDeposit: nil,
	}

	msg2 := vm.MsgAddPackage{
		Creator: caller.GetAddress(),
		Package: &std.MemPackage{
			Name: "hello",
			Path: deploymentPath2,
			Files: []*std.MemFile{
				{
					Name: "gnomod.toml",
					Body: gnolang.GenGnoModLatest(deploymentPath2),
				},
				{
					Name: "hello.gno",
					Body: body2,
				},
			},
		},
		Send:       send,
		MaxDeposit: deposit,
	}

	// Verify initial balance of deployer's account
	baseAcc, _, err := client.QueryAccount(caller.GetAddress())
	require.NoError(t, err)
	assert.Equal(t, std.Coins{std.Coin{Denom: "ugnot", Amount: 10000000000000}}, baseAcc.GetCoins())
	// Execute AddPackage
	_, err = client.AddPackage(baseCfg, msg1, msg2)
	assert.NoError(t, err)

	// Check Package #1
	query, err := client.Query(QueryCfg{
		Path: "vm/qfile",
		Data: []byte(deploymentPath1),
	})
	require.NoError(t, err)
	assert.Equal(t, string(query.Response.Data), "echo.gno\ngnomod.toml")

	// Query balance to validate deposit
	baseAcc, _, err = client.QueryAccount(gnolang.DeriveStorageDepositCryptoAddr(deploymentPath1))
	require.NoError(t, err)
	assert.Equal(t, "177600ugnot", baseAcc.GetCoins().String())

	// Check Package #2
	query, err = client.Query(QueryCfg{
		Path: "vm/qfile",
		Data: []byte(deploymentPath2),
	})
	require.NoError(t, err)
	assert.Contains(t, string(query.Response.Data), "hello.gno")
	assert.Contains(t, string(query.Response.Data), "gnomod.toml")

	// Query storage deposit balance to validate deposit
	baseAcc, _, err = client.QueryAccount(gnolang.DeriveStorageDepositCryptoAddr(deploymentPath2))
	require.NoError(t, err)
	assert.Equal(t, std.Coins{std.Coin{Denom: "ugnot", Amount: 178700}}, baseAcc.GetCoins())

	// Verify the realm account balance received from the send
	baseAcc, _, err = client.QueryAccount(gnolang.DerivePkgCryptoAddr(deploymentPath2))
	require.NoError(t, err)
	assert.Equal(t, std.Coins{std.Coin{Denom: "ugnot", Amount: 1000000}}, baseAcc.GetCoins())

	// Verify remaining balance of deployer's account
	baseAcc, _, err = client.QueryAccount(caller.GetAddress())
	require.NoError(t, err)
	// 999999654370 = 10000000000000 - (GasFee 2100000 + Storage Deposit 177600 + Storage Deposit 178700 + Send 1000000)
	assert.Equal(t, std.Coins{std.Coin{Denom: "ugnot", Amount: 9999996543700}}, baseAcc.GetCoins())

	// Test signing separately (using a different deployment path)
	deploymentPath1B := "gno.land/p/demo/integration/test/echo2"
	deploymentPath2B := "gno.land/p/demo/integration/test/hello2"
	msg1.Package.Path = deploymentPath1B
	msg2.Package.Path = deploymentPath2B
	_, err = addPackageSigningSeparately(t, client, baseCfg, msg1, msg2)
	assert.NoError(t, err)
	query, err = client.Query(QueryCfg{
		Path: "vm/qfile",
		Data: []byte(deploymentPath1B),
	})
	require.NoError(t, err)
	assert.Equal(t, string(query.Response.Data), "echo.gno\ngnomod.toml")
	query, err = client.Query(QueryCfg{
		Path: "vm/qfile",
		Data: []byte(deploymentPath2B),
	})
	require.NoError(t, err)
	assert.Contains(t, string(query.Response.Data), "hello.gno")
	assert.Contains(t, string(query.Response.Data), "gnomod.toml")
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

func loadpkgs(t *testing.T, rootdir string, paths ...string) []gnoland.TxWithMetadata {
	t.Helper()

	loader := integration.NewPkgsLoader()
	examplesDir := filepath.Join(rootdir, "examples")
	for _, path := range paths {
		path = filepath.Clean(path)
		path = filepath.Join(examplesDir, path)
		err := loader.LoadPackage(examplesDir, path, "")
		require.NoErrorf(t, err, "`loadpkg` unable to load package(s) from %q: %s", path, err)
	}
	privKey, err := integration.GeneratePrivKeyFromMnemonic(integration.DefaultAccount_Seed, "", 0, 0)
	require.NoError(t, err)

	defaultFee := std.NewFee(50000, std.MustParseCoin(ugnot.ValueString(1000000)))

	meta, err := loader.GenerateTxs(privKey, defaultFee, nil)
	require.NoError(t, err)
	return meta
}

func TestCallVariadicFunc_Integration(t *testing.T) {
	// Setup packages
	rootdir := gnoenv.RootDir()
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	meta := loadpkgs(t, rootdir, "gno.land/r/tests/vm/variadic")
	state := config.Genesis.AppState.(gnoland.GnoGenesisState)
	state.Txs = append(state.Txs, meta...)
	config.Genesis.AppState = state

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
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      21000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	// Make Msg config
	msg := vm.MsgCall{
		Caller:  caller.GetAddress(),
		PkgPath: "gno.land/r/tests/vm/variadic",
		Func:    "Echo",
		Args:    []string{"test", "argument"},
		Send:    nil,
	}

	// Execute call
	res, err := client.Call(baseCfg, msg)
	require.NoError(t, err)

	expected := "(\"test argument\" string)\n\n"
	got := string(res.DeliverTx.Data)

	assert.Equal(t, expected, got)

	res, err = callSigningSeparately(t, client, baseCfg, msg)
	require.NoError(t, err)
	got = string(res.DeliverTx.Data)
	assert.Equal(t, expected, got)
}
func TestCallVariadicZeroVariadicArgsFunc_Integration(t *testing.T) {
	// Setup packages
	rootdir := gnoenv.RootDir()
	config := integration.TestingMinimalNodeConfig(gnoenv.RootDir())
	meta := loadpkgs(t, rootdir, "gno.land/r/tests/vm/variadic")
	state := config.Genesis.AppState.(gnoland.GnoGenesisState)
	state.Txs = append(state.Txs, meta...)
	config.Genesis.AppState = state

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
		GasFee:         ugnot.ValueString(2100000),
		GasWanted:      21000000,
		AccountNumber:  0,
		SequenceNumber: 0,
		Memo:           "",
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	// Make Msg config
	msg := vm.MsgCall{
		Caller:  caller.GetAddress(),
		PkgPath: "gno.land/r/tests/vm/variadic",
		Func:    "Echo",
		Args:    []string{},
		Send:    nil,
	}

	// Execute call
	res, err := client.Call(baseCfg, msg)
	require.NoError(t, err)

	expected := "(\"\" string)\n\n"
	got := string(res.DeliverTx.Data)

	assert.Equal(t, expected, got)

	res, err = callSigningSeparately(t, client, baseCfg, msg)
	require.NoError(t, err)
	got = string(res.DeliverTx.Data)
	assert.Equal(t, expected, got)
}
