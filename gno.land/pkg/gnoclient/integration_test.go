package gnoclient

import (
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	core_types "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testChainID string = "tendermint_test"

func TestCallSingle_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	signer := defaultInMemorySigner(t, "tendermint_test")
	rpcClient := rpcclient.NewHTTP(remoteAddr, "/websocket")

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
	signer := defaultInMemorySigner(t, "tendermint_test")
	rpcClient := rpcclient.NewHTTP(remoteAddr, "/websocket")

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

func TestMultiTxTimestamp_Integration(t *testing.T) {
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	keybase := keys.NewInMemory()
	signerOne := defaultInMemorySigner(t, testChainID)

	newAccountMnemonic := "when school roof tomato organ middle bring smile rebuild faith chase fragile increase paddle cool pink model become nation abuse advice sword mimic reduce"
	signerTwo := newInMemorySigner(t, testChainID, keybase, newAccountMnemonic, "test2")
	config.AddGenesisBalances(
		gnoland.Balance{
			Address: signerTwo.Address,
			Amount: std.Coins{
				std.Coin{
					Denom:  "ugnot",
					Amount: 1000000000,
				},
			},
		},
	)

	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()
	rpcClient := rpcclient.NewHTTP(remoteAddr, "/websocket")

	clientOne := Client{
		Signer:    signerOne,
		RPCClient: rpcClient,
	}
	clientTwo := Client{
		Signer:    signerTwo,
		RPCClient: rpcClient,
	}

	type callRes struct {
		result *core_types.ResultBroadcastTxCommit
		err    error
	}

	maxTries := 5
	for i := 1; i <= maxTries; i++ {
		resCh := make(chan callRes, 2)
		wg := new(sync.WaitGroup)
		wg.Add(2)

		sendTx := func(client Client) {
			baseCfg := BaseTxCfg{
				GasFee:    "10000ugnot",
				GasWanted: 8000000,
			}

			// Make Msg configs
			msg := MsgCall{
				PkgPath:  "gno.land/r/demo/deep/very/deep",
				FuncName: "CurrentTimeUnixNano",
			}

			res, err := client.Call(baseCfg, msg)
			resCh <- callRes{result: res, err: err}
			wg.Done()
		}

		go sendTx(clientOne)
		go sendTx(clientTwo)
		wg.Wait()
		close(resCh)

		results := make([]*core_types.ResultBroadcastTxCommit, 0, 2)
		for res := range resCh {
			if res.err != nil {
				t.Errorf("unexpected error %v", res.err)
			}
			results = append(results, res.result)
		}

		if len(results) != 2 {
			t.Errorf("expected 2 results, got %d", len(results))
		}

		if results[0].Height != results[1].Height {
			if i < maxTries {
				continue
			}
			t.Errorf("expected same height, got %d and %d", results[0].Height, results[1].Height)
		}

		extractInt := func(data []byte) int64 {
			parts := strings.Split(string(data), " ")
			numStr := parts[0][1:]
			num, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil {
				t.Errorf("unable to parse number from string %s", string(data))
			}

			return num
		}

		time1, time2 := extractInt(results[0].DeliverTx.Data), extractInt(results[1].DeliverTx.Data)
		diff := time1 - time2
		if diff < 0 {
			diff *= -1
		}

		if diff != 100 {
			if i < maxTries {
				continue
			}
			t.Errorf("expected time difference to be 100, got %d", diff)
		}

		break
	}
}

func TestSendSingle_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	signer := defaultInMemorySigner(t, "tendermint_test")
	rpcClient := rpcclient.NewHTTP(remoteAddr, "/websocket")

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
	signer := defaultInMemorySigner(t, "tendermint_test")
	rpcClient := rpcclient.NewHTTP(remoteAddr, "/websocket")

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
	signer := defaultInMemorySigner(t, "tendermint_test")
	rpcClient := rpcclient.NewHTTP(remoteAddr, "/websocket")

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
	signer := defaultInMemorySigner(t, "tendermint_test")
	rpcClient := rpcclient.NewHTTP(remoteAddr, "/websocket")

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

// todo add more integration tests:
// MsgCall with Send field populated (single/multiple)
// MsgRun with Send field populated (single/multiple)

func defaultInMemorySigner(t *testing.T, chainID string) *SignerFromKeybase {
	t.Helper()

	keybase := keys.NewInMemory()
	return newInMemorySigner(t, chainID, keybase, integration.DefaultAccount_Seed, integration.DefaultAccount_Name)
}

func newInMemorySigner(t *testing.T, chainID string, keybase keys.Keybase, mnemonic, name string) *SignerFromKeybase {
	t.Helper()

	info, err := keybase.CreateAccount(name, mnemonic, "", "", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error getting new signer: %v", err)
	}

	return &SignerFromKeybase{
		Keybase:  keybase, // Stores keys in memory or on disk
		Account:  name,    // Account name or bech32 format
		Address:  info.GetAddress(),
		Password: "",      // Password for encryption
		ChainID:  chainID, // Chain ID for transaction signing
	}
}
