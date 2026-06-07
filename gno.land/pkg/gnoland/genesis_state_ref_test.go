package gnoland

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	core_types "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTestCache creates a synthetic cache directory in t.TempDir() with the
// structure expected by OpenGenesisStateRef. Returns the cache directory
// path. The fixture is intentionally tiny so tests stay fast.
func writeTestCache(t *testing.T, balances, txs []string, smallFields map[string]string) string {
	t.Helper()
	return writeTestCacheWithVersion(t, balances, txs, smallFields, []byte(strconv.Itoa(cacheSchemaVersion)+"\n"))
}

// writeTestCacheWithVersion is like writeTestCache but lets the test specify
// raw bytes for the version file. Pass nil to omit the file entirely.
func writeTestCacheWithVersion(t *testing.T, balances, txs []string, smallFields map[string]string, versionFile []byte) string {
	t.Helper()
	dir := t.TempDir()

	if versionFile != nil {
		require.NoError(t, os.WriteFile(filepath.Join(dir, versionFilename), versionFile, 0o644))
	}

	manifest := genesisCacheManifest{
		SourceHash:   "test-hash",
		BalanceCount: len(balances),
		TxCount:      len(txs),
	}
	manifestBytes, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, manifestFilename), manifestBytes, 0o644))

	envelope := map[string]json.RawMessage{}
	for k, v := range smallFields {
		envelope[k] = json.RawMessage(v)
	}
	envelopeBytes, err := json.Marshal(envelope)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, envelopeFilename), envelopeBytes, 0o644))

	writeJSONL := func(name string, lines []string) {
		var buf bytes.Buffer
		for _, line := range lines {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), buf.Bytes(), 0o644))
	}
	writeJSONL(balancesFilename, balances)
	writeJSONL(txsFilename, txs)

	return dir
}

func TestOpenGenesisStateRef_ValidCache(t *testing.T) {
	dir := writeTestCache(t,
		[]string{`"g1aaa=10ugnot"`, `"g1bbb=20ugnot"`, `"g1ccc=30ugnot"`},
		[]string{`{"msg":"tx1"}`, `{"msg":"tx2"}`},
		map[string]string{
			"@type": `"/gno.GenesisState"`,
			"auth":  `{"params":{}}`,
		},
	)

	ref, err := OpenGenesisStateRef(dir)
	require.NoError(t, err)
	require.NotNil(t, ref)

	assert.Equal(t, 3, ref.BalanceCount())
	assert.Equal(t, 2, ref.TxCount())

	atype, ok := ref.SmallField("@type")
	require.True(t, ok)
	assert.JSONEq(t, `"/gno.GenesisState"`, string(atype))

	auth, ok := ref.SmallField("auth")
	require.True(t, ok)
	assert.JSONEq(t, `{"params":{}}`, string(auth))

	_, ok = ref.SmallField("nonexistent")
	assert.False(t, ok)
}

func TestOpenGenesisStateRef_MissingDir(t *testing.T) {
	_, err := OpenGenesisStateRef("/nonexistent/path/that/does/not/exist")
	require.Error(t, err)
}

func TestOpenGenesisStateRef_MissingManifest(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "envelope.json"), []byte("{}"), 0o644))

	_, err := OpenGenesisStateRef(dir)
	require.Error(t, err)
}

func TestOpenGenesisStateRef_MissingVersion(t *testing.T) {
	// Build a cache with everything except the version file.
	dir := writeTestCacheWithVersion(t, nil, nil, nil, nil)

	_, err := OpenGenesisStateRef(dir)
	require.Error(t, err)
}

func TestOpenGenesisStateRef_VersionMismatch(t *testing.T) {
	// Build a cache stamped with a schema version we don't support.
	bogus := []byte(strconv.Itoa(cacheSchemaVersion+1) + "\n")
	dir := writeTestCacheWithVersion(t, nil, nil, nil, bogus)

	_, err := OpenGenesisStateRef(dir)
	require.Error(t, err)
}

func TestOpenGenesisStateRef_VersionUnparseable(t *testing.T) {
	dir := writeTestCacheWithVersion(t, nil, nil, nil, []byte("not-a-number\n"))

	_, err := OpenGenesisStateRef(dir)
	require.Error(t, err)
}

