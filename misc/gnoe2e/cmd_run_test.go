package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"log/slog"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/misc/gnoe2e/internal/cluster"
	integ "github.com/gnolang/gno/misc/gnoe2e/internal/integration"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestGpaoApproverAddressMatchesImportedKey pins that the address placed in
// Genesis.PkgApprovers and the address of the key imported into the keybase
// are always the same key. Both come from one derivation, so a future edit
// cannot reintroduce two that silently drift apart and leave the chain with an
// approver nobody can sign for.
func TestGpaoApproverAddressMatchesImportedKey(t *testing.T) {
	gpaoKey, err := deriveGpaoKey(defaultGpaoMnemonic)
	require.NoError(t, err)
	require.NotNil(t, gpaoKey)
	genesisAddr := gpaoKey.PubKey().Address()

	userKey, err := integration.GeneratePrivKeyFromMnemonic(defaultMnemonic, "", 0, 0)
	require.NoError(t, err)

	// Through the call execRun actually makes, so this covers the path it
	// names rather than a hand-rolled stand-in for it.
	cfg := cluster.DefaultClusterConfig()
	spec := integ.ClusterSpec{Validators: 1, CodeSubmissionPolicy: "inert"}
	require.NoError(t, spec.ApplyTo(&cfg, userKey.PubKey().Address(), genesisAddr))

	kb, err := keys.NewKeyBaseFromDir(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, kb.ImportPrivKey(gpaoKeyName, gpaoKey, ""))

	imported, err := kb.GetByName(gpaoKeyName)
	require.NoError(t, err)

	require.Equal(t, genesisAddr, imported.GetAddress())
	require.Contains(t, cfg.Genesis.PkgApprovers, imported.GetAddress())
}

// The oracle's identity belongs to the run rather than to the scenarios that
// happen to start gpao: what shapes a chain is its submission policy and its
// approver set, and every script reads the address out of GPAO_ADDR whether it
// runs the oracle or enables by hand.
func TestOracleIdentityIsAlwaysProvisioned(t *testing.T) {
	cfg := &runCfg{
		keyName:      defaultKeyName,
		mnemonic:     defaultMnemonic,
		gpaoMnemonic: defaultGpaoMnemonic,
		cluster:      cluster.DefaultClusterConfig(),
	}

	ids, cleanup, err := setupIdentities(cfg, discardLogger())
	require.NoError(t, err)
	defer cleanup()

	assert.False(t, ids.userAddr.IsZero(), "the run's own key is provisioned too")
	require.False(t, ids.gpaoAddr.IsZero(), "no scenario has to ask for the oracle's address")

	// Usable as a signer, not merely present in genesis: the control arm of an
	// oracle scenario enables a package with this key directly.
	kb, err := keys.NewKeyBaseFromDir(ids.gnoHome)
	require.NoError(t, err)
	imported, err := kb.GetByName(gpaoKeyName)
	require.NoError(t, err)
	assert.Equal(t, ids.gpaoAddr, imported.GetAddress())
}

// An oracle whose key is absent from an inert chain's approver set looks
// healthy: it starts, follows blocks, verifies and reports, and every package
// it approves stays inert. The run's warning is the only symptom, so it has to
// fire where that holds and stay silent everywhere else -- the oracle's key is
// provisioned for every run now, and most runs are on chains that read no
// approver set at all.
func TestOracleWarningFiresOnlyWhereItCannotActivate(t *testing.T) {
	oracle, err := crypto.AddressFromBech32("g19rl4cm2hmr8afy4kldpxz3fka4jguq0a0u3773")
	require.NoError(t, err)
	user, err := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	require.NoError(t, err)

	tests := map[string]struct {
		genesis cluster.GenesisConfig
		want    bool
	}{
		"an inert chain the oracle approves on": {
			genesis: cluster.GenesisConfig{
				CodeSubmissionPolicy: vm.CodeSubmissionPolicyInert,
				PkgApprovers:         []crypto.Address{oracle},
			},
			want: false,
		},
		"an inert chain somebody else approves on": {
			genesis: cluster.GenesisConfig{
				CodeSubmissionPolicy: vm.CodeSubmissionPolicyInert,
				PkgApprovers:         []crypto.Address{user},
			},
			want: true,
		},
		"a chain that parks nothing to activate": {
			genesis: cluster.GenesisConfig{CodeSubmissionPolicy: vm.CodeSubmissionPolicyPermissionless},
			want:    false,
		},
		"the chain default, which is not inert either": {
			genesis: cluster.GenesisConfig{},
			want:    false,
		},
		"an inert chain reached through a path, somebody else approving": {
			genesis: cluster.GenesisConfig{
				PkgApprovers: []crypto.Address{user},
				Params:       []cluster.Override{{Key: "vm.code_submission_policy", Value: "inert"}},
			},
			want: true,
		},
		"an inert chain reached through a path, the oracle approving": {
			genesis: cluster.GenesisConfig{
				PkgApprovers: []crypto.Address{oracle},
				Params:       []cluster.Override{{Key: "vm.code_submission_policy", Value: "inert"}},
			},
			want: false,
		},
		"an approver set replaced through a path": {
			genesis: cluster.GenesisConfig{
				CodeSubmissionPolicy: vm.CodeSubmissionPolicyInert,
				PkgApprovers:         []crypto.Address{oracle},
				Params: []cluster.Override{
					{Key: "vm.pkg_approvers", Value: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},
				},
			},
			want: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, oracleCannotActivate(tt.genesis, oracle))
		})
	}
}

