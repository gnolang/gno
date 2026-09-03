package integration

import (
	"testing"

	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/misc/gnoe2e/internal/cluster"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testAddrs(t *testing.T) (user, oracle crypto.Address) {
	t.Helper()
	user, err := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	require.NoError(t, err)
	oracle, err = crypto.AddressFromBech32("g19rl4cm2hmr8afy4kldpxz3fka4jguq0a0u3773")
	require.NoError(t, err)
	return user, oracle
}

func TestClusterSpecApplyTo(t *testing.T) {
	user, oracle := testAddrs(t)

	t.Run("an unset option leaves the chain default alone", func(t *testing.T) {
		cfg := cluster.DefaultClusterConfig()
		defaultGas := cfg.Genesis.MaxGas

		require.NoError(t, ClusterSpec{Validators: 3}.ApplyTo(&cfg, user, oracle))

		assert.Equal(t, 3, cfg.NumValidators)
		assert.Equal(t, defaultGas, cfg.Genesis.MaxGas)
		assert.Empty(t, string(cfg.Genesis.CodeSubmissionPolicy))
	})

	t.Run("options reach the genesis they belong to", func(t *testing.T) {
		cfg := cluster.DefaultClusterConfig()
		spec := ClusterSpec{
			Validators:           1,
			Oracle:               true,
			CodeSubmissionPolicy: "inert",
			BlockMaxGas:          20_000_000,
		}
		require.NoError(t, spec.ApplyTo(&cfg, user, oracle))

		assert.Equal(t, 1, cfg.NumValidators)
		assert.Equal(t, vm.CodeSubmissionPolicyInert, cfg.Genesis.CodeSubmissionPolicy)
		assert.Equal(t, int64(20_000_000), cfg.Genesis.MaxGas)
	})

	// The generic keys are the one part of the spec the runner does not resolve
	// further: they reach the cluster as the text a scenario wrote, in order,
	// for the cluster to apply to the node config and to genesis.
	t.Run("the generic keys reach the cluster as declared", func(t *testing.T) {
		cfg := cluster.DefaultClusterConfig()
		spec := ClusterSpec{
			Validators: 1,
			NodeConfig: []cluster.Override{
				{Key: "consensus.timeout_commit", Value: "2s"},
				{Key: "mempool.size", Value: "200"},
			},
			GenesisParams: []cluster.Override{{Key: "vm.chain_domain", Value: "example.gno.land"}},
		}
		require.NoError(t, spec.ApplyTo(&cfg, user, oracle))

		assert.Equal(t, spec.NodeConfig, cfg.NodeConfig)
		assert.Equal(t, spec.GenesisParams, cfg.Genesis.Params)
	})

	t.Run("the oracle approves when no other role is named", func(t *testing.T) {
		cfg := cluster.DefaultClusterConfig()
		spec := ClusterSpec{Validators: 1, Oracle: true, CodeSubmissionPolicy: "inert"}
		require.NoError(t, spec.ApplyTo(&cfg, user, oracle))

		assert.Equal(t, []crypto.Address{oracle}, cfg.Genesis.PkgApprovers)
	})

	t.Run("naming the user as approver is what leaves the oracle unauthorized", func(t *testing.T) {
		cfg := cluster.DefaultClusterConfig()
		spec := ClusterSpec{Validators: 1, Oracle: true, CodeSubmissionPolicy: "inert", PkgApprover: "user"}
		require.NoError(t, spec.ApplyTo(&cfg, user, oracle))

		assert.Equal(t, []crypto.Address{user}, cfg.Genesis.PkgApprovers,
			"the oracle must be absent, or the scenario cannot separate refusal from inability")
	})

	// An inert chain nobody can enable on is a chain where nothing submitted
	// after genesis can ever go live, so it is refused rather than booted.
	t.Run("inert without an oracle needs an approver named", func(t *testing.T) {
		cfg := cluster.DefaultClusterConfig()
		spec := ClusterSpec{Validators: 1, CodeSubmissionPolicy: "inert"}
		err := spec.ApplyTo(&cfg, user, oracle)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "inert")
		assert.Contains(t, err.Error(), "pkg-approver")
	})

	// A command-line override names an address where a script names a role, so
	// both have to resolve.
	t.Run("an approver given as an address is used as given", func(t *testing.T) {
		cfg := cluster.DefaultClusterConfig()
		third, err := crypto.AddressFromBech32("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq")
		require.NoError(t, err)

		spec := ClusterSpec{Validators: 1, CodeSubmissionPolicy: "inert", PkgApprover: third.String()}
		require.NoError(t, spec.ApplyTo(&cfg, user, oracle))

		assert.Equal(t, []crypto.Address{third}, cfg.Genesis.PkgApprovers)
	})

	t.Run("neither a role nor an address is refused rather than ignored", func(t *testing.T) {
		cfg := cluster.DefaultClusterConfig()
		spec := ClusterSpec{Validators: 1, PkgApprover: "treasurer"}
		err := spec.ApplyTo(&cfg, user, oracle)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `"treasurer"`)
	})

	t.Run("the resulting config is one the cluster will accept", func(t *testing.T) {
		cfg := cluster.DefaultClusterConfig()
		spec := ClusterSpec{Validators: 4, Oracle: true, CodeSubmissionPolicy: "inert"}
		require.NoError(t, spec.ApplyTo(&cfg, user, oracle))
		assert.NoError(t, cfg.Validate())
	})
}
