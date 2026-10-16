package gnoland

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rpccore "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rs "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/log"
)

const setHaltScript = "misc/govdao-scripts/set-halt.sh"

// set-halt.sh refuses a floor the binary at REMOTE already meets (see
// gno.land/cmd/gnoland/UPGRADES.md, "Restart gates"). It mirrors meetsMinVersion in shell; this runs
// it for every pair below and requires a refusal exactly where meetsMinVersion
// says the running binary meets the floor.
func TestSetHaltRefusesAFloorTheRunningBinaryMeets(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("bash"); err != nil {
		t.Skipf("set-halt.sh needs bash: %v", err)
	}
	_, curlErr := exec.LookPath("curl")
	_, wgetErr := exec.LookPath("wget")
	if curlErr != nil && wgetErr != nil {
		t.Skip("set-halt.sh needs curl or wget to read the remote's version")
	}
	gnokey := gnokeyStub(t)

	running := []string{
		"v1.5.1", "v1.5.1+deadbeef", "v1.6.0-rc.2", "v1.6.0-rc.10", "v1.10.0",
		"v1.0.0-alpha.beta", "v1.0.0-1", "v1.0.0-Alpha", "v99999999999999999999.0.0",
		"chain/gnoland1.1",
		// Betanet's prefix with more after the minor: the node widens whatever
		// follows the first dot, so an off-tag build on a chain/gnolandX.Y
		// branch parses.
		"chain/gnoland1.1.0-rc", "chain/gnoland1.2.3335+bc43a5fb7",
		// Not release versions, so they meet no floor.
		"develop", "master.3335+bc43a5fb7", "heads/chain/mainnet.3444+e75fef82c",
		"v1.5.1+", "v1.02.0", "chain/gnoland1.1+deadbeef", "chain/gnoland1.1-rc.1",
	}
	floors := []string{
		"v1.5.0", "v1.5.1", "v1.5.2", "v1.5.1-rc.1", "v1.5.2-rc.1", "v1.6.0", "v2.0.0",
		"v1.6.0-rc.1", "v1.6.0-rc.2", "v1.6.0-rc.9", "v1.6.0-rc.10", "v1.6.0-rc.2.1",
		"v1.6.0-rc", "v1.6.0-alpha", "v1.6.0-rc.a", "v1.9.0", "v1.10.1",
		"v1.0.0-alpha", "v1.0.0-alpha.1", "v1.0.0-alpha.beta", "v1.0.0-0", "v1.0.0-1", "v1.0.0-Alpha",
		"v99999999999999999998.0.0", "v100000000000000000000.0.0",
		"v1.1.0", "v1.1.1", "chain/gnoland1.0", "chain/gnoland1.1", "chain/gnoland1.2",
	}

	for _, version := range running {
		remote := statusServer(t, version)
		for _, floor := range floors {
			t.Run(version+"/"+floor, func(t *testing.T) {
				t.Parallel()

				res := runSetHalt(t, gnokey, remote, "1000", floor)
				if meetsMinVersion(version, floor) {
					assertRefused(t, res, "already meets")
					return
				}
				assert.NoError(t, res.err, res.out)
				assert.True(t, res.reachedGnokey, "a floor %s does not meet must be proposed:\n%s", version, res.out)
				if _, parses := parseReleaseVersion(version); !parses {
					assert.Contains(t, res.out, "not a release version", "an unparseable running version must be reported")
				} else {
					assert.NotContains(t, res.out, "not a release version")
				}
			})
		}
	}

	// Every REMOTE form gnokey takes reaches the same /status: the refusal
	// proves the version was read through it.
	live := statusServer(t, "v1.5.1")
	hostPort := strings.TrimPrefix(live, "http://")
	for _, tc := range []struct{ name, remote string }{
		{"a trailing slash", live + "/"},
		{"the tcp scheme", "tcp://" + hostPort},
		{"no scheme", hostPort},
	} {
		t.Run("a REMOTE with "+tc.name+" is read over http", func(t *testing.T) {
			t.Parallel()
			assertRefused(t, runSetHalt(t, gnokey, tc.remote, "1000", "v1.5.1"), "already meets")
		})
	}

	t.Run("a REMOTE over another scheme stops the proposal", func(t *testing.T) {
		t.Parallel()
		assertRefused(t, runSetHalt(t, gnokey, "ws://"+hostPort, "1000", "v1.5.2"), "cannot read /status over ws://")
	})

	// A listener that drops every request, rather than a closed port the OS
	// could hand to another test's server.
	dropped := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))
	t.Cleanup(dropped.Close)

	t.Run("a remote that does not answer stops the proposal", func(t *testing.T) {
		t.Parallel()
		assertRefused(t, runSetHalt(t, gnokey, dropped.URL, "1000", "v1.5.2"), "cannot read the version REMOTE runs")
	})

	// gnokey dials REMOTE directly, so the version must be read the same way:
	// a proxy in the environment would block the check or read another node.
	t.Run("a proxy in the environment is not used to read the version", func(t *testing.T) {
		t.Parallel()

		proxy := []string{"NO_PROXY=", "no_proxy="}
		for _, name := range []string{"http_proxy", "HTTP_PROXY", "https_proxy", "HTTPS_PROXY", "all_proxy", "ALL_PROXY"} {
			proxy = append(proxy, name+"="+dropped.URL)
		}
		assertRefused(t, runSetHalt(t, gnokey, live, "1000", "v1.5.1", proxy...), "already meets")
	})

	t.Run("a status without a build version stops the proposal", func(t *testing.T) {
		t.Parallel()
		assertRefused(t, runSetHalt(t, gnokey, statusServer(t, ""), "1000", "v1.5.2"), "build_version")
	})

	t.Run("cancelling sets no floor, so nothing is compared", func(t *testing.T) {
		t.Parallel()

		res := runSetHalt(t, gnokey, dropped.URL, "0", "")
		assert.NoError(t, res.err, res.out)
		assert.True(t, res.reachedGnokey, res.out)
	})
}

