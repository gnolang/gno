package history

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testChainID = "history-test"

// tx builds a distinguishable transaction: the memo carries n, so a test can
// assert on both content and ordering after a round trip.
func tx(n int) gnoland.TxWithMetadata {
	addr := crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	return gnoland.TxWithMetadata{
		Tx: std.Tx{
			Msgs: []std.Msg{bank.MsgSend{
				FromAddress: addr,
				ToAddress:   addr,
				Amount:      std.Coins{std.NewCoin(ugnot.Denom, int64(n))},
			}},
			Fee:  std.NewFee(1000, std.NewCoin(ugnot.Denom, 1000)),
			Memo: fmt.Sprintf("tx-%d", n),
		},
		Metadata: &gnoland.GnoTxMetadata{Timestamp: int64(1700000000 + n)},
	}
}

func memos(txs []gnoland.TxWithMetadata) []string {
	out := make([]string, len(txs))
	for i, t := range txs {
		out[i] = t.Tx.Memo
	}
	return out
}

func open(t *testing.T, dir string) *Store {
	t.Helper()
	s, err := Open(dir, Config{ChainID: testChainID, Logger: log.NewTestingLogger(t)})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// onlySegment returns the single segment in dir, failing if there is not
// exactly one.
func onlySegment(t *testing.T, dir string) string {
	t.Helper()
	segments, err := filepath.Glob(filepath.Join(dir, historyDirname, "*"+segmentExt))
	require.NoError(t, err)
	require.Len(t, segments, 1)
	return segments[0]
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := open(t, dir)

	loaded, err := s.Load(context.Background())
	require.NoError(t, err)
	assert.Empty(t, loaded, "a fresh state dir has no history")

	for i := 1; i <= 3; i++ {
		require.NoError(t, s.Record(tx(i)))
	}
	require.NoError(t, s.Close())

	// A second process opens the same directory and reads it back.
	s2 := open(t, dir)
	loaded, err = s2.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"tx-1", "tx-2", "tx-3"}, memos(loaded))
	assert.Equal(t, int64(1700000001), loaded[0].Metadata.Timestamp, "metadata survives the round trip")
}

// TestEpochsAppendAcrossSegments is the property the whole design rests on:
// reopening never rewrites or duplicates an earlier run's segment.
func TestEpochsAppendAcrossSegments(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	for epoch := range 3 {
		s := open(t, dir)
		loaded, err := s.Load(context.Background())
		require.NoError(t, err)
		assert.Len(t, loaded, epoch, "epoch %d replays every earlier transaction exactly once", epoch)

		require.NoError(t, s.Record(tx(epoch+1)))
		require.NoError(t, s.Close())
	}

	segments, err := filepath.Glob(filepath.Join(dir, historyDirname, "*"+segmentExt))
	require.NoError(t, err)
	assert.Len(t, segments, 3, "one segment per epoch")

	s := open(t, dir)
	loaded, err := s.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"tx-1", "tx-2", "tx-3"}, memos(loaded))
}

// TestLoadOrdersSegmentsNumerically guards against lexicographic ordering,
// which silently replays a history out of order once it passes ten segments
// of differing digit counts.
func TestLoadOrdersSegmentsNumerically(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	histDir := filepath.Join(dir, historyDirname)
	require.NoError(t, os.MkdirAll(histDir, 0o755))

	// 9, 10 and 1000000 sort the wrong way as strings once zero padding runs out.
	writeLine := func(name string, n int) {
		raw, err := encodeLine(tx(n))
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(histDir, name), raw, 0o644))
	}
	writeLine("000009"+segmentExt, 9)
	writeLine("000010"+segmentExt, 10)
	writeLine("1000000"+segmentExt, 1000000)

	s := open(t, dir)
	loaded, err := s.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"tx-9", "tx-10", "tx-1000000"}, memos(loaded))
}

func TestOpenRefusesAnotherChain(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := open(t, dir)
	require.NoError(t, s.Record(tx(1)))
	require.NoError(t, s.Close())

	_, err := Open(dir, Config{ChainID: "some-other-chain", Logger: log.NewTestingLogger(t)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "history-test")
	assert.Contains(t, err.Error(), "some-other-chain")
}

func TestOpenRefusesNewerSchema(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, writeManifest(dir, &Manifest{
		SchemaVersion: SchemaVersion + 1,
		ChainID:       testChainID,
	}))

	_, err := Open(dir, Config{ChainID: testChainID, Logger: log.NewTestingLogger(t)})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "schema version")
}

// TestLoadToleratesTruncatedTail is what a hard kill mid-flush leaves behind.
// Dropping the one partial transaction is a strictly smaller loss than
// refusing to boot on the whole history, which for a container that gets
// SIGKILLed would mean the state dir is unusable.
func TestLoadToleratesTruncatedTail(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := open(t, dir)
	require.NoError(t, s.Record(tx(1)))
	require.NoError(t, s.Record(tx(2)))
	require.NoError(t, s.Close())

	segment := onlySegment(t, dir)
	raw, err := os.ReadFile(segment)
	require.NoError(t, err)

	lines := strings.SplitAfter(string(raw), "\n")
	require.GreaterOrEqual(t, len(lines), 2)
	// First line whole, second line cut in half, no terminating newline.
	truncated := lines[0] + lines[1][:len(lines[1])/2]
	require.NotContains(t, truncated[len(lines[0]):], "\n")
	require.NoError(t, os.WriteFile(segment, []byte(truncated), 0o644))

	s2 := open(t, dir)
	loaded, err := s2.Load(context.Background())
	require.NoError(t, err, "a partial final line must not make the state dir unopenable")
	assert.Equal(t, []string{"tx-1"}, memos(loaded))
}

