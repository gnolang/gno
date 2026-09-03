package cluster

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/stretchr/testify/require"
)

// TestGenesisCreation tests the genesis file creation process
func TestGenesisCreation(t *testing.T) {
	tempDir := t.TempDir()

	validator, err := SetupValidatorNode(tempDir, 0)
	require.NoError(t, err)
	defer validator.Cleanup()

	validators := []*Node{validator}
	err = BuildGenesis(tempDir, DefaultGenesisConfig(), validators)
	require.NoError(t, err)

	genesisPath := tempDir + "/shared_genesis.json"
	_, err = os.Stat(genesisPath)
	require.NoError(t, err, "shared genesis file should exist")

	err = CopySharedGenesis(tempDir, validator)
	require.NoError(t, err)
	_, err = os.Stat(validator.Genesis)
	require.NoError(t, err, "validator genesis file should exist")
}

func TestBuildGenesisWritesInertPolicy(t *testing.T) {
	tempDir := t.TempDir()

	validator, err := SetupValidatorNode(tempDir, 0)
	require.NoError(t, err)
	defer validator.Cleanup()

	approver := crypto.MustAddressFromString("g19rl4cm2hmr8afy4kldpxz3fka4jguq0a0u3773")

	cfg := DefaultGenesisConfig()
	cfg.LoadExamples = false
	cfg.CodeSubmissionPolicy = vm.CodeSubmissionPolicyInert
	cfg.PkgApprovers = []crypto.Address{approver}

	require.NoError(t, BuildGenesis(tempDir, cfg, []*Node{validator}))

	// Decode with the same loader the node uses to boot, rather than
	// substring-matching the serialized JSON.
	doc, err := bft.GenesisDocFromFile(filepath.Join(tempDir, "shared_genesis.json"))
	require.NoError(t, err)

	genState, ok := doc.AppState.(gnoland.GnoGenesisState)
	require.True(t, ok, "expected AppState to decode as gnoland.GnoGenesisState, got %T", doc.AppState)

	require.Equal(t, vm.CodeSubmissionPolicyInert, genState.VM.Params.CodeSubmissionPolicy)
	require.Contains(t, genState.VM.Params.PkgApprovers, approver)
}

func TestValidateRejectsInertWithoutApprovers(t *testing.T) {
	cfg := DefaultGenesisConfig()
	cfg.CodeSubmissionPolicy = vm.CodeSubmissionPolicyInert

	err := cfg.Validate()

	require.Error(t, err)
	require.Contains(t, err.Error(), "pkg-approver")
}

func TestValidateRejectsUnknownPolicy(t *testing.T) {
	cfg := DefaultGenesisConfig()
	cfg.CodeSubmissionPolicy = "sometimes"

	err := cfg.Validate()

	require.Error(t, err)
	require.Contains(t, err.Error(), "sometimes")
}
