package client

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/testutils"
	"github.com/stretchr/testify/require"
)

func Test_execBroadcast(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for test files
	kbHome, cleanup := testutils.NewTestCaseDir(t)
	defer cleanup()

	// Create a test transaction JSON file
	txFile := filepath.Join(kbHome, "test_tx.json")
	txJSON := `{
		"type": "cosmos-sdk/StdTx",
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
				Hash:   []byte{0xab, 0xcd},
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
		cli: cli,
	}

	// Create a new test IO
	io := commands.NewTestIO()

	// Test: Invalid number of arguments
	args := []string{}
	err = execBroadcast(cfg, args, io)
	require.Error(t, err)
	require.Equal(t, err, flag.ErrHelp)

	// Test: File not found
	args = []string{"non_existent_file.json"}
	err = execBroadcast(cfg, args, io)
	require.Error(t, err)
	require.Contains(t, err.Error(), "reading tx document file non_existent_file.json")

	// Test: Successful broadcast
	args = []string{txFile}
	err = execBroadcast(cfg, args, io)
	require.NoError(t, err)

	output := io.Out()
	require.Contains(t, output, "test data")
	require.Contains(t, output, "OK!")
	require.Contains(t, output, "GAS WANTED: 100")
	require.Contains(t, output, "GAS USED: 90")
	require.Contains(t, output, "HEIGHT: 12345")
	require.Contains(t, output, "TX HASH: abcd")
}

func Test_execBroadcast_CheckTxError(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for test files
	kbHome, cleanup := testutils.NewTestCaseDir(t)
	defer cleanup()

	// Create a test transaction JSON file
	txFile := filepath.Join(kbHome, "test_tx.json")
	txJSON := `{
		"type": "cosmos-sdk/StdTx",
		"value": {
			"msg": [],
			"fee": {},
			"signatures": [],
			"memo": ""
		}
	}`
	err := os.WriteFile(txFile, []byte(txJSON), 0644)
	require.NoError(t, err)

	// Initialize test configuration
	cfg := &BroadcastCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Home:                  kbHome,
				InsecurePasswordStdin: true,
			},
		},
	}

	// Create a new test IO
	io := commands.NewTestIO()

	// Mock BroadcastHandler to return a CheckTx error
	originalBroadcastHandler := BroadcastHandler
	defer func() { BroadcastHandler = originalBroadcastHandler }()
	BroadcastHandler = func(cfg *BroadcastCfg) (*ctypes.ResultBroadcastTxCommit, error) {
		return &ctypes.ResultBroadcastTxCommit{
			CheckTx: abci.ResponseCheckTx{
				Code: 1, // non-zero indicates error
				Log:  "CheckTx failed",
			},
			DeliverTx: abci.ResponseDeliverTx{},
			Hash:      []byte{0x01, 0x02, 0x03},
		}, nil
	}

	// Test: CheckTx error
	args := []string{txFile}
	err = execBroadcast(cfg, args, io)
	require.Error(t, err)
	require.Contains(t, err.Error(), "transaction failed")
	require.Contains(t, err.Error(), "CheckTx failed")
}

func Test_execBroadcast_DeliverTxError(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for test files
	kbHome, cleanup := testutils.NewTestCaseDir(t)
	defer cleanup()

	// Create a test transaction JSON file
	txFile := filepath.Join(kbHome, "test_tx.json")
	txJSON := `{
		"type": "cosmos-sdk/StdTx",
		"value": {
			"msg": [],
			"fee": {},
			"signatures": [],
			"memo": ""
		}
	}`
	err := os.WriteFile(txFile, []byte(txJSON), 0644)
	require.NoError(t, err)

	// Initialize test configuration
	cfg := &BroadcastCfg{
		RootCfg: &BaseCfg{
			BaseOptions: BaseOptions{
				Home:                  kbHome,
				InsecurePasswordStdin: true,
			},
		},
	}

	// Create a new test IO
	io := commands.NewTestIO()

	// Mock BroadcastHandler to return a DeliverTx error
	originalBroadcastHandler := BroadcastHandler
	defer func() { BroadcastHandler = originalBroadcastHandler }()
	BroadcastHandler = func(cfg *BroadcastCfg) (*ctypes.ResultBroadcastTxCommit, error) {
		return &ctypes.ResultBroadcastTxCommit{
			CheckTx: abci.ResponseCheckTx{},
			DeliverTx: abci.ResponseDeliverTx{
				Code: 1, // non-zero indicates error
				Log:  "DeliverTx failed",
			},
			Hash: []byte{0x01, 0x02, 0x03},
		}, nil
	}

	// Test: DeliverTx error
	args := []string{txFile}
	err = execBroadcast(cfg, args, io)
	require.Error(t, err)
	require.Contains(t, err.Error(), "transaction failed")
	require.Contains(t, err.Error(), "DeliverTx failed")
}
