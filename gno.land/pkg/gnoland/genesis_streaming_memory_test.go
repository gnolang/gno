package gnoland

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	rpccore "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rs "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixtureEnvVar names the env var that locates a local-only large genesis
// fixture used to verify the streaming path stays bounded under production-
// shape input. The fixture is intentionally outside the repo (too large) —
// tests gated on it skip when the env var is unset.
const fixtureEnvVar = "GNO_GENESIS_MEMORY_FIXTURE"

// requireRealFixture skips the test when running with -short or when the
// fixture path env var is unset / unreadable. Returns the fixture path.
func requireRealFixture(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("memory-bound test is opt-in: skipped under -short")
	}
	path := os.Getenv(fixtureEnvVar)
	if path == "" {
		t.Skipf("memory-bound test skipped: set %s to a large genesis.json", fixtureEnvVar)
	}
	if _, err := os.Stat(path); err != nil {
		t.Skipf("fixture at %s (%s) not readable: %v", fixtureEnvVar, path, err)
	}
	return path
}

// peakHeapSampler runs a goroutine that calls runtime.ReadMemStats on a
// tight-ish interval and tracks the maximum HeapInuse seen. It returns a
// stop function that joins the goroutine and returns the peak in bytes.
//
// The interval is small enough to catch transient allocation spikes on the
// streaming path but large enough to not dominate runtime.GC scheduling.
func peakHeapSampler(interval time.Duration) (stop func() uint64) {
	var peak atomic.Uint64
	done := make(chan struct{})
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)
		var ms runtime.MemStats
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				// One last read so the very tail of the run is included.
				runtime.ReadMemStats(&ms)
				if ms.HeapInuse > peak.Load() {
					peak.Store(ms.HeapInuse)
				}
				return
			case <-ticker.C:
				runtime.ReadMemStats(&ms)
				if ms.HeapInuse > peak.Load() {
					peak.Store(ms.HeapInuse)
				}
			}
		}
	}()

	return func() uint64 {
		close(done)
		<-stopped
		return peak.Load()
	}
}

// settledBaseline returns a HeapInuse reading after forcing GC and releasing
// memory to the OS. Used to produce a per-test baseline so prior allocations
// (test setup, fixture warm-up) don't bleed into the bound assertion.
func settledBaseline() uint64 {
	runtime.GC()
	debug.FreeOSMemory()
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.HeapInuse
}

// TestLoadStreamingGenesisDoc_PeakHeapBound runs the cold-path preprocessing
// pass on the 192 MB real fixture and asserts that peak heap stays well
// under the would-be in-memory size of the source. The token-walking decoder
// should remain O(1) regardless of fixture size; this guards against a
// future regression that buffers the bulk arrays.
func TestLoadStreamingGenesisDoc_PeakHeapBound(t *testing.T) {
	src := requireRealFixture(t)

	// Force GC + release before measuring so prior test setup doesn't bleed
	// in. baseline is the floor we subtract from peak.
	baseline := settledBaseline()
	stop := peakHeapSampler(2 * time.Millisecond)

	cacheRoot := t.TempDir()
	doc, err := LoadStreamingGenesisDoc(src, cacheRoot, nil)
	peak := stop()
	require.NoError(t, err)
	require.NotNil(t, doc.AppState, "streaming load must attach a non-nil AppState")

	delta := peak - baseline
	t.Logf("LoadStreamingGenesisDoc cold path: baseline=%d peak=%d delta=%d (fixture=%dB)",
		baseline, peak, delta, fileSize(t, src))

	const bound = 50 * 1024 * 1024
	require.Less(t, delta, uint64(bound),
		"cold-path peak heap delta %d B exceeds bound %d B — token walker may be buffering bulk arrays",
		delta, bound)
}

