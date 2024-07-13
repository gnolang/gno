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

var addr, _ = crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

func Test_SignAndBroadcastHandler(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for test files
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	defer kbCleanUp()

	// Check the keybase
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	require.NoError(t, err)

	_, err = kb.CreateAccount("test", testMnemonic, "", "", 0, 0)
	require.NoError(t, err)

	// Define test cases
	testCases := []struct {
		name      string
		cfg       MakeTxCfg
		keyName   string
		expectErr bool
		errMsg    string
	}{
		{
			name:    "Successful sign and broadcast",
			keyName: "test",
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
				Client: &mockRPCClient{
					abciQueryWithOptions: func(path string, data []byte, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
						jsonData := amino.MustMarshalJSON(&std.BaseAccount{
							Address: addr,
						})

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
			name:    "Failed to get key by name",
			keyName: "nonexistent", // non-existent key name

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
			},
			expectErr: true,
			errMsg:    "Key nonexistent not found",
		},
		{
			name:    "Failed to query account",
			keyName: "test",
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
				Client: &mockRPCClient{
					abciQueryWithOptions: func(path string, data []byte, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
						return nil, errors.New("account not found")
					},
				},
			},
			expectErr: true,
			errMsg:    "account not found",
		},
		{
			name:    "Failed to parse account data",
			keyName: "test",
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
				Client: &mockRPCClient{
					abciQueryWithOptions: func(path string, data []byte, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
						return &ctypes.ResultABCIQuery{
							Response: abci.ResponseQuery{
								ResponseBase: abci.ResponseBase{
									Data: []byte(""),
								},
							},
						}, nil
					},
				},
			},
			expectErr: true,
			errMsg:    "cannot decode empty bytes",
		},
		{
			name:    "Failed to sign transaction",
			keyName: "test",
			cfg: MakeTxCfg{
				RootCfg: &BaseCfg{
					BaseOptions: BaseOptions{
						Remote:                "localhost:26657",
						Home:                  kbHome,
						InsecurePasswordStdin: true,
					},
				},
				GasWanted: int64(1 << 60), // gas over flow
				GasFee:    "1000ugnot",
				Memo:      "test memo",
				Broadcast: true,
				Simulate:  SimulateSkip,
				ChainID:   "test-chain",
				Client: &mockRPCClient{
					abciQueryWithOptions: func(path string, data []byte, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
						jsonData := amino.MustMarshalJSON(&std.BaseAccount{
							Address: addr,
						})

						return &ctypes.ResultABCIQuery{
							Response: abci.ResponseQuery{
								ResponseBase: abci.ResponseBase{
									Data: jsonData,
								},
							},
						}, nil
					},
				},
			},
			expectErr: true,
			errMsg:    "unable to sign transaction, unable to validate transaction, gas overflow error",
		},
		{
			name:    "Failed to broadcast transaction",
			keyName: "test",
			cfg: MakeTxCfg{
				RootCfg: &BaseCfg{
					BaseOptions: BaseOptions{
						Remote:                "localhost:26657",
						Home:                  kbHome,
						InsecurePasswordStdin: true,
					},
				},
				GasWanted: 1000,
				GasFee:    "1000ugnot",
				Memo:      "test memo",
				Broadcast: true,
				Simulate:  SimulateSkip,
				ChainID:   "test-chain",
				Client: &mockRPCClient{
					abciQueryWithOptions: func(path string, data []byte, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
						jsonData := amino.MustMarshalJSON(&std.BaseAccount{
							Address: addr,
						})

						return &ctypes.ResultABCIQuery{
							Response: abci.ResponseQuery{
								ResponseBase: abci.ResponseBase{
									Data: jsonData,
								},
							},
						}, nil
					},
					broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
						return nil, errors.New("broadcast failed") // failed to broadcast
					},
				},
			},
			expectErr: true,
			errMsg:    "broadcast failed",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

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

			_, err := SignAndBroadcastHandler(&tc.cfg, tc.keyName, tx, "")

			if tc.expectErr {
				require.Error(t, err)
				require.Equal(t, tc.errMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_ExecSignAndBroadcast(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for test files
	kbHome, kbCleanUp := testutils.NewTestCaseDir(t)
	defer kbCleanUp()

	// Check the keybase
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	require.NoError(t, err)

	_, err = kb.CreateAccount("test", testMnemonic, "", "", 0, 0)
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
				Client: &mockRPCClient{
					abciQueryWithOptions: func(path string, data []byte, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
						jsonData := amino.MustMarshalJSON(&std.BaseAccount{
							Address: addr,
						})

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
			errMsg:    "RPC client has not been initialized",
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
				Client:    nil, // empty RPCClient
			},
		},
		{
			name:      "SignAndBroadcastHandler error",
			args:      []string{"test"},
			expectErr: true,
			errMsg:    "Key test not found",
			cfg: MakeTxCfg{
				RootCfg: &BaseCfg{
					BaseOptions: BaseOptions{
						Home:                  "", // keybase dir is not set => no keys found
						InsecurePasswordStdin: true,
					},
				},
				GasWanted: 100000,
				GasFee:    "1000ugnot",
				Memo:      "test memo",
				Broadcast: true,
				Simulate:  SimulateTest,
				ChainID:   "test-chain",
				Client:    nil,
			},
		},
		{
			name:      "CheckTx error",
			args:      []string{"test"},
			expectErr: true,
			output:    "",
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
				Client: &mockRPCClient{
					abciQueryWithOptions: func(path string, data []byte, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
						jsonData := amino.MustMarshalJSON(&std.BaseAccount{
							Address: addr,
						})

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
			output:    "",
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
				Client: &mockRPCClient{
					abciQueryWithOptions: func(path string, data []byte, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
						jsonData := amino.MustMarshalJSON(&std.BaseAccount{
							Address: addr,
						})

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
