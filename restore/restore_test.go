package restore

import (
	"context"
	"io"
	"sync/atomic"
	"testing"
	"time"

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
			nextFn: func(_ context.Context) (*std.Tx, error) {
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
		s.ExecuteRestore(context.Background(), false),
	)

	// Verify the restore was correct
	assert.Len(t, sentTxs, exampleTxCount)

	for _, tx := range sentTxs {
		assert.Equal(t, exampleTx, tx)
	}
}

func TestRestore_ExecuteRestore_Watch(t *testing.T) {
	t.Parallel()

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	var (
		exampleTxCount = 20
		exampleTxGiven = 0

		simulateEOF atomic.Bool

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
			nextFn: func(_ context.Context) (*std.Tx, error) {
				if simulateEOF.Load() {
					return nil, io.EOF
				}

				// ~ the half mark, cut off the tx stream
				// by simulating the end of the stream (temporarily)
				if exampleTxGiven == exampleTxCount/2 {
					// Simulate EOF, but after some time
					// make sure the Next call returns an actual transaction
					simulateEOF.Store(true)

					time.AfterFunc(
						50*time.Millisecond,
						func() {
							simulateEOF.Store(false)
						},
					)

					exampleTxGiven++

					return exampleTx, nil
				}

				if exampleTxGiven == exampleTxCount {
					// All transactions parsed, simulate
					// the user cancelling the context
					cancelFn()

					return nil, io.EOF
				}

				exampleTxGiven++

				return exampleTx, nil
			},
		}
	)

	s := NewService(mockClient, mockSource, WithLogger(noop.New()))
	s.watchInterval = 10 * time.Millisecond // make the interval almost instant for the test

	// Execute the restore
	assert.NoError(
		t,
		s.ExecuteRestore(
			ctx,
			true, // Enable watch
		),
	)

	// Verify the restore was correct
	assert.Len(t, sentTxs, exampleTxCount)

	for _, tx := range sentTxs {
		assert.Equal(t, exampleTx, tx)
	}
}
