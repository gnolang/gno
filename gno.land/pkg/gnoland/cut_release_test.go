package gnoland

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/upgrades"
)

// runCutReleaseFuncs is runScriptFuncs on misc/release/cut-release.sh.
func runCutReleaseFuncs(t *testing.T, vars map[string]string, funcs ...string) (string, bool) {
	t.Helper()

	return runScriptFuncs(t, "misc/release/cut-release.sh", vars, funcs...)
}

// runScriptFuncs makes calls to functions of a release script in bash, in
// order, with those functions and whichever of the script's die, info, ok and
// warn it defines, and the script's globals set from vars. A call is a function name,
// optionally followed by shell-quoted arguments. It returns their combined
// output and whether they all returned (die exits non-zero).
func runScriptFuncs(t *testing.T, rel string, vars map[string]string, calls ...string) (string, bool) {
	t.Helper()

	src := repoFile(t, rel)
	var script strings.Builder
	script.WriteString("set -euo pipefail\n")
	for _, helper := range []string{"die", "info", "ok", "warn"} {
		script.WriteString(findShellFunc(src, helper) + "\n")
	}
	for _, call := range calls {
		script.WriteString(shellFunc(t, rel, src, strings.Fields(call)[0]) + "\n")
	}
	script.WriteString(strings.Join(calls, "\n") + "\n")

	// cut-release.sh's check_ledger_entry builds the ledger tool with go run; a
	// module fetch that hangs fails here rather than at the test binary's own
	// timeout.
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

// shellFunc returns the definition of a function in the shell script rel, read
// as src, and fails the test if the script defines no such function.
func shellFunc(t *testing.T, rel, src, name string) string {
	t.Helper()

	def := findShellFunc(src, name)
	require.NotEmpty(t, def, "%s defines no %s()", rel, name)
	return def
}

// findShellFunc returns the definition of a function in a shell script's
// source, either a one-liner or a block closed by a "}" on its own line, or ""
// when the script defines none.
func findShellFunc(src, name string) string {
	q := regexp.QuoteMeta(name)
	for _, re := range []string{`(?ms)^` + q + `\(\) \{\n.*?^\}$`, `(?m)^` + q + `\(\) \{ .* \}$`} {
		if def := regexp.MustCompile(re).FindString(src); def != "" {
			return def
		}
	}
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

// A tag never moves, so its message must not record a halt height: the height
// passed to cut-release.sh can still change until the halt proposal executes.
func TestCutReleaseTagMessage(t *testing.T) {
	t.Parallel()

	out, returned := runCutReleaseFuncs(t, map[string]string{
		"CHAIN": "mainnet", "VERSION": "v1.6.0", "COMMIT": strings.Repeat("c", 40),
		"PREVIOUS": "v1.5.1", "HALT_HEIGHT": "848000",
	}, "tag_message")
	require.True(t, returned, out)
	assert.NotContains(t, out, "848000", "the tag message records the halt height")
	assert.Contains(t, out, "halt_min_version v1.6.0")
}

// bash 3.2 ends a script aborted by set -u with the EXIT trap's status, 0
// unless the trap says otherwise; the operator must still see a failure. This
// bites on bash 3.2 only, macOS's /bin/bash: bash 5 exits non-zero from the
// abort with or without the fix. TestReleaseScriptsFailAnyOtherEnding pins
// the fix's shape on any bash.
func TestCutReleaseAbortExitsNonZero(t *testing.T) {
	t.Parallel()

	src := repoFile(t, "misc/release/cut-release.sh")
	broken := strings.Replace(src, "\tparse_args \"$@\"\n", "\tparse_args \"$@\"\n\t: \"${UNSET_BY_THE_TEST}\"\n", 1)
	require.NotEqual(t, src, broken, "could not inject the abort")
	// The script finds the repository two levels up from itself.
	root := t.TempDir()
	copied := filepath.Join(root, "misc", "release", "cut-release.sh")
	require.NoError(t, os.MkdirAll(filepath.Dir(copied), 0o700))
	require.NoError(t, os.WriteFile(copied, []byte(broken), 0o700))

	out, err := exec.Command("bash", copied, "v99.0.0").CombinedOutput()
	assert.Error(t, err, "an unset variable aborts the run:\n%s", out)
	assert.Contains(t, string(out), "UNSET_BY_THE_TEST")

	out, err = exec.Command("bash", copied, "--help").CombinedOutput()
	assert.NoError(t, err, "--help exits zero:\n%s", out)
}

// Candidates are cut from chain/mainnet too, so CHAIN names mainnet for every
// cut, while a candidate's halt is the testnet's. The proposal printed for a
// candidate must not be a command that, pasted, schedules the halt on mainnet.
func TestCutReleaseHaltProposalNamesTheNetwork(t *testing.T) {
	t.Parallel()

	for _, c := range []struct{ version, says, never string }{
		{"v1.8.0-rc.1", "misc/deployments/<chain>.gno.land/govdao-exec.sh set-halt 120000 v1.8.0-rc.1", "misc/deployments/mainnet.gno.land/"},
		{"v1.8.0", "misc/deployments/mainnet.gno.land/govdao-exec.sh set-halt 120000 v1.8.0", "<chain>"},
	} {
		t.Run(c.version, func(t *testing.T) {
			t.Parallel()

			out, returned := runCutReleaseFuncs(t, map[string]string{
				"CHAIN": "mainnet", "VERSION": c.version, "HALT_HEIGHT": "120000",
			}, "print_halt_proposal")
			require.True(t, returned, out)
			assert.Contains(t, out, c.says)
			assert.NotContains(t, out, c.never)
		})
	}
}

// Each release script marks the ends it means, --help and main returning,
// with DONE=1, and its EXIT trap fails any other ending. This is the shape
// TestCutReleaseAbortExitsNonZero and TestPromoteReleaseRuns's abort case
// exercise on bash 3.2, read from the scripts so it holds on any bash.
func TestReleaseScriptsFailAnyOtherEnding(t *testing.T) {
	t.Parallel()

	const fails = "[[ ${DONE} -eq 1 ]] || exit 1"
	for _, rel := range []string{"misc/release/cut-release.sh", promoteReleaseScript} {
		t.Run(rel, func(t *testing.T) {
			t.Parallel()

			src := repoFile(t, rel)
			trap := grepOne(t, rel, `(?m)^trap (.+) EXIT$`)
			// An inline trap ends with the check; a named one is a function
			// whose last statement is the check.
			if body, ok := strings.CutPrefix(trap, "'"); ok {
				assert.True(t, strings.HasSuffix(body, "; "+fails+"'"), "%s: the EXIT trap must end with %s", rel, fails)
			} else {
				def := shellFunc(t, rel, src, trap)
				assert.True(t, strings.HasSuffix(def, "\t"+fails+"\n}"), "%s: %s() must end with %s", rel, trap, fails)
			}
			assert.Less(t, strings.Index(src, "\nDONE=0\n"), strings.Index(src, "\ntrap "), "%s must set DONE=0 before its trap", rel)
			assert.True(t, strings.HasSuffix(src, "\nmain \"$@\"\nDONE=1\n"), "%s must end with main \"$@\" and then DONE=1", rel)
			// The only other end it means is --help's.
			exits := regexp.MustCompile(`(?m)^(.*)\n\s*exit 0$`).FindAllStringSubmatch(src, -1)
			require.NotEmpty(t, exits, "%s: no exit 0 for --help", rel)
			for _, m := range exits {
				assert.Equal(t, "DONE=1", strings.TrimSpace(m[1]), "%s: an exit 0 must follow DONE=1", rel)
			}
		})
	}
}
