package fork

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/stretchr/testify/require"

	vmm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
)

// writeTestGenesis writes a minimal but valid genesis.json to a temp file.
// It uses a fresh private validator so the genesis is self-contained.
func writeTestGenesis(t *testing.T, appState gnoland.GnoGenesisState) string {
	t.Helper()

	pv := bft.NewMockPV()
	pk := pv.PubKey()

	genDoc := bft.GenesisDoc{
		GenesisTime: time.Now(),
		ChainID:     "test-hardfork-1",
		ConsensusParams: abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxTxBytes:   1_000_000,
				MaxDataBytes: 2_000_000,
				MaxGas:       3_000_000_000,
				TimeIotaMS:   100,
			},
		},
		Validators: []bft.GenesisValidator{
			{
				Address: pk.Address(),
				PubKey:  pk,
				Power:   10,
				Name:    "test-validator",
			},
		},
		AppState: appState,
	}

	data, err := amino.MarshalJSONIndent(genDoc, "", "  ")
	require.NoError(t, err)

	dir := t.TempDir()
	path := filepath.Join(dir, "genesis.json")
	require.NoError(t, os.WriteFile(path, data, 0o644))
	return path
}

func minimalAppState() gnoland.GnoGenesisState {
	return gnoland.GnoGenesisState{
		Balances: []gnoland.Balance{},
		Txs:      []gnoland.TxWithMetadata{},
		Auth:     auth.DefaultGenesisState(),
		Bank:     bank.DefaultGenesisState(),
		VM:       vmm.DefaultGenesisState(),
	}
}

// TestTestCfg_FlagDefaults asserts the default values of fork test's
// command-line flags. The defaults must match production node behavior:
// sig verification skipped (mirroring -skip-genesis-sig-verification on
// gnoland), but failing txs do fail the test (explicit opt-in required
// via --skip-failing-genesis-txs for parity with the cluster's permissive
// mode).
func TestTestCfg_FlagDefaults(t *testing.T) {
	t.Parallel()

	cfg := &testCfg{}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg.RegisterFlags(fs)
	require.NoError(t, fs.Parse(nil))

	// Default ON: sig verification must be skipped to not trip historical
	// txs whose deployer keys don't match msg.Creator (manfred et al).
	require.True(t, cfg.skipGenesisSigVerification,
		"skip-genesis-sig-verification should default to true")

	// Default OFF: failing txs should still fail the test by default
	// (strictest posture for CI). Operators match the cluster flag
	// explicitly when the chain tolerates genesis tx failures.
	require.False(t, cfg.skipFailingTxs,
		"skip-failing-genesis-txs should default to false")

	// Override parse.
	require.NoError(t, fs.Parse([]string{
		"--skip-genesis-sig-verification=false",
		"--skip-failing-genesis-txs=true",
	}))
	require.False(t, cfg.skipGenesisSigVerification)
	require.True(t, cfg.skipFailingTxs)
}

// TestExecTest_MissingGenesis verifies that a missing genesis file is caught.
func TestExecTest_MissingGenesis(t *testing.T) {
	io := commands.NewTestIO()
	cfg := &testCfg{
		genesis: "/nonexistent/path/genesis.json",
		timeout: 5 * time.Second,
	}
	err := execTest(context.Background(), cfg, io)
	require.ErrorContains(t, err, "reading genesis file")
}

// TestExecTest_InvalidGenesis verifies that a malformed genesis file is caught.
func TestExecTest_InvalidGenesis(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.json")
	require.NoError(t, os.WriteFile(bad, []byte(`{"not_valid": "json"`), 0o644))

	io := commands.NewTestIO()
	cfg := &testCfg{
		genesis: bad,
		timeout: 5 * time.Second,
	}
	err := execTest(context.Background(), cfg, io)
	require.ErrorContains(t, err, "parsing genesis")
}

// TestExecTest_EmptyGenesis runs a full in-process replay with an empty genesis
// (no transactions). This verifies the happy path without requiring network access.
//
// This test is skipped in short mode (-short) because loading stdlibs takes ~30s.
func TestExecTest_EmptyGenesis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode — requires loading stdlibs (~30s)")
	}

	// Ensure GNOROOT is set (required for stdlibs).
	// If running from the repo root, gnoenv.GuessRootDir() will find it via go list.
	path := writeTestGenesis(t, minimalAppState())

	io := commands.NewTestIO()
	cfg := &testCfg{
		genesis: path,
		timeout: 3 * time.Minute,
	}

	err := execTest(context.Background(), cfg, io)
	require.NoError(t, err, "empty genesis replay should succeed")
}

// TestExecTest_HardforkGenesis builds a minimal hardfork genesis (with
// PastChainIDs and InitialHeight set) and verifies it can be replayed.
//
// This test is skipped in short mode (-short) because loading stdlibs takes ~30s.
func TestExecTest_HardforkGenesis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode — requires loading stdlibs (~30s)")
	}

	appState := minimalAppState()
	appState.PastChainIDs = []string{"test-hardfork-source"}

	pv := bft.NewMockPV()
	pk := pv.PubKey()

	genDoc := bft.GenesisDoc{
		GenesisTime:   time.Now(),
		ChainID:       "test-hardfork-1",
		InitialHeight: 100, // hardfork starts at block 100
		ConsensusParams: abci.ConsensusParams{
			Block: &abci.BlockParams{
				MaxTxBytes:   1_000_000,
				MaxDataBytes: 2_000_000,
				MaxGas:       3_000_000_000,
				TimeIotaMS:   100,
			},
		},
		Validators: []bft.GenesisValidator{
			{
				Address: pk.Address(),
				PubKey:  pk,
				Power:   10,
				Name:    "test-validator",
			},
		},
		AppState: appState,
	}

	data, err := amino.MarshalJSONIndent(genDoc, "", "  ")
	require.NoError(t, err)

	dir := t.TempDir()
	path := filepath.Join(dir, "genesis.json")
	require.NoError(t, os.WriteFile(path, data, 0o644))

	io := commands.NewTestIO()
	cfg := &testCfg{
		genesis: path,
		timeout: 3 * time.Minute,
	}

	err = execTest(context.Background(), cfg, io)
	require.NoError(t, err, "hardfork genesis replay should succeed")
}
