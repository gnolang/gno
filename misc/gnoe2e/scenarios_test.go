package main

import (
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	integ "github.com/gnolang/gno/misc/gnoe2e/internal/integration"
	"github.com/gnolang/gno/misc/gnoe2e/internal/termlog"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

// scriptT adapts *testing.T to testscript.T, whose Run takes a
// func(testscript.T) that *testing.T's own Run cannot satisfy. go-internal has
// the same shim but keeps it unexported behind testscript.Run, which builds its
// own params; the ones here come from integ.RunT.
type scriptT struct {
	*testing.T
}

func (t scriptT) Run(name string, f func(testscript.T)) {
	t.T.Run(name, func(sub *testing.T) { f(scriptT{T: sub}) })
}

func (t scriptT) Verbose() bool { return testing.Verbose() }

// testLogWriter sends the suite logger's output to t.Log, so cluster and
// oracle diagnostics land in the test's own output rather than on a stderr
// go test attributes to no test at all.
type testLogWriter struct {
	t *testing.T
}

func (w testLogWriter) Write(p []byte) (int, error) {
	// The handler ends every record with a newline and t.Log adds its own.
	w.t.Log(strings.TrimSuffix(string(p), "\n"))
	return len(p), nil
}

// TestScenarios runs every txtar scenario through the same suite and scenario
// setup the run command uses, so the go test route and the CLI route exercise
// one implementation rather than two that can drift.
func TestScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("boots real gnoland clusters; run without -short")
	}

	// go test sets the working directory to the module root.
	paths, err := filepath.Glob("testdata/*/*.txtar")
	require.NoError(t, err)
	require.NotEmpty(t, paths)

	cfg := defaultRunCfg()
	cfg.verbose = testing.Verbose()

	// No overrides: there are no flags here, so every scenario gets the
	// cluster it declares.
	scenarios, err := integ.ResolveScenarios(paths, integ.ClusterOverrides{})
	require.NoError(t, err)

	logger := slog.New(termlog.NewHandler(testLogWriter{t: t}, testing.Verbose()))

	s, err := prepareSuite(t.Context(), cfg, scenarios, logger) // builds the binaries once

	require.NoError(t, err)
	t.Cleanup(s.cleanup)

	for _, scen := range scenarios {
		name := strings.TrimSuffix(strings.TrimPrefix(scen.Path, "testdata/"), ".txtar")
		t.Run(name, func(t *testing.T) {
			// No t.Parallel: every script shares one goleveldb keybase under
			// GNOHOME, and a 4-validator cluster plus gpao already fills a
			// 4-vCPU runner.
			rc, teardown, err := prepareScenario(t.Context(), cfg, s, scen, logger)
			require.NoError(t, err)
			// t.Cleanup, not defer: testscript runs the script in a parallel
			// subtest, so a defer here would fire before the script does.
			t.Cleanup(teardown)

			integ.RunT(scriptT{T: t}, rc)
		})
	}
}
