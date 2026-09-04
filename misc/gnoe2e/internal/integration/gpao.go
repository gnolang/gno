package integration

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gnolang/gno/misc/gnoe2e/internal/cluster"
	"github.com/gnolang/gno/misc/gnoe2e/internal/daemon"
	"github.com/rogpeppe/go-internal/testscript"
)

const gpaoReadyWait = 30 * time.Second

// GpaoConfig carries what stays fixed for a whole run. Per-invocation flags
// come from the script.
type GpaoConfig struct {
	// Binary yields the oracle binary, building it if the run has not needed
	// one yet. A provider rather than a path, because a run whose scripts
	// never start the oracle must not pay to build it.
	Binary   func() (string, error)
	Mnemonic string
	ChainID  string
	Remote   string // -remote used when the script names no node
	GnoRoot  string
}

// GpaoTSCmd returns the "gpao" testscript command, which accepts start, stop
// and restart.
//
// The oracle's lifetime is owned here rather than in the script so that a
// script that fails part way still leaves no daemon behind, and so that
// readiness is a probe rather than a sleep.
func GpaoTSCmd(cfg GpaoConfig) func(*testscript.TestScript, bool, []string) {
	r := &gpaoRunner{cfg: cfg}
	return r.command
}

type gpaoRunner struct {
	cfg GpaoConfig

	mu sync.Mutex
	d  *daemon.Daemon
}

func (r *gpaoRunner) command(ts *testscript.TestScript, neg bool, args []string) {
	if len(args) == 0 {
		ts.Fatalf("usage: gpao start|stop|restart [flags...]")
	}

	switch args[0] {
	case "start":
		r.start(ts, neg, args[1:])
	case "stop":
		if neg {
			ts.Fatalf("gpao stop does not support negation")
		}
		r.stop(ts)
	case "restart":
		if neg {
			ts.Fatalf("gpao restart does not support negation")
		}
		r.stop(ts)
		r.start(ts, false, args[1:])
	default:
		ts.Fatalf("unknown gpao subcommand %q", args[0])
	}
}

// start launches the oracle, building the binary if this is the run's first
// start. Under neg the script is asserting that the oracle itself refuses to
// come up, so only the launch is negatable: the guards below fail the script
// whatever neg says, because a script that is already running an oracle, or a
// binary that will not compile, is a fault in the scenario or the checkout
// rather than the failure it meant to observe.
func (r *gpaoRunner) start(ts *testscript.TestScript, neg bool, flags []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.d != nil {
		ts.Fatalf("gpao: already running; stop it first")
	}
	if r.cfg.Binary == nil {
		ts.Fatalf("gpao: no binary provider configured for this run")
	}
	binary, err := r.cfg.Binary()
	if err != nil {
		ts.Fatalf("gpao: build: %v", err)
	}

	port, err := cluster.FindAvailablePort()
	if err != nil {
		ts.Fatalf("gpao: %v", err)
	}
	listen := fmt.Sprintf("127.0.0.1:%d", port)

	d, err := daemon.Start(context.Background(), daemon.Config{
		Name:       "gpao",
		BinaryPath: binary,
		Args:       gpaoArgs(r.cfg, listen, flags),
		Env:        []string{"GPAO_MNEMONIC=" + r.cfg.Mnemonic},
		Ready:      statusProbe(statusURL(listen)),
		ReadyWait:  gpaoReadyWait,
	})
	if err == nil && neg {
		// The script expected this to fail. Stop the daemon we just started
		// before TSValidateError reports the unexpected success below, or a
		// mistaken "! gpao start" would leak a running oracle for the rest of
		// the suite -- nothing else in this function will ever register it
		// for teardown.
		stopDaemon(ts, d)
	}
	TSValidateError(ts, "gpao start", neg, err)
	if err != nil {
		return
	}

	r.d = d
	ts.Setenv("GPAO_STATUS", statusURL(listen))
	// Captures d rather than calling r.stop: a bare ts.Defer(r.stop) would
	// target whatever daemon is registered in r.d when teardown runs, not the
	// one this script started. Harmless today because scripts run
	// sequentially, but this closure stops the right daemon regardless.
	ts.Defer(func() {
		r.mu.Lock()
		if r.d == d {
			r.d = nil
		}
		r.mu.Unlock()
		stopDaemon(ts, d)
	})
}

func (r *gpaoRunner) stop(ts *testscript.TestScript) {
	r.mu.Lock()
	d := r.d
	r.d = nil
	r.mu.Unlock()
	stopDaemon(ts, d)
}

// stopDaemon terminates d and logs what it wrote, so a scenario that fails an
// assertion still has the oracle's own evidence to go with the diff -- Output()
// otherwise has no reader once the daemon starts cleanly.
//
// Stopped before its output is read, because the oracle writes as it goes down:
// the run's tally, the refusal it was in the middle of, a panic raised during
// shutdown. Read first, none of that has been written yet.
//
// Uses ts.Logf rather than ts.Stderr/Stdout: those panic outside a builtin
// command's own execution, which the ts.Defer teardown path is.
func stopDaemon(ts *testscript.TestScript, d *daemon.Daemon) {
	if d == nil {
		return
	}
	if err := d.Stop(); err != nil {
		ts.Logf("gpao: stop: %v", err)
	}
	if out := d.Output(); out != "" {
		ts.Logf("gpao output:\n%s", out)
	}
}

// gpaoArgs builds the oracle's argv. Script flags come last so that a script
// naming a node overrides the run's default, and the default is left out
// entirely when the script names one.
func gpaoArgs(cfg GpaoConfig, statusListen string, flags []string) []string {
	args := []string{
		"-chain-id", cfg.ChainID,
		"-status-listen", statusListen,
	}
	if cfg.GnoRoot != "" {
		args = append(args, "-gno-root", cfg.GnoRoot)
	}
	if !namesFlag(flags, "remote") && cfg.Remote != "" {
		args = append(args, "-remote", cfg.Remote)
	}
	return append(args, flags...)
}

func namesFlag(flags []string, name string) bool {
	return slices.ContainsFunc(flags, func(f string) bool {
		return f == "-"+name || f == "--"+name ||
			strings.HasPrefix(f, "-"+name+"=") || strings.HasPrefix(f, "--"+name+"=")
	})
}

func statusURL(listen string) string { return "http://" + listen }

// statusProbe reports ready once the status board answers, which proves only
// that the HTTP listener is up: its handler never touches the RPC client, so
// a 200 here does not prove the run loop survived the start-height query that
// follows. daemon.Start's race against process exit is what catches that.
func statusProbe(base string) daemon.Probe {
	return func(ctx context.Context) error {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/status", nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status board answered %s", resp.Status)
		}
		return nil
	}
}
