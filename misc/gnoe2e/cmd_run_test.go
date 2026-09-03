package main

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/integration"
	"github.com/gnolang/gno/misc/gnoe2e/internal/cluster"
	integ "github.com/gnolang/gno/misc/gnoe2e/internal/integration"
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
	spec := integ.ClusterSpec{Validators: 1, Oracle: true, CodeSubmissionPolicy: "inert"}
	require.NoError(t, spec.ApplyTo(&cfg, userKey.PubKey().Address(), genesisAddr))

	kb, err := keys.NewKeyBaseFromDir(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, kb.ImportPrivKey(gpaoKeyName, gpaoKey, ""))

	imported, err := kb.GetByName(gpaoKeyName)
	require.NoError(t, err)

	require.Equal(t, genesisAddr, imported.GetAddress())
	require.Contains(t, cfg.Genesis.PkgApprovers, imported.GetAddress())
}

// Whether an oracle is provisioned follows what the scenarios ask for, not a
// flag: deriving a key nobody wants funds an account and builds a binary for
// nothing.
func TestOracleKeyFollowsWhatTheScenariosAskFor(t *testing.T) {
	cfg := &runCfg{
		keyName:      defaultKeyName,
		mnemonic:     defaultMnemonic,
		gpaoMnemonic: defaultGpaoMnemonic,
		cluster:      cluster.DefaultClusterConfig(),
	}

	t.Run("no scenario asking for one leaves it unprovisioned", func(t *testing.T) {
		scenarios := []integ.Scenario{{Spec: integ.ClusterSpec{Validators: 1}}}
		ids, cleanup, err := setupIdentities(cfg, scenarios, discardLogger())
		require.NoError(t, err)
		defer cleanup()

		assert.Nil(t, ids.gpaoKey)
		assert.True(t, ids.gpaoAddr.IsZero())
		assert.False(t, ids.userAddr.IsZero(), "the run's own key is provisioned regardless")
	})

	t.Run("one scenario asking for one is enough", func(t *testing.T) {
		scenarios := []integ.Scenario{
			{Spec: integ.ClusterSpec{Validators: 1}},
			{Spec: integ.ClusterSpec{Validators: 3, Oracle: true}},
		}
		ids, cleanup, err := setupIdentities(cfg, scenarios, discardLogger())
		require.NoError(t, err)
		defer cleanup()

		require.NotNil(t, ids.gpaoKey)
		assert.Equal(t, ids.gpaoKey.PubKey().Address(), ids.gpaoAddr)

		// Usable as a signer, not merely present in genesis: the control arm
		// of an oracle scenario enables a package with this key directly.
		kb, err := keys.NewKeyBaseFromDir(ids.gnoHome)
		require.NoError(t, err)
		imported, err := kb.GetByName(gpaoKeyName)
		require.NoError(t, err)
		assert.Equal(t, ids.gpaoAddr, imported.GetAddress())
	})
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
			"-oracle",
			"-code-submission-policy", "inert",
			"-block-max-gas", "20000000",
		)
		require.NotNil(t, o.Validators)
		require.NotNil(t, o.Oracle)
		require.NotNil(t, o.CodeSubmissionPolicy)
		require.NotNil(t, o.BlockMaxGas)
		assert.Equal(t, 4, *o.Validators)
		assert.True(t, *o.Oracle)
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
