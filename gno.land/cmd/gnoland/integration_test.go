package main

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/require"
)

var integrationChainID = "intnet"

func TestIntegration_gasSponsorship(t *testing.T) {
	// Set up in-memory node
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNoopLogger(), config)
	defer node.Stop()

	// Init Signer & RPCClient
	signer1 := newInMemorySigner(t, "test1")
	signer2 := newInMemorySigner(t, "test2")
	fmt.Println(signer1)
	fmt.Println(signer2)
	rpcClient, err := rpcclient.NewHTTPClient(remoteAddr)
	require.NoError(t, err)

	_ = rpcClient

	/*
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

		expected := "(\"it works!\" string)\n\n(\"hi test argument\" string)\n\n"

		// Execute call
		res, err := client.Call(baseCfg, msg1, msg2)

		got := string(res.DeliverTx.Data)
		assert.Nil(t, err)
		assert.Equal(t, expected, got)
	*/
}

func newInMemorySigner(t *testing.T, name string) *gnoclient.SignerFromKeybase {
	t.Helper()

	entropy, err := bip39.NewEntropy(256)
	require.NoError(t, err)

	mnemonic, err := bip39.NewMnemonic(entropy)
	require.NoError(t, err)

	kb := keys.NewInMemory()
	_, err = kb.CreateAccount(name, mnemonic, "", "", uint32(0), uint32(0))
	require.NoError(t, err)

	return &gnoclient.SignerFromKeybase{
		Keybase:  kb,                 // Stores keys in memory or on disk
		Account:  name,               // Account name or bech32 format
		Password: "",                 // Password for encryption
		ChainID:  integrationChainID, // Chain ID for transaction signing
	}
}
