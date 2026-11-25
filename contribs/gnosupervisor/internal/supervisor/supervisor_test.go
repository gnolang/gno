package supervisor

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/semver"
)

// The supervisor tests run a fake gnoland: this test binary, re-executed
// with fakeEnv set. It keeps its height in <data-dir>/height and behaves
// like gnoland at a governance halt (see gno.land/cmd/gnoland/UPGRADES.md):
// whatever its version, it commits the halt height and exits with code 2;
// it refuses to start below the floor after a halt, or at or above the
// floor before it; and it fails a block its version does not run, as a
// wrong app hash would.
const (
	fakeEnv      = "GNOSUPERVISOR_FAKE_GNOLAND"
	fakeChainEnv = "GNOSUPERVISOR_FAKE_CHAIN"
)

type fakeChain struct {
	Halts      []fakeHalt  `json:"halts"`
	Ranges     []fakeRange `json:"ranges"`
	Tip        int64       `json:"tip"`      // the network's height: idle there
	CrashAt    int64       `json:"crash_at"` // exit 1 once before this block
	Past       int64       `json:"past"`     // keep committing past halts up to this
	IgnoreTerm bool        `json:"ignore_term"`
	StayAtHalt bool        `json:"stay_at_halt"` // a node in consensus mode
	StayAt     int64       `json:"stay_at"`      // the same, at this halt only
}

type fakeHalt struct {
	Height int64  `json:"height"`
	Floor  string `json:"floor"`
}

type fakeRange struct {
	From, To int64
	Versions []string
}

func TestMain(m *testing.M) {
	if os.Getenv(fakeEnv) == "1" {
		os.Exit(fakeGnoland(os.Args[1:]))
	}
	os.Exit(m.Run())
}

func fakeGnoland(args []string) int {
	if len(args) == 0 || args[0] != "start" {
		return 64
	}
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	dataDir := fs.String("data-dir", "", "")
	fs.String("genesis", "", "")
	gnoroot := fs.String("gnoroot-dir", "", "")
	if err := fs.Parse(args[1:]); err != nil {
		return 64
	}
	var chain fakeChain
	if err := json.Unmarshal([]byte(os.Getenv(fakeChainEnv)), &chain); err != nil {
		return 65
	}
	vb, err := os.ReadFile(filepath.Join(*gnoroot, "VERSION"))
	if err != nil || os.Getenv("GNOROOT") != *gnoroot {
		return 66
	}
	version := string(vb)
	h := readHeight(*dataDir)
	logRun(*dataDir, fmt.Sprintf("start %s %d", version, h))

	// Startup gates, as checkNodeStartupParams.
	for _, halt := range chain.Halts {
		if halt.Floor == "" {
			continue
		}
		if h >= halt.Height && semver.Compare(version, halt.Floor) < 0 {
			fmt.Fprintf(os.Stderr, "binary version %q does not meet the minimum version %q\n", version, halt.Floor)
			return 1
		}
		if h < halt.Height && semver.Compare(version, halt.Floor) >= 0 && h+1 >= halt.Height-5 {
			// The proposal executed a few blocks before the halt.
			fmt.Fprintf(os.Stderr, "binary version %q is an upgrade intended for halt height %d\n", version, halt.Height)
			return 1
		}
	}

	term := make(chan os.Signal, 1)
	signal.Notify(term, syscall.SIGTERM)
	crashMarker := filepath.Join(*dataDir, "crashed")
	for b := h + 1; ; b++ {
		if b > chain.Tip {
			// At the network's tip: run until stopped.
			for {
				<-term
				if !chain.IgnoreTerm {
					return 0
				}
			}
		}
		if b == chain.CrashAt {
			if _, err := os.Stat(crashMarker); err != nil {
				os.WriteFile(crashMarker, nil, 0o644)
				return 1
			}
		}
		if !runs(chain, b, version) {
			fmt.Fprintf(os.Stderr, "wrong Block.Header.AppHash at %d\n", b)
			return 1
		}
		writeHeight(*dataDir, b)
		for _, halt := range chain.Halts {
			if halt.Height == b && b >= chain.Past {
				if chain.StayAtHalt || chain.StayAt == b {
					fmt.Fprintf(os.Stderr, "CONSENSUS FAILURE!!! halt height %d reached\n", b)
					<-term
					return 0
				}
				fmt.Fprintf(os.Stderr, "panic: halt height %d reached, node shutting down\n", b)
				return 2
			}
		}
	}
}

