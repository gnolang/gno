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

// A generic genesis param has to reach the document the node boots from, not
// just the config struct.
func TestBuildGenesisWritesParamOverrides(t *testing.T) {
	tempDir := t.TempDir()

	validator, err := SetupValidatorNode(tempDir, 0)
	require.NoError(t, err)
	defer validator.Cleanup()

	cfg := DefaultGenesisConfig()
	cfg.LoadExamples = false
	cfg.Params = []Override{
		{Key: "vm.chain_domain", Value: "example.gno.land"},
		{Key: "auth.max_memo_bytes", Value: "128"},
	}

	require.NoError(t, BuildGenesis(tempDir, cfg, []*Node{validator}))

	doc, err := bft.GenesisDocFromFile(filepath.Join(tempDir, "shared_genesis.json"))
	require.NoError(t, err)
	genState, ok := doc.AppState.(gnoland.GnoGenesisState)
	require.True(t, ok, "expected AppState to decode as gnoland.GnoGenesisState, got %T", doc.AppState)

	require.Equal(t, "example.gno.land", genState.VM.Params.ChainDomain)
	require.Equal(t, int64(128), genState.Auth.Params.MaxMemoBytes)
}

// The named keys are sugar over paths a scenario can also spell out, so the
// spelled-out one has to be applied last.
func TestBuildGenesisAppliesParamOverridesAfterTheNamedKeys(t *testing.T) {
	tempDir := t.TempDir()

	validator, err := SetupValidatorNode(tempDir, 0)
	require.NoError(t, err)
	defer validator.Cleanup()

	approver := crypto.MustAddressFromString("g19rl4cm2hmr8afy4kldpxz3fka4jguq0a0u3773")

	cfg := DefaultGenesisConfig()
	cfg.LoadExamples = false
	cfg.CodeSubmissionPolicy = vm.CodeSubmissionPolicyPermissioned
	cfg.PkgApprovers = []crypto.Address{approver}
	cfg.Params = []Override{{Key: "vm.code_submission_policy", Value: "inert"}}

	require.NoError(t, BuildGenesis(tempDir, cfg, []*Node{validator}))

	doc, err := bft.GenesisDocFromFile(filepath.Join(tempDir, "shared_genesis.json"))
	require.NoError(t, err)
	genState, ok := doc.AppState.(gnoland.GnoGenesisState)
	require.True(t, ok, "expected AppState to decode as gnoland.GnoGenesisState, got %T", doc.AppState)

	require.Equal(t, vm.CodeSubmissionPolicyInert, genState.VM.Params.CodeSubmissionPolicy)
}

// A param the module refuses is refused here rather than at boot, where it
// reads as a node that died for no stated reason.
func TestBuildGenesisRejectsAParamTheModuleRefuses(t *testing.T) {
	tempDir := t.TempDir()

	validator, err := SetupValidatorNode(tempDir, 0)
	require.NoError(t, err)
	defer validator.Cleanup()

	cfg := DefaultGenesisConfig()
	cfg.LoadExamples = false
	cfg.Params = []Override{{Key: "vm.chain_domain", Value: "not a domain"}}

	err = BuildGenesis(tempDir, cfg, []*Node{validator})

	require.Error(t, err)
	require.Contains(t, err.Error(), "genesis.vm.chain_domain")
	require.Contains(t, err.Error(), "not a domain")
}

func TestBuildGenesisRejectsAnUnknownParamKey(t *testing.T) {
	tempDir := t.TempDir()

	validator, err := SetupValidatorNode(tempDir, 0)
	require.NoError(t, err)
	defer validator.Cleanup()

	cfg := DefaultGenesisConfig()
	cfg.LoadExamples = false
	cfg.Params = []Override{{Key: "vm.nope", Value: "1"}}

	err = BuildGenesis(tempDir, cfg, []*Node{validator})

	require.Error(t, err)
	require.Contains(t, err.Error(), "genesis.vm.nope")
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
