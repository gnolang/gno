package gnoclient

import (
	"github.com/jaekwon/testify/assert"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/jaekwon/testify/require"
)

func newInMemorySigner(t *testing.T, chainid string) *SignerFromKeybase {
	t.Helper()

	mmeonic := integration.DefaultAccount_Seed
	name := integration.DefaultAccount_Name

	kb := keys.NewInMemory()
	_, err := kb.CreateAccount(name, mmeonic, "", "", uint32(0), uint32(0))
	require.NoError(t, err)

	return &SignerFromKeybase{
		Keybase:  kb,      // Stores keys in memory or on disk
		Account:  name,    // Account name or bech32 format
		Password: "",      // Password for encryption
		ChainID:  chainid, // Chain ID for transaction signing
	}
}

func TestClient_Request(t *testing.T) {
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNopLogger(), config)
	defer node.Stop()

	signer := newInMemorySigner(t, config.TMConfig.ChainID())

	client := Client{
		Signer:    signer,
		RPCClient: rpcclient.NewHTTP(remoteAddr, "/websocket"),
	}

	data, res, err := client.Render("gno.land/r/demo/boards", "")
	require.NoError(t, err)
	require.NotEmpty(t, data)

	require.NotNil(t, res)
	require.NotEmpty(t, res.Response.Data)

	// XXX: need more test
}

func TestClient_Call(t *testing.T) {
	t.Parallel()

	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNopLogger(), config)
	defer node.Stop()

	signer := newInMemorySigner(t, config.TMConfig.ChainID())
	rpcClient := rpcclient.NewHTTP(remoteAddr, "/websocket")

	client := Client{
		Signer:    signer,
		RPCClient: rpcClient,
	}

	cfg := BaseTxCfg{
		GasWanted:      100000,
		GasFee:         "10000ugnot",
		AccountNumber:  1,
		SequenceNumber: 1,
		Memo:           "Test memo",
	}

	msg := []MsgCall{
		{
			PkgPath:  "gno.land/r/demo/deep/very/deep",
			FuncName: "Render",
			Args:     []string{""},
			Send:     "100ugnot",
		},
	}

	res, err := client.Call(cfg, msg...)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestClient_Call_Errors(t *testing.T) {
	t.Parallel()

	// todo Replace with mock client
	config, _ := integration.TestingNodeConfig(t, gnoenv.RootDir())
	node, remoteAddr := integration.TestingInMemoryNode(t, log.NewNopLogger(), config)
	defer node.Stop()

	signer := newInMemorySigner(t, config.TMConfig.ChainID())
	rpcClient := rpcclient.NewHTTP(remoteAddr, "/websocket")

	testCases := []struct {
		name          string
		client        Client
		cfg           BaseTxCfg
		msgs          []MsgCall
		expectedError error
	}{
		{
			name: "Invalid Signer",
			client: Client{
				nil,
				rpcClient,
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []MsgCall{
				{
					PkgPath:  "random/path",
					FuncName: "RandomName",
					Send:     "",
					Args:     []string{},
				},
			},
			expectedError: errMissingSigner,
		},
		{
			name: "Invalid RPCClient",
			client: Client{
				signer,
				nil,
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []MsgCall{
				{
					PkgPath:  "random/path",
					FuncName: "RandomName",
					Send:     "",
					Args:     []string{},
				},
			},
			expectedError: errMissingRPCClient,
		},
		{
			name: "Invalid Gas Fee",
			client: Client{
				signer,
				rpcClient,
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         "",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []MsgCall{
				{
					PkgPath:  "random/path",
					FuncName: "RandomName",
				},
			},
			expectedError: errInvalidGasFee,
		},
		{
			name: "Negative Gas Wanted",
			client: Client{
				signer,
				rpcClient,
			},
			cfg: BaseTxCfg{
				GasWanted:      -1,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []MsgCall{
				{
					PkgPath:  "random/path",
					FuncName: "RandomName",
					Send:     "",
					Args:     []string{},
				},
			},
			expectedError: errInvalidGasWanted,
		},
		{
			name: "0 Gas Wanted",
			client: Client{
				signer,
				rpcClient,
			},
			cfg: BaseTxCfg{
				GasWanted:      0,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []MsgCall{
				{
					PkgPath:  "random/path",
					FuncName: "RandomName",
					Send:     "",
					Args:     []string{},
				},
			},
			expectedError: errInvalidGasWanted,
		},
		{
			name: "Invalid PkgPath",
			client: Client{
				signer,
				rpcClient,
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []MsgCall{
				{
					PkgPath:  "",
					FuncName: "RandomName",
					Send:     "",
					Args:     []string{},
				},
			},
			expectedError: errInvalidPkgPath,
		},
		{
			name: "Invalid FuncName",
			client: Client{
				signer,
				rpcClient,
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []MsgCall{
				{
					PkgPath:  "random/path",
					FuncName: "",
					Send:     "",
					Args:     []string{},
				},
			},
			expectedError: errInvalidFuncName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.Call(tc.cfg, tc.msgs...)
			assert.Equal(t, err, tc.expectedError)
			assert.Nil(t, res)
		})
	}
}
