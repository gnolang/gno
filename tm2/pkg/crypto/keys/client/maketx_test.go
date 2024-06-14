package client

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func Test_ExecSignAndBroadcast(t *testing.T) {
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
	cfg := &MakeTxCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Home:                  kbHome,
				InsecurePasswordStdin: true,
				Remote:                "",
			},
		},
		GasWanted: 100,
		GasFee:    "1",
		Memo:      "test memo",
		Broadcast: true,
		Simulate:  SimulateTest,
		ChainID:   "test-chain",
		cli:       cli,
	}

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
			errMsg:    "arguments",
		},
		{
			name:      "File not found",
			args:      []string{"non_existent_file.json"},
			expectErr: true,
			errMsg:    "open non_existent_file.json: no such file or directory",
		},
		{
			name:      "Successful sign and broadcast",
			args:      []string{"test-account", txFile},
			expectErr: false,
			output:    "test data\nOK!\nGAS WANTED: 100\nGAS USED:   90\nHEIGHT:     12345\nEVENTS:     []\nTX HASH:    q80=\n",
		},
	}

	// Create a new test IO
	io := commands.NewTestIO()

	mockOutput := new(bytes.Buffer)
	io.SetOut(commands.WriteNopCloser(mockOutput))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tx := std.Tx{} // Initialize a proper transaction object
			err = ExecSignAndBroadcast(cfg, tc.args, tx, io)

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

func Test_ExecSignAndBroadcast_CheckTxError(t *testing.T) {
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
	cfg := &MakeTxCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Home:                  kbHome,
				InsecurePasswordStdin: true,
			},
		},
		GasWanted: 100,
		GasFee:    "1",
		Memo:      "test memo",
		Broadcast: true,
		Simulate:  SimulateTest,
		ChainID:   "test-chain",
		cli:       cli,
	}

	// Create a new test IO
	io := commands.NewTestIO()

	// Test: CheckTx error
	args := []string{"test-account", txFile}
	tx := std.Tx{} // Initialize a proper transaction object
	err = ExecSignAndBroadcast(cfg, args, tx, io)
	require.Error(t, err)
	require.Contains(t, err.Error(), expectedError.Error())
}

func Test_ExecSignAndBroadcast_DeliverTxError(t *testing.T) {
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
	cfg := &MakeTxCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Home:                  kbHome,
				InsecurePasswordStdin: true,
			},
		},
		GasWanted: 100,
		GasFee:    "1",
		Memo:      "test memo",
		Broadcast: true,
		Simulate:  SimulateTest,
		ChainID:   "test-chain",
		cli:       cli,
	}

	// Create a new test IO
	io := commands.NewTestIO()

	// Test: DeliverTx error
	args := []string{"test-account", txFile}
	tx := std.Tx{} // Initialize a proper transaction object
	err = ExecSignAndBroadcast(cfg, args, tx, io)
	require.Error(t, err)
	require.Contains(t, err.Error(), expectedError.Error())
}
