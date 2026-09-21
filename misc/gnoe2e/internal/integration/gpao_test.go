package integration

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
	"github.com/stretchr/testify/require"
)

// TestMain lets the compiled test binary impersonate gpao when
// GNOE2E_FAKE_GPAO is set, the same trick internal/daemon's tests use to
// exercise a real child process without building the actual oracle binary.
func TestMain(m *testing.M) {
	// Stands in for a gnoland that dies during boot, the way a node does when
	// VMKeeper.Initialize panics over its own data dir: one line on stderr,
	// non-zero exit, never opens RPC.
	if os.Getenv("GNOE2E_FAKE_GNOLAND") == "die" {
		fmt.Fprintln(os.Stderr, "fake gnoland: refusing to boot")
		os.Exit(1)
	}
	switch os.Getenv("GNOE2E_FAKE_GPAO") {
	case "":
		os.Exit(m.Run())
	case "die":
		// Stands in for the real oracle's fatal path (oracle.go:265-269): dies
		// immediately, the way daemon.Start sees a -start-height 0 run against
		// a dead endpoint, without needing that endpoint to exist at all.
		fmt.Fprintln(os.Stderr, "failed to query node status: dial tcp: connection refused")
		os.Exit(1)
	case "farewell":
		// Stands in for the oracle's own shutdown path, which writes its final
		// tally as it goes: a fake that says nothing on the way down cannot
		// tell whether the harness read its output before or after stopping it.
		serveFakeGpaoUntilSignalled()
	case "serve":
		// Stands in for a gpao that comes up and stays up: serves the
		// -status-listen address gpaoArgs always injects, then blocks --
		// the negation test needs a daemon that survives, not one that races
		// to exit before the probe can see it.
		serveFakeGpaoStatus()
	default:
		os.Exit(9)
	}
}

// serveFakeGpaoUntilSignalled answers the status probe, then writes one last
// line when it is asked to stop.
func serveFakeGpaoUntilSignalled() {
	startFakeGpaoStatus()
	stopping := make(chan os.Signal, 1)
	signal.Notify(stopping, syscall.SIGTERM, os.Interrupt)
	<-stopping
	fmt.Fprintln(os.Stderr, "fake gpao: FINAL_TALLY on the way down")
	os.Exit(0)
}

func serveFakeGpaoStatus() {
	startFakeGpaoStatus()
	time.Sleep(time.Hour)
}

func startFakeGpaoStatus() {
	listen := flagValue(os.Args[1:], "-status-listen")
	if listen == "" {
		fmt.Fprintln(os.Stderr, "fake gpao: no -status-listen given")
		os.Exit(9)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fake gpao: listen: %v\n", err)
		os.Exit(9)
	}
	go http.Serve(ln, mux) //nolint:errcheck // the fake serves until the process exits
}

// fakeGpaoBinary stands in for the run's lazy build. The compiled test binary
// impersonates gpao under GNOE2E_FAKE_GPAO, so no oracle has to be built to
// exercise a real child process.
func fakeGpaoBinary() (string, error) { return os.Args[0], nil }

func flagValue(args []string, name string) string {
	for i, a := range args {
		if a == name && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// A daemon that dies before becoming ready lets "! gpao start" pass rather
// than aborting the script, which is the only way a scenario can assert that
// the oracle refuses to come up.
func TestGpaoStartNegationSucceedsWhenStartFails(t *testing.T) {
	t.Setenv("GNOE2E_FAKE_GPAO", "die")

	adapter := NewTestscriptT(testLogger(t), false)
	cfg := GpaoConfig{Binary: fakeGpaoBinary, ChainID: "test-e2e"}
	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"gpao": GpaoTSCmd(cfg),
	}

	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, "! gpao start\n")},
		Cmds:  cmds,
	})

	require.False(t, adapter.Failed, "! gpao start must pass when the daemon fails to become ready")
}

// TestGpaoStartNegationFailsWhenTheStartUnexpectedlySucceeds guards the same
// hazard TestGnokeyNegationFailsWhenTheCommandUnexpectedlySucceeds guards for
// gnokey: a negated command whose underlying operation succeeds must fail the
// script, not silently pass it. A negation that can never fail is worse than
// no negation.
func TestGpaoStartNegationFailsWhenTheStartUnexpectedlySucceeds(t *testing.T) {
	t.Setenv("GNOE2E_FAKE_GPAO", "serve")

	adapter := NewTestscriptT(testLogger(t), false)
	cfg := GpaoConfig{Binary: fakeGpaoBinary, ChainID: "test-e2e"}
	cmds := map[string]func(*testscript.TestScript, bool, []string){
		"gpao": GpaoTSCmd(cfg),
	}

	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, "! gpao start\n")},
		Cmds:  cmds,
	})

	require.True(t, adapter.Failed, "! gpao start must fail the script when the start unexpectedly succeeds")
}

