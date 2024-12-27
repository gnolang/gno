package integration

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

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

// RunSeqShimTestscripts force running txtar integration tests sequentially using in-memory nodes
// ignoring parallel testing.
func RunInMemoryTestscripts(t *testing.T, p testscript.Params) {
	t.Helper()

	// If there's an original setup, execute it
	origSetup := p.Setup
	p.Setup = func(env *testscript.Env) error {
		if origSetup != nil {
			if err := origSetup(env); err != nil {
				return err
			}
		}

		env.Setenv("TS_INMEMORY", "true")
		env.Values[envKeyInMemory] = true
		return nil
	}

	testscript.RunT(tSeqShim{t}, p)
}
