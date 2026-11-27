package integration

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno_integration "github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

func TestTestdata(t *testing.T) {
	t.Parallel()

	flagSeqTS, _ := strconv.ParseBool(os.Getenv("SEQ_TS"))

	// A script about one realm or package lives next to it under examples/;
	// testdata/ holds the ones about the VM, the node or gnokey itself. See
	// examples/README.md.
	p, err := gno_integration.NewTestingParamsFromRoots(t,
		"testdata", filepath.Join(gnoenv.RootDir(), "examples"))
	require.NoError(t, err)

	if coverdir, ok := gno_integration.ResolveCoverageDir(); ok {
		err := gno_integration.SetupTestscriptsCoverage(&p, coverdir)
		require.NoError(t, err)
	}

	// Set up gnoland for testscript. Integration txtars run against in-memory
	// nodes (SetupGnolandTestscript's default): they share the process-global
	// stdlib/typecheck caches (no per-node cold reload) and are safe to run in
	// parallel, which is dramatically faster than spawning a subprocess node
	// per txtar.
	err = SetupGnolandTestscript(t, &p)
	require.NoError(t, err)

	// Parallel by default. SEQ_TS forces sequential execution, which is
	// occasionally useful for debugging or profiling a single txtar.
	if flagSeqTS {
		testscript.RunT(tSeqShim{t}, p)
	} else {
		testscript.Run(t, p)
	}
}

type tSeqShim struct{ *testing.T }

// noop Parallel method allow us to run test sequentially
func (tSeqShim) Parallel() {}

func (t tSeqShim) Run(name string, f func(testscript.T)) {
	t.T.Run(name, func(t *testing.T) {
		f(tSeqShim{t})
	})
}

func (t tSeqShim) Verbose() bool {
	return testing.Verbose()
}
