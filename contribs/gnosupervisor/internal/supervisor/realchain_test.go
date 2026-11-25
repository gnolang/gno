//go:build realchain

package supervisor

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/contribs/gnosupervisor/internal/oci"
	"github.com/gnolang/gno/contribs/gnosupervisor/internal/oci/ocitest"
	"github.com/gnolang/gno/gno.land/pkg/upgrades"
)

// TestRealChainReplay replays a real local chain across a real GovDAO halt.
// scripts/devchain.sh up starts that chain: a validator that ran OLD_BIN up
// to the halt HALT_HEIGHT (floor HALT_MIN_VERSION) and NEW_BIN after it. The
// test serves both binaries as release images from an in-process registry,
// then runs the supervisor on a fresh node that syncs from genesis: OLD_BIN
// must stop at the halt with code 2, and NEW_BIN must carry the node past it.
//
//	REALCHAIN_ENV=<dir>/run/nodeb.env \
//	    go test -tags realchain -run TestRealChainReplay -v ./internal/supervisor/
func TestRealChainReplay(t *testing.T) {
	envPath := os.Getenv("REALCHAIN_ENV")
	if envPath == "" {
		t.Skip("set REALCHAIN_ENV to the nodeb.env scripts/devchain.sh writes")
	}
	env := readEnv(t, envPath)
	for _, k := range []string{"GNOROOT", "CHAIN_ID", "GENESIS", "HALT_HEIGHT", "HALT_MIN_VERSION", "OLD_BIN", "NEW_BIN", "NODE_A_P2P"} {
		require.NotEmpty(t, env[k], k)
	}
	halt := mustInt(t, env["HALT_HEIGHT"])
	oldVersion := binaryVersion(t, env["OLD_BIN"])
	newVersion := binaryVersion(t, env["NEW_BIN"])
	require.Equal(t, env["HALT_MIN_VERSION"], newVersion)

	// Release images, built the way the gnoland Dockerfile lays them out.
	reg := ocitest.New(t)
	stdlibs := treeLayer(t, env["GNOROOT"], "gnovm/stdlibs")
	image := func(bin string) string {
		b, err := os.ReadFile(bin)
		require.NoError(t, err)
		return reg.PushImage(t, []string{"linux/amd64", "linux/arm64"},
			ocitest.Layer(t, ocitest.File{Name: "usr/bin/gnoland", Body: string(b), Mode: 0o755}),
			stdlibs)
	}
	oldDigest, newDigest := image(env["OLD_BIN"]), image(env["NEW_BIN"])

	ledgerPath := filepath.Join(t.TempDir(), upgrades.LedgerFile)
	writeLedger(t, ledgerPath, env["CHAIN_ID"], env["GENESIS"], halt,
		[2]string{oldVersion, oldDigest}, [2]string{newVersion, newDigest})
	plan, err := LoadPlan(ledgerPath, env["GENESIS"])
	require.NoError(t, err)

	// A fresh node that only knows node A as a peer.
	dataDir := t.TempDir()
	gnoland := func(args ...string) {
		out, err := exec.Command(env["OLD_BIN"], args...).CombinedOutput()
		require.NoError(t, err, string(out))
	}
	cfgPath := filepath.Join(dataDir, "config", "config.toml")
	gnoland("config", "init", "--config-path", cfgPath)
	gnoland("secrets", "init", "--data-dir", filepath.Join(dataDir, "secrets"))
	for k, v := range map[string]string{
		"rpc.laddr":              "tcp://127.0.0.1:36757",
		"p2p.laddr":              "tcp://127.0.0.1:36756",
		"p2p.pex":                "false",
		"p2p.allow_duplicate_ip": "true",
		"p2p.persistent_peers":   env["NODE_A_P2P"],
	} {
		gnoland("config", "set", "--config-path", cfgPath, k, v)
	}
	node, err := OpenNode(dataDir)
	require.NoError(t, err)

	client := oci.NewClient()
	client.BaseURL = reg.URL
	lock, err := Lock(dataDir)
	require.NoError(t, err)
	defer lock.Close()
	logs := &syncBuffer{}
	cfg := Config{
		Plan:        plan,
		Node:        node,
		Artifacts:   &Store{Dir: t.TempDir(), Platform: "linux/amd64", Client: client, Log: slogTo(logs)},
		DataDir:     dataDir,
		Genesis:     env["GENESIS"],
		ExtraArgs:   []string{"--skip-genesis-sig-verification", "--log-level", "info"},
		LockFile:    lock,
		StopTimeout: 30 * time.Second,
		MaxRestarts: 3,
		BackoffMin:  time.Second,
		BackoffMax:  5 * time.Second,
		// The real RPC watchers: the chain's halt param stays at the halt
		// the ledger records, which must not be taken for an unknown one.
		Live:         node.LiveHeight(),
		HaltParam:    node.HaltParam(),
		PollInterval: time.Second,
		StuckAfter:   2 * time.Minute,
		Stdout:       logs,
		Stderr:       logs,
		Log:          slogTo(logs),
	}
	defer func() {
		if t.Failed() {
			t.Logf("supervisor and node logs (tail):\n%s", tail(logs.String(), 80))
		}
	}()

	runUntil := func(target int64) {
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() { done <- Run(ctx, cfg) }()
		deadline := time.After(5 * time.Minute)
		for {
			select {
			case err := <-done:
				cancel()
				require.NoError(t, err)
				t.Fatalf("supervisor returned before height %d", target)
			case <-deadline:
				cancel()
				<-done
				t.Fatalf("height %d not reached in time", target)
			case <-time.After(time.Second):
				if h, _ := rpcHeight(context.Background(), http.DefaultClient, "http://127.0.0.1:36757/status"); h >= target {
					cancel()
					require.NoError(t, <-done)
					return
				}
			}
		}
	}

	// From genesis across the halt.
	runUntil(halt + 10)
	out := logs.String()
	assert.Contains(t, out, "msg=\"starting gnoland\" version="+oldVersion+" next_block=1 ")
	assert.Contains(t, out, "panic: halt height ")
	assert.Contains(t, out, "msg=\"halt height reached; switching release\" version="+oldVersion+" halt_height="+env["HALT_HEIGHT"])
	assert.Contains(t, out, "msg=\"starting gnoland\" version="+newVersion+" next_block="+itoa(halt+1)+" ")
	// Block sync defers its writes and a stop drops the ones not yet flushed
	// (tm2/adr/pr6334_deferred_block_sync_writes.md), so the state on disk
	// may be behind what RPC served; the flush point at the halt height
	// guarantees the halt itself is on disk.
	h, err := node.CommittedHeight()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, h, halt)

	// A restarted supervisor resumes on the new release, from the cache.
	served := reg.Total.Load()
	runUntil(halt + 15)
	assert.Contains(t, logs.String()[len(out):], "msg=\"starting gnoland\" version="+newVersion+" next_block=")
	assert.NotContains(t, logs.String()[len(out):], "version="+oldVersion+" next_block")
	assert.Equal(t, served, reg.Total.Load(), "cached artifacts are not pulled again")
	assert.NoFileExists(t, filepath.Join(dataDir, HaltMarkerFile), "the ledger's halt is not an unknown one")
}