func runs(chain fakeChain, b int64, version string) bool {
	for _, r := range chain.Ranges {
		if b >= r.From && (r.To == 0 || b <= r.To) {
			for _, v := range r.Versions {
				if v == version {
					return true
				}
			}
			return false
		}
	}
	return false
}

func readHeight(dataDir string) int64 {
	b, err := os.ReadFile(filepath.Join(dataDir, "height"))
	if err != nil {
		return 0
	}
	h, _ := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64)
	return h
}

func writeHeight(dataDir string, h int64) {
	os.WriteFile(filepath.Join(dataDir, "height"), []byte(strconv.FormatInt(h, 10)), 0o644)
}

func logRun(dataDir, line string) {
	f, _ := os.OpenFile(filepath.Join(dataDir, "runs.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	fmt.Fprintln(f, line)
	f.Close()
}

type fakeHeights string

func (d fakeHeights) CommittedHeight() (int64, error) { return readHeight(string(d)), nil }

// fakeArtifacts hands out this test binary with a GNOROOT that names the
// version, and records what it prepared.
type fakeArtifacts struct {
	dir      string
	mu       sync.Mutex
	prepared []string
	fail     map[string]error
}

func (f *fakeArtifacts) Prepare(_ context.Context, leg Leg) (Artifact, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.prepared = append(f.prepared, leg.Version)
	if err := f.fail[leg.Version]; err != nil {
		return Artifact{}, err
	}
	root := filepath.Join(f.dir, leg.Version)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return Artifact{}, err
	}
	if err := os.WriteFile(filepath.Join(root, "VERSION"), []byte(leg.Version), 0o644); err != nil {
		return Artifact{}, err
	}
	exe, err := os.Executable()
	if err != nil {
		return Artifact{}, err
	}
	return Artifact{Version: leg.Version, Digest: *leg.Entry.Image.Digest, Binary: exe, GnoRoot: root}, nil
}

// syncBuffer is a goroutine-safe log sink.
type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

type harness struct {
	dataDir string
	cfg     Config
	logs    *syncBuffer
	arts    *fakeArtifacts
}

// The ledger every supervisor test replays: three coordinated releases and a
// rolling patch, with floors on the last two halts as on mainnet since
// v1.7.0.
var testEntries = []fixtureEntry{
	{version: "v1.0.0"},
	{version: "v1.1.0", halt: 30},
	{version: "v1.1.1", halt: rolling},
	{version: "v1.2.0", halt: 60},
}

var testChain = fakeChain{
	Halts:  []fakeHalt{{Height: 30}, {Height: 60, Floor: "v1.2.0"}},
	Ranges: []fakeRange{{1, 30, []string{"v1.0.0"}}, {31, 60, []string{"v1.1.0", "v1.1.1"}}, {61, 0, []string{"v1.2.0"}}},
	Tip:    90,
}

func newHarness(t *testing.T, chain fakeChain) *harness {
	t.Helper()
	data, err := json.Marshal(chain)
	require.NoError(t, err)
	t.Setenv(fakeEnv, "1")
	t.Setenv(fakeChainEnv, string(data))

	dataDir := t.TempDir()
	logs := &syncBuffer{}
	arts := &fakeArtifacts{dir: t.TempDir()}
	return &harness{
		dataDir: dataDir,
		logs:    logs,
		arts:    arts,
		cfg: Config{
			Plan:        newPlan(fixtureLedger(t, testSHA, testEntries...)),
			Node:        fakeHeights(dataDir),
			Artifacts:   arts,
			DataDir:     dataDir,
			Genesis:     filepath.Join(dataDir, "genesis.json"),
			StopTimeout: 5 * time.Second,
			MaxRestarts: 3,
			BackoffMin:  time.Millisecond,
			BackoffMax:  10 * time.Millisecond,
			Stdout:      logs,
			Stderr:      logs,
			Log:         slogTo(logs),
		},
	}
}

// run runs the supervisor until the node idles at the tip (then stops it
// as an operator would) or until Run returns on its own.
func (h *harness) run(t *testing.T, tip int64) error {
	t.Helper()
	return h.runUntil(t, func() bool { return readHeight(h.dataDir) >= tip })
}

func (h *harness) runUntil(t *testing.T, stop func() bool) error {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- Run(ctx, h.cfg) }()
	deadline := time.After(30 * time.Second)
	for {
		select {
		case err := <-done:
			return err
		case <-deadline:
			cancel()
			<-done
			t.Fatalf("timeout; logs:\n%s", h.logs.String())
		case <-time.After(5 * time.Millisecond):
			if stop() {
				cancel()
				return <-done
			}
		}
	}
}

