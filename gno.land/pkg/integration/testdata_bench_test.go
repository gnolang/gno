//go:build gnobench

package integration

import (
	"os"
	"path/filepath"
	"testing"

	gnointegration "github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

func TestBenchOpsIntegration(t *testing.T) {
	// Bench tests run sequentially to avoid profiler conflicts
	// (benchops uses global state)

	testdir := filepath.Join("testdata", "bench")

	// Skip if bench directory doesn't exist
	if _, err := os.Stat(testdir); os.IsNotExist(err) {
		t.Skip("testdata/bench directory not found")
	}

	// Check if we should update scripts
	updateScripts := os.Getenv("UPDATE_SCRIPTS") != ""

	p := gnointegration.NewTestingParams(t, testdir)
	p.UpdateScripts = updateScripts

	// Coverage setup
	if coverdir, ok := gnointegration.ResolveCoverageDir(); ok {
		err := gnointegration.SetupTestscriptsCoverage(&p, coverdir)
		require.NoError(t, err)
	}

	// Standard gnoland setup
	err := SetupGnolandTestscript(t, &p)
	require.NoError(t, err)

	// Add benchops profiling support with auto-update of txtar files
	SetupGnolandBenchInMemory(&p, testdir, updateScripts)

	// Force in-memory mode for bench tests (no RPC, runs in same process)
	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		env.Values[envKeyExecCommand] = commandKindInMemory
		if origSetup != nil {
			return origSetup(env)
		}
		return nil
	}

	// Run tests sequentially using tSeqShim (benchops uses global state)
	testscript.RunT(tSeqShim{t}, p)
}