func TestGenesisStateRef_IterBalances(t *testing.T) {
	balances := []string{`"g1aaa=10ugnot"`, `"g1bbb=20ugnot"`, `"g1ccc=30ugnot"`}
	dir := writeTestCache(t, balances, nil, nil)

	ref, err := OpenGenesisStateRef(dir)
	require.NoError(t, err)

	var got []string
	for line, err := range ref.IterBalances(context.Background()) {
		require.NoError(t, err)
		got = append(got, string(line))
	}
	assert.Equal(t, balances, got)
}

func TestGenesisStateRef_IterBalances_EmptyFile(t *testing.T) {
	dir := writeTestCache(t, nil, nil, nil)

	ref, err := OpenGenesisStateRef(dir)
	require.NoError(t, err)

	var count int
	for _, err := range ref.IterBalances(context.Background()) {
		require.NoError(t, err)
		count++
	}
	assert.Equal(t, 0, count)
}

func TestGenesisStateRef_IterTxs(t *testing.T) {
	txs := []string{`{"msg":"tx1"}`, `{"msg":"tx2"}`}
	dir := writeTestCache(t, nil, txs, nil)

	ref, err := OpenGenesisStateRef(dir)
	require.NoError(t, err)

	var got []string
	for line, err := range ref.IterTxs(context.Background()) {
		require.NoError(t, err)
		got = append(got, string(line))
	}
	assert.Equal(t, txs, got)
}

func TestGenesisStateRef_IterBalances_ContextCancelled(t *testing.T) {
	balances := []string{`"g1aaa=10ugnot"`, `"g1bbb=20ugnot"`, `"g1ccc=30ugnot"`}
	dir := writeTestCache(t, balances, nil, nil)

	ref, err := OpenGenesisStateRef(dir)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var sawCancellation bool
	var consumed int
	for _, err := range ref.IterBalances(ctx) {
		if err != nil {
			assert.ErrorIs(t, err, context.Canceled)
			sawCancellation = true
			break
		}
		consumed++
	}
	assert.True(t, sawCancellation, "iterator must yield context.Canceled when ctx is cancelled before iteration")
	assert.Equal(t, 0, consumed, "no lines should be consumed when ctx is cancelled before iteration starts")
}

// slimFixturePath is a real-shape gnoland genesis.json trimmed to 5
// balances and 2 txs. The fixture lives on disk; tests copy it into a
// temp dir to avoid mutating the testdata file. The slim fixture has the
// same structural shape as a 200 MB production genesis (typed validators
// with PubKeys, ConsensusParams, ed25519 etc.), so streaming behavior is
// exercised authentically.
const slimFixturePath = "testdata/slim_genesis.json"

const (
	slimFixtureBalanceCount = 5
	slimFixtureTxCount      = 2
)

// copySlimFixture copies the slim genesis.json fixture into a fresh temp
// dir and returns the destination path. Streaming copy — the fixture
// itself is small but the test never assumes it fits in memory.
func copySlimFixture(t *testing.T) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "genesis.json")
	src, err := os.Open(slimFixturePath)
	require.NoError(t, err)
	defer src.Close()
	out, err := os.Create(dst)
	require.NoError(t, err)
	defer out.Close()
	_, err = io.Copy(out, src)
	require.NoError(t, err)
	return dst
}

