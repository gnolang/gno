package cluster

import (
	"reflect"
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The node config is where a scenario reaches for a different chain rhythm, so
// the kinds that matter there are nested sections, durations, numbers, bools
// and comma-separated lists.
func TestApplyOverrideOnNodeConfig(t *testing.T) {
	tests := map[string]struct {
		override Override
		check    func(t *testing.T, cfg *config.Config)
	}{
		"a duration in a nested section": {
			override: Override{Key: "consensus.timeout_commit", Value: "2s"},
			check: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				assert.Equal(t, 2*time.Second, cfg.Consensus.TimeoutCommit)
			},
		},
		"an int": {
			override: Override{Key: "mempool.size", Value: "200"},
			check: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				assert.Equal(t, 200, cfg.Mempool.Size)
			},
		},
		"an int64": {
			override: Override{Key: "p2p.send_rate", Value: "1024"},
			check: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				assert.Equal(t, int64(1024), cfg.P2P.SendRate)
			},
		},
		"a uint64": {
			override: Override{Key: "p2p.max_num_inbound_peers", Value: "5"},
			check: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				assert.Equal(t, uint64(5), cfg.P2P.MaxNumInboundPeers)
			},
		},
		"a bool the default leaves off": {
			override: Override{Key: "consensus.skip_timeout_commit", Value: "true"},
			check: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				assert.True(t, cfg.Consensus.SkipTimeoutCommit)
			},
		},
		"a string": {
			override: Override{Key: "p2p.seeds", Value: "abc@1.2.3.4:26656"},
			check: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				assert.Equal(t, "abc@1.2.3.4:26656", cfg.P2P.Seeds)
			},
		},
		"a list, split on commas": {
			override: Override{Key: "rpc.cors_allowed_origins", Value: "a.example,b.example"},
			check: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				assert.Equal(t, []string{"a.example", "b.example"}, cfg.RPC.CORSAllowedOrigins)
			},
		},
		"a list drops the empty items rather than keeping blanks": {
			override: Override{Key: "rpc.cors_allowed_origins", Value: "a.example,,b.example,"},
			check: func(t *testing.T, cfg *config.Config) {
				t.Helper()
				assert.Equal(t, []string{"a.example", "b.example"}, cfg.RPC.CORSAllowedOrigins)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			require.NoError(t, applyOverride(reflect.ValueOf(cfg).Elem(), "toml", tt.override))
			tt.check(t, cfg)
		})
	}
}

// genesisParams is the shape BuildGenesis exposes the three settable modules
// under, so the overrides here are spelled exactly as a scenario spells them.
type genesisParams struct {
	Auth *auth.Params `json:"auth"`
	VM   *vm.Params   `json:"vm"`
	Bank *bank.Params `json:"bank"`
}

func TestApplyOverrideOnGenesisParams(t *testing.T) {
	const (
		addrA = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
		addrB = "g19rl4cm2hmr8afy4kldpxz3fka4jguq0a0u3773"
	)

	tests := map[string]struct {
		override Override
		check    func(t *testing.T, authP *auth.Params, vmP *vm.Params, bankP *bank.Params)
	}{
		"a named string type": {
			override: Override{Key: "vm.code_submission_policy", Value: "inert"},
			check: func(t *testing.T, _ *auth.Params, vmP *vm.Params, _ *bank.Params) {
				t.Helper()
				assert.Equal(t, vm.CodeSubmissionPolicyInert, vmP.CodeSubmissionPolicy)
			},
		},
		"a plain string": {
			override: Override{Key: "vm.chain_domain", Value: "example.gno.land"},
			check: func(t *testing.T, _ *auth.Params, vmP *vm.Params, _ *bank.Params) {
				t.Helper()
				assert.Equal(t, "example.gno.land", vmP.ChainDomain)
			},
		},
		"an int64": {
			override: Override{Key: "auth.max_memo_bytes", Value: "128"},
			check: func(t *testing.T, authP *auth.Params, _ *vm.Params, _ *bank.Params) {
				t.Helper()
				assert.Equal(t, int64(128), authP.MaxMemoBytes)
			},
		},
		"an address": {
			override: Override{Key: "vm.storage_fee_collector", Value: addrA},
			check: func(t *testing.T, _ *auth.Params, vmP *vm.Params, _ *bank.Params) {
				t.Helper()
				assert.Equal(t, crypto.MustAddressFromString(addrA), vmP.StorageFeeCollector)
			},
		},
		"a list of addresses": {
			override: Override{Key: "vm.pkg_approvers", Value: addrA + "," + addrB},
			check: func(t *testing.T, _ *auth.Params, vmP *vm.Params, _ *bank.Params) {
				t.Helper()
				assert.Equal(t, []crypto.Address{
					crypto.MustAddressFromString(addrA),
					crypto.MustAddressFromString(addrB),
				}, vmP.PkgApprovers)
			},
		},
		"a gas price": {
			override: Override{Key: "auth.initial_gasprice", Value: "10ugnot/1gas"},
			check: func(t *testing.T, authP *auth.Params, _ *vm.Params, _ *bank.Params) {
				t.Helper()
				assert.Equal(t, std.GasPrice{Gas: 1, Price: std.MustParseCoin("10ugnot")}, authP.InitialGasPrice)
			},
		},
		"a list of strings": {
			override: Override{Key: "bank.restricted_denoms", Value: "ugnot,foo"},
			check: func(t *testing.T, _ *auth.Params, _ *vm.Params, bankP *bank.Params) {
				t.Helper()
				assert.Equal(t, []string{"ugnot", "foo"}, bankP.RestrictedDenoms)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			authP, vmP, bankP := auth.DefaultParams(), vm.DefaultParams(), bank.DefaultParams()
			root := genesisParams{Auth: &authP, VM: &vmP, Bank: &bankP}
			require.NoError(t, applyOverride(reflect.ValueOf(&root).Elem(), "json", tt.override))
			tt.check(t, &authP, &vmP, &bankP)
		})
	}
}