// TestLoadRejectsCorruptionMidFile is the other half: a bad line that is not
// the unterminated tail is real corruption, and replaying a history with a
// hole punched in it would produce a chain nobody can reason about.
func TestLoadRejectsCorruptionMidFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := open(t, dir)
	for i := 1; i <= 3; i++ {
		require.NoError(t, s.Record(tx(i)))
	}
	require.NoError(t, s.Close())

	segment := onlySegment(t, dir)
	raw, err := os.ReadFile(segment)
	require.NoError(t, err)

	lines := strings.SplitAfter(string(raw), "\n")
	require.GreaterOrEqual(t, len(lines), 3)
	lines[1] = "{\"tx\": this is not json}\n"
	require.NoError(t, os.WriteFile(segment, []byte(strings.Join(lines, "")), 0o644))

	s2 := open(t, dir)
	_, err = s2.Load(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), ":2:", "the error names the offending line")
}

// TestLoadRejectsTruncatedTailInSealedSegment: an earlier epoch's segment that
// lost its tail is corruption too, because every later segment replays after
// it. Only a missing newline saves it, and a sealed segment always has one.
func TestLoadRejectsCorruptSealedSegment(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := open(t, dir)
	require.NoError(t, s.Record(tx(1)))
	require.NoError(t, s.Close())

	segment := onlySegment(t, dir)
	raw, err := os.ReadFile(segment)
	require.NoError(t, err)
	// Keep the newline, break the content: not a partial write.
	require.NoError(t, os.WriteFile(segment, []byte("{oops}\n"+string(raw)), 0o644))

	s2 := open(t, dir)
	_, err = s2.Load(context.Background())
	require.Error(t, err)
}

func TestLoadIgnoresForeignFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := open(t, dir)
	require.NoError(t, s.Record(tx(1)))
	require.NoError(t, s.Close())

	histDir := filepath.Join(dir, historyDirname)
	require.NoError(t, os.WriteFile(filepath.Join(histDir, "README.md"), []byte("notes\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(histDir, "backup.jsonl.bak"), []byte("junk\n"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(histDir, "quarantine"), 0o755))

	s2 := open(t, dir)
	loaded, err := s2.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"tx-1"}, memos(loaded))
}

func TestRecordRotatesOnSize(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s, err := Open(dir, Config{
		ChainID:         testChainID,
		MaxSegmentBytes: 1, // every transaction overflows, so every one rotates
		Logger:          log.NewTestingLogger(t),
	})
	require.NoError(t, err)

	for i := 1; i <= 3; i++ {
		require.NoError(t, s.Record(tx(i)))
	}
	require.NoError(t, s.Close())

	segments, err := filepath.Glob(filepath.Join(dir, historyDirname, "*"+segmentExt))
	require.NoError(t, err)
	assert.Len(t, segments, 3, "a segment per transaction at a 1-byte cap")

	s2 := open(t, dir)
	loaded, err := s2.Load(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"tx-1", "tx-2", "tx-3"}, memos(loaded), "rotation preserves order")
}

func TestManifestCounts(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := open(t, dir)
	require.NoError(t, s.Record(tx(1)))
	require.NoError(t, s.Record(tx(2)))
	require.NoError(t, s.Sync())

	assert.Equal(t, 2, s.Recorded())
	assert.Equal(t, 2, s.Total())

	m, err := readManifest(dir)
	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Equal(t, SchemaVersion, m.SchemaVersion)
	assert.Equal(t, testChainID, m.ChainID)
	assert.Equal(t, 2, m.Txs)
	assert.Equal(t, 1, m.Epochs)
	require.NoError(t, s.Close())

	// A second epoch counts the replayed transactions once, not twice.
	s2 := open(t, dir)
	_, err = s2.Load(context.Background())
	require.NoError(t, err)
	require.NoError(t, s2.Record(tx(3)))
	require.NoError(t, s2.Sync())
	assert.Equal(t, 1, s2.Recorded())
	assert.Equal(t, 3, s2.Total())

	m, err = readManifest(dir)
	require.NoError(t, err)
	assert.Equal(t, 3, m.Txs)
	assert.Equal(t, 2, m.Epochs)
}

func TestManifestIsWorldReadable(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	s := open(t, dir)
	require.NoError(t, s.Sync())

	info, err := os.Stat(manifestPath(dir))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm(),
		"the state dir is meant to be inspected and archived, not private to one uid")
}

func TestCloseIsIdempotent(t *testing.T) {
	t.Parallel()

	s := open(t, t.TempDir())
	require.NoError(t, s.Record(tx(1)))
	require.NoError(t, s.Close())
	require.NoError(t, s.Close())

	err := s.Record(tx(2))
	require.Error(t, err, "recording after close must not silently drop a transaction")
}

func TestOpenRejectsEmptyDir(t *testing.T) {
	t.Parallel()

	_, err := Open("", Config{ChainID: testChainID})
	require.Error(t, err)
}