func TestLoadStreamingGenesisDoc_PopulatesSmallFields(t *testing.T) {
	src := copySlimFixture(t)
	cacheRoot := t.TempDir()

	doc, err := LoadStreamingGenesisDoc(src, cacheRoot, nil)
	require.NoError(t, err)
	require.NotNil(t, doc)

	// Top-level small fields are decoded onto the doc, not stuffed in the envelope.
	assert.Equal(t, "test-13", doc.ChainID)
	assert.False(t, doc.GenesisTime.IsZero())
	require.NotEmpty(t, doc.Validators, "real fixture has validators")
	assert.NotZero(t, doc.Validators[0].Power)
	assert.False(t, doc.Validators[0].Address.IsZero())
	require.NotNil(t, doc.Validators[0].PubKey, "validator pub_key decoded via amino interface dispatch")

	// AppState is the streaming ref, not an in-memory typed struct.
	ref, ok := doc.AppState.(*GenesisStateRef)
	require.True(t, ok, "AppState should be *GenesisStateRef, got %T", doc.AppState)
	assert.Equal(t, slimFixtureBalanceCount, ref.BalanceCount())
	assert.Equal(t, slimFixtureTxCount, ref.TxCount())

	// app_state @type marker (amino's interface tag) lands in the envelope.
	atype, ok := ref.SmallField("@type")
	require.True(t, ok)
	assert.JSONEq(t, `"/gno.GenesisState"`, string(atype))

	// Streaming iteration over balances returns one non-empty JSON value per slot.
	count := 0
	for line, err := range ref.IterBalances(context.Background()) {
		require.NoError(t, err)
		assert.NotEmpty(t, line)
		count++
	}
	assert.Equal(t, slimFixtureBalanceCount, count)
}

func TestLoadStreamingGenesisDoc_CacheHitSkipsRewrite(t *testing.T) {
	src := copySlimFixture(t)
	cacheRoot := t.TempDir()

	doc1, err := LoadStreamingGenesisDoc(src, cacheRoot, nil)
	require.NoError(t, err)
	ref1 := doc1.AppState.(*GenesisStateRef)

	manifestPath := filepath.Join(ref1.cacheDir, manifestFilename)
	info1, err := os.Stat(manifestPath)
	require.NoError(t, err)

	// Sleep so a second write would observe a different mtime.
	time.Sleep(20 * time.Millisecond)

	doc2, err := LoadStreamingGenesisDoc(src, cacheRoot, nil)
	require.NoError(t, err)
	assert.Equal(t, doc1.ChainID, doc2.ChainID, "warm path still populates small fields")

	info2, err := os.Stat(manifestPath)
	require.NoError(t, err)
	assert.Equal(t, info1.ModTime(), info2.ModTime(), "cache hit must not rewrite files")
}

func TestLoadStreamingGenesisDoc_GCRemovesStaleCaches(t *testing.T) {
	src := copySlimFixture(t)
	cacheRoot := t.TempDir()

	staleHash := filepath.Join(cacheRoot, "deadbeef")
	require.NoError(t, os.MkdirAll(staleHash, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(staleHash, "marker"), []byte("x"), 0o644))
	staleTmp := filepath.Join(cacheRoot, ".tmp-leftover")
	require.NoError(t, os.MkdirAll(staleTmp, 0o755))

	doc, err := LoadStreamingGenesisDoc(src, cacheRoot, nil)
	require.NoError(t, err)
	ref := doc.AppState.(*GenesisStateRef)

	entries, err := os.ReadDir(cacheRoot)
	require.NoError(t, err)
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	assert.Equal(t, []string{filepath.Base(ref.cacheDir)}, names, "GC must leave only the current cache directory")
}

func TestLoadStreamingGenesisDoc_RebuildsCorruptCache(t *testing.T) {
	src := copySlimFixture(t)
	cacheRoot := t.TempDir()

	doc1, err := LoadStreamingGenesisDoc(src, cacheRoot, nil)
	require.NoError(t, err)
	ref1 := doc1.AppState.(*GenesisStateRef)

	// Corrupt the cache by removing the version file; the reader will
	// reject it on next open, so a rebuild must happen.
	require.NoError(t, os.Remove(filepath.Join(ref1.cacheDir, versionFilename)))

	doc2, err := LoadStreamingGenesisDoc(src, cacheRoot, nil)
	require.NoError(t, err)
	ref2 := doc2.AppState.(*GenesisStateRef)
	assert.Equal(t, ref1.cacheDir, ref2.cacheDir, "same source produces same cache dir")
	assert.Equal(t, slimFixtureBalanceCount, ref2.BalanceCount())
}