// TestServeGenesis_PeakHeapBound spins up the actual RPC handler against a
// preprocessed fixture and serves the body to a deliberately-slow reader.
// The slow reader maintains backpressure so the streaming path is actively
// running while the sampler observes the heap — without it, the OS socket
// buffer absorbs the entire response and we'd be measuring nothing.
func TestServeGenesis_PeakHeapBound(t *testing.T) {
	src := requireRealFixture(t)

	// Pre-warm the cache so the test only measures the serve path. Cold-path
	// memory is covered by TestLoadStreamingGenesisDoc_PeakHeapBound.
	cacheRoot := t.TempDir()
	doc, err := LoadStreamingGenesisDoc(src, cacheRoot, nil)
	require.NoError(t, err)

	// Register a synthetic /genesis method that returns a *ResultGenesis
	// wrapping our streaming-loaded doc. This mirrors what the production
	// rpc/core.Genesis function does, minus the global state.
	res := &rpccore.ResultGenesis{Genesis: doc}
	funcMap := map[string]*rs.RPCFunc{
		"genesis": rs.NewRPCFunc(func(_ *rpctypes.Context) (*rpccore.ResultGenesis, error) {
			return res, nil
		}, ""),
	}
	mux := http.NewServeMux()
	rs.RegisterRPCFuncs(mux, funcMap, log.NewNoopLogger())

	srv := httptest.NewServer(mux)
	defer srv.Close()

	baseline := settledBaseline()
	stop := peakHeapSampler(2 * time.Millisecond)

	resp, err := http.Get(srv.URL + "/genesis")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	// Slow reader: 64 KB chunks with a tiny sleep between reads. This keeps
	// the server actively writing during the measurement window so the
	// sampler sees actual streaming-path allocations rather than
	// already-flushed-and-GC'd transients.
	total := drainSlow(t, resp.Body, 64*1024, 200*time.Microsecond)
	peak := stop()

	delta := peak - baseline
	t.Logf("Serve /genesis on real fixture: baseline=%d peak=%d delta=%d body=%d",
		baseline, peak, delta, total)

	const bound = 50 * 1024 * 1024
	require.Less(t, delta, uint64(bound),
		"/genesis serve peak heap delta %d B exceeds bound %d B — streaming path may be buffering",
		delta, bound)
}

// TestServeGenesis_ConcurrentPeakHeapBound is the test that pins down the
// original DoS scenario: many simultaneous /genesis requests. The pre-
// streaming code path allocated ~170 MB per request via json.MarshalIndent;
// 8 concurrent requests would have peaked near 1.4 GB of transient heap.
// With the streaming hook, each request should peak around the single-
// request delta (~11 MB on the production fixture), so the concurrent
// peak should scale roughly linearly without the per-request blowup.
//
// The slow-reader pattern is critical here: without backpressure the
// requests finish staggered, and the sampler may not catch the moment
// where all 8 are mid-stream simultaneously.
func TestServeGenesis_ConcurrentPeakHeapBound(t *testing.T) {
	src := requireRealFixture(t)

	cacheRoot := t.TempDir()
	doc, err := LoadStreamingGenesisDoc(src, cacheRoot, nil)
	require.NoError(t, err)

	res := &rpccore.ResultGenesis{Genesis: doc}
	funcMap := map[string]*rs.RPCFunc{
		"genesis": rs.NewRPCFunc(func(_ *rpctypes.Context) (*rpccore.ResultGenesis, error) {
			return res, nil
		}, ""),
	}
	mux := http.NewServeMux()
	rs.RegisterRPCFuncs(mux, funcMap, log.NewNoopLogger())

	srv := httptest.NewServer(mux)
	defer srv.Close()

	const concurrency = 8

	// Force GC + release before measuring so the cold load above doesn't
	// bleed into the bound.
	baseline := settledBaseline()
	stop := peakHeapSampler(2 * time.Millisecond)

	var wg sync.WaitGroup
	wg.Add(concurrency)
	totals := make([]int64, concurrency)
	errs := make([]error, concurrency)

	for i := range concurrency {
		go func(idx int) {
			defer wg.Done()
			resp, err := http.Get(srv.URL + "/genesis")
			if err != nil {
				errs[idx] = err
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				errs[idx] = io.EOF
				return
			}
			totals[idx] = drainSlow(t, resp.Body, 64*1024, 200*time.Microsecond)
		}(i)
	}
	wg.Wait()
	peak := stop()

	for i, err := range errs {
		require.NoError(t, err, "concurrent request %d", i)
	}
	for i, n := range totals {
		require.Greater(t, n, int64(0), "concurrent request %d returned empty body", i)
	}

	delta := peak - baseline
	t.Logf("Serve /genesis ×%d concurrent: baseline=%d peak=%d delta=%d (per-request avg body=%d)",
		concurrency, baseline, peak, delta, totals[0])

	// Per-request peak on the real fixture is ~11 MB. 8× = ~90 MB. Set a
	// 250 MB ceiling: triple the linear extrapolation, but still 5× tighter
	// than the pre-streaming behavior (which would have allocated
	// ~170 MB × 8 = 1.4 GB transiently). A regression that re-introduces
	// per-request buffering will trip this trivially.
	const bound = 250 * 1024 * 1024
	require.Less(t, delta, uint64(bound),
		"concurrent /genesis (×%d) peak heap delta %d B exceeds bound %d B — streaming path may be buffering per request",
		concurrency, delta, bound)
}

