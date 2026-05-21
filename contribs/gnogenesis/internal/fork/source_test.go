package fork

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- txs source flag validation

// TestOpenTxsSource_NoFlag asserts that openTxsSource refuses to run when
// no --source-txs-* flag is set, listing all three so the user knows which
// one to pick.
func TestOpenTxsSource_NoFlag(t *testing.T) {
	t.Parallel()

	src, err := (&generateCfg{rpcWorkersPerEndpoint: defaultWorkersPerEndpoint}).openTxsSource()
	require.Nil(t, src)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one txs source flag is required")
	assert.Contains(t, err.Error(), "--source-txs-rpc")
	assert.Contains(t, err.Error(), "--source-txs-jsonl-file")
	assert.Contains(t, err.Error(), "--source-txs-data-dir")
}

// TestOpenTxsSource_MutuallyExclusive asserts that setting more than one
// --source-txs-* flag at once is rejected with the offenders named.
func TestOpenTxsSource_MutuallyExclusive(t *testing.T) {
	t.Parallel()

	cfg := &generateCfg{
		rpcWorkersPerEndpoint: defaultWorkersPerEndpoint,
		sourceTxsRPC:          "http://example.invalid:26657",
		sourceTxsJSONLFile:    "/tmp/whatever.jsonl",
	}
	src, err := cfg.openTxsSource()
	require.Nil(t, src)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
	assert.Contains(t, err.Error(), "--source-txs-rpc")
	assert.Contains(t, err.Error(), "--source-txs-jsonl-file")
}

// ---- genesis source flag validation

func TestOpenGenesisSource_NoFlag(t *testing.T) {
	t.Parallel()

	src, err := (&generateCfg{}).openGenesisSource()
	require.Nil(t, src)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one genesis source flag is required")
	assert.Contains(t, err.Error(), "--source-genesis-rpc")
	assert.Contains(t, err.Error(), "--source-genesis-file")
}

func TestOpenGenesisSource_MutuallyExclusive(t *testing.T) {
	t.Parallel()

	cfg := &generateCfg{
		sourceGenesisRPC:  "http://example.invalid:26657",
		sourceGenesisFile: "/tmp/whatever.json",
	}
	src, err := cfg.openGenesisSource()
	require.Nil(t, src)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
	assert.Contains(t, err.Error(), "--source-genesis-rpc")
	assert.Contains(t, err.Error(), "--source-genesis-file")
}

// ---- individual source constructors (jsonl + data-dir + file-genesis)

func TestNewJSONLFileTxsSource_Missing(t *testing.T) {
	t.Parallel()

	src, err := newJSONLFileTxsSource(filepath.Join(t.TempDir(), "nope.jsonl"))
	require.Nil(t, src)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "txs jsonl file")
}

func TestNewJSONLFileTxsSource_RejectsDirectory(t *testing.T) {
	t.Parallel()

	src, err := newJSONLFileTxsSource(t.TempDir())
	require.Nil(t, src)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expects a file, got directory")
}

func TestNewDataDirTxsSource_MissingDBs(t *testing.T) {
	t.Parallel()

	src, err := newDataDirTxsSource(t.TempDir())
	require.Nil(t, src)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "blockstore.db")
}

func TestNewFileGenesisSource_Missing(t *testing.T) {
	t.Parallel()

	src, err := newFileGenesisSource(filepath.Join(t.TempDir(), "nope.json"))
	require.Nil(t, src)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "genesis file")
}

// TestFileGenesisSource_ReadsFromDisk verifies the file source actually
// parses a real on-disk genesis (basic round-trip).
func TestFileGenesisSource_ReadsFromDisk(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "genesis.json")
	require.NoError(t, os.WriteFile(
		path,
		[]byte(`{"chain_id":"file-chain","app_state":null}`),
		0o644,
	))

	src, err := newFileGenesisSource(path)
	require.NoError(t, err)
	defer src.Close()

	doc, err := src.FetchGenesis(t.Context())
	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, "file-chain", doc.ChainID)
	assert.Equal(t, "genesis file", src.Description())
}

// ---- multi-endpoint RPC URL parsing (txs + genesis sources)

func TestNewRPCTxsSource_SingleURL(t *testing.T) {
	t.Parallel()

	src, err := newRPCTxsSource("http://localhost:26657", defaultWorkersPerEndpoint)
	require.NoError(t, err)
	defer src.Close()

	assert.Equal(t, []string{"http://localhost:26657"}, src.rpcURLs)
	assert.Len(t, src.clients, 1)
	assert.Equal(t, "RPC", src.Description())
}

func TestNewRPCTxsSource_MultipleURLs(t *testing.T) {
	t.Parallel()

	src, err := newRPCTxsSource("http://a:26657,http://b:26657,http://c:26657", defaultWorkersPerEndpoint)
	require.NoError(t, err)
	defer src.Close()

	assert.Equal(t, []string{"http://a:26657", "http://b:26657", "http://c:26657"}, src.rpcURLs)
	assert.Len(t, src.clients, 3)
	assert.Equal(t, "RPC (3 endpoints)", src.Description())
}

func TestNewRPCTxsSource_TrimsWhitespaceAndSkipsEmpty(t *testing.T) {
	t.Parallel()

	src, err := newRPCTxsSource("  http://a:26657 , , http://b:26657  ", defaultWorkersPerEndpoint)
	require.NoError(t, err)
	defer src.Close()

	assert.Equal(t, []string{"http://a:26657", "http://b:26657"}, src.rpcURLs)
}

func TestNewRPCTxsSource_RejectsNonHTTPSchemeInList(t *testing.T) {
	t.Parallel()

	for _, s := range []string{
		"ws://localhost:26657,http://b:26657",
		"http://a:26657,tcp://b:26657",
		"http://a:26657,ftp://example.org",
		"http://a:26657,/some/path",
	} {
		_, err := newRPCTxsSource(s, defaultWorkersPerEndpoint)
		require.Error(t, err, "input: %s", s)
		assert.Contains(t, err.Error(), "http")
	}
}

func TestNewRPCTxsSource_RejectsAllEmpty(t *testing.T) {
	t.Parallel()

	_, err := newRPCTxsSource(",,", defaultWorkersPerEndpoint)
	require.Error(t, err)
}

func TestNewRPCGenesisSource_MultipleURLs(t *testing.T) {
	t.Parallel()

	src, err := newRPCGenesisSource("http://a:26657,http://b:26657")
	require.NoError(t, err)
	defer src.Close()

	assert.Equal(t, []string{"http://a:26657", "http://b:26657"}, src.rpcURLs)
	assert.Len(t, src.clients, 2)
	assert.Equal(t, "RPC (2 endpoints)", src.Description())
}