func TestLoadStreamingGenesisDoc_DifferentSourceDifferentCache(t *testing.T) {
	srcA := copySlimFixture(t)
	// Tweak the second copy: append a single byte of whitespace before the
	// final newline so the SHA-256 differs while the JSON stays valid.
	srcB := copySlimFixture(t)
	appendByte(t, srcB, ' ')

	cacheRoot := t.TempDir()
	docA, err := LoadStreamingGenesisDoc(srcA, cacheRoot, nil)
	require.NoError(t, err)
	refA := docA.AppState.(*GenesisStateRef)

	docB, err := LoadStreamingGenesisDoc(srcB, cacheRoot, nil)
	require.NoError(t, err)
	refB := docB.AppState.(*GenesisStateRef)

	assert.NotEqual(t, refA.cacheDir, refB.cacheDir)
	_, err = os.Stat(refA.cacheDir)
	assert.True(t, os.IsNotExist(err), "stale cache dir must be GC'd")
}

func appendByte(t *testing.T, path string, b byte) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0o644)
	require.NoError(t, err)
	defer f.Close()
	_, err = f.Write([]byte{b})
	require.NoError(t, err)
}

// TestLoadStreamingGenesisDoc_IndentedSourceProducesValidJSONL guards against
// a regression where an indented source file (with newlines inside each tx
// object) caused streamArrayToJSONL to write multi-line records into the
// JSONL file, breaking IterTxs which splits on '\n'. The slim fixture is
// indented; this test asserts each yielded line round-trips through
// json.Unmarshal as a single value.
func TestLoadStreamingGenesisDoc_IndentedSourceProducesValidJSONL(t *testing.T) {
	src := copySlimFixture(t)
	doc, err := LoadStreamingGenesisDoc(src, t.TempDir(), nil)
	require.NoError(t, err)
	ref := doc.AppState.(*GenesisStateRef)

	count := 0
	for line, err := range ref.IterTxs(context.Background()) {
		require.NoError(t, err)
		var raw json.RawMessage
		require.NoError(t, json.Unmarshal(line, &raw),
			"each JSONL line must be a single JSON value (line %d): %q", count, string(line))
		count++
	}
	assert.Equal(t, slimFixtureTxCount, count)
}

// TestLoadStreamingGenesisDoc_TolerantOfUnknownTopLevelFields verifies that
// the streaming loader accepts top-level fields that GenesisDoc does not
// model (e.g. initial_height as emitted by hardfork tooling). The amino
// path used by GenesisDocFromJSON has always silently ignored such fields;
// the streaming path must match that tolerance or it'll reject real
// production genesis files at startup.
func TestLoadStreamingGenesisDoc_TolerantOfUnknownTopLevelFields(t *testing.T) {
	src := filepath.Join(t.TempDir(), "genesis.json")
	// initial_height is the field the real gnoland1 genesis carries; pick
	// it specifically so the regression matches the production failure.
	const body = `{
		"genesis_time": "2026-04-28T00:00:00Z",
		"chain_id": "test-chain",
		"initial_height": "815001",
		"app_hash": "",
		"app_state": {
			"balances": [],
			"txs": []
		}
	}`
	require.NoError(t, os.WriteFile(src, []byte(body), 0o644))

	doc, err := LoadStreamingGenesisDoc(src, t.TempDir(), nil)
	require.NoError(t, err, "streaming loader must tolerate unknown top-level fields")
	assert.Equal(t, "test-chain", doc.ChainID)
	require.NotNil(t, doc.AppState)
}

// TestLoadStreamingGenesisDoc_WarnsOnUnknownTopLevelFields verifies that
// while the streaming loader tolerates unknown fields (as the existing
// amino path does), it surfaces them as warnings so a misspelled known
// field — say "consensus_paramz" — cannot silently disappear into the
// cache without operator visibility.
func TestLoadStreamingGenesisDoc_WarnsOnUnknownTopLevelFields(t *testing.T) {
	src := filepath.Join(t.TempDir(), "genesis.json")
	const body = `{
		"genesis_time": "2026-04-28T00:00:00Z",
		"chain_id": "test-chain",
		"initial_height": "815001",
		"consensus_paramz": {"oops": "typo"},
		"app_hash": "",
		"app_state": {"balances": [], "txs": []}
	}`
	require.NoError(t, os.WriteFile(src, []byte(body), 0o644))

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))

	_, err := LoadStreamingGenesisDoc(src, t.TempDir(), logger)
	require.NoError(t, err)

	logged := buf.String()
	assert.Contains(t, logged, "initial_height",
		"unknown field initial_height must be logged as a warning so misspellings surface: %s", logged)
	assert.Contains(t, logged, "consensus_paramz",
		"unknown field consensus_paramz must be logged: %s", logged)
}

