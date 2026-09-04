package cluster

import (
	"reflect"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/stretchr/testify/require"
)

// The listing is only useful if every key it prints is a key a scenario can
// set, so each one is resolved against the target it was read from.
func TestConfigDefaultsResolveAgainstTheNodeConfig(t *testing.T) {
	root := reflect.ValueOf(DefaultNodeConfig()).Elem()

	defaults := ConfigDefaults()
	require.NotEmpty(t, defaults)

	for _, o := range defaults {
		_, err := resolveField(root, nodeConfigSelector, o.Key)
		require.NoError(t, err, "config.%s", o.Key)
	}
}

func TestGenesisDefaultsResolveAgainstTheGenesisParams(t *testing.T) {
	genState := gnoland.DefaultGenState()
	root := reflect.ValueOf(newSettableParams(&genState)).Elem()

	defaults := GenesisDefaults()
	require.NotEmpty(t, defaults)

	for _, o := range defaults {
		_, err := resolveField(root, genesisParamsSelector, o.Key)
		require.NoError(t, err, "genesis.%s", o.Key)
	}
}

// The values are the ones a cluster boots with, not the ones tm2 ships: the
// harness overwrites the consensus timings before a node starts.
func TestConfigDefaultsReportTheHarnessTimings(t *testing.T) {
	defaults := make(map[string]string, len(ConfigDefaults()))
	for _, o := range ConfigDefaults() {
		defaults[o.Key] = o.Value
	}

	require.Equal(t, "10ms", defaults["consensus.timeout_commit"])
	require.Equal(t, "true", defaults["consensus.create_empty_blocks"])
	require.Equal(t, "true", defaults["p2p.allow_duplicate_ip"])
}

func TestGenesisDefaultsReportTheChainDomain(t *testing.T) {
	defaults := make(map[string]string, len(GenesisDefaults()))
	for _, o := range GenesisDefaults() {
		defaults[o.Key] = o.Value
	}

	require.Equal(t, "gno.land", defaults["vm.chain_domain"])
	require.Equal(t, "65536", defaults["auth.max_memo_bytes"])
}

// A list is stated as one comma-separated value, so it has to print that way
// to be pasted back.
func TestFieldDefaultsPrintListsCommaSeparated(t *testing.T) {
	target := struct {
		Names []string `json:"names"`
	}{Names: []string{"first", "second"}}

	defaults := fieldDefaults(reflect.ValueOf(&target).Elem(), "json")

	require.Equal(t, []Override{{Key: "names", Value: "first,second"}}, defaults)
}
