package integration

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/misc/gnoe2e/internal/cluster"
)

// TestValidatorCmdRefusesWhatItCannotDo covers the ways a script can ask for
// something the harness must not silently ignore.
//
// Every case here fails the script rather than passing quietly, and that is
// the point: a scenario whose outage never happened goes on to assert against
// a cluster that never lost a node, and reports success for behaviour it did
// not observe. The nil-cluster case is the one a reader is most likely to hit
// by accident -- a script run without a local cluster has none to take down.
func TestValidatorCmdRefusesWhatItCannotDo(t *testing.T) {
	tests := []struct {
		name   string
		script string
		want   string
	}{
		{
			name:   "no local cluster",
			script: "validator stop 0\n",
			want:   "no local cluster",
		},
		{
			name:   "missing index",
			script: "validator stop\n",
			want:   "usage:",
		},
		{
			name:   "unknown subcommand",
			script: "validator pause 0\n",
			want:   "unknown validator subcommand",
		},
		{
			name:   "index is not a number",
			script: "validator stop one\n",
			want:   "invalid index",
		},
		{
			name:   "negation",
			script: "! validator stop 0\n",
			want:   "does not support negation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, logBuf := bufferedTestLogger(t)
			adapter := NewTestscriptT(logger, true)

			// A nil cluster stands in for a run that started none. Everything
			// that can be judged from the script alone is checked before the
			// cluster is consulted, so those cases report their own fault
			// here rather than the missing cluster.
			cmds := map[string]func(*testscript.TestScript, bool, []string){
				"validator": ValidatorTSCmd(nil),
			}
			testscript.RunT(adapter, testscript.Params{
				Files: []string{writeScript(t, tt.script)},
				Cmds:  cmds,
			})

			require.True(t, adapter.Failed, "the script must fail")
			require.Contains(t, logBuf.String(), tt.want,
				"the failure has to say which of these went wrong")
		})
	}
}

// TestValidatorRestartNegationSucceedsWhenTheNodeDies lets a scenario assert
// that a node cannot come back: "! validator restart N" passes when the
// restarted process exits before it is ready, and the node's own stderr tail
// reaches the script's stderr so the scenario can name the reason it died.
//
// The unexpected-success half (a negated restart whose node does come up must
// fail the script) is the shared TSValidateError contract, pinned by the
// gnokey and gpao negation tests; bringing a real node to RPC readiness here
// would need a gnoland stand-in that answers amino-encoded ABCIInfo.
func TestValidatorRestartNegationSucceedsWhenTheNodeDies(t *testing.T) {
	t.Setenv("GNOE2E_FAKE_GNOLAND", "die")

	cl := &cluster.Cluster{
		BinaryPath: os.Args[0],
		Validators: []*cluster.Node{{
			Index:   0,
			DataDir: t.TempDir(),
			RPCAddr: "tcp://127.0.0.1:1",
			Genesis: "unused",
		}},
	}
	adapter := NewTestscriptT(testLogger(t), false)
	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"validator": ValidatorTSCmd(cl),
	}

	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, "! validator restart 0\nstderr 'fake gnoland: refusing to boot'\n")},
		Cmds:  cmds,
	})

	require.False(t, adapter.Failed, "! validator restart must pass when the node exits before it is ready, and its stderr must reach the script")
}