// TestCacheWriter_FinalizeClosesUnderlyingFiles verifies that finalize takes
// responsibility for closing the JSONL files BEFORE returning success.
// Pre-fix, finalize only flushed bufio buffers and the deferred writer.close()
// ran after os.Rename — meaning Close() errors (ENOSPC/EIO that only surface
// at close, common on NFS and overlay filesystems) were swallowed AND the
// renamed cache was already in its final location, looking valid to
// OpenGenesisStateRef.
//
// We assert the contract structurally: after finalize returns, writes to
// the underlying *os.File must fail with "use of closed file" — proving
// finalize actually closed it. The complementary assertion is that the
// cached JSONL content was flushed before close (otherwise we'd be silently
// truncating the cache).
func TestCacheWriter_FinalizeClosesUnderlyingFiles(t *testing.T) {
	tmpDir := t.TempDir()
	w, err := newCacheWriter(tmpDir, "test-hash")
	require.NoError(t, err)

	const balanceLine = `{"a":1}` + "\n"
	const txLine = `{"t":1}` + "\n"
	_, err = w.balancesBuf.Write([]byte(balanceLine))
	require.NoError(t, err)
	_, err = w.txsBuf.Write([]byte(txLine))
	require.NoError(t, err)
	w.manifest.BalanceCount = 1
	w.manifest.TxCount = 1

	require.NoError(t, w.finalize(tmpDir))

	// Files must be closed by finalize; subsequent writes must fail.
	_, balErr := w.balancesFile.Write([]byte("after"))
	require.Error(t, balErr, "balancesFile must be closed by finalize")
	_, txErr := w.txsFile.Write([]byte("after"))
	require.Error(t, txErr, "txsFile must be closed by finalize")

	// And the buffered data must have made it to disk before close.
	got, err := os.ReadFile(filepath.Join(tmpDir, balancesFilename))
	require.NoError(t, err)
	assert.Equal(t, balanceLine, string(got))
}

// TestGenesisStateRef_StreamJSON_RehydratesAppStateShape verifies that the
// streamed output mirrors the original app_state object: the small sibling
// fields (auth, bank, vm, …) appear at the top level, and the bulk arrays
// (balances, txs) are reconstituted as JSON arrays whose elements come line
// by line from the JSONL caches. This is the wire-format contract for
// /genesis when AppState is a *GenesisStateRef.
func TestGenesisStateRef_StreamJSON_RehydratesAppStateShape(t *testing.T) {
	balances := []string{
		`{"address":"g1aaa","amount":"10ugnot"}`,
		`{"address":"g1bbb","amount":"20ugnot"}`,
	}
	txs := []string{
		`{"msg":[{"type":"send","from":"g1aaa"}]}`,
		`{"msg":[{"type":"send","from":"g1bbb"}]}`,
	}
	smallFields := map[string]string{
		"auth": `{"params":{"max_memo":256}}`,
		"bank": `{"params":{}}`,
		"vm":   `{"params":{"chain_domain":"test"}}`,
	}
	dir := writeTestCache(t, balances, txs, smallFields)
	ref, err := OpenGenesisStateRef(dir)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, ref.StreamJSON(context.Background(), &buf))

	// Output must be valid JSON.
	var got map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got),
		"streamed output must parse as a JSON object: %s", buf.String())

	// Small fields are rehydrated verbatim.
	assert.JSONEq(t, smallFields["auth"], string(got["auth"]))
	assert.JSONEq(t, smallFields["bank"], string(got["bank"]))
	assert.JSONEq(t, smallFields["vm"], string(got["vm"]))

	// balances and txs arrays are reconstituted from the JSONL files.
	var gotBalances []json.RawMessage
	require.NoError(t, json.Unmarshal(got["balances"], &gotBalances))
	require.Len(t, gotBalances, len(balances))
	for i, line := range balances {
		assert.JSONEq(t, line, string(gotBalances[i]))
	}

	var gotTxs []json.RawMessage
	require.NoError(t, json.Unmarshal(got["txs"], &gotTxs))
	require.Len(t, gotTxs, len(txs))
	for i, line := range txs {
		assert.JSONEq(t, line, string(gotTxs[i]))
	}
}

