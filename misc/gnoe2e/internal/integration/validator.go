package integration

import (
	"context"
	"errors"
	"strconv"

	"github.com/gnolang/gno/misc/gnoe2e/internal/cluster"
	"github.com/rogpeppe/go-internal/testscript"
)

// ValidatorTSCmd returns a testscript command that stops and restarts
// individual validators, so a scenario can inject a node outage.
//
// Lifecycle is driven from Go rather than through txtar's exec primitive: a
// script says which node should be down and for how long, and the harness owns
// signalling it, waiting for it to be gone, and bringing it back ready. A
// script that spawned processes itself would also have to clean them up, and
// would leak one on every failed assertion.
//
// Indexing matches the RPC_ADDR_N the scripts already use.
//
// Only restart is negatable: "! validator restart N" asserts that the node
// cannot come back, and the error, which carries the node's stderr tail,
// reaches the script's stderr so the scenario can name the reason it died. It
// asserts that narrowly: a node the cluster does not have, or one that was
// never stopped, fails the script in either mode. A stop that fails is a
// harness fault, not something a scenario observes.
func ValidatorTSCmd(cl *cluster.Cluster) func(ts *testscript.TestScript, neg bool, args []string) {
	return func(ts *testscript.TestScript, neg bool, args []string) {
		if len(args) != 2 {
			ts.Fatalf("usage: validator stop|restart <index>")
		}

		// The script is checked before the environment is: a malformed
		// command is the author's mistake in any mode, and reporting it as
		// "no local cluster" would send them after the wrong thing entirely.
		if args[0] != "stop" && args[0] != "restart" {
			ts.Fatalf("unknown validator subcommand %q", args[0])
		}
		if neg && args[0] == "stop" {
			ts.Fatalf("validator stop does not support negation")
		}
		index, err := strconv.Atoi(args[1])
		if err != nil {
			ts.Fatalf("validator: invalid index %q: %v", args[1], err)
		}

		if cl == nil {
			ts.Fatalf("validator: no local cluster; this script has no cluster to act on")
		}

		// Background, not the script's context: RestartValidator's context
		// governs the restarted process, so a node revived here has to outlive
		// the command that revived it.
		switch args[0] {
		case "stop":
			if err := cl.StopValidator(context.Background(), index); err != nil {
				ts.Fatalf("validator stop %d: %v", index, err)
			}
		case "restart":
			err := cl.RestartValidator(context.Background(), index)
			// A script naming a node the cluster does not have, or one it never
			// stopped, is the script's own mistake in either mode. Letting a
			// negation stand on it would pass a scenario in which no node ever
			// went away and none ever came back.
			if errors.Is(err, cluster.ErrUnknownValidator) || errors.Is(err, cluster.ErrValidatorRunning) {
				ts.Fatalf("validator restart %d: %v", index, err)
			}
			TSValidateError(ts, "validator restart", neg, err)
		}
	}
}
