package fork

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- helpers ----

func histTx(height int64, signer crypto.Address, seq uint64, memo string) gnoland.TxWithMetadata {
	return gnoland.TxWithMetadata{
		Tx: sampleTx(memo),
		Metadata: &gnoland.GnoTxMetadata{
			BlockHeight: height,
			ChainID:     "gnoland-1",
			SignerInfo:  []gnoland.SignerAccountInfo{{Address: signer, AccountNum: 1, Sequence: seq}},
		},
	}
}

func patchEntry(height int64, signer crypto.Address, seq uint64, memo, reason string) AnnotatedTx {
	return AnnotatedTx{
		Tx: sampleTx(memo),
		Metadata: &gnoland.GnoTxMetadata{
			BlockHeight: height,
			ChainID:     "gnoland-1",
			SignerInfo:  []gnoland.SignerAccountInfo{{Address: signer, AccountNum: 1, Sequence: seq}},
		},
		Reason: reason,
	}
}

func writePatchFile(t *testing.T, path string, entries ...AnnotatedTx) {
	t.Helper()
	var buf strings.Builder
	for _, e := range entries {
		bz, err := amino.MarshalJSON(e)
		require.NoError(t, err)
		buf.Write(bz)
		buf.WriteByte('\n')
	}
	require.NoError(t, os.WriteFile(path, []byte(buf.String()), 0o644))
}

// ---- tests ----

func TestApplyPatchTxs_AppliesMatchingPatch(t *testing.T) {
	t.Parallel()
	manfred := crypto.Address{0xAA}
	dir := t.TempDir()
	patchPath := filepath.Join(dir, "patches.jsonl")
	writePatchFile(t, patchPath, patchEntry(1950, manfred, 42, "patched-body", "API drift fix"))

	txs := []gnoland.TxWithMetadata{
		histTx(1949, manfred, 41, "earlier"),
		histTx(1950, manfred, 42, "original-body"),
		histTx(1966, manfred, 43, "later"),
	}

	n, err := applyPatchTxs(txs, []string{patchPath}, commands.NewTestIO())
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	// untouched
	assert.Equal(t, "earlier", txs[0].Tx.Memo)
	assert.Equal(t, "later", txs[2].Tx.Memo)
	assert.Empty(t, txs[0].Metadata.Source)
	assert.Empty(t, txs[2].Metadata.Source)

	// patched
	assert.Equal(t, "patched-body", txs[1].Tx.Memo)
	assert.Equal(t, gnoland.SourcePatched, txs[1].Metadata.Source)
	assert.Equal(t, "API drift fix", txs[1].Metadata.Note)
	require.NotNil(t, txs[1].Metadata.OriginalTx)
	assert.Equal(t, "original-body", txs[1].Metadata.OriginalTx.Memo)
	// existing metadata preserved
	assert.Equal(t, int64(1950), txs[1].Metadata.BlockHeight)
	require.Len(t, txs[1].Metadata.SignerInfo, 1)
}

func TestApplyPatchTxs_AppliesAcrossMultipleFiles(t *testing.T) {
	t.Parallel()
	manfred := crypto.Address{0xAA}
	dir := t.TempDir()
	pathA := filepath.Join(dir, "a.jsonl")
	pathB := filepath.Join(dir, "b.jsonl")
	writePatchFile(t, pathA, patchEntry(1950, manfred, 42, "fix-a", "reason A"))
	writePatchFile(t, pathB, patchEntry(1966, manfred, 43, "fix-b", "reason B"))

	txs := []gnoland.TxWithMetadata{
		histTx(1950, manfred, 42, "old-a"),
		histTx(1966, manfred, 43, "old-b"),
	}

	n, err := applyPatchTxs(txs, []string{pathA, pathB}, commands.NewTestIO())
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, "fix-a", txs[0].Tx.Memo)
	assert.Equal(t, "reason A", txs[0].Metadata.Note)
	assert.Equal(t, "fix-b", txs[1].Tx.Memo)
	assert.Equal(t, "reason B", txs[1].Metadata.Note)
}

func TestApplyPatchTxs_RejectsCrossFileDuplicate(t *testing.T) {
	t.Parallel()
	manfred := crypto.Address{0xAA}
	dir := t.TempDir()
	pathA := filepath.Join(dir, "a.jsonl")
	pathB := filepath.Join(dir, "b.jsonl")
	writePatchFile(t, pathA, patchEntry(1950, manfred, 42, "a", "ra"))
	writePatchFile(t, pathB, patchEntry(1950, manfred, 42, "b", "rb"))

	txs := []gnoland.TxWithMetadata{histTx(1950, manfred, 42, "src")}

	_, err := applyPatchTxs(txs, []string{pathA, pathB}, commands.NewTestIO())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate patch key")
}

func TestApplyPatchTxs_RejectsInFileDuplicate(t *testing.T) {
	t.Parallel()
	manfred := crypto.Address{0xAA}
	dir := t.TempDir()
	patchPath := filepath.Join(dir, "dupes.jsonl")
	writePatchFile(t, patchPath,
		patchEntry(1950, manfred, 42, "a", "ra"),
		patchEntry(1950, manfred, 42, "b", "rb"),
	)

	txs := []gnoland.TxWithMetadata{histTx(1950, manfred, 42, "src")}

	_, err := applyPatchTxs(txs, []string{patchPath}, commands.NewTestIO())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate patch key")
}

func TestApplyPatchTxs_RejectsUnmatched(t *testing.T) {
	t.Parallel()
	manfred := crypto.Address{0xAA}
	dir := t.TempDir()
	patchPath := filepath.Join(dir, "stray.jsonl")
	writePatchFile(t, patchPath, patchEntry(9999, manfred, 42, "stray", "doesn't exist"))

	txs := []gnoland.TxWithMetadata{histTx(1950, manfred, 42, "src")}

	_, err := applyPatchTxs(txs, []string{patchPath}, commands.NewTestIO())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "did not match")
}

func TestApplyPatchTxs_RejectsMissingKey(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	patchPath := filepath.Join(dir, "nokey.jsonl")
	// patch entry without signer_info / block_height
	writePatchFile(t, patchPath, AnnotatedTx{
		Tx:       sampleTx("nokey"),
		Metadata: &gnoland.GnoTxMetadata{ChainID: "gnoland-1"},
		Reason:   "missing key fields",
	})

	manfred := crypto.Address{0xAA}
	txs := []gnoland.TxWithMetadata{histTx(1950, manfred, 42, "src")}

	_, err := applyPatchTxs(txs, []string{patchPath}, commands.NewTestIO())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "block_height")
}

func TestApplyPatchTxs_NoopWithEmptyPatchList(t *testing.T) {
	t.Parallel()
	manfred := crypto.Address{0xAA}
	txs := []gnoland.TxWithMetadata{histTx(1950, manfred, 42, "untouched")}

	n, err := applyPatchTxs(txs, nil, commands.NewTestIO())
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, "untouched", txs[0].Tx.Memo)
	assert.Empty(t, txs[0].Metadata.Source)
}
