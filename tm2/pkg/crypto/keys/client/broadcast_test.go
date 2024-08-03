package client

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func Test_execBroadcast(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		errMsg    string
		output    string
	}{
		{
			name:      "Invalid number of arguments",
			args:      []string{},
			expectErr: true,
			errMsg:    flag.ErrHelp.Error(),
		},
		{
			name:      "File not found",
			args:      []string{"non_existent_file.json"},
			expectErr: true,
			errMsg:    "open non_existent_file.json: no such file or directory",
		},
		{
			name:      "Successful broadcast",
			args:      []string{"existed_file.json"},
			expectErr: false,
			output:    "test data\nOK!\nGAS WANTED: 100\nGAS USED:   90\nHEIGHT:     12345\nEVENTS:     []\nTX HASH:    q80=\n",
		},
	}

	// Create a new test IO
	io := commands.NewTestIO()

	mockOutput := new(bytes.Buffer)
	io.SetOut(commands.WriteNopCloser(mockOutput))

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create a temporary directory for test files
			kbHome, cleanup := testutils.NewTestCaseDir(t)
			defer cleanup()

			// Create a test transaction JSON file
			txFile := filepath.Join(kbHome, "test_tx.json")
			txJSON := `{
			"type": "StdTx",
			"value": {
				"msg": [],
				"fee": {},
				"signatures": [],
				"memo": ""
			}
		}`

			err := os.WriteFile(txFile, []byte(txJSON), 0644)
			require.NoError(t, err)

			cli := &mockRPCClient{
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
			}

			// Initialize test configuration
			cfg := &BroadcastCfg{
				RootCfg: &BaseCfg{
					BaseOptions: BaseOptions{
						Home:                  kbHome,
						InsecurePasswordStdin: true,
						Remote:                "",
					},
				},
				client: cli,
			}

			if len(tc.args) > 0 && tc.args[0] == "existed_file.json" {
				tc.args[0] = txFile
			}

			err = execBroadcast(cfg, tc.args, io)

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

func Test_execBroadcast_CheckTxError(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for test files
	kbHome, cleanup := testutils.NewTestCaseDir(t)
	defer cleanup()

	// Create a test transaction JSON file
	txFile := filepath.Join(kbHome, "test_tx.json")
	txJSON := `{
		"type": "StdTx",
		"value": {
			"msg": [],
			"fee": {},
			"signatures": [],
			"memo": ""
		}
	}`

	err := os.WriteFile(txFile, []byte(txJSON), 0644)
	require.NoError(t, err)

	expectedError := errors.New("CheckTx failed")

	cli := &mockRPCClient{
		broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
			return &ctypes.ResultBroadcastTxCommit{
				CheckTx: abci.ResponseCheckTx{
					ResponseBase: abci.ResponseBase{
						Error: abci.ABCIErrorOrStringError(expectedError),
						Log:   expectedError.Error(),
					},
				},
				DeliverTx: abci.ResponseDeliverTx{},
				Hash:      []byte{0x01, 0x02, 0x03},
				Height:    12345,
			}, nil
		},
	}

	// Initialize test configuration
	cfg := &BroadcastCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Home:                  kbHome,
				InsecurePasswordStdin: true,
			},
		},
		client: cli,
	}

	// Create a new test IO
	io := commands.NewTestIO()

	// Test: CheckTx error
	args := []string{txFile}
	err = execBroadcast(cfg, args, io)
	require.Error(t, err)
	require.Contains(t, err.Error(), expectedError.Error(), "transaction failed")
}

func Test_execBroadcast_DeliverTxError(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for test files
	kbHome, cleanup := testutils.NewTestCaseDir(t)
	defer cleanup()

	// Create a test transaction JSON file
	txFile := filepath.Join(kbHome, "test_tx.json")
	txJSON := `{
		"type": "StdTx",
		"value": {
			"msg": [],
			"fee": {},
			"signatures": [],
			"memo": ""
		}
	}`

	err := os.WriteFile(txFile, []byte(txJSON), 0644)
	require.NoError(t, err)

	expectedError := errors.New("DeliverTx failed")

	cli := &mockRPCClient{
		broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
			return &ctypes.ResultBroadcastTxCommit{
				CheckTx: abci.ResponseCheckTx{},
				DeliverTx: abci.ResponseDeliverTx{
					ResponseBase: abci.ResponseBase{
						Error: abci.ABCIErrorOrStringError(expectedError),
						Log:   expectedError.Error(),
					},
				},
				Hash:   []byte{0x01, 0x02, 0x03},
				Height: 12345,
			}, nil
		},
	}

	// Initialize test configuration
	cfg := &BroadcastCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Home:                  kbHome,
				InsecurePasswordStdin: true,
			},
		},
		client: cli,
	}

	// Create a new test IO
	io := commands.NewTestIO()

	// Test: DeliverTx error
	args := []string{txFile}
	err = execBroadcast(cfg, args, io)
	require.Error(t, err)
	require.Contains(t, err.Error(), expectedError.Error(), "transaction failed")
}

