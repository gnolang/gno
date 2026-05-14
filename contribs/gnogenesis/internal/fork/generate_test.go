package fork

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	bftypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteTxsJSONL_RoundTrip verifies amino round-trip preserves
// std.Msg interface types in JSONL output.
func TestWriteTxsJSONL_RoundTrip(t *testing.T) {
	t.Parallel()

	// Create a tx with a concrete Msg (bank.MsgSend).
	msg := bank.MsgSend{
		FromAddress: crypto.AddressFromPreimage([]byte("sender")),
		ToAddress:   crypto.AddressFromPreimage([]byte("receiver")),
		Amount:      std.NewCoins(std.NewCoin("ugnot", 1000)),
	}
	tx := std.Tx{
		Msgs: []std.Msg{msg},
		Fee:  std.NewFee(50000, std.NewCoin("ugnot", 1000)),
	}
	original := []gnoland.TxWithMetadata{
		{
			Tx: tx,
			Metadata: &gnoland.GnoTxMetadata{
				Timestamp:   1234567890,
				BlockHeight: 42,
				ChainID:     "test-chain",
			},
		},
	}

	// Write to JSONL.
	dir := t.TempDir()
	path := filepath.Join(dir, "txs.jsonl")
	require.NoError(t, writeTxsJSONL(path, original))

	// Read back line-by-line using amino.UnmarshalJSON (the correct decoder
	// for amino-registered interfaces like std.Msg).
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	var decoded []gnoland.TxWithMetadata
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var tx gnoland.TxWithMetadata
		require.NoError(t, amino.UnmarshalJSON(line, &tx), "amino should unmarshal JSONL line")
		decoded = append(decoded, tx)
	}
	require.NoError(t, scanner.Err())

	require.Len(t, decoded, 1, "should decode exactly one tx")

	// The Msg should round-trip correctly with its type preserved.
	require.Len(t, decoded[0].Tx.Msgs, 1, "should have one msg")
	_, ok := decoded[0].Tx.Msgs[0].(bank.MsgSend)
	require.True(t, ok, "Msg should be bank.MsgSend after round-trip, got %T", decoded[0].Tx.Msgs[0])

	// Metadata should survive.
	require.NotNil(t, decoded[0].Metadata)
	assert.Equal(t, int64(42), decoded[0].Metadata.BlockHeight)
	assert.Equal(t, "test-chain", decoded[0].Metadata.ChainID)
}

// TestBuildHardforkGenesis_DefaultsGasParams asserts that buildHardforkGenesis
// populates the new gas-storage params (min_*/fixed_*_depth_100,
// iter_next_cost_flat) with code defaults when the source genesis has them
// all at zero (i.e. was generated before the gas-storage refactor). Without
// this, the resulting genesis would fail Params.Validate() on any post-
// refactor node (iter_next_cost_flat must be > 0).
func TestBuildHardforkGenesis_DefaultsGasParams(t *testing.T) {
	t.Parallel()

	// Source genesis mimicking a pre-refactor gnoland1: vm.params has the
	// original 6 fields set but none of the 7 new gas-storage fields.
	src := &bftypes.GenesisDoc{
		ChainID: "gnoland1",
		AppState: gnoland.GnoGenesisState{
			VM: vm.GenesisState{
				Params: vm.Params{
					SysNamesPkgPath:     "gno.land/r/sys/names",
					SysCLAPkgPath:       "gno.land/r/sys/cla",
					ChainDomain:         "gno.land",
					DefaultDeposit:      "600000000ugnot",
					StoragePrice:        "100ugnot",
					StorageFeeCollector: crypto.AddressFromPreimage([]byte("storage_fee_collector")),
				},
			},
		},
	}

	_, appState, err := buildHardforkGenesis(src, nil, "test-13", "gnoland1", 813643)
	require.NoError(t, err)
	require.NotNil(t, appState)

	defaults := vm.DefaultParams()
	assert.Equal(t, defaults.MinGetReadDepth100, appState.VM.Params.MinGetReadDepth100, "MinGetReadDepth100 should be defaulted")
	assert.Equal(t, defaults.MinSetReadDepth100, appState.VM.Params.MinSetReadDepth100, "MinSetReadDepth100 should be defaulted")
	assert.Equal(t, defaults.MinWriteDepth100, appState.VM.Params.MinWriteDepth100, "MinWriteDepth100 should be defaulted")
	assert.Equal(t, defaults.FixedGetReadDepth100, appState.VM.Params.FixedGetReadDepth100, "FixedGetReadDepth100 should be defaulted")
	assert.Equal(t, defaults.FixedSetReadDepth100, appState.VM.Params.FixedSetReadDepth100, "FixedSetReadDepth100 should be defaulted")
	assert.Equal(t, defaults.FixedWriteDepth100, appState.VM.Params.FixedWriteDepth100, "FixedWriteDepth100 should be defaulted")
	assert.Equal(t, defaults.IterNextCostFlat, appState.VM.Params.IterNextCostFlat, "IterNextCostFlat should be defaulted")

	// Pre-existing fields from the source must survive untouched.
	assert.Equal(t, "gno.land/r/sys/names", appState.VM.Params.SysNamesPkgPath)
	assert.Equal(t, "gno.land", appState.VM.Params.ChainDomain)

	// Validate() must now pass.
	require.NoError(t, appState.VM.Params.Validate(),
		"defaulted params should pass Validate()")
}