// TestGenesisStateRef_StreamJSON_EmptyArrays verifies that the streaming path
// emits valid empty `[]` arrays when the cache has zero balances / zero txs,
// matching the shape of the source app_state object even when sparse.
func TestGenesisStateRef_StreamJSON_EmptyArrays(t *testing.T) {
	dir := writeTestCache(t, nil, nil, map[string]string{
		"auth": `{"params":{"max_memo":1}}`,
	})
	ref, err := OpenGenesisStateRef(dir)
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, ref.StreamJSON(context.Background(), &buf))

	var got map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	assert.JSONEq(t, `[]`, string(got["balances"]))
	assert.JSONEq(t, `[]`, string(got["txs"]))
}

// TestGenesisStateRef_StreamJSON_DeterministicSmallFieldOrder verifies that
// small fields are emitted in a stable (sorted) order across runs. Map
// iteration in Go is intentionally randomized, so without explicit sorting
// the wire format would jitter and break consumers that rely on the bytes
// (e.g., naive byte-equality cache keys, signing/hashing).
func TestGenesisStateRef_StreamJSON_DeterministicSmallFieldOrder(t *testing.T) {
	dir := writeTestCache(t, nil, nil, map[string]string{
		"vm":     `{"v":1}`,
		"auth":   `{"a":1}`,
		"bank":   `{"b":1}`,
		"params": `{"p":1}`,
	})
	ref, err := OpenGenesisStateRef(dir)
	require.NoError(t, err)

	var first bytes.Buffer
	require.NoError(t, ref.StreamJSON(context.Background(), &first))
	for i := 0; i < 10; i++ {
		var again bytes.Buffer
		require.NoError(t, ref.StreamJSON(context.Background(), &again))
		assert.Equal(t, first.Bytes(), again.Bytes(),
			"StreamJSON output must be byte-stable across runs (iteration %d)", i)
	}
}

// TestGenesisStateRef_StreamJSON_HonorsContextCancellation verifies that an
// already-cancelled context aborts streaming before the bulk arrays are
// fully emitted. This matters in production: a long /genesis stream tied to
// a disconnected client must release server resources promptly.
func TestGenesisStateRef_StreamJSON_HonorsContextCancellation(t *testing.T) {
	balances := []string{
		`{"a":1}`, `{"a":2}`, `{"a":3}`, `{"a":4}`, `{"a":5}`,
	}
	dir := writeTestCache(t, balances, nil, nil)
	ref, err := OpenGenesisStateRef(dir)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = ref.StreamJSON(ctx, &bytes.Buffer{})
	require.Error(t, err, "StreamJSON must return ctx.Err when ctx is already cancelled")
	assert.ErrorIs(t, err, context.Canceled)
}

// TestGenesisStateRef_StreamJSON_ConcurrentSafe pins down the concurrent-
// access contract that the original DoS scenario depends on: many
// simultaneous /genesis serves call StreamJSON against the same *ref.
// A future regression that introduces shared mutable state (a buffer
// pool, a counter, a write-cached field) would silently break this and
// the env-var-gated memory tests in CI never catch it. This test runs
// in the default `go test -race` matrix.
func TestGenesisStateRef_StreamJSON_ConcurrentSafe(t *testing.T) {
	balances := []string{
		`{"address":"g1aaa","amount":"10ugnot"}`,
		`{"address":"g1bbb","amount":"20ugnot"}`,
		`{"address":"g1ccc","amount":"30ugnot"}`,
	}
	txs := []string{
		`{"msg":[{"type":"send","from":"g1aaa"}]}`,
	}
	dir := writeTestCache(t, balances, txs, map[string]string{
		"auth": `{"params":{"max_memo":256}}`,
		"vm":   `{"params":{"chain_domain":"test"}}`,
	})
	ref, err := OpenGenesisStateRef(dir)
	require.NoError(t, err)

	const concurrency = 16

	// Single-threaded baseline so we can assert byte-equal output across
	// all concurrent callers — proves no per-call state leaked between
	// goroutines.
	var baseline bytes.Buffer
	require.NoError(t, ref.StreamJSON(context.Background(), &baseline))
	expected := baseline.Bytes()

	var wg sync.WaitGroup
	wg.Add(concurrency)
	results := make([][]byte, concurrency)
	errs := make([]error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			var buf bytes.Buffer
			if err := ref.StreamJSON(context.Background(), &buf); err != nil {
				errs[idx] = err
				return
			}
			results[idx] = buf.Bytes()
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		require.NoError(t, err, "concurrent StreamJSON #%d", i)
	}
	for i, got := range results {
		assert.Equal(t, expected, got,
			"concurrent StreamJSON output #%d differs from single-threaded baseline", i)
	}
}

