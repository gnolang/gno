package restore

import (
	"context"
	"io"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/tx-archive/log/noop"
	"github.com/stretchr/testify/assert"
)

func TestRestore_ExecuteRestore(t *testing.T) {
	t.Parallel()

	var (
		exampleTxCount = 10
		exampleTxGiven = 0

		exampleTx = &std.Tx{
			Memo: "example tx",
		}

		sentTxs = make([]*std.Tx, 0)

		mockClient = &mockClient{
			sendTransactionFn: func(tx *std.Tx) error {
				sentTxs = append(sentTxs, tx)

				return nil
			},
		}
		mockSource = &mockSource{
			nextFn: func(ctx context.Context) (*std.Tx, error) {
				if exampleTxGiven == exampleTxCount {
					return nil, io.EOF
				}

				exampleTxGiven++

				return exampleTx, nil
			},
		}
	)

	s := NewService(mockClient, mockSource, WithLogger(noop.New()))

	// Execute the restore
	assert.NoError(
		t,
		s.ExecuteRestore(context.Background()),
	)

	// Verify the restore was correct
	assert.Len(t, sentTxs, exampleTxCount)

	for _, tx := range sentTxs {
		assert.Equal(t, exampleTx, tx)
	}
}