// A run with -timeout 0 is a run with no deadline, the way `go test -timeout 0`
// is. Handing zero to context.WithTimeout instead yields a context that is
// already past its deadline, which takes every node process down before the
// first scenario boots.
func TestRunContextTreatsZeroTimeoutAsNoLimit(t *testing.T) {
	t.Run("zero means no deadline", func(t *testing.T) {
		ctx, cancel := runContext(context.Background(), 0)
		defer cancel()

		require.NoError(t, ctx.Err())
		_, ok := ctx.Deadline()
		assert.False(t, ok, "zero must not set a deadline at all")
	})

	t.Run("the default is bounded", func(t *testing.T) {
		ctx, cancel := runContext(context.Background(), defaultRunCfg().timeout)
		defer cancel()

		require.NoError(t, ctx.Err())
		deadline, ok := ctx.Deadline()
		require.True(t, ok)
		assert.WithinDuration(t, time.Now().Add(10*time.Minute), deadline, time.Minute)
	})
}

// A flag left alone must not override what a script declared, which is why the
// overrides are read from the flags actually given rather than from their
// values.
func TestClusterOverridesReadsOnlyTheFlagsGiven(t *testing.T) {
	parse := func(args ...string) integ.ClusterOverrides {
		t.Helper()
		clusterCfg := cluster.DefaultClusterConfig()
		cfg := &runCfg{cluster: clusterCfg}
		fs := flag.NewFlagSet("run", flag.ContinueOnError)
		cfg.RegisterFlags(fs)
		require.NoError(t, fs.Parse(args))
		return cfg.clusterOverrides()
	}

	t.Run("no flags names nothing", func(t *testing.T) {
		assert.Equal(t, integ.ClusterOverrides{}, parse())
	})

	t.Run("a flag set to its own default is still an override", func(t *testing.T) {
		// The default validator count, passed explicitly. Reading the value
		// alone could not tell this from the flag being absent.
		o := parse("-validators", "2")
		require.NotNil(t, o.Validators)
		assert.Equal(t, 2, *o.Validators)
	})

	t.Run("unrelated flags name nothing", func(t *testing.T) {
		o := parse("-verbose", "-keyname", "someone")
		assert.Equal(t, integ.ClusterOverrides{}, o)
	})

	t.Run("each cluster flag reaches its own setting", func(t *testing.T) {
		o := parse(
			"-validators", "4",
			"-code-submission-policy", "inert",
			"-block-max-gas", "20000000",
		)
		require.NotNil(t, o.Validators)
		require.NotNil(t, o.CodeSubmissionPolicy)
		require.NotNil(t, o.BlockMaxGas)
		assert.Equal(t, 4, *o.Validators)
		assert.Equal(t, "inert", *o.CodeSubmissionPolicy)
		assert.Equal(t, int64(20_000_000), *o.BlockMaxGas)
	})

	t.Run("an approver given as an address is passed through", func(t *testing.T) {
		o := parse("-pkg-approver", "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
		require.NotNil(t, o.PkgApprover)
		assert.Equal(t, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", *o.PkgApprover)
	})
}

// TestDefaultMnemonicDerivesTheDocumentedApprover pins the test user's
// address, which the unauthorized-oracle suite depends on as the one approver
// its chain has. Nothing at runtime couples the two, so without this a change
// to defaultMnemonic would leave that suite with an approver nobody holds --
// failing far from the cause, inside a multi-minute scenario.
func TestDefaultMnemonicDerivesTheDocumentedApprover(t *testing.T) {
	userKey, err := integration.GeneratePrivKeyFromMnemonic(defaultMnemonic, "", 0, 0)
	require.NoError(t, err)

	require.Equal(t, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", userKey.PubKey().Address().String())
}

func TestGpaoApproverIsDistinctFromTheTestUser(t *testing.T) {
	gpaoKey, err := integration.GeneratePrivKeyFromMnemonic(defaultGpaoMnemonic, "", 0, 0)
	require.NoError(t, err)
	userKey, err := integration.GeneratePrivKeyFromMnemonic(defaultMnemonic, "", 0, 0)
	require.NoError(t, err)

	// gpao derives its signer at account 0 index 0 and offers no way to change
	// that, so sharing the run's mnemonic would make the oracle and the
	// submitter one account and put their sequence numbers in contention.
	require.NotEqual(t,
		userKey.PubKey().Address().String(),
		gpaoKey.PubKey().Address().String())
}

// A run parses its flags into one template and fills the approver set per
// scenario, from the oracle key it provisions. Validating the cross-field rule
// against that template refuses `-code-submission-policy inert` before it can
// ever be satisfied: nothing on the command line names the oracle, and
// -pkg-approver takes a bech32 address rather than a role.
func TestValidateAcceptsAnInertRunTemplate(t *testing.T) {
	cfg := defaultRunCfg()
	cfg.cluster.Genesis.CodeSubmissionPolicy = vm.CodeSubmissionPolicyInert

	require.NoError(t, cfg.validate())
}

// The cluster package reports through package-level slog calls, so the run's
// handler has to be the default one as well as the injected one. Without that,
// every diagnostic those calls make is dropped by the handler slog starts with,
// -verbose included.
func TestNewRunLoggerInstallsTheDefault(t *testing.T) {
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })

	var out bytes.Buffer
	newRunLogger(&out, true)

	slog.Debug("a package-level diagnostic")

	assert.Contains(t, out.String(), "a package-level diagnostic")
}

// TestRunScenariosStopsWhenTheRunIsCancelled pins what a run reports once its
// deadline has passed or Ctrl-C has been pressed.
//
// Every scenario after that point still builds a temp dir, validator keys and a
// genesis before os/exec refuses to start a process under a dead context, and
// each is then reported as a failure. Interrupting a fourteen-scenario run at
// the second prints twelve failures for scenarios that never ran, which is the
// opposite of what the run promises.
//
// The detail of a failure is logged where it happens, so the error that ends
// the run counts rather than repeats them: a scenario reported twice reads as
// two different failures.
func TestRunScenariosStopsWhenTheRunIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	scenarios := []integ.Scenario{{Path: "first"}, {Path: "second"}, {Path: "third"}}

	var attempted []string
	err := runScenarios(ctx, scenarios, discardLogger(), func(scen integ.Scenario) error {
		attempted = append(attempted, scen.Path)
		cancel() // the run's deadline passes while the first scenario is running
		return errors.New("the cluster went away under it")
	})

	require.Error(t, err)
	assert.Equal(t, []string{"first"}, attempted,
		"a cancelled run must not attempt the scenarios it has no time for")
	assert.Contains(t, err.Error(), "after 1 of 3 scenarios",
		"the count is what was attempted, out of what was listed")
	assert.ErrorIs(t, err, context.Canceled,
		"a run cut off partway has not answered the question it was asked")
	assert.NotContains(t, err.Error(), "the cluster went away under it",
		"the detail was already reported where it happened")
}

