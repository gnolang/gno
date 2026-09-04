package main

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
)

func defaultsOutput(t *testing.T, args []string) string {
	t.Helper()

	var buf bytes.Buffer
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(&buf))

	require.NoError(t, execDefaults(io, args))
	return buf.String()
}

// The listing answers one question -- what can this scenario declare -- so
// both key families and the named keys are in it.
func TestDefaultsListsEveryClusterKey(t *testing.T) {
	out := defaultsOutput(t, nil)

	require.Contains(t, out, "validators:")
	require.Contains(t, out, "block-max-gas: 3000000000")
	require.Contains(t, out, "genesis.vm.chain_domain: gno.land")
	require.Contains(t, out, "config.consensus.timeout_commit: 10ms")
}

// A refused key is listed rather than left out: a scenario author who tried it
// needs to read why, not wonder whether the key exists.
func TestDefaultsMarksTheAddressesTheHarnessAssigns(t *testing.T) {
	out := defaultsOutput(t, nil)

	require.Regexp(t, `config\.rpc\.laddr:.*cannot be set`, out)
	require.Regexp(t, `config\.p2p\.laddr:.*cannot be set`, out)
	require.Regexp(t, `config\.p2p\.persistent_peers:.*cannot be set`, out)
}

func TestDefaultsPrintsOneFamilyOnRequest(t *testing.T) {
	out := defaultsOutput(t, []string{"genesis"})

	require.Contains(t, out, "genesis.vm.chain_domain: gno.land")
	require.NotContains(t, out, "config.")
	require.NotContains(t, out, "validators:")
}

func TestDefaultsRefusesAnUnknownFamily(t *testing.T) {
	err := execDefaults(commands.NewTestIO(), []string{"nodes"})

	require.Error(t, err)
	require.Contains(t, err.Error(), `"nodes"`)
}
