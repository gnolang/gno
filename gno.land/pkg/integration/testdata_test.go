package integration

import (
	"os"
	"strconv"
	"testing"

	gno_integration "github.com/gnolang/gno/gnovm/pkg/integration"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

func TestTestdata(t *testing.T) {
	t.Parallel()

	flagSeqTS, _ := strconv.ParseBool(os.Getenv("SEQ_TS"))

	p := gno_integration.NewTestingParams(t, "testdata")

	if coverdir, ok := gno_integration.ResolveCoverageDir(); ok {
		err := gno_integration.SetupTestscriptsCoverage(&p, coverdir)
		require.NoError(t, err)
	}

	// Set up gnoland for testscript
	err := SetupGnolandTestscript(t, &p)
	require.NoError(t, err)

	// Integration txtars run against in-memory nodes: they share the
	// process-global stdlib/typecheck caches (no per-node cold reload) and
	// are safe to run in parallel, which is dramatically faster than spawning
	// a subprocess node per txtar.
	mode := commandKindInMemory

	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		env.Values[envKeyExecCommand] = mode
		if origSetup != nil {
			if err := origSetup(env); err != nil {
				return err
			}
		}

		return nil
	}

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