func readEnv(t *testing.T, path string) map[string]string {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	env := map[string]string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if k, v, ok := strings.Cut(line, "="); ok && !strings.HasPrefix(line, "#") {
			env[k] = v
		}
	}
	require.NoError(t, s.Err())
	return env
}

func binaryVersion(t *testing.T, bin string) string {
	t.Helper()
	// The version the node reports is the one its startup gates compare.
	out, err := exec.Command(bin, "version").CombinedOutput()
	if err == nil {
		for f := range strings.FieldsSeq(string(out)) {
			if v, ok := upgrades.ParseVersion(f); ok {
				return v
			}
		}
	}
	base := filepath.Base(bin)
	v, ok := upgrades.ParseVersion(base[strings.LastIndex(base, "-v")+1:])
	require.True(t, ok, "cannot tell the version of %s: %s", bin, out)
	return v
}

// treeLayer packs dir/sub of a gno checkout as gnoroot/sub.
func treeLayer(t *testing.T, root, sub string) []byte {
	t.Helper()
	var files []ocitest.File
	err := filepath.WalkDir(filepath.Join(root, sub), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		files = append(files, ocitest.File{Name: "gnoroot/" + filepath.ToSlash(rel), Body: string(b)})
		return nil
	})
	require.NoError(t, err)
	return ocitest.Layer(t, files...)
}

func writeLedger(t *testing.T, path, chainID, genesis string, halt int64, oldRel, newRel [2]string) {
	t.Helper()
	g, err := os.ReadFile(genesis)
	require.NoError(t, err)
	sum := sha256.Sum256(g)
	at := time.Now().UTC().Truncate(time.Second)
	floor := newRel[0]
	proposal := int64(0)
	entry := func(kind upgrades.Kind, rel [2]string) upgrades.Entry {
		d := rel[1]
		return upgrades.Entry{
			Kind:     kind,
			Version:  rel[0],
			Commit:   strings.Repeat("0", 40),
			Image:    upgrades.Image{Ref: "ghcr.io/gnolang/gno/gnoland:" + rel[0], Digest: &d},
			Binaries: map[string]string{"linux/amd64": "https://example.invalid/gnoland?checksum=sha256:" + strings.Repeat("0", 64)},
			Release:  "https://github.com/gnolang/gno/releases/tag/" + rel[0],
		}
	}
	up := entry(upgrades.KindUpgrade, newRel)
	up.HaltHeight, up.HaltTime, up.HaltMinVersion, up.Proposal = &halt, &at, &floor, &proposal
	l := upgrades.Ledger{
		SchemaVersion: upgrades.SchemaVersion,
		ChainID:       chainID,
		GenesisSHA256: hex.EncodeToString(sum[:]),
		GenesisTime:   at.Add(-time.Hour),
		Upgrades:      []upgrades.Entry{entry(upgrades.KindGenesis, oldRel), up},
	}
	require.NoError(t, l.Validate())
	b, err := json.MarshalIndent(l, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, b, 0o644))
}

func mustInt(t *testing.T, s string) int64 {
	t.Helper()
	n, err := strconv.ParseInt(s, 10, 64)
	require.NoError(t, err)
	return n
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

func tail(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
