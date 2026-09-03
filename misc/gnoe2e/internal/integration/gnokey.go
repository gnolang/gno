package integration

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/keyscli"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/rogpeppe/go-internal/testscript"
)

// GnokeyTSCmd returns a testscript command handler for gnokey.
// It auto-injects --home, --remote, --chain-id, and --insecure-password-stdin.
func GnokeyTSCmd() func(ts *testscript.TestScript, neg bool, args []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		gnoHomeDir := ts.Getenv("GNOHOME")
		rpcAddr := ts.Getenv("RPC_ADDR")

		io := commands.NewTestIO()
		io.SetOut(commands.WriteNopCloser(ts.Stdout()))
		io.SetErr(commands.WriteNopCloser(ts.Stderr()))
		io.SetIn(strings.NewReader("\n"))

		cmd := keyscli.NewRootCmd(io, client.DefaultBaseOptions)

		// Note: --chainid is a subcommand flag (maketx), not a root flag.
		// Scripts should use -chainid $CHAIN_ID when needed.

		args = gnokeyArgs(gnoHomeDir, rpcAddr, args)

		// reported is set just before the non-deferred TSValidateError call
		// below. A panic reaching the deferred recover with reported already
		// true did not come from cmd.ParseAndRun -- it is ts.Fatalf's own
		// failNow, raised by that same TSValidateError call (e.g. a negated
		// command that unexpectedly succeeded). Re-validating it would have
		// the recover's call see a non-nil err and, when neg is true, treat
		// that as satisfying the negation -- silently turning a reported
		// failure into a pass. Re-panicking instead lets it propagate as
		// testscript itself expects.
		var err error
		reported := false
		defer func() {
			if r := recover(); r != nil {
				if reported {
					panic(r)
				}
				fmt.Fprintf(ts.Stderr(), "gnokey panic: %v\n%s\n", r, debug.Stack())
				switch val := r.(type) {
				case error:
					err = val
				case string:
					err = fmt.Errorf("panic: %s", val)
				default:
					err = fmt.Errorf("panic: %v", val)
				}
				TSValidateError(ts, "gnokey", neg, err)
			}
		}()

		err = cmd.ParseAndRun(context.Background(), args)
		reported = true
		TSValidateError(ts, "gnokey", neg, err)
	}
}

// gnokeyArgs assembles gnokey's argv.
//
// Caller arguments come last so a script naming its own -remote overrides the
// injected default: the flag parser keeps the final occurrence. -remote is a
// root flag, so a caller must place it before the subcommand.
func gnokeyArgs(gnoHome, rpcAddr string, args []string) []string {
	out := []string{"-home", gnoHome, "-insecure-password-stdin=true"}
	if rpcAddr != "" {
		out = append(out, "-remote", rpcAddr)
	}
	return append(out, args...)
}

// TSValidateError checks a command result against the negation flag.
// If err != nil and neg is false, it fatals. If err == nil and neg is true, it fatals.
func TSValidateError(ts *testscript.TestScript, cmd string, neg bool, err error) {
	if err != nil {
		fmt.Fprintf(ts.Stderr(), "%q error: %+v\n", cmd, err)
		if !neg {
			ts.Fatalf("unexpected %q command failure: %s", cmd, err)
		}
	} else {
		if neg {
			ts.Fatalf("unexpected %q command success", cmd)
		}
	}
}
