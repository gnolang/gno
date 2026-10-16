package gnoland

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/upgrades"
)

// runCutReleaseFuncs runs the named functions of misc/release/cut-release.sh in
// bash, in order, with the script's globals set from vars. It returns their
// combined output and whether they all returned (die exits non-zero).
func runCutReleaseFuncs(t *testing.T, vars map[string]string, funcs ...string) (string, bool) {
	t.Helper()

	src := repoFile(t, "misc/release/cut-release.sh")
	var script strings.Builder
	script.WriteString("set -euo pipefail\n")
	for _, name := range []string{"die", "ok", "warn"} {
		script.WriteString(shellFunc(t, src, name) + "\n")
	}
	for _, name := range funcs {
		script.WriteString(shellFunc(t, src, name) + "\n")
	}
	script.WriteString(strings.Join(funcs, "\n") + "\n")

	// check_ledger_entry builds the ledger tool with go run; a module fetch that
	// hangs fails here rather than at the test binary's own timeout.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-c", script.String())
	cmd.Env = append(os.Environ(), "REPO_ROOT="+repoPath(t, "."))
	for k, v := range vars {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		t.Fatalf("could not run bash: %v", err)
	}
	return string(out), err == nil
}

// shellFunc returns the definition of a function in a shell script, either a
// one-liner or a block closed by a "}" on its own line.
func shellFunc(t *testing.T, src, name string) string {
	t.Helper()

	q := regexp.QuoteMeta(name)
	for _, re := range []string{`(?ms)^` + q + `\(\) \{\n.*?^\}$`, `(?m)^` + q + `\(\) \{ .* \}$`} {
		if def := regexp.MustCompile(re).FindString(src); def != "" {
			return def
		}
	}
	require.Failf(t, "function not found", "cut-release.sh defines no %s()", name)
	return ""
}

// A ledger row records what a chain runs, and what it runs is known only once
// it has upgraded: the halt height is an estimate until the proposal executes,
// the halt time is known once the chain has halted, and the image digest and
// binaries do not exist before the tag is built. So a final release is cut
// before its row is written, and cut-release.sh must let it through; what it
// still refuses is a row that names other code.
func TestCutReleaseLedgerGate(t *testing.T) {
	t.Parallel()

	const dir = "misc/deployments/mainnet.gno.land"
	l, err := upgrades.Load(repoPath(t, dir))
	require.NoError(t, err)
	var last upgrades.Entry
	for _, e := range l.Upgrades {
		if !strings.Contains(e.Version, "-") {
			last = e
		}
	}
	require.NotEmpty(t, last.Version, "%s has no final release", dir)

	t.Run("a final without its row is cut with a warning", func(t *testing.T) {
		t.Parallel()

		out, returned := runCutReleaseFuncs(t, map[string]string{
			"CHAIN": "mainnet", "VERSION": "v99.0.0", "COMMIT": strings.Repeat("a", 40),
		}, "check_ledger_entry")
		assert.True(t, returned, "check_ledger_entry refused a final without its row:\n%s", out)
		assert.Contains(t, out, "no entry for v99.0.0")
	})

	t.Run("a final whose row names other code is refused", func(t *testing.T) {
		t.Parallel()

		out, returned := runCutReleaseFuncs(t, map[string]string{
			"CHAIN": "mainnet", "VERSION": last.Version, "COMMIT": strings.Repeat("0", 40),
		}, "check_ledger_entry")
		assert.False(t, returned, "check_ledger_entry accepted %s on a commit its row does not name:\n%s", last.Version, out)
		assert.Contains(t, out, "names "+last.Commit[:9])
	})

	t.Run("a final whose row names the tagged commit is cut", func(t *testing.T) {
		t.Parallel()

		out, returned := runCutReleaseFuncs(t, map[string]string{
			"CHAIN": "mainnet", "VERSION": last.Version, "COMMIT": last.Commit,
		}, "check_ledger_entry")
		assert.True(t, returned, "check_ledger_entry refused %s on the commit its row names:\n%s", last.Version, out)
		assert.Contains(t, out, "the entry names the commit being tagged")
	})

	// A candidate is a rehearsal: the final release cut after it, on its commit,
	// is checked the same way and refused.
	t.Run("a candidate without a row is cut with a warning", func(t *testing.T) {
		t.Parallel()

		out, returned := runCutReleaseFuncs(t, map[string]string{
			"CHAIN": "mainnet", "VERSION": "v99.0.0-rc.1", "COMMIT": strings.Repeat("a", 40),
		}, "check_ledger_entry")
		assert.True(t, returned, "check_ledger_entry refused a candidate without a row:\n%s", out)
		assert.Contains(t, out, "no entry for v99.0.0 yet")
	})

	t.Run("a candidate whose final's row names other code is cut with a warning", func(t *testing.T) {
		t.Parallel()

		out, returned := runCutReleaseFuncs(t, map[string]string{
			"CHAIN": "mainnet", "VERSION": last.Version + "-rc.99", "COMMIT": strings.Repeat("0", 40),
		}, "check_ledger_entry")
		assert.True(t, returned, "check_ledger_entry refused a candidate:\n%s", out)
		assert.Contains(t, out, "names "+last.Commit[:9])
	})
}
