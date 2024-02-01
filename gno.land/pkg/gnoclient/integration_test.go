package gnoclient

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Call_Single_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	signer := newInMemorySigner(t, "tendermint_test")
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

func TestClient_Call_Multiple_Integration(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	signer := newInMemorySigner(t, "tendermint_test")
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

// todo add more integration tests.

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
