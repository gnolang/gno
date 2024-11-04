package legacy

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriter_Legacy(t *testing.T) {
	t.Parallel()

	var (
		b bytes.Buffer

		txData = &gnoland.TxWithMetadata{
			Tx: std.Tx{
				Memo: "example tx",
			},
		}
	)

	// Create a new standard writer
	w := NewWriter(&b)

	// Write example tx data
	require.NoError(t, w.WriteTxData(txData))

	var readTx std.Tx

	readErr := amino.UnmarshalJSON(b.Bytes(), &readTx)
	require.NoError(t, readErr)

	assert.Equal(t, txData.Tx, readTx)
}
