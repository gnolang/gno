package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/amino"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
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

	tx := std.Tx{
		Fee: std.Fee{
			GasWanted: 10,
			GasFee: std.Coin{
				Amount: 10,
				Denom:  "ugnot",
			},
		},
		Signatures: nil, // no sigs
	}

	tx.Msgs = []std.Msg{
		bank.MsgSend{FromAddress: signer1.Info().GetAddress()}, // XXX: replace with NoopMsg
		bank.MsgSend{FromAddress: signer2.Info().GetAddress()},
	}

	signbz1, err := tx.GetSignBytes(integrationChainID, 1, 0)

	sig1, pubkey1, err := signer1.Keybase.Sign(signer1.Account, signer1.Password, signbz1)
	require.NoError(t, err)

	tx.Signatures = []std.Signature{
		{
			PubKey:    pubkey1,
			Signature: sig1,
		},
	}

	txFile, err := os.CreateTemp("", "")
	require.NoError(t, err)

	encodedTx, err := amino.MarshalJSON(tx)
	require.NoError(t, err)
	fmt.Println(string(encodedTx))

	_, err = txFile.Write(encodedTx)
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
