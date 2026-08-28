package fork

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMigrationTxs_ReasonCopiedToNote(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "mig.jsonl")

	at, err := amino.MarshalJSON(AnnotatedTx{
		Tx:       sampleTx("payload"),
		Metadata: &gnoland.GnoTxMetadata{ChainID: "test-13"},
		Reason:   "valoper-seed: bootstrap operator g1abc",
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, append(at, '\n'), 0o644))

	got, err := loadMigrationTxs(path)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "valoper-seed: bootstrap operator g1abc", got[0].Metadata.Note)
	assert.Equal(t, "test-13", got[0].Metadata.ChainID)
}

func TestLoadMigrationTxs_PlainTxWithMetadataParses(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "legacy.jsonl")

	bz, err := amino.MarshalJSON(gnoland.TxWithMetadata{
		Tx:       sampleTx("legacy"),
		Metadata: &gnoland.GnoTxMetadata{ChainID: "test-13"},
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, append(bz, '\n'), 0o644))

	got, err := loadMigrationTxs(path)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "", got[0].Metadata.Note)
}

func TestLoadMigrationTxs_ForcesBlockHeightZero(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "bh.jsonl")

	at, err := amino.MarshalJSON(AnnotatedTx{
		Tx:       sampleTx("bh"),
		Metadata: &gnoland.GnoTxMetadata{BlockHeight: 12345},
		Reason:   "r",
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, append(at, '\n'), 0o644))

	got, err := loadMigrationTxs(path)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, int64(0), got[0].Metadata.BlockHeight, "block_height must be forced to 0 for migration txs")
}

func TestLoadMigrationTxs_CreatesMetadataWhenAbsent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "nometa.jsonl")

	at, err := amino.MarshalJSON(AnnotatedTx{
		Tx:     sampleTx("nm"),
		Reason: "x",
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, append(at, '\n'), 0o644))

	got, err := loadMigrationTxs(path)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.NotNil(t, got[0].Metadata)
	assert.Equal(t, "x", got[0].Metadata.Note)
}

func TestLoadMigrationTxs_DoesNotMutateInput(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "two.jsonl")

	at1 := AnnotatedTx{Tx: sampleTx("a"), Metadata: &gnoland.GnoTxMetadata{BlockHeight: 100}, Reason: "ra"}
	at2 := AnnotatedTx{Tx: sampleTx("b"), Metadata: &gnoland.GnoTxMetadata{BlockHeight: 200}, Reason: "rb"}
	bz1, err := amino.MarshalJSON(at1)
	require.NoError(t, err)
	bz2, err := amino.MarshalJSON(at2)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, []byte(string(bz1)+"\n"+string(bz2)+"\n"), 0o644))

	got, err := loadMigrationTxs(path)
	require.NoError(t, err)
	require.Len(t, got, 2)
	// Both should have BlockHeight=0 regardless of input ordering.
	assert.Equal(t, int64(0), got[0].Metadata.BlockHeight)
	assert.Equal(t, int64(0), got[1].Metadata.BlockHeight)
	assert.Equal(t, "ra", got[0].Metadata.Note)
	assert.Equal(t, "rb", got[1].Metadata.Note)
	// Verify Tx bodies preserved.
	assert.Equal(t, "a", got[0].Tx.Memo)
	assert.Equal(t, "b", got[1].Tx.Memo)
}
