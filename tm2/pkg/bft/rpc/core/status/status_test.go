package status

import (
	"errors"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_StatusHandler(t *testing.T) {
	t.Parallel()

	t.Run("Invalid GTE param", func(t *testing.T) {
		t.Parallel()

		h := NewHandler(func() (*ResultStatus, error) {
			t.FailNow()

			return nil, nil
		})

		res, err := h.StatusHandler(nil, []any{"not-an-int"})
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Build status error", func(t *testing.T) {
		t.Parallel()

		buildErr := errors.New("build failed")

		h := NewHandler(func() (*ResultStatus, error) {
			return nil, buildErr
		})

		res, err := h.StatusHandler(nil, nil)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
		assert.Contains(t, err.Message, buildErr.Error())
	})

	t.Run("heightGte not satisfied", func(t *testing.T) {
		t.Parallel()

		h := NewHandler(func() (*ResultStatus, error) {
			return &ResultStatus{
				SyncInfo: SyncInfo{
					LatestBlockHeight: 5,
				},
			}, nil
		})

		res, err := h.StatusHandler(nil, []any{int64(10)})
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidRequestErrorCode, err.Code)
	})

	t.Run("Valid status, no heightGte", func(t *testing.T) {
		t.Parallel()

		expected := &ResultStatus{
			SyncInfo: SyncInfo{
				LatestBlockHeight: 12,
			},
		}

		h := NewHandler(func() (*ResultStatus, error) {
			return expected, nil
		})

		res, err := h.StatusHandler(nil, nil)
		require.Nil(t, err)
		require.NotNil(t, res)

		out, ok := res.(*ResultStatus)
		require.True(t, ok)

		assert.Equal(t, expected, out)
	})

	t.Run("Valid status, heightGte satisfied", func(t *testing.T) {
		t.Parallel()

		expected := &ResultStatus{
			SyncInfo: SyncInfo{
				LatestBlockHeight: 10,
			},
		}

		h := NewHandler(func() (*ResultStatus, error) {
			return expected, nil
		})

		res, err := h.StatusHandler(nil, []any{int64(10)})
		require.Nil(t, err)
		require.NotNil(t, res)

		out, ok := res.(*ResultStatus)
		require.True(t, ok)

		assert.Equal(t, expected, out)
	})
}

func TestHandler_StatusIndexerLegacy(t *testing.T) {
	t.Parallel()

	var status *ResultStatus
	assert.False(t, status.TxIndexEnabled())

	status = &ResultStatus{}
	assert.False(t, status.TxIndexEnabled())

	status.NodeInfo = types.NodeInfo{}
	assert.False(t, status.TxIndexEnabled())

	cases := []struct {
		expected bool
		other    types.NodeInfoOther
	}{
		{false, types.NodeInfoOther{}},
		{false, types.NodeInfoOther{TxIndex: "aa"}},
		{false, types.NodeInfoOther{TxIndex: "off"}},
		{true, types.NodeInfoOther{TxIndex: "on"}},
	}

	for _, tc := range cases {
		status.NodeInfo.Other = tc.other
		assert.Equal(t, tc.expected, status.TxIndexEnabled())
	}
}
