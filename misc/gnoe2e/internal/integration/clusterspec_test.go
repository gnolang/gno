package integration

import (
	"testing"

	"github.com/gnolang/gno/misc/gnoe2e/internal/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func script(clusterSection string) []byte {
	return []byte("# a scenario\ngnokey query auth/accounts\n\n" +
		"-- cluster --\n" + clusterSection)
}

func TestParseClusterSpec(t *testing.T) {
	tests := []struct {
		name    string
		section string
		want    ClusterSpec
		wantErr string
	}{
		{
			name:    "validators alone leaves every option at its zero value",
			section: "validators: 3\n",
			want:    ClusterSpec{Validators: 3},
		},
		{
			name:    "every option set",
			section: "validators: 1\ncode-submission-policy: inert\npkg-approver: user\nblock-max-gas: 20000000\n",
			want: ClusterSpec{
				Validators:           1,
				CodeSubmissionPolicy: "inert",
				PkgApprover:          "user",
				BlockMaxGas:          20_000_000,
			},
		},
		{
			name:    "blank lines and comments are ignored",
			section: "\n# how many\nvalidators: 4\n\n# and the policy\ncode-submission-policy: inert\n",
			want:    ClusterSpec{Validators: 4, CodeSubmissionPolicy: "inert"},
		},
		{
			name:    "surrounding space is not part of the value",
			section: "  validators :  2  \n  pkg-approver :  user  \n",
			want:    ClusterSpec{Validators: 2, PkgApprover: "user"},
		},
		{
			name:    "a config key reaches the node config with its prefix stripped",
			section: "validators: 1\nconfig.consensus.timeout_commit: 500ms\nconfig.mempool.size: 200\n",
			want: ClusterSpec{
				Validators: 1,
				NodeConfig: []cluster.Override{
					{Key: "consensus.timeout_commit", Value: "500ms"},
					{Key: "mempool.size", Value: "200"},
				},
			},
		},
		{
			name:    "a genesis key reaches the genesis params with its prefix stripped",
			section: "validators: 1\ngenesis.vm.chain_domain: test.gno.land\ngenesis.auth.max_memo_bytes: 128\n",
			want: ClusterSpec{
				Validators: 1,
				GenesisParams: []cluster.Override{
					{Key: "vm.chain_domain", Value: "test.gno.land"},
					{Key: "auth.max_memo_bytes", Value: "128"},
				},
			},
		},
		{
			// Kept rather than deduplicated: the later line wins because it is
			// applied last, and that is only true if both survive parsing in
			// the order they were written.
			name:    "a repeated key keeps both entries in declaration order",
			section: "validators: 1\nconfig.mempool.size: 200\nconfig.mempool.size: 300\n",
			want: ClusterSpec{
				Validators: 1,
				NodeConfig: []cluster.Override{
					{Key: "mempool.size", Value: "200"},
					{Key: "mempool.size", Value: "300"},
				},
			},
		},
		{
			name:    "the RPC listen address belongs to the harness",
			section: "validators: 1\nconfig.rpc.laddr: tcp://0.0.0.0:26657\n",
			wantErr: `cluster section: config.rpc.laddr cannot be set: the harness assigns each node its listen ports and its peers`,
		},
		{
			name:    "the P2P listen address belongs to the harness",
			section: "validators: 1\nconfig.p2p.laddr: tcp://0.0.0.0:26656\n",
			wantErr: `cluster section: config.p2p.laddr cannot be set: the harness assigns each node its listen ports and its peers`,
		},
		{
			// ConfigureP2PTopology writes each validator a peer list of its
			// own, and the scenario's overrides are applied after it, once,
			// identically to every node. A scenario that set this would hand
			// all of them the same peers, and a set that cannot reach quorum
			// commits no blocks at all.
			name:    "the peer list belongs to the harness",
			section: "validators: 2\nconfig.p2p.persistent_peers: node@localhost:26656\n",
			wantErr: `cluster section: config.p2p.persistent_peers cannot be set: the harness assigns each node its listen ports and its peers`,
		},
		{
			name:    "a prefix with no path after it names nothing",
			section: "validators: 1\nconfig.: 200\n",
			wantErr: `cluster section: unknown key "config."`,
		},
		{
			name:    "validators is the one key with no usable zero value",
			section: "code-submission-policy: inert\n",
			wantErr: `cluster section: "validators" is required`,
		},
		{
			name:    "a cluster of no validators is not a cluster",
			section: "validators: 0\n",
			wantErr: `cluster section: validators must be at least 1, got 0`,
		},
		{
			name:    "a negative count is rejected rather than clamped",
			section: "validators: -1\n",
			wantErr: `cluster section: validators must be at least 1, got -1`,
		},
		{
			name:    "a count past what the harness will start is a typo",
			section: "validators: 17\n",
			wantErr: `cluster section: validators must be at most 16, got 17`,
		},
		{
			name:    "an unknown key is a typo, not something to ignore",
			section: "validators: 1\nvalidatorz: 2\n",
			wantErr: `cluster section: unknown key "validatorz"`,
		},
		{
			// The oracle identity is provisioned for every run and gpao is
			// started by the script, so a section still declaring it is a
			// scenario written against the older harness. Loudly refused
			// rather than accepted as a no-op.
			name:    "the retired oracle key is refused like any other unknown one",
			section: "validators: 1\noracle: true\n",
			wantErr: `cluster section: unknown key "oracle"`,
		},
		{
			name:    "a non-numeric count names the offending value",
			section: "validators: three\n",
			wantErr: `cluster section: validators "three": invalid syntax`,
		},
		{
			name:    "a line with no separator cannot be a setting",
			section: "validators: 1\ninert\n",
			wantErr: `cluster section: expected "key: value", got "inert"`,
		},
		{
			// Both generic families are text until they are applied, which
			// happens when this scenario's own cluster is built. Left to that,
			// a typo in a fourteen-scenario run is reported after the thirteen
			// before it have each booted a chain.
			name:    "a config path that resolves to nothing is a typo",
			section: "validators: 1\nconfig.mempool.sizz: 200\n",
			wantErr: `cluster section: config.mempool.sizz`,
		},
		{
			name:    "a config value the field cannot hold is caught at parse time",
			section: "validators: 1\nconfig.mempool.size: plenty\n",
			wantErr: `cluster section: config.mempool.size`,
		},
		{
			name:    "a genesis path that resolves to nothing is a typo",
			section: "validators: 1\ngenesis.vm.chain_domainn: tour.gno.land\n",
			wantErr: `cluster section: genesis.vm.chain_domainn`,
		},
		{
			name:    "a genesis value the module refuses is caught at parse time",
			section: "validators: 1\ngenesis.auth.max_memo_bytes: -1\n",
			wantErr: `cluster section: invalid params after setting genesis.auth.max_memo_bytes`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseClusterSpec(script(tt.section))
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// The section is required rather than optional, so that reading a scenario
// never leaves the cluster it runs against implicit.
func TestParseClusterSpecRequiresTheSection(t *testing.T) {
	_, err := ParseClusterSpec([]byte("# a scenario\ngnokey query auth/accounts\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), `no "cluster" section`)
}
