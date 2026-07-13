package fork

import (
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/stretchr/testify/assert"
)

func TestInspectReport_GroupsByCategoryAndCounts(t *testing.T) {
	t.Parallel()

	manfred := crypto.Address{0x42}
	state := &gnoland.GnoGenesisState{
		Txs: []gnoland.TxWithMetadata{
			{Tx: sampleTx("base-1"), Metadata: &gnoland.GnoTxMetadata{Source: gnoland.SourceBase}},
			{Tx: sampleTx("base-2"), Metadata: &gnoland.GnoTxMetadata{Source: gnoland.SourceBase}},
			{Tx: sampleTx("hist-1"), Metadata: &gnoland.GnoTxMetadata{Source: gnoland.SourceHistorical, BlockHeight: 100}},
			{Tx: sampleTx("hist-2"), Metadata: &gnoland.GnoTxMetadata{Source: gnoland.SourceHistorical, BlockHeight: 200}},
			{Tx: sampleTx("hist-3"), Metadata: &gnoland.GnoTxMetadata{Source: gnoland.SourceHistorical, BlockHeight: 300}},
			{Tx: sampleTx("patched-body"), Metadata: &gnoland.GnoTxMetadata{
				Source:      gnoland.SourcePatched,
				BlockHeight: 1950,
				Note:        "API drift on unrestrict.gno",
				SignerInfo:  []gnoland.SignerAccountInfo{{Address: manfred, Sequence: 42}},
			}},
			{Tx: sampleTx("mig-1"), Metadata: &gnoland.GnoTxMetadata{Source: gnoland.SourceMigration, Note: "addpkg: gno.land/r/sys/validators/v3"}},
			{Tx: sampleTx("mig-2"), Metadata: &gnoland.GnoTxMetadata{Source: gnoland.SourceMigration, Note: "valoper-seed: register g1abc"}},
		},
	}

	report := inspectReport(state)

	// counts
	assert.Contains(t, report, "Base genesis: 2")
	assert.Contains(t, report, "Historical:   3")
	assert.Contains(t, report, "Patched:      1")
	assert.Contains(t, report, "Migration:    2")
	assert.Contains(t, report, "Total:        8")

	// patched detail: shows height, sender, sequence, reason
	assert.Contains(t, report, "h=1950")
	assert.Contains(t, report, "seq=42")
	assert.Contains(t, report, "API drift on unrestrict.gno")

	// migration detail: shows the Note text
	assert.Contains(t, report, "addpkg: gno.land/r/sys/validators/v3")
	assert.Contains(t, report, "valoper-seed: register g1abc")
}

func TestInspectReport_UnannotatedCount(t *testing.T) {
	t.Parallel()

	state := &gnoland.GnoGenesisState{
		Txs: []gnoland.TxWithMetadata{
			{Tx: sampleTx("legacy-no-meta")},
			{Tx: sampleTx("legacy-empty-source"), Metadata: &gnoland.GnoTxMetadata{}},
		},
	}

	report := inspectReport(state)
	assert.Contains(t, report, "Unannotated:  2")
	assert.Contains(t, report, "Total:        2")
}

func TestInspectReport_EmptyState(t *testing.T) {
	t.Parallel()

	report := inspectReport(&gnoland.GnoGenesisState{})
	assert.Contains(t, report, "Total:        0")
}