// drainSlow reads body in chunks of chunkSize, sleeping pause between chunks,
// and returns the total bytes read. Failure to read EOF is fatal.
func drainSlow(t *testing.T, body io.Reader, chunkSize int, pause time.Duration) int64 {
	t.Helper()
	buf := make([]byte, chunkSize)
	var total int64
	for {
		n, err := body.Read(buf)
		total += int64(n)
		if err == io.EOF {
			return total
		}
		require.NoError(t, err)
		if pause > 0 {
			time.Sleep(pause)
		}
		// ctx isn't needed here; the test's overall timeout bounds the run.
		_ = context.Background()
	}
}

func fileSize(t *testing.T, path string) int64 {
	t.Helper()
	st, err := os.Stat(path)
	require.NoError(t, err)
	return st.Size()
}

// TestServeGenesis_SemanticEquivalenceWithSource is the cross-system
// correctness check: load the real fixture through the streaming path,
// serve it through /genesis, and assert the response is semantically
// equivalent to the source genesis.json.
//
// "Semantically equivalent" means: same fields, same values, same array
// orderings — but NOT byte-equivalent. Whitespace, key order in objects,
// and unknown top-level fields ignored by the walker (initial_height) are
// expected differences. Anything else is a bug.
func TestServeGenesis_SemanticEquivalenceWithSource(t *testing.T) {
	src := requireRealFixture(t)

	// 1. Load the source as the reference value.
	srcBytes, err := os.ReadFile(src)
	require.NoError(t, err)
	source := decodeJSONNumberAware(t, srcBytes)

	// 2. Run the source through the full streaming + serve pipeline.
	cacheRoot := t.TempDir()
	doc, err := LoadStreamingGenesisDoc(src, cacheRoot, nil)
	require.NoError(t, err)

	res := &rpccore.ResultGenesis{Genesis: doc}
	funcMap := map[string]*rs.RPCFunc{
		"genesis": rs.NewRPCFunc(func(_ *rpctypes.Context) (*rpccore.ResultGenesis, error) {
			return res, nil
		}, ""),
	}
	mux := http.NewServeMux()
	rs.RegisterRPCFuncs(mux, funcMap, log.NewNoopLogger())
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/genesis")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// /genesis returns {"jsonrpc":"2.0","id":...,"result":{"genesis":{...}}}.
	// Unwrap two layers to get to the doc.
	var envelope struct {
		Result struct {
			Genesis json.RawMessage `json:"genesis"`
		} `json:"result"`
	}
	require.NoError(t, json.Unmarshal(bodyBytes, &envelope))
	require.NotEmpty(t, envelope.Result.Genesis, "envelope.result.genesis must be non-empty")

	served := decodeJSONNumberAware(t, envelope.Result.Genesis)

	// 3. Strip known-ignored top-level fields from the source so we compare
	// apples to apples. The walker silently consumes these (they don't
	// belong to GenesisDoc); the existing GenesisDocFromJSON path does the
	// same.
	ignoredTopLevel := map[string]struct{}{
		"initial_height": {},
	}
	srcMap, ok := source.(map[string]any)
	require.True(t, ok, "source genesis must be a JSON object at top level")
	for k := range ignoredTopLevel {
		delete(srcMap, k)
	}

	// 4. Verify balance and tx arrays match in length AND order before
	// the structural comparison — this gives a clear failure message if
	// the streaming path drops or reorders elements (which would be a
	// catastrophic bug for genesis tx ordering).
	srcAppState, ok := srcMap["app_state"].(map[string]any)
	require.True(t, ok, "source app_state must be a JSON object")
	servedMap, ok := served.(map[string]any)
	require.True(t, ok, "served genesis must be a JSON object at top level")
	servedAppState, ok := servedMap["app_state"].(map[string]any)
	require.True(t, ok, "served app_state must be a JSON object")

	assertSliceLenEqual(t, srcAppState, servedAppState, "balances")
	assertSliceLenEqual(t, srcAppState, servedAppState, "txs")
	assertOrderedArrayEqual(t, srcAppState, servedAppState, "balances")
	assertOrderedArrayEqual(t, srcAppState, servedAppState, "txs")

	// 5. Full structural equality after the targeted checks above. If we
	// reach here and these still differ, it'll be in app_state's small
	// fields (auth/bank/vm) or a top-level field we missed.
	if !reflect.DeepEqual(source, served) {
		// Don't dump the full ~200 MB diff — point at the first divergent
		// top-level key for diagnosis.
		t.Fatalf("served /genesis differs from source: first divergent top-level key = %q",
			firstDivergentKey(srcMap, servedMap))
	}
}