func Test_BroadcastHandler(t *testing.T) {
	t.Parallel()

	cfg := &BroadcastCfg{
		tx: &std.Tx{
			Msgs:       []std.Msg{},
			Fee:        std.Fee{},
			Signatures: []std.Signature{},
			Memo:       "",
		},
		DryRun:       false,
		testSimulate: false,
		client: &mockRPCClient{
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
					Hash:   []byte{0xab, 0xcd},
					Height: 12345,
				}, nil
			},
		},
	}

	// Test: Successful broadcast
	bres, err := BroadcastHandler(cfg)
	require.NoError(t, err)
	require.NotNil(t, bres)
	require.Equal(t, []byte("test data"), bres.DeliverTx.Data)
	require.Equal(t, int64(100), bres.DeliverTx.GasWanted)
	require.Equal(t, int64(90), bres.DeliverTx.GasUsed)
	require.Equal(t, int64(12345), bres.Height)
	require.Equal(t, []byte{0xab, 0xcd}, bres.Hash)

	// Test: Invalid transaction
	cfg.tx = nil
	bres, err = BroadcastHandler(cfg)
	require.Error(t, err)
	require.Nil(t, bres)
	require.Equal(t, "invalid tx", err.Error())
}

func Test_SimulateTx(t *testing.T) {
	t.Parallel()

	cli := &mockRPCClient{
		abciQuery: func(path string, data []byte) (*ctypes.ResultABCIQuery, error) {
			return &ctypes.ResultABCIQuery{
				Response: abci.ResponseQuery{
					Value: amino.MustMarshal(&abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte("simulation data"),
						},
						GasWanted: 100,
						GasUsed:   90,
					}),
				},
			}, nil
		},
	}

	// Test: Successful simulation
	tx := []byte(`{
		"type": "StdTx",
		"value": {
			"msg": [],
			"fee": {},
			"signatures": [],
			"memo": ""
		}
	}`)
	res, err := SimulateTx(cli, tx)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, []byte("simulation data"), res.DeliverTx.Data)
	require.Equal(t, int64(100), res.DeliverTx.GasWanted)
	require.Equal(t, int64(90), res.DeliverTx.GasUsed)

	// Test: Error during simulation
	expectedError := errors.New("simulate tx")
	cli.abciQuery = func(path string, data []byte) (*ctypes.ResultABCIQuery, error) {
		return nil, expectedError
	}
	res, err = SimulateTx(cli, tx)
	require.Error(t, err)
	require.Nil(t, res)
	require.Equal(t, expectedError.Error(), err.Error())
}

func Test_BroadcastHandler_DryRun(t *testing.T) {
	t.Parallel()

	cfg := &BroadcastCfg{
		tx: &std.Tx{
			Msgs:       []std.Msg{},
			Fee:        std.Fee{},
			Signatures: []std.Signature{},
			Memo:       "",
		},
		DryRun: true,
		client: &mockRPCClient{
			abciQuery: func(path string, data []byte) (*ctypes.ResultABCIQuery, error) {
				return &ctypes.ResultABCIQuery{
					Response: abci.ResponseQuery{
						Value: amino.MustMarshal(&abci.ResponseDeliverTx{
							ResponseBase: abci.ResponseBase{
								Data: []byte("simulation data"),
							},
							GasWanted: 100,
							GasUsed:   90,
						}),
					},
				}, nil
			},
		},
	}

	// Test: DryRun mode
	bres, err := BroadcastHandler(cfg)
	require.NoError(t, err)
	require.NotNil(t, bres)
	require.Equal(t, []byte("simulation data"), bres.DeliverTx.Data)
	require.Equal(t, int64(100), bres.DeliverTx.GasWanted)
	require.Equal(t, int64(90), bres.DeliverTx.GasUsed)
}

func Test_BroadcastHandler_SimulateError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("simulate tx")
	cfg := &BroadcastCfg{
		tx: &std.Tx{
			Msgs:       []std.Msg{},
			Fee:        std.Fee{},
			Signatures: []std.Signature{},
			Memo:       "",
		},
		DryRun:       false,
		testSimulate: true,
		client: &mockRPCClient{
			abciQuery: func(path string, data []byte) (*ctypes.ResultABCIQuery, error) {
				return nil, expectedError
			},
		},
	}

	// Test: Simulation error
	bres, err := BroadcastHandler(cfg)
	require.Error(t, err)
	require.Nil(t, bres)
	require.Equal(t, expectedError.Error(), err.Error())
}
