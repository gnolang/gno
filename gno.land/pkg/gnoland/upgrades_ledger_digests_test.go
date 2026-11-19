package gnoland

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/upgrades"
)

const ledgerDigestsWorkflow = ".github/workflows/ci-upgrades-ledger-digests.yml"

// ledgerDigestsScript returns the script of the daily ledger job's run step,
// without its YAML indentation.
func ledgerDigestsScript(t *testing.T) string {
	t.Helper()

	lines := strings.Split(repoFile(t, ledgerDigestsWorkflow), "\n")
	var script []string
	runs := 0
	for i, line := range lines {
		key := strings.TrimLeft(line, " ")
		if key != "run: |" {
			continue
		}
		runs++
		indent := strings.Repeat(" ", len(line)-len(key)+2)
		for _, body := range lines[i+1:] {
			if body != "" && !strings.HasPrefix(body, indent) {
				break
			}
			script = append(script, strings.TrimPrefix(body, indent))
		}
	}
	require.Equal(t, 1, runs, "%s: the test expects one run step", ledgerDigestsWorkflow)
	return strings.Join(script, "\n")
}

// runLedgerDigestsJob runs the daily ledger job's script as GitHub Actions
// runs a step with no shell: key (bash -e), in a checkout holding one ledger
// with the given entries. It returns the output and whether the job passed.
func runLedgerDigestsJob(t *testing.T, entries ...upgrades.Entry) (string, bool) {
	t.Helper()

	checkout := t.TempDir()
	dir := filepath.Join(checkout, "misc", "deployments", "test.gno.land")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	ledger, err := json.Marshal(upgrades.Ledger{ChainID: "test-1", Upgrades: entries})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, upgrades.LedgerFile), ledger, 0o644))

	script := filepath.Join(t.TempDir(), "job.sh")
	require.NoError(t, os.WriteFile(script, []byte(ledgerDigestsScript(t)), 0o644))

	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-e", script)
	cmd.Dir = checkout
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		t.Fatalf("could not run bash: %v", err)
	}
	return string(out), err == nil
}

// requireGitHubAPI skips the test unless gh can read the GitHub API, through
// which the daily ledger job reads tags, and jq is installed. CI's Go test
// step has no GitHub token, so the test runs where gh is logged in.
func requireGitHubAPI(t *testing.T) {
	t.Helper()

	for _, tool := range []string{"gh", "jq"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}
	if err := exec.Command("gh", "auth", "status").Run(); err != nil {
		t.Skip("gh is not authenticated")
	}
}

// An entry is written once its chain runs the release, after the cut, so the
// commit check cut-release.sh makes at the cut does not see it. The daily job
// checks every entry's commit against the commit its version's tag points to.
func TestLedgerDigestsJobChecksCommits(t *testing.T) {
	t.Parallel()
	requireGitHubAPI(t)

	l, err := upgrades.Load(repoPath(t, "misc/deployments/mainnet.gno.land"))
	require.NoError(t, err)
	// With the digest and binaries null, pending the build, the job checks
	// neither the registry nor the release assets: the commit alone is left.
	entry := func(v string) upgrades.Entry {
		e, ok := l.Entry(v)
		require.True(t, ok, "the mainnet ledger has no %s", v)
		e.Image.Digest, e.Binaries, e.RanAs = nil, nil, nil
		return e
	}
	v150, v151 := entry("v1.5.0"), entry("v1.5.1")

	t.Run("the commit its version's tag points to passes", func(t *testing.T) {
		t.Parallel()

		out, ok := runLedgerDigestsJob(t, v151)
		assert.True(t, ok, out)
		assert.Regexp(t, `ok +commit +v1\.5\.1 `, out)
	})

	t.Run("another commit fails, and the entries after it are still checked", func(t *testing.T) {
		t.Parallel()

		wrong := v151
		wrong.Commit = v150.Commit
		out, ok := runLedgerDigestsJob(t, wrong, v150)
		assert.False(t, ok, out)
		assert.Contains(t, out, "::error::")
		assert.Contains(t, out, "v1.5.1 points to commit "+v151.Commit)
		assert.Contains(t, out, "the ledger records commit "+v150.Commit)
		assert.Regexp(t, `ok +commit +v1\.5\.0 `, out)
	})

	t.Run("a row copied from the one before it fails", func(t *testing.T) {
		t.Parallel()

		// Only the version was bumped: the commit and the release link still
		// name v1.5.0, so they agree with each other but not with the version.
		copied := v150
		copied.Version = v151.Version
		out, ok := runLedgerDigestsJob(t, copied)
		assert.False(t, ok, out)
		assert.Contains(t, out, "v1.5.1 points to commit "+v151.Commit)
	})

	t.Run("a lightweight tag names the commit itself", func(t *testing.T) {
		t.Parallel()

		// cut-release.sh annotates every tag it pushes; chain/onyx, the tag
		// onyx's genesis was built from, is lightweight.
		e := v150
		e.Version = "chain/onyx"
		e.Commit = "5cdbc25fcde0b7569911a4e308ae5d2f6e96c399"
		out, ok := runLedgerDigestsJob(t, e)
		assert.True(t, ok, out)
		assert.Regexp(t, `ok +commit +chain/onyx `, out)
	})

	t.Run("a tag that does not exist fails, and the entries after it are still checked", func(t *testing.T) {
		t.Parallel()

		// v1.5 prefixes existing tags; a lookup must not match them.
		e := v151
		e.Version = "v1.5"
		out, ok := runLedgerDigestsJob(t, e, v150)
		assert.False(t, ok, out)
		assert.Contains(t, out, "cannot read tag v1.5")
		assert.Regexp(t, `ok +commit +v1\.5\.0 `, out)
	})
}