// A run that is not cancelled attempts every scenario, whatever the ones before
// it did: each owns its cluster, so one failure says nothing about the rest.
func TestRunScenariosRunsThemAllWhenOneFails(t *testing.T) {
	scenarios := []integ.Scenario{{Path: "first"}, {Path: "second"}, {Path: "third"}}

	attempted := 0
	err := runScenarios(t.Context(), scenarios, discardLogger(), func(scen integ.Scenario) error {
		attempted++
		if scen.Path == "first" {
			return errors.New("failed")
		}
		return nil
	})

	require.Error(t, err)
	assert.Equal(t, 3, attempted)
	assert.Contains(t, err.Error(), "1 of 3")
}

// A run has to unwind on the signal a CI runner cancels a job with, not only on
// the one a terminal sends. Left to the default disposition, SIGTERM kills the
// run where it stands: the validators of the scenario in flight keep running
// and every temp directory it made stays behind.
func TestRunSignalsCoverACancelledJob(t *testing.T) {
	ctx, stop := signal.NotifyContext(t.Context(), runSignals...)
	defer stop()

	require.NoError(t, syscall.Kill(syscall.Getpid(), syscall.SIGTERM))

	select {
	case <-ctx.Done():
	case <-time.After(10 * time.Second):
		t.Fatal("SIGTERM did not end the run")
	}
}