// A collating locale must not widen the floor's shape check: under en_US.UTF-8
// on glibc, [A-Za-z] matches é, and a floor the node cannot parse refuses the
// upgraded binary too. Where en_US.UTF-8 does not widen the range, as on macOS,
// or is not installed, there is nothing to prove and the test skips.
func TestSetHaltChecksTheFloorShapeInTheCLocale(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("bash"); err != nil {
		t.Skipf("set-halt.sh needs bash: %v", err)
	}
	probe := exec.Command("bash", "-c", `[[ é =~ ^[A-Za-z]$ ]]`)
	probe.Env = append(os.Environ(), "LC_ALL=en_US.UTF-8")
	if probe.Run() != nil {
		t.Skip("en_US.UTF-8 does not widen [A-Za-z] here, so the C locale has nothing to fix")
	}
	_, parses := parseReleaseVersion("v1.0.0-é")
	require.False(t, parses)

	res := runSetHalt(t, gnokeyStub(t), statusServer(t, "v1.5.1"), "1000", "v1.0.0-é", "LC_ALL=en_US.UTF-8")
	assertRefused(t, res, "MIN_VERSION must be vMAJOR.MINOR.PATCH")
}

// statusServer serves the node's RPC /status, reporting version as its
// build_version.
func statusServer(t *testing.T, version string) string {
	t.Helper()

	funcMap := map[string]*rs.RPCFunc{
		"status": rs.NewRPCFunc(func(_ *rpctypes.Context) (*rpccore.ResultStatus, error) {
			return &rpccore.ResultStatus{BuildVersion: version}, nil
		}, ""),
	}
	mux := http.NewServeMux()
	rs.RegisterRPCFuncs(mux, funcMap, log.NewNoopLogger())

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv.URL
}

// gnokeyStub writes a gnokey that only touches $GNOKEY_REACHED, so nothing is
// signed or broadcast, and returns its directory. One stub serves every run:
// macOS assesses each new executable on its first launch, which costs about
// as much as the run itself.
func gnokeyStub(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	stub := "#!/bin/sh\ntouch \"$GNOKEY_REACHED\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "gnokey"), []byte(stub), 0o755))
	return dir
}

type setHaltResult struct {
	out           string
	err           error
	reachedGnokey bool
}

// runSetHalt runs set-halt.sh with the gnokey stub in gnokeyDir first on PATH
// and env added to the environment. An empty floor is the same as none.
func runSetHalt(t *testing.T, gnokeyDir, remote, height, floor string, env ...string) setHaltResult {
	t.Helper()

	marker := filepath.Join(t.TempDir(), "gnokey-reached")
	cmd := exec.Command("bash", repoPath(setHaltScript), height, floor)
	cmd.Env = append(os.Environ(),
		"PATH="+gnokeyDir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"GNOKEY_NAME=test", "CHAIN_ID=test-1", "REMOTE="+remote, "GNOKEY_REACHED="+marker,
		// curl sends even 127.0.0.1 through a configured proxy.
		"NO_PROXY=127.0.0.1", "no_proxy=127.0.0.1",
	)
	cmd.Env = append(cmd.Env, env...)
	out, err := cmd.CombinedOutput()

	_, statErr := os.Stat(marker)
	return setHaltResult{out: string(out), err: err, reachedGnokey: statErr == nil}
}

// assertRefused checks that set-halt.sh stopped before gnokey, saying reason.
func assertRefused(t *testing.T, res setHaltResult, reason string) {
	t.Helper()

	assert.Error(t, res.err, res.out)
	assert.False(t, res.reachedGnokey, "the proposal must not be made:\n%s", res.out)
	assert.Contains(t, res.out, reason)
}