// failAfterNWriter accepts up to budget bytes, then returns the configured
// error on the next write. Used to verify the streaming code surfaces
// mid-body write failures (the common case under broken-pipe, OS write
// errors, or a client disconnect partway through a 200 MB response)
// rather than panicking or silently truncating.
type failAfterNWriter struct {
	budget int
	err    error
}

func (w *failAfterNWriter) Write(p []byte) (int, error) {
	if w.budget <= 0 {
		return 0, w.err
	}
	if len(p) <= w.budget {
		w.budget -= len(p)
		return len(p), nil
	}
	n := w.budget
	w.budget = 0
	return n, w.err
}

// TestGenesisStateRef_StreamJSON_PropagatesMidStreamWriteError verifies
// that a write failure mid-body (after the envelope prefix has gone out)
// is returned as an error rather than panicking or silently truncating.
// The pre-fix WriteRPCResponseHTTP panicked on broken-pipe writes and
// crashed the goroutine; this test pins down that the streaming path
// stays panic-free across the same condition.
func TestGenesisStateRef_StreamJSON_PropagatesMidStreamWriteError(t *testing.T) {
	balances := []string{
		`{"address":"g1aaa","amount":"10ugnot"}`,
		`{"address":"g1bbb","amount":"20ugnot"}`,
		`{"address":"g1ccc","amount":"30ugnot"}`,
		`{"address":"g1ddd","amount":"40ugnot"}`,
	}
	dir := writeTestCache(t, balances, nil, map[string]string{
		"auth": `{"params":{"max_memo":256}}`,
	})
	ref, err := OpenGenesisStateRef(dir)
	require.NoError(t, err)

	// 32 bytes is enough to get past the opening "{" + small fields header
	// but well short of the full balances array — the failure lands while
	// the iterator is mid-body. ErrClosedPipe matches the real-world
	// signal that motivated the panic-to-error conversion.
	w := &failAfterNWriter{budget: 32, err: io.ErrClosedPipe}
	require.NotPanics(t, func() {
		err = ref.StreamJSON(context.Background(), w)
	})
	require.Error(t, err, "mid-body write failure must be returned, not swallowed")
	assert.ErrorIs(t, err, io.ErrClosedPipe,
		"the underlying write error must be wrapped through, not flattened")
}

// TestResultGenesis_StreamJSON_PropagatesMidStreamWriteError exercises the
// same write-failure contract one layer up — through the full
// *ResultGenesis.StreamJSON path including the amino-marshaled envelope
// and the streamed AppState splice. Confirms that errors from the inner
// AppState.StreamJSON propagate through the outer wrapper without panic.
func TestResultGenesis_StreamJSON_PropagatesMidStreamWriteError(t *testing.T) {
	balances := []string{
		`{"address":"g1aaa","amount":"10ugnot"}`,
		`{"address":"g1bbb","amount":"20ugnot"}`,
	}
	dir := writeTestCache(t, balances, nil, map[string]string{
		"auth": `{"params":{}}`,
	})
	ref, err := OpenGenesisStateRef(dir)
	require.NoError(t, err)

	doc := &bft.GenesisDoc{
		ChainID:  "fail-test",
		AppState: ref,
	}
	res := &core_types.ResultGenesis{Genesis: doc}

	// Tighter budget so the failure lands inside the AppState body
	// (amino-marshaled envelope alone is ~80–120 bytes for this doc).
	w := &failAfterNWriter{budget: 200, err: io.ErrClosedPipe}
	require.NotPanics(t, func() {
		err = res.StreamJSON(context.Background(), w)
	})
	require.Error(t, err, "mid-body write failure in nested StreamJSON must be returned")
	assert.ErrorIs(t, err, io.ErrClosedPipe)
}

