package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeScriptFile(t *testing.T, dir, name, clusterSection string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, script(clusterSection), 0o644))
	return path
}

// Two scenarios asking for the same chain still get one each. Sharing a chain
// would carry deployed packages, balances and heights from whichever ran first
// into the next, making a result depend on its neighbours.
func TestResolveScenariosGivesEachScriptItsOwnCluster(t *testing.T) {
	dir := t.TempDir()
	a := writeScriptFile(t, dir, "a.txtar", "validators: 3\noracle: true\n")
	b := writeScriptFile(t, dir, "b.txtar", "validators: 3\noracle: true\n")
	c := writeScriptFile(t, dir, "c.txtar", "validators: 1\n")

	scenarios, err := ResolveScenarios([]string{a, b, c}, ClusterOverrides{})
	require.NoError(t, err)
	require.Len(t, scenarios, 3)

	assert.Equal(t, []Scenario{
		{Path: a, Spec: ClusterSpec{Validators: 3, Oracle: true}},
		{Path: b, Spec: ClusterSpec{Validators: 3, Oracle: true}},
		{Path: c, Spec: ClusterSpec{Validators: 1}},
	}, scenarios)
}

// The caller's order is the run order, and it does not vary between runs.
// Nothing carries between scenarios now, so this is about a readable log and a
// reproducible report rather than about what any assertion sees.
func TestResolveScenariosKeepsTheCallersOrder(t *testing.T) {
	dir := t.TempDir()
	third := writeScriptFile(t, dir, "3_third.txtar", "validators: 1\n")
	first := writeScriptFile(t, dir, "1_first.txtar", "validators: 1\n")
	other := writeScriptFile(t, dir, "0_other.txtar", "validators: 2\n")

	for range 5 {
		scenarios, err := ResolveScenarios([]string{third, other, first}, ClusterOverrides{})
		require.NoError(t, err)
		require.Len(t, scenarios, 3)
		assert.Equal(t, []string{third, other, first}, scenarioPaths(scenarios))
	}
}

// An override beats what a scenario declared, and still leaves it a chain of
// its own.
func TestResolveScenariosAppliesOverridesPerScenario(t *testing.T) {
	dir := t.TempDir()
	three := writeScriptFile(t, dir, "a_three.txtar", "validators: 3\noracle: true\n")
	one := writeScriptFile(t, dir, "b_one.txtar", "validators: 1\noracle: true\n")

	scenarios, err := ResolveScenarios([]string{three, one}, ClusterOverrides{Validators: ptr(2)})
	require.NoError(t, err)
	require.Len(t, scenarios, 2)
	for _, s := range scenarios {
		assert.Equal(t, 2, s.Spec.Validators, s.Path)
	}
}

func TestResolveScenariosReportsTheOffendingFile(t *testing.T) {
	dir := t.TempDir()
	good := writeScriptFile(t, dir, "good.txtar", "validators: 1\n")

	undeclared := filepath.Join(dir, "undeclared.txtar")
	require.NoError(t, os.WriteFile(undeclared, []byte("# no cluster here\ngnokey query auth/accounts\n"), 0o644))

	_, err := ResolveScenarios([]string{good, undeclared}, ClusterOverrides{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "undeclared.txtar")
	assert.Contains(t, err.Error(), `no "cluster" section`)

	_, err = ResolveScenarios([]string{filepath.Join(dir, "absent.txtar")}, ClusterOverrides{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "absent.txtar")
}

func scenarioPaths(scenarios []Scenario) []string {
	paths := make([]string, len(scenarios))
	for i, s := range scenarios {
		paths[i] = s.Path
	}
	return paths
}