// A binary that will not build is a fault in the checkout rather than the
// failure a script meant to observe, so it fails the script whatever the
// negation on the line says, and it says which build failed.
func TestGpaoStartFailsTheScriptWhenTheBuildFails(t *testing.T) {
	logger, logged := bufferedTestLogger(t)
	adapter := NewTestscriptT(logger, false)
	cfg := GpaoConfig{
		Binary:  func() (string, error) { return "", errors.New("build gpao: exit status 1") },
		ChainID: "test-e2e",
	}

	testscript.RunT(adapter, testscript.Params{
		Files: []string{writeScript(t, "! gpao start\n")},
		Cmds:  map[string]func(*testscript.TestScript, bool, []string){"gpao": GpaoTSCmd(cfg)},
	})

	require.True(t, adapter.Failed, "a build that fails fails the script, negation or not")
	require.Contains(t, logged.String(), "gpao: build: build gpao: exit status 1")
}

func TestGpaoArgsInjectDefaults(t *testing.T) {
	cfg := GpaoConfig{ChainID: "test-e2e", Remote: "tcp://127.0.0.1:26657", GnoRoot: "/gno"}

	args := gpaoArgs(cfg, "127.0.0.1:9999", nil)

	require.Equal(t, []string{
		"-chain-id", "test-e2e",
		"-status-listen", "127.0.0.1:9999",
		"-gno-root", "/gno",
		"-remote", "tcp://127.0.0.1:26657",
	}, args)
}

func TestGpaoScriptRemoteWins(t *testing.T) {
	cfg := GpaoConfig{ChainID: "test-e2e", Remote: "tcp://default:26657"}

	args := gpaoArgs(cfg, "127.0.0.1:9999", []string{"-remote", "tcp://chosen:26657"})

	require.NotContains(t, args, "tcp://default:26657",
		"a script that names a node must not also get the default")
	require.Contains(t, args, "tcp://chosen:26657")
}

func TestGpaoScriptRemoteEqualsFormWins(t *testing.T) {
	cfg := GpaoConfig{ChainID: "test-e2e", Remote: "tcp://default:26657"}

	args := gpaoArgs(cfg, "127.0.0.1:9999", []string{"-remote=tcp://chosen:26657"})

	require.NotContains(t, args, "tcp://default:26657",
		"the -flag=value form must also suppress the injected default")
	require.Contains(t, args, "-remote=tcp://chosen:26657")
}

// A GpaoConfig with no way to obtain the binary is a fault in the run's
// wiring, not in the scenario, so it fails the script with a message that
// names the gap rather than panicking on a nil func.
func TestGpaoStartFailsTheScriptWhenNoBinaryProviderIsConfigured(t *testing.T) {
	logger, logged := bufferedTestLogger(t)
	adapter := NewTestscriptT(logger, false)
	cfg := GpaoConfig{ChainID: "test-e2e"}

	require.NotPanics(t, func() {
		testscript.RunT(adapter, testscript.Params{
			Files: []string{writeScript(t, "gpao start\n")},
			Cmds:  map[string]func(*testscript.TestScript, bool, []string){"gpao": GpaoTSCmd(cfg)},
		})
	})

	require.True(t, adapter.Failed, "a run with no binary provider fails the script")
	require.Contains(t, logged.String(), "gpao: no binary provider")
}

// TestStopDaemonKeepsWhatTheOracleSaidOnItsWayDown pins the order teardown
// reads in.
//
// The oracle writes its last words as it stops, and this log is the only reader
// its output ever gets: taken before the stop, the tally, the final refusal and
// any panic raised during shutdown are all still unwritten and go nowhere.
func TestStopDaemonKeepsWhatTheOracleSaidOnItsWayDown(t *testing.T) {
	t.Setenv("GNOE2E_FAKE_GPAO", "farewell")

	logger, logged := bufferedTestLogger(t)
	adapter := NewTestscriptT(logger, true)
	cfg := GpaoConfig{Binary: fakeGpaoBinary, ChainID: "test-e2e"}

	testscript.RunT(adapter, testscript.Params{
		// No "gpao stop" line: the run's own teardown is the path every
		// scenario takes, and the only one that reads the oracle's output.
		Files: []string{writeScript(t, "gpao start\n")},
		Cmds: map[string]func(*testscript.TestScript, bool, []string){
			"gpao": GpaoTSCmd(cfg),
		},
	})

	require.False(t, adapter.Failed)
	require.Contains(t, logged.String(), "FINAL_TALLY")
}
