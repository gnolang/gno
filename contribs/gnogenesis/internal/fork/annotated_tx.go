package fork

import (
	"fmt"
	"os"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// AnnotatedTx is the schema for --patch-txs and --migration-tx jsonl inputs.
// It mirrors gnoland.TxWithMetadata with one extra field — Reason — that the
// user writes to document why this tx exists. The reason flows into the
// produced genesis as GnoTxMetadata.Note when the tx is materialized.
//
// Files written in the plain gnoland.TxWithMetadata shape (no "reason" key)
// still parse here; Reason defaults to "" in that case.
type AnnotatedTx struct {
	Tx       std.Tx                 `json:"tx"`
	Metadata *gnoland.GnoTxMetadata `json:"metadata,omitempty"`
	Reason   string                 `json:"reason,omitempty"`
}

// readAnnotatedTxs parses a .jsonl file of AnnotatedTx entries. Blank lines and
// lines starting with '#' are ignored. Returns an error that names the
// offending line number on the first malformed entry.
func readAnnotatedTxs(path string) ([]AnnotatedTx, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	out := make([]AnnotatedTx, 0, len(lines))
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		var at AnnotatedTx
		if err := amino.UnmarshalJSON([]byte(line), &at); err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		out = append(out, at)
	}
	return out, nil
}