func (h *harness) runs(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(h.dataDir, "runs.log"))
	require.NoError(t, err)
	return strings.Split(strings.TrimSpace(string(b)), "\n")
}

func slogTo(w io.Writer) *slog.Logger { return slog.New(slog.NewTextHandler(w, nil)) }

func TestRunReplaysAcrossHalts(t *testing.T) {
	h := newHarness(t, testChain)
	require.NoError(t, h.run(t, testChain.Tip))

	assert.Equal(t, []string{"start v1.0.0 0", "start v1.1.1 30", "start v1.2.0 60"}, h.runs(t),
		"genesis release to 30, the last rolling patch to 60, then v1.2.0; logs:\n%s", h.logs.String())
	assert.Equal(t, testChain.Tip, readHeight(h.dataDir))
	assert.Equal(t, 2, strings.Count(h.logs.String(), "halt height reached; switching release"))
	// The next release is prepared before the halt that needs it.
	assert.Equal(t, []string{"v1.0.0", "v1.1.1", "v1.1.1", "v1.2.0", "v1.2.0"}, h.arts.prepared)
}

func TestRunResumesFromTheDataDirectory(t *testing.T) {
	h := newHarness(t, testChain)
	writeHeight(h.dataDir, 45)
	require.NoError(t, h.run(t, testChain.Tip))
	assert.Equal(t, []string{"start v1.1.1 45", "start v1.2.0 60"}, h.runs(t))
}

func TestRunRestartsAfterACrash(t *testing.T) {
	chain := testChain
	chain.CrashAt = 40
	h := newHarness(t, chain)
	require.NoError(t, h.run(t, chain.Tip))
	assert.Equal(t, []string{"start v1.0.0 0", "start v1.1.1 30", "start v1.1.1 39", "start v1.2.0 60"}, h.runs(t))
	assert.Contains(t, h.logs.String(), "gnoland exited unexpectedly; restarting the same release")
}

// A halt the ledger does not record stops a node that is catching up, like
// any halt. It may be an upgrade missing from this copy of the ledger, so the
// supervisor stops; the operator says it was a pause with
// --continue-after-halt, and the same release goes on.
func TestRunPauseHaltNotInLedger(t *testing.T) {
	chain := testChain
	chain.Halts = append([]fakeHalt{{Height: 20}}, chain.Halts...)
	h := newHarness(t, chain)
	err := h.run(t, chain.Tip)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--continue-after-halt 20")
	assert.Equal(t, []string{"start v1.0.0 0"}, h.runs(t))

	// The refusal outlives the process: running the same command again
	// (by hand, or a restart policy) refuses too, and starts nothing.
	err = h.run(t, chain.Tip)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--continue-after-halt 20")
	assert.Equal(t, []string{"start v1.0.0 0"}, h.runs(t))

	h.cfg.ContinueAfterHalt = 20
	require.NoError(t, h.run(t, chain.Tip))
	assert.Equal(t, []string{"start v1.0.0 0", "start v1.0.0 20", "start v1.1.1 30", "start v1.2.0 60"}, h.runs(t))
}