// A path that runs past a leaf comes straight out of a scenario file, and
// nothing between here and the boot recovers a panic, so it has to be an error
// on both targets.
func TestApplyOverrideRefusesAPathPastALeaf(t *testing.T) {
	t.Run("node config", func(t *testing.T) {
		cfg := config.DefaultConfig()
		var err error
		require.NotPanics(t, func() {
			err = applyOverride(reflect.ValueOf(cfg).Elem(), "toml", Override{Key: "mempool.size.max", Value: "1"})
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mempool.size.max")
	})

	t.Run("genesis params", func(t *testing.T) {
		authP, vmP, bankP := auth.DefaultParams(), vm.DefaultParams(), bank.DefaultParams()
		root := genesisParams{Auth: &authP, VM: &vmP, Bank: &bankP}
		var err error
		require.NotPanics(t, func() {
			err = applyOverride(reflect.ValueOf(&root).Elem(), "json", Override{Key: "vm.chain_domain.tld", Value: "gno"})
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "vm.chain_domain.tld")
	})
}

// A key that does not resolve, or a value the field cannot hold, is the
// scenario author's mistake: it has to name what they wrote, not just fail.
func TestApplyOverrideRefusesWhatItCannotSet(t *testing.T) {
	tests := map[string]struct {
		override Override
		wantErr  []string
	}{
		"an unknown leaf names the key": {
			override: Override{Key: "p2p.nope", Value: "1"},
			wantErr:  []string{"p2p.nope"},
		},
		"an unknown section names the key": {
			override: Override{Key: "nope.nope", Value: "1"},
			wantErr:  []string{"nope.nope"},
		},
		"a non-numeric value names the value and the field type": {
			override: Override{Key: "mempool.size", Value: "many"},
			wantErr:  []string{`"many"`, "int"},
		},
		"an unparseable duration names the value and the field type": {
			override: Override{Key: "consensus.timeout_commit", Value: "soon"},
			wantErr:  []string{`"soon"`, "time.Duration"},
		},
		"a field of a type nothing converts to names that type": {
			override: Override{Key: "tx_event_store.event_store_params", Value: "a=b"},
			wantErr:  []string{"tx_event_store.event_store_params", "EventStoreParams"},
		},
		"a segment past a scalar leaf names the key": {
			override: Override{Key: "mempool.size.max", Value: "1"},
			wantErr:  []string{"mempool.size.max", "not a section"},
		},
		"a segment past a duration leaf names the key": {
			override: Override{Key: "consensus.timeout_commit.foo", Value: "1s"},
			wantErr:  []string{"consensus.timeout_commit.foo", "not a section"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			err := applyOverride(reflect.ValueOf(cfg).Elem(), "toml", tt.override)
			require.Error(t, err)
			for _, want := range tt.wantErr {
				assert.Contains(t, err.Error(), want)
			}
		})
	}
}