// decodeJSONNumberAware decodes JSON into a generic map[string]any (or
// []any), but with json.Decoder.UseNumber() so numeric values stay as
// json.Number instead of being coerced to float64. This is critical for
// genesis: many fields are stringly-encoded int64 ("815001"), and any
// float64 round-tripping would lose precision and false-fail equality.
func decodeJSONNumberAware(t *testing.T, raw []byte) any {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	require.NoError(t, dec.Decode(&v))
	return v
}

// assertSliceLenEqual asserts that the given key in both maps holds an
// array of the same length. Logs the actual lengths on failure.
func assertSliceLenEqual(t *testing.T, src, served map[string]any, key string) {
	t.Helper()
	srcArr, _ := src[key].([]any)
	servedArr, _ := served[key].([]any)
	require.Equal(t, len(srcArr), len(servedArr),
		"app_state.%s length mismatch: source=%d served=%d", key, len(srcArr), len(servedArr))
}

// assertOrderedArrayEqual asserts that the given key in both maps holds
// an array whose elements are reflect.DeepEqual in the same positions.
// Reports the index of the first divergence on failure.
func assertOrderedArrayEqual(t *testing.T, src, served map[string]any, key string) {
	t.Helper()
	srcArr, _ := src[key].([]any)
	servedArr, _ := served[key].([]any)
	for i := range srcArr {
		if !reflect.DeepEqual(srcArr[i], servedArr[i]) {
			t.Fatalf("app_state.%s[%d] differs between source and served output", key, i)
		}
	}
	// Also exercise the assert package so tooling sees a non-fatal happy path
	// — the require above already short-circuits failures.
	assert.Equal(t, len(srcArr), len(servedArr))
}

// firstDivergentKey returns the first top-level key whose value differs
// between src and served (by reflect.DeepEqual), or "" if none differ at
// the top level (meaning a deeper divergence exists). Used only for
// failure-mode diagnostics, not the assertion itself.
func firstDivergentKey(src, served map[string]any) string {
	for k, v := range src {
		if !reflect.DeepEqual(v, served[k]) {
			return k
		}
	}
	for k := range served {
		if _, ok := src[k]; !ok {
			return "(extra in served) " + k
		}
	}
	return ""
}