// A halt beyond the ledger's last range (the ledger lags the chain) stops the
// supervisor instead of restarting the old release, with a floor or without
// one: without one, nothing in the node would refuse the old release, which
// would run the next release's blocks.
func TestRunStopsAtAHaltTheLedgerDoesNotKnow(t *testing.T) {
	for _, floor := range []string{"v1.3.0", ""} {
		t.Run("floor "+floor, func(t *testing.T) {
			chain := testChain
			chain.Halts = append(chain.Halts, fakeHalt{Height: 75, Floor: floor})
			h := newHarness(t, chain)
			err := h.run(t, chain.Tip)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "does not record as a halt")
			assert.Equal(t, int64(75), readHeight(h.dataDir))
			assert.Equal(t, []string{"start v1.0.0 0", "start v1.1.1 30", "start v1.2.0 60"}, h.runs(t), "never restarted past 75")
		})
	}
}

// Once the ledger records the halt the supervisor stopped at, it goes on
// with the next release, and the record of the stop goes away.
func TestRunResumesOnceTheLedgerRecordsTheHalt(t *testing.T) {
	chain := testChain
	chain.Halts = append(chain.Halts, fakeHalt{Height: 75, Floor: "v1.3.0"})
	chain.Ranges = []fakeRange{chain.Ranges[0], chain.Ranges[1], {61, 75, []string{"v1.2.0"}}, {76, 0, []string{"v1.3.0"}}}
	h := newHarness(t, chain)
	require.Error(t, h.run(t, chain.Tip))

	h.cfg.Plan = newPlan(fixtureLedger(t, testSHA, append(testEntries, fixtureEntry{version: "v1.3.0", halt: 75})...))
	require.NoError(t, h.run(t, chain.Tip))
	assert.Equal(t, []string{"start v1.0.0 0", "start v1.1.1 30", "start v1.2.0 60", "start v1.3.0 75"}, h.runs(t))
	assert.NoFileExists(t, filepath.Join(h.dataDir, HaltMarkerFile))
}

// A node at the tip (in consensus) that halts at a height the ledger does not
// record stays up. The supervisor reads the chain's halt height, records the
// stop, and a restarted supervisor refuses to run the release past it.
func TestRunRecordsAnUnknownHaltAtTheTip(t *testing.T) {
	chain := testChain
	chain.StayAtHalt = true
	chain.Halts = append(chain.Halts, fakeHalt{Height: 75}) // no floor
	h := newHarness(t, chain)
	h.cfg.Live = func(context.Context) (int64, error) { return readHeight(h.dataDir), nil }
	h.cfg.HaltParam = func(context.Context) (int64, error) {
		if readHeight(h.dataDir) >= 75 {
			return 75, nil
		}
		return 60, nil
	}
	h.cfg.PollInterval = 5 * time.Millisecond
	h.cfg.StuckAfter = 50 * time.Millisecond
	marker := filepath.Join(h.dataDir, HaltMarkerFile)
	require.NoError(t, h.runUntil(t, func() bool { _, err := os.Stat(marker); return err == nil }))
	assert.Contains(t, h.logs.String(), "the chain halted at a height the ledger does not record")

	err := h.run(t, chain.Tip)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--continue-after-halt 75")
	assert.Equal(t, []string{"start v1.0.0 0", "start v1.1.1 30", "start v1.2.0 60"}, h.runs(t), "not restarted past 75")
}

// The stop is recorded as soon as the node reaches the chain's halt height,
// without waiting for it to look idle; and when the supervisor is told to
// stop first, it records the stop before stopping the node.
func TestRunRecordsAnUnknownHaltAtOnce(t *testing.T) {
	for name, poll := range map[string]time.Duration{"by the watcher": 5 * time.Millisecond, "on stop": time.Hour} {
		t.Run(name, func(t *testing.T) {
			chain := testChain
			chain.StayAt = 75
			chain.Halts = append(chain.Halts, fakeHalt{Height: 75}) // no floor
			h := newHarness(t, chain)
			h.cfg.Live = func(context.Context) (int64, error) { return readHeight(h.dataDir), nil }
			h.cfg.HaltParam = func(context.Context) (int64, error) {
				if readHeight(h.dataDir) >= 75 {
					return 75, nil
				}
				return 60, nil // the last halt's param stays set
			}
			h.cfg.PollInterval = poll
			h.cfg.StuckAfter = time.Hour
			marker := filepath.Join(h.dataDir, HaltMarkerFile)
			if poll == time.Hour {
				require.NoError(t, h.run(t, 75)) // the operator stops it at the halt
			} else {
				require.NoError(t, h.runUntil(t, func() bool { _, err := os.Stat(marker); return err == nil }))
			}
			assert.FileExists(t, marker)
			assert.Equal(t, 1, strings.Count(h.logs.String(), "does not record; this release will not be restarted"),
				"no false alarm at the ledger's halts (30, 60)")
			err := h.run(t, chain.Tip)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "--continue-after-halt 75")
		})
	}
}

