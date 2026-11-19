package fork

import "github.com/gnolang/gno/gno.land/pkg/gnoland"

// annotateSource tags every tx's metadata.Source with the given value, creating
// metadata if nil. Called at well-defined points during fork assembly so each
// tx in the resulting genesis carries provenance for the inspect subcommand.
func annotateSource(txs []gnoland.TxWithMetadata, source string) {
	for i := range txs {
		if txs[i].Metadata == nil {
			txs[i].Metadata = &gnoland.GnoTxMetadata{}
		}
		txs[i].Metadata.Source = source
	}
}