// TestBuildHardforkGenesis_PreservesTunedGasParams asserts that operator-tuned
// gas params (any one of the 7 non-zero) disable the default-fill entirely,
// preserving the operator's intent.
func TestBuildHardforkGenesis_PreservesTunedGasParams(t *testing.T) {
	t.Parallel()

	// Source with only IterNextCostFlat set (simulating operator who tuned
	// one field). The other 6 must stay at zero (no partial defaulting).
	src := &bftypes.GenesisDoc{
		ChainID: "gnoland1",
		AppState: gnoland.GnoGenesisState{
			VM: vm.GenesisState{
				Params: vm.Params{
					SysNamesPkgPath:  "gno.land/r/sys/names",
					SysCLAPkgPath:    "gno.land/r/sys/cla",
					ChainDomain:      "gno.land",
					DefaultDeposit:   "600000000ugnot",
					StoragePrice:     "100ugnot",
					IterNextCostFlat: 500, // operator override
				},
			},
		},
	}

	_, appState, err := buildHardforkGenesis(src, nil, "test-13", "gnoland1", 813643)
	require.NoError(t, err)
	assert.Equal(t, int64(500), appState.VM.Params.IterNextCostFlat,
		"operator tuning should be preserved")
	assert.Equal(t, int64(0), appState.VM.Params.MinGetReadDepth100,
		"defaulting should NOT kick in when any field is set")
	assert.Equal(t, int64(0), appState.VM.Params.MinWriteDepth100)
}

// TestBuildHardforkGenesis_DefaultsGasReplayMode asserts that buildHardforkGenesis
// sets GasReplayMode = "source" when the source genesis leaves it empty.
// "source" is the safe default for hardfork replay: historical txs preserve
// their original outcome rather than being re-gassed under the new VM's meter.
func TestBuildHardforkGenesis_DefaultsGasReplayMode(t *testing.T) {
	t.Parallel()

	src := &bftypes.GenesisDoc{
		ChainID:  "gnoland1",
		AppState: gnoland.GnoGenesisState{
			// GasReplayMode left unset in source
		},
	}
	_, appState, err := buildHardforkGenesis(src, nil, "test-13", "gnoland1", 813643)
	require.NoError(t, err)
	assert.Equal(t, "source", appState.GasReplayMode)
}

// TestBuildHardforkGenesis_PreservesExplicitGasReplayMode asserts that an
// explicit GasReplayMode in the source (e.g. "strict" for comparison testing
// or an operator override) is not overwritten.
func TestBuildHardforkGenesis_PreservesExplicitGasReplayMode(t *testing.T) {
	t.Parallel()

	src := &bftypes.GenesisDoc{
		ChainID: "gnoland1",
		AppState: gnoland.GnoGenesisState{
			GasReplayMode: "strict",
		},
	}
	_, appState, err := buildHardforkGenesis(src, nil, "test-13", "gnoland1", 813643)
	require.NoError(t, err)
	assert.Equal(t, "strict", appState.GasReplayMode,
		"explicit GasReplayMode must not be overwritten")
}

// TestVerifyGenesisFile_Invalid verifies that verifyGenesisFile returns an
// error for a malformed genesis file (so the calling tool can abort).
func TestVerifyGenesisFile_Invalid(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		err := verifyGenesisFile(filepath.Join(dir, "does-not-exist.json"))
		require.Error(t, err)
	})

	t.Run("malformed json", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(dir, "bad.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"not_valid": `), 0o644))
		err := verifyGenesisFile(path)
		require.Error(t, err)
	})
}