func TestHaltMarkerKeptUntilTheHalt(t *testing.T) {
	dir := t.TempDir()
	c := Config{DataDir: dir, Plan: newPlan(fixtureLedger(t, testSHA, testEntries...))}
	require.NoError(t, writeHaltMarker(dir, 75))
	require.NoError(t, c.checkHaltMarker(70), "below the halt: go on, the replay stops there again")
	assert.FileExists(t, filepath.Join(dir, HaltMarkerFile))
	assert.Error(t, c.checkHaltMarker(75))
	require.NoError(t, c.checkHaltMarker(80), "past it")
	assert.NoFileExists(t, filepath.Join(dir, HaltMarkerFile))
}

// A node that left block sync for consensus before a halt catches the halt
// panic and stays up. The supervisor notices it idle at the range's end,
// stops it and switches release.
func TestRunStopsANodeIdleAtTheHalt(t *testing.T) {
	chain := testChain
	chain.StayAtHalt = true
	h := newHarness(t, chain)
	h.cfg.Live = func(context.Context) (int64, error) { return readHeight(h.dataDir), nil }
	h.cfg.PollInterval = 5 * time.Millisecond
	h.cfg.StuckAfter = 50 * time.Millisecond
	require.NoError(t, h.run(t, chain.Tip))
	assert.Equal(t, []string{"start v1.0.0 0", "start v1.1.1 30", "start v1.2.0 60"}, h.runs(t))
	assert.Equal(t, 2, strings.Count(h.logs.String(), "idle at its range's halt height"))
}

func TestRunGivesUpOnACrashLoop(t *testing.T) {
	// The fake refuses every block of the first range: no progress at all.
	chain := testChain
	chain.Ranges = append([]fakeRange{{1, 30, []string{"v0.0.0"}}}, chain.Ranges[1:]...)
	h := newHarness(t, chain)
	err := h.run(t, chain.Tip)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exited 4 times in a row at height 0")
	assert.Len(t, h.runs(t), 4)
}

func TestRunRefusesAHeightPastTheRange(t *testing.T) {
	// The fake does not halt at 30: the node ran blocks that belong to the
	// next release. The supervisor stops rather than guess.
	chain := testChain
	chain.Ranges = []fakeRange{{1, 0, []string{"v1.0.0"}}}
	chain.Halts = []fakeHalt{{Height: 35}}
	h := newHarness(t, chain)
	err := h.run(t, chain.Tip)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "committed block 35, past the end of its range (30)")
}

func TestRunStopsOnArtifactFailure(t *testing.T) {
	h := newHarness(t, testChain)
	h.arts.fail = map[string]error{"v1.0.0": fmt.Errorf("digest mismatch")}
	err := h.run(t, testChain.Tip)
	assert.ErrorContains(t, err, "digest mismatch")
	_, statErr := os.Stat(filepath.Join(h.dataDir, "runs.log"))
	assert.ErrorIs(t, statErr, os.ErrNotExist, "nothing started")
}

func TestRunKillsAChildThatIgnoresSIGTERM(t *testing.T) {
	chain := testChain
	chain.IgnoreTerm = true
	h := newHarness(t, chain)
	h.cfg.StopTimeout = 200 * time.Millisecond
	start := time.Now()
	require.NoError(t, h.run(t, chain.Tip))
	assert.Less(t, time.Since(start), 20*time.Second)
	assert.Contains(t, h.logs.String(), "gnoland did not stop in time; killing it")
}
