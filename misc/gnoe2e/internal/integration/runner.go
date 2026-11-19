package integration

import (
	"fmt"
	"log/slog"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/misc/gnoe2e/internal/cluster"
	"github.com/rogpeppe/go-internal/testscript"
)

// RunConfig holds resolved values for running one txtar script.
type RunConfig struct {
	// ScriptPath is the single script this run drives. One script per run,
	// because the cluster it names in RPCAddr was booted for it alone.
	ScriptPath string
	RPCAddr    string
	// RPCAddrs holds every validator's RPC address in index order. Scripts see
	// them as RPC_ADDR_0 upward. RPCAddr stays the first validator.
	RPCAddrs []string
	ChainID  string
	GnoHome  string
	UserAddr string
	KeyName  string
	Verbose  bool
	Logger   *slog.Logger
	Gpao     GpaoConfig
	// GpaoKeyName is the keybase name under which the oracle's key was
	// imported, exported to scripts as GPAO_KEY_NAME.
	GpaoKeyName string
	// Cluster is the local cluster the scripts run against, so a scenario can
	// stop and restart individual validators. Nil only when no local cluster
	// was started for the script.
	Cluster *cluster.Cluster
	// GpaoAddr is the oracle's approver address, exported to scripts as
	// GPAO_ADDR. GpaoKeyName names the keybase entry; a scenario querying the
	// approver's own balance needs the address itself.
	GpaoAddr string
}

// Run sets up testscript params and runs the script against the configured
// node.
func Run(cfg RunConfig) error {
	return runWithAdapter(NewTestscriptT(cfg.Logger, cfg.Verbose), cfg)
}

// RunT drives one script through testscript.T, which a real *testing.T
// satisfies: a go test driver can hand its own T here and get the script's
// transcript in the test output rather than through a logger of our own.
func RunT(t testscript.T, cfg RunConfig) {
	tsCmds := testscriptCmds(cfg)

	// Build testscript params
	params := testscript.Params{
		// Files rather than Dir: testscript reads a whole directory, and this
		// run owns exactly one script.
		Files: []string{cfg.ScriptPath},
		Setup: func(env *testscript.Env) error {
			env.Setenv("RPC_ADDR", cfg.RPCAddr)
			env.Setenv("CHAIN_ID", cfg.ChainID)
			env.Setenv("USER_ADDR", cfg.UserAddr)
			env.Setenv("USER_NAME", cfg.KeyName)
			env.Setenv("GNOHOME", cfg.GnoHome)
			env.Setenv("GNOROOT", gnoenv.RootDir())
			env.Setenv("GPAO_KEY_NAME", cfg.GpaoKeyName)
			env.Setenv("GPAO_ADDR", cfg.GpaoAddr)
			for i, addr := range cfg.RPCAddrs {
				env.Setenv(fmt.Sprintf("RPC_ADDR_%d", i), addr)
			}
			return nil
		},
		Cmds: tsCmds,
	}

	testscript.RunT(t, params)
}

func runWithAdapter(adapter *TestscriptT, cfg RunConfig) error {
	cfg.Logger.Info("running script", "path", cfg.ScriptPath)

	// Run in a goroutine so runtime.Goexit() from the adapter
	// only terminates that goroutine, not the caller.
	done := make(chan struct{})
	go func() {
		defer close(done)
		RunT(adapter, cfg)
	}()
	<-done

	if adapter.Failed {
		return fmt.Errorf("script failed: %s", cfg.ScriptPath)
	}

	return nil
}

// testscriptCmds builds the txtar command map. Extracted from RunT so tests
// can assert its exact registration set without running a script.
func testscriptCmds(cfg RunConfig) map[string]func(*testscript.TestScript, bool, []string) {
	tsCmds := make(map[string]func(*testscript.TestScript, bool, []string))
	tsCmds["gnokey"] = GnokeyTSCmd()
	tsCmds["sleep"] = SleepCmd()
	tsCmds["repeat"] = RepeatCmd(tsCmds)
	tsCmds["eventually"] = EventuallyCmd(tsCmds)
	tsCmds["http_get"] = HTTPGetCmd()
	tsCmds["gpao"] = GpaoTSCmd(cfg.Gpao)
	tsCmds["validator"] = ValidatorTSCmd(cfg.Cluster)
	return tsCmds
}
