package cluster

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

// TestBuildGenesisRefusesAChainNothingCanEverGoLiveOn pins the invariant on the
// params the chain really boots with, in every spelling that reaches them.
//
// The generic `genesis.` family is applied after the named fields, so a
// scenario can set the policy and the approver set without either field ever
// changing. A chain that boots inert with an empty approver set parks every
// package submitted after genesis with nobody able to enable it, and the only
// symptom is a scenario whose deploys never activate.
func TestBuildGenesisRefusesAChainNothingCanEverGoLiveOn(t *testing.T) {
	approver := crypto.MustAddressFromString("g19rl4cm2hmr8afy4kldpxz3fka4jguq0a0u3773")

	tests := map[string]func(cfg *GenesisConfig){
		"the named key, with nobody named to enable": func(cfg *GenesisConfig) {
			cfg.CodeSubmissionPolicy = vm.CodeSubmissionPolicyInert
		},
		"the path spelling, which never touches the named field": func(cfg *GenesisConfig) {
			cfg.Params = []Override{{Key: "vm.code_submission_policy", Value: "inert"}}
		},
		"a path that empties the approver set the named key filled": func(cfg *GenesisConfig) {
			cfg.CodeSubmissionPolicy = vm.CodeSubmissionPolicyInert
			cfg.PkgApprovers = []crypto.Address{approver}
			cfg.Params = []Override{{Key: "vm.pkg_approvers", Value: ""}}
		},
	}

	for name, declare := range tests {
		t.Run(name, func(t *testing.T) {
			tempDir := t.TempDir()
			validator, err := SetupValidatorNode(tempDir, 0)
			require.NoError(t, err)
			defer validator.Cleanup()

			cfg := DefaultGenesisConfig()
			cfg.LoadExamples = false
			declare(&cfg)

			err = BuildGenesis(tempDir, cfg, []*Node{validator})

			require.Error(t, err)
			require.Contains(t, err.Error(), "approver")
		})
	}
}

// The run command parses its flags into one template and fills the approver set
// per scenario, from the oracle key it provisions. The template legitimately
// holds an inert policy and nobody to enable on it, so the field-level check
// cannot be the one that enforces the invariant.
func TestValidateAcceptsInertBeforeTheApproverIsKnown(t *testing.T) {
	cfg := DefaultGenesisConfig()
	cfg.CodeSubmissionPolicy = vm.CodeSubmissionPolicyInert

	require.NoError(t, cfg.Validate())
}

func TestValidateRejectsUnknownPolicy(t *testing.T) {
	cfg := DefaultGenesisConfig()
	cfg.CodeSubmissionPolicy = "sometimes"

	err := cfg.Validate()

	require.Error(t, err)
	require.Contains(t, err.Error(), "sometimes")
}

// A cluster funds the accounts it runs with and nothing else, so a scenario
// that reads a balance sees only what the harness put there.
func TestBuildGenesisFundsOnlyTheClusterAccounts(t *testing.T) {
	tempDir := t.TempDir()

	validator, err := SetupValidatorNode(tempDir, 0)
	require.NoError(t, err)
	defer validator.Cleanup()

	user := crypto.MustAddressFromString("g19rl4cm2hmr8afy4kldpxz3fka4jguq0a0u3773")

	cfg := DefaultGenesisConfig()
	cfg.LoadExamples = false
	cfg.Balances = map[string]int64{user.String(): 1_000_000_000}

	require.NoError(t, BuildGenesis(tempDir, cfg, []*Node{validator}))

	doc, err := bft.GenesisDocFromFile(filepath.Join(tempDir, "shared_genesis.json"))
	require.NoError(t, err)
	genState, ok := doc.AppState.(gnoland.GnoGenesisState)
	require.True(t, ok, "expected AppState to decode as gnoland.GnoGenesisState, got %T", doc.AppState)

	funded := make(map[string]int64, len(genState.Balances))
	for _, balance := range genState.Balances {
		funded[balance.Address.String()] = balance.Amount.AmountOf("ugnot")
	}

	require.Len(t, funded, 2)
	require.Equal(t, int64(1_000_000_000), funded[user.String()])
	require.Contains(t, funded, doc.Validators[0].Address.String())
}

// BuildGenesis stores the app state as a value, so a summary that asks for a
// pointer prints nothing. It is the one line saying how much the chain was
// given, and a run whose packages fail to deploy for want of a balance has
// nothing else to read.
func TestPrintGenesisConfigReportsTheAppState(t *testing.T) {
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	var logged bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&logged, &slog.HandlerOptions{Level: slog.LevelDebug})))

	tempDir := t.TempDir()
	validator, err := SetupValidatorNode(tempDir, 0)
	require.NoError(t, err)
	defer validator.Cleanup()

	cfg := DefaultGenesisConfig()
	cfg.LoadExamples = false
	require.NoError(t, BuildGenesis(tempDir, cfg, []*Node{validator}))

	require.Contains(t, logged.String(), "genesis state")
}

// The deployer of the genesis packages is validator 0, which clusterBalances
// has already funded. Replacing that balance rather than adding to it leaves
// the validator holding exactly the deploy budget and nothing of its own, which
// is invisible while examples is large and a starved signer the day it is not.
func TestFundDeployerKeepsWhatTheAddressAlreadyHeld(t *testing.T) {
	addr := crypto.MustAddressFromString("g19rl4cm2hmr8afy4kldpxz3fka4jguq0a0u3773")
	balances := gnoland.NewBalances()
	balances.Set(addr, std.NewCoins(std.NewCoin("ugnot", validatorBalance)))

	fundDeployer(balances, addr, 3)

	held, ok := balances.Get(addr)
	require.True(t, ok)
	assert.Equal(t, int64(validatorBalance)+3*genesisDeployBudget, held.Amount.AmountOf("ugnot"))
}
