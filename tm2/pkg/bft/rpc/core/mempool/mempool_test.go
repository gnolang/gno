package mempool

import (
	"errors"
	"testing"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mock"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_BroadcastTxAsyncHandler(t *testing.T) {
	t.Parallel()

	t.Run("Missing tx param", func(t *testing.T) {
		t.Parallel()

		h := &Handler{
			mempool: &mock.Mempool{},
		}

		res, err := h.BroadcastTxAsyncHandler(nil, nil)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("CheckTx error", func(t *testing.T) {
		t.Parallel()

		var (
			checkErr = errors.New("mempool error")
			mp       = &mock.Mempool{
				CheckTxFn: func(tx types.Tx, cb func(abci.Response)) error {
					return checkErr
				},
			}

			h = &Handler{
				mempool: mp,
			}

			txBytes = []byte("tx-bytes")
			params  = []any{txBytes}
		)

		res, err := h.BroadcastTxAsyncHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
		assert.Contains(t, err.Message, checkErr.Error())
	})

	t.Run("Valid broadcast", func(t *testing.T) {
		t.Parallel()

		var (
			capturedTx types.Tx
			mp         = &mock.Mempool{
				CheckTxFn: func(tx types.Tx, cb func(abci.Response)) error {
					capturedTx = tx
					return nil
				},
			}

			h = &Handler{
				mempool: mp,
			}

			txBytes = []byte("some-tx")
			params  = []any{txBytes}
		)

		res, err := h.BroadcastTxAsyncHandler(nil, params)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultBroadcastTx)
		require.True(t, ok)

		expectedHash := types.Tx(txBytes).Hash()
		assert.Equal(t, expectedHash, result.Hash)
		assert.Equal(t, types.Tx(txBytes), capturedTx)
	})
}

func TestHandler_BroadcastTxSyncHandler(t *testing.T) {
	t.Parallel()

	t.Run("Missing tx param", func(t *testing.T) {
		t.Parallel()

		h := &Handler{
			mempool: &mock.Mempool{},
		}

		res, err := h.BroadcastTxSyncHandler(nil, nil)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("CheckTx error", func(t *testing.T) {
		t.Parallel()

		var (
			checkErr = errors.New("sync mempool error")

			mp = &mock.Mempool{
				CheckTxFn: func(tx types.Tx, cb func(abci.Response)) error {
					return checkErr
				},
			}

			h = &Handler{
				mempool: mp,
			}

			params = []any{[]byte("tx")}
		)

		res, err := h.BroadcastTxSyncHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
		assert.Contains(t, err.Message, checkErr.Error())
	})

	t.Run("Valid CheckTx response", func(t *testing.T) {
		t.Parallel()

		var (
			txBytes = []byte("sync-tx")
			tx      = types.Tx(txBytes)

			checkResp = abci.ResponseCheckTx{
				ResponseBase: abci.ResponseBase{
					Data:  []byte("data"),
					Log:   "log-message",
					Error: nil,
				},
			}

			mp = &mock.Mempool{
				CheckTxFn: func(txArg types.Tx, cb func(abci.Response)) error {
					assert.Equal(t, tx, txArg)

					cb(checkResp)
					return nil
				},
			}

			h = &Handler{
				mempool: mp,
			}

			params = []any{txBytes}
		)

		res, err := h.BroadcastTxSyncHandler(nil, params)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultBroadcastTx)
		require.True(t, ok)

		assert.Equal(t, checkResp.Error, result.Error)
		assert.Equal(t, checkResp.Data, result.Data)
		assert.Equal(t, checkResp.Log, result.Log)
		assert.Equal(t, tx.Hash(), result.Hash)
	})
}

func TestHandler_BroadcastTxCommitHandler(t *testing.T) {
	t.Parallel()

	t.Run("Missing tx param", func(t *testing.T) {
		t.Parallel()

		h := &Handler{
			mempool: &mock.Mempool{},
		}

		res, err := h.BroadcastTxCommitHandler(nil, nil)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("CheckTx call error", func(t *testing.T) {
		t.Parallel()

		var (
			checkErr = errors.New("commit mempool error")
			mp       = &mock.Mempool{
				CheckTxFn: func(tx types.Tx, cb func(abci.Response)) error {
					return checkErr
				},
			}
			h = &Handler{
				mempool: mp,
			}

			params = []any{[]byte("tx")}
		)

		res, err := h.BroadcastTxCommitHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
		assert.Contains(t, err.Message, "error on BroadcastTxCommit")
		assert.Contains(t, err.Message, checkErr.Error())
	})

	t.Run("CheckTx response error", func(t *testing.T) {
		t.Parallel()

		var (
			txBytes = []byte("commit-tx")
			tx      = types.Tx(txBytes)

			checkResp = abci.ResponseCheckTx{
				ResponseBase: abci.ResponseBase{
					Error: testABCIError{msg: "check failed"},
					Data:  []byte("ignored"),
					Log:   "ignored",
				},
			}

			mp = &mock.Mempool{
				CheckTxFn: func(txArg types.Tx, cb func(abci.Response)) error {
					assert.Equal(t, tx, txArg)
					cb(checkResp)

					return nil
				},
			}

			h = &Handler{
				mempool:    mp,
				dispatcher: nil, // explicit
			}

			params = []any{txBytes}
		)

		res, err := h.BroadcastTxCommitHandler(nil, params)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultBroadcastTxCommit)
		require.True(t, ok)

		assert.Equal(t, checkResp, result.CheckTx)
		assert.Equal(t, abci.ResponseDeliverTx{}, result.DeliverTx)
		assert.Equal(t, tx.Hash(), result.Hash)
		assert.Equal(t, int64(0), result.Height)
	})

	t.Run("Successful commit", func(t *testing.T) {
		t.Parallel()

		var (
			txBytes = []byte("commit-success-tx")
			tx      = types.Tx(txBytes)

			checkResp = abci.ResponseCheckTx{
				ResponseBase: abci.ResponseBase{
					Error: nil,
					Data:  []byte("check-data"),
					Log:   "check-log",
				},
			}

			expectedDeliver = abci.ResponseDeliverTx{
				ResponseBase: abci.ResponseBase{
					Data: []byte("deliver-data"),
					Log:  "deliver-log",
				},
			}
			expectedHeight = int64(42)
			params         = []any{txBytes}
		)

		waiter := newTxWaiter()
		waiter.res = types.TxResult{
			Height:   expectedHeight,
			Response: expectedDeliver,
		}
		close(waiter.done)

		dispatcher := &txDispatcher{
			timeout: time.Minute,
			waiters: map[string]*txWaiter{
				string(tx): waiter,
			},
		}

		mp := &mock.Mempool{
			CheckTxFn: func(txArg types.Tx, cb func(abci.Response)) error {
				assert.Equal(t, tx, txArg)
				cb(checkResp)

				return nil
			},
		}

		h := &Handler{
			mempool:    mp,
			dispatcher: dispatcher,
		}

		res, err := h.BroadcastTxCommitHandler(nil, params)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultBroadcastTxCommit)
		require.True(t, ok)

		assert.Equal(t, checkResp, result.CheckTx)
		assert.Equal(t, expectedDeliver, result.DeliverTx)
		assert.Equal(t, expectedHeight, result.Height)
		assert.Equal(t, tx.Hash(), result.Hash)
	})
}
func TestHandler_UnconfirmedTxsHandler(t *testing.T) {
	t.Parallel()

	t.Run("Invalid limit param", func(t *testing.T) {
		t.Parallel()

		var (
			h = &Handler{
				mempool: &mock.Mempool{},
			}

			params = []any{"not-an-int"}
		)

		res, err := h.UnconfirmedTxsHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Valid limit and mempool data", func(t *testing.T) {
		t.Parallel()

		var (
			expectedTxs = []types.Tx{
				[]byte("tx1"),
				[]byte("tx2"),
				[]byte("tx3"),
			}

			mp = &mock.Mempool{
				ReapMaxTxsFn: func(maxTxs int) types.Txs {
					assert.Equal(t, 10, maxTxs)

					return expectedTxs
				},
				SizeFn: func() int {
					return 5
				},
				TxsBytesFn: func() int64 {
					return 123
				},
			}

			h = &Handler{
				mempool: mp,
			}

			params = []any{int64(10)}
		)

		res, err := h.UnconfirmedTxsHandler(nil, params)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultUnconfirmedTxs)
		require.True(t, ok)

		assert.Equal(t, len(expectedTxs), result.Count)
		assert.Equal(t, 5, result.Total)
		assert.Equal(t, int64(123), result.TotalBytes)
		assert.Equal(t, expectedTxs, result.Txs)
	})
}

func TestHandler_NumUnconfirmedTxsHandler(t *testing.T) {
	t.Parallel()

	t.Run("Unexpected params", func(t *testing.T) {
		t.Parallel()

		h := &Handler{
			mempool: &mock.Mempool{},
		}

		res, err := h.NumUnconfirmedTxsHandler(nil, []any{"extra"})
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Valid call", func(t *testing.T) {
		t.Parallel()

		var (
			size     = 7
			txsBytes = int64(456)

			mp = &mock.Mempool{
				SizeFn: func() int {
					return size
				},
				TxsBytesFn: func() int64 {
					return txsBytes
				},
			}

			h = &Handler{
				mempool: mp,
			}
		)

		res, err := h.NumUnconfirmedTxsHandler(nil, nil)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultUnconfirmedTxs)
		require.True(t, ok)

		assert.Equal(t, size, result.Count)
		assert.Equal(t, size, result.Total)
		assert.Equal(t, txsBytes, result.TotalBytes)
	})
}

type testABCIError struct {
	msg string
}

func (e testABCIError) Error() string {
	return e.msg
}

func (e testABCIError) AssertABCIError() {}
