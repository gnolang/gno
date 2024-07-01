package client

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/require"
)

var (
	addr, _  = crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	mnemonic = "process giant gadget pet latin sock receive exercise arctic indoor clump transfer zero increase version model defense teach hole program economy bridge enhance fade"
)

func Test_ExecSignAndBroadcast(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for test files
	kbHome, cleanup := testutils.NewTestCaseDir(t)
	defer cleanup()

	// Check the keybase
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	require.NoError(t, err)

	_, err = kb.CreateAccount("test", mnemonic, "", "", 0, 0)
	require.NoError(t, err)

	// Define test cases
	testCases := []struct {
		name      string
		cfg       MakeTxCfg
		args      []string
		expectErr bool
		errMsg    string
		output    string
	}{
		{
			name:      "Successful sign and broadcast",
			args:      []string{"test"},
			expectErr: false,
			output:    "test data\nOK!\nGAS WANTED: 100\nGAS USED:   90\nHEIGHT:     12345\nEVENTS:     []\nTX HASH:    q80=\n",
			cfg: MakeTxCfg{
				RootCfg: &BaseCfg{
					BaseOptions: BaseOptions{
						Remote:                "localhost:26657",
						Home:                  kbHome,
						InsecurePasswordStdin: true,
					},
				},
				GasWanted: 100000,
				GasFee:    "1000ugnot",
				Memo:      "test memo",
				Broadcast: true,
				Simulate:  SimulateSkip,
				ChainID:   "test-chain",
				cli: &mockRPCClient{
					abciQueryWithOptions: func(path string, data []byte, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
						var acc = std.NewBaseAccountWithAddress(addr)

						jsonData := amino.MustMarshalJSON(&acc)

						return &ctypes.ResultABCIQuery{
							Response: abci.ResponseQuery{
								ResponseBase: abci.ResponseBase{
									Data: jsonData,
								},
							},
						}, nil
					},
					broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
						return &ctypes.ResultBroadcastTxCommit{
							CheckTx: abci.ResponseCheckTx{},
							DeliverTx: abci.ResponseDeliverTx{
								ResponseBase: abci.ResponseBase{
									Data:   []byte("test data"),
									Events: []abci.Event{},
								},
								GasWanted: 100,
								GasUsed:   90,
							},
							Hash:   []byte{0xab, 0xcd}, // "q80=" at base64 format
							Height: 12345,
						}, nil
					},
				},
			},
		},
		{
			name:      "Invalid simulation mode",
			args:      []string{"test"},
			expectErr: true,
			errMsg:    "invalid simulate option: \"invalid\"",
			cfg: MakeTxCfg{
				RootCfg: &BaseCfg{
					BaseOptions: BaseOptions{
						Home:                  kbHome,
						InsecurePasswordStdin: true,
					},
				},
				GasWanted: 100000,
				GasFee:    "1000ugnot",
				Memo:      "test memo",
				Broadcast: true,
				Simulate:  "invalid", // invalid simulation mode
				ChainID:   "test-chain",
			},
		},
		{
			name:      "Invalid number of arguments",
			args:      []string{}, // empty arguments
			expectErr: true,
			errMsg:    flag.ErrHelp.Error(),
			cfg: MakeTxCfg{
				RootCfg: &BaseCfg{
					BaseOptions: BaseOptions{
						Home:                  kbHome,
						InsecurePasswordStdin: true,
					},
				},
				GasWanted: 100000,
				GasFee:    "1000ugnot",
				Memo:      "test memo",
				Broadcast: true,
				Simulate:  SimulateTest,
				ChainID:   "test-chain",
			},
		},
		{
			name:      "RPC client not initialized",
			args:      []string{"test"},
			expectErr: true,
			errMsg:    "rpcClient hasn't been initialized",
			cfg: MakeTxCfg{
				RootCfg: &BaseCfg{
					BaseOptions: BaseOptions{
						Home:                  kbHome,
						InsecurePasswordStdin: true,
					},
				},
				GasWanted: 100000,
				GasFee:    "1000ugnot",
				Memo:      "test memo",
				Broadcast: true,
				Simulate:  SimulateTest,
				ChainID:   "test-chain",
				cli:       nil, // empty RPCClient
			},
		},
		{
			name:      "SignAndBroadcastHandler error",
			args:      []string{"test"},
			expectErr: true,
			errMsg:    "rpcClient hasn't been initialized",
			cfg: MakeTxCfg{
				RootCfg: &BaseCfg{
					BaseOptions: BaseOptions{
						Home:                  "", // keybase dir is not set
						InsecurePasswordStdin: true,
					},
				},
				GasWanted: 100000,
				GasFee:    "1000ugnot",
				Memo:      "test memo",
				Broadcast: true,
				Simulate:  SimulateTest,
				ChainID:   "test-chain",
				cli:       nil,
			},
		},
		{
			name:      "CheckTx error",
			args:      []string{"test"},
			expectErr: true,
			output:    "test data\nOK!\nGAS WANTED: 100\nGAS USED:   90\nHEIGHT:     12345\nEVENTS:     []\nTX HASH:    q80=\n",
			cfg: MakeTxCfg{
				RootCfg: &BaseCfg{
					BaseOptions: BaseOptions{
						Remote:                "localhost:26657",
						Home:                  kbHome,
						InsecurePasswordStdin: true,
					},
				},
				GasWanted: 100000,
				GasFee:    "1000ugnot",
				Memo:      "test memo",
				Broadcast: true,
				Simulate:  SimulateSkip,
				ChainID:   "test-chain",
				cli: &mockRPCClient{
					abciQueryWithOptions: func(path string, data []byte, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
						var acc = std.NewBaseAccountWithAddress(addr)

						jsonData := amino.MustMarshalJSON(&acc)

						return &ctypes.ResultABCIQuery{
							Response: abci.ResponseQuery{
								ResponseBase: abci.ResponseBase{
									Data: jsonData,
								},
							},
						}, nil
					},
					broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
						err := errors.New("failed to checkTx")

						return &ctypes.ResultBroadcastTxCommit{
							CheckTx: abci.ResponseCheckTx{
								ResponseBase: abci.ResponseBase{
									Error: abci.ABCIErrorOrStringError(err),
									Log:   err.Error(),
								},
							},
						}, nil
					},
				},
			},
		},
		{
			name:      "DeliverTx error",
			args:      []string{"test"},
			expectErr: true,
			output:    "test data\nOK!\nGAS WANTED: 100\nGAS USED:   90\nHEIGHT:     12345\nEVENTS:     []\nTX HASH:    q80=\n",
			cfg: MakeTxCfg{
				RootCfg: &BaseCfg{
					BaseOptions: BaseOptions{
						Remote:                "localhost:26657",
						Home:                  kbHome,
						InsecurePasswordStdin: true,
					},
				},
				GasWanted: 100000,
				GasFee:    "1000ugnot",
				Memo:      "test memo",
				Broadcast: true,
				Simulate:  SimulateSkip,
				ChainID:   "test-chain",
				cli: &mockRPCClient{
					abciQueryWithOptions: func(path string, data []byte, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
						var acc = std.NewBaseAccountWithAddress(addr)

						jsonData := amino.MustMarshalJSON(&acc)

						return &ctypes.ResultABCIQuery{
							Response: abci.ResponseQuery{
								ResponseBase: abci.ResponseBase{
									Data: jsonData,
								},
							},
						}, nil
					},
					broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
						return &ctypes.ResultBroadcastTxCommit{
							DeliverTx: abci.ResponseDeliverTx{
								ResponseBase: abci.ResponseBase{
									Error: abci.ABCIErrorOrStringError(err),
									Log:   err.Error(),
								},
							},
						}, nil
					},
				},
			},
		},
	}

	// Create a new test IO
	io := commands.NewTestIO()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockOutput := new(bytes.Buffer)

			io.SetIn(strings.NewReader("\n"))
			io.SetOut(commands.WriteNopCloser(mockOutput))

			// Initialize a proper transaction object
			tx := std.Tx{
				Msgs: []std.Msg{
					bank.MsgSend{
						FromAddress: addr,
						ToAddress:   addr,
						Amount:      std.MustParseCoins("1000ugnot"),
					},
				},
				Fee: std.NewFee(tc.cfg.GasWanted, std.MustParseCoin(tc.cfg.GasFee)),
			}

			err = ExecSignAndBroadcast(&tc.cfg, tc.args, tx, io)

			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)

				actualOutput := mockOutput.String()

				require.Equal(t, tc.output, actualOutput)
			}
		})
	}
}
