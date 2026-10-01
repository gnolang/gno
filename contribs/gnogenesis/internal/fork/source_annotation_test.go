package fork

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnotateSource_SetsOnAllTxs(t *testing.T) {
	t.Parallel()

	txs := []gnoland.TxWithMetadata{
		{Tx: sampleTx("a"), Metadata: &gnoland.GnoTxMetadata{BlockHeight: 100}},
		{Tx: sampleTx("b"), Metadata: &gnoland.GnoTxMetadata{BlockHeight: 200, ChainID: "gnoland-1"}},
	}

	annotateSource(txs, gnoland.SourceHistorical)

	assert.Equal(t, gnoland.SourceHistorical, txs[0].Metadata.Source)
	assert.Equal(t, gnoland.SourceHistorical, txs[1].Metadata.Source)
	// existing fields preserved
	assert.Equal(t, int64(100), txs[0].Metadata.BlockHeight)
	assert.Equal(t, "gnoland-1", txs[1].Metadata.ChainID)
}

func TestAnnotateSource_CreatesMetadataIfNil(t *testing.T) {
	t.Parallel()

	txs := []gnoland.TxWithMetadata{{Tx: sampleTx("no-meta")}}
	require.Nil(t, txs[0].Metadata)

	annotateSource(txs, gnoland.SourceBase)

	require.NotNil(t, txs[0].Metadata)
	assert.Equal(t, gnoland.SourceBase, txs[0].Metadata.Source)
}

func TestAnnotateSource_OverwritesExisting(t *testing.T) {
	t.Parallel()

	txs := []gnoland.TxWithMetadata{
		{Tx: sampleTx("x"), Metadata: &gnoland.GnoTxMetadata{Source: gnoland.SourceHistorical}},
	}

	annotateSource(txs, gnoland.SourcePatched)

	assert.Equal(t, gnoland.SourcePatched, txs[0].Metadata.Source)
}

func TestAnnotateSource_EmptySliceIsNoop(t *testing.T) {
	t.Parallel()
	annotateSource(nil, gnoland.SourceBase)
	annotateSource([]gnoland.TxWithMetadata{}, gnoland.SourceBase)
}
