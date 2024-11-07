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

// RunGnolandTestscripts sets up and runs txtar integration tests for gnoland nodes.
// It prepares an in-memory gnoland node and initializes the necessary environment and custom commands.
// The function adapts the test setup for use with the testscript package, enabling
// the execution of gnoland and gnokey commands within txtar scripts.
//
// Refer to package documentation in doc.go for more information on commands and example txtar scripts.
func RunSeqShimTestscripts(t *testing.T, p testscript.Params) {
	testscript.RunT(tSeqShim{t}, p)
}
