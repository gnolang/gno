package fork

import "github.com/gnolang/gno/gno.land/pkg/gnoland"

// loadMigrationTxs reads a .jsonl of AnnotatedTx entries (also accepts plain
// TxWithMetadata, with Reason defaulting to "") and converts each to a
// TxWithMetadata ready for appState injection:
//   - Reason is copied into metadata.Note (free-form provenance description)
//   - metadata.BlockHeight is forced to 0 so the chain replays the entry in
//     genesis-mode (sig verify skipped under --skip-genesis-sig-verification;
//     chain-id falls back to PastChainIDs[0])
//
// Callers should follow up with annotateSource(_, SourceMigration) to tag
// provenance.
func loadMigrationTxs(path string) ([]gnoland.TxWithMetadata, error) {
	ats, err := readAnnotatedTxs(path)
	if err != nil {
		return nil, err
	}
	out := make([]gnoland.TxWithMetadata, len(ats))
	for i, at := range ats {
		meta := &gnoland.GnoTxMetadata{}
		if at.Metadata != nil {
			tmp := *at.Metadata
			meta = &tmp
		}
		meta.BlockHeight = 0
		meta.Note = at.Reason
		out[i] = gnoland.TxWithMetadata{Tx: at.Tx, Metadata: meta}
	}
	return out, nil
}