// TestResultGenesis_StreamJSON_SlimFixtureEndToEnd is the cross-package
// integration test that pins down the contract /genesis is supposed to
// satisfy: load a real-shape genesis through the streaming loader, wrap it
// in a ResultGenesis, and serialize it through the same StreamJSON path the
// RPC handler will invoke. The output must be valid JSON containing every
// balance and tx from the source — without ever buffering app_state in
// memory.
func TestResultGenesis_StreamJSON_SlimFixtureEndToEnd(t *testing.T) {
	src := copySlimFixture(t)
	cacheRoot := t.TempDir()

	doc, err := LoadStreamingGenesisDoc(src, cacheRoot, nil)
	require.NoError(t, err)

	_, ok := doc.AppState.(*GenesisStateRef)
	require.True(t, ok, "LoadStreamingGenesisDoc must attach *GenesisStateRef as AppState")

	res := &core_types.ResultGenesis{Genesis: doc}

	var buf bytes.Buffer
	require.NoError(t, res.StreamJSON(context.Background(), &buf))

	// The wire body must parse as a single JSON object with one key "genesis".
	var envelope struct {
		Genesis json.RawMessage `json:"genesis"`
	}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &envelope),
		"streamed /genesis body must be valid JSON: %s", buf.String())

	// Inside, the doc must carry the chain id from the fixture and an
	// app_state that contains balances/txs as proper arrays. The
	// validators slice must preserve amino's polymorphic @type/value
	// markers — without them, a /genesis client cannot decode the
	// PubKey field, breaking the RPC contract.
	var inner struct {
		ChainID    string            `json:"chain_id"`
		Validators []json.RawMessage `json:"validators"`
		AppState   struct {
			Balances []json.RawMessage `json:"balances"`
			Txs      []json.RawMessage `json:"txs"`
			Auth     json.RawMessage   `json:"auth"`
			Bank     json.RawMessage   `json:"bank"`
			VM       json.RawMessage   `json:"vm"`
		} `json:"app_state"`
	}
	require.NoError(t, json.Unmarshal(envelope.Genesis, &inner))
	assert.NotEmpty(t, inner.ChainID, "chain_id must round-trip from the source fixture")
	assert.Equal(t, slimFixtureBalanceCount, len(inner.AppState.Balances),
		"every source balance must appear in the streamed app_state")
	assert.Equal(t, slimFixtureTxCount, len(inner.AppState.Txs),
		"every source tx must appear in the streamed app_state")
	assert.NotEmpty(t, inner.AppState.Auth, "auth small field must be rehydrated")
	assert.NotEmpty(t, inner.AppState.Bank, "bank small field must be rehydrated")
	assert.NotEmpty(t, inner.AppState.VM, "vm small field must be rehydrated")

	// Validator polymorphism check: every validator's pub_key must carry
	// the amino "@type" and "value" markers. Without these, a client of
	// /genesis can't decode the polymorphic crypto.PubKey field — which
	// is the most catastrophic silent regression possible for this code
	// path because it breaks chain bootstrap for every downstream tool.
	require.NotEmpty(t, inner.Validators, "real fixture has validators")
	for i, raw := range inner.Validators {
		var v struct {
			PubKey json.RawMessage `json:"pub_key"`
		}
		require.NoError(t, json.Unmarshal(raw, &v),
			"validator %d must be a valid JSON object", i)
		assert.Contains(t, string(v.PubKey), `"@type"`,
			"validator[%d].pub_key must carry amino @type marker after streaming round-trip", i)
		assert.Contains(t, string(v.PubKey), `"value"`,
			"validator[%d].pub_key must carry amino value marker after streaming round-trip", i)
	}
}
