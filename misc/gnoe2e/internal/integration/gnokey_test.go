package integration

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

// TestGnokeyNegationFailsWhenTheCommandUnexpectedlySucceeds guards the
// recover path in GnokeyTSCmd: a negated gnokey command whose underlying call
// succeeds must fail the script, not pass it.
func TestGnokeyNegationFailsWhenTheCommandUnexpectedlySucceeds(t *testing.T) {
	gnoHomeDir := t.TempDir()

	adapter := NewTestscriptT(testLogger(t), false)
	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"gnokey": GnokeyTSCmd(),
	}

	// "list" against a freshly created, empty keybase succeeds
	// deterministically with no RPC involved, so it stands in for any gnokey
	// subcommand that unexpectedly succeeds.
	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, "! gnokey list\n")},
		Cmds:  cmds,
		Setup: func(env *testscript.Env) error {
			env.Setenv("GNOHOME", gnoHomeDir)
			return nil
		},
	})

	require.True(t, adapter.Failed, "! gnokey must fail the script when the command unexpectedly succeeds")
}
