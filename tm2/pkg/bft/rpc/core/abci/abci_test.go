package abci

import (
	"errors"
	"testing"

	abciTypes "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mock"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_QueryHandler(t *testing.T) {
	t.Parallel()

	t.Run("Missing data param", func(t *testing.T) {
		t.Parallel()

		var (
			mockQuery = &mock.AppConn{
				QuerySyncFn: func(_ abciTypes.RequestQuery) (abciTypes.ResponseQuery, error) {
					t.FailNow()

					return abciTypes.ResponseQuery{}, nil
				},
			}

			params = []any{
				"some/path",
			}
		)

		h := NewHandler(mockQuery)

		res, err := h.QueryHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Query sync error", func(t *testing.T) {
		t.Parallel()

		var (
			queryErr = errors.New("app query error")
			params   = []any{
				"some/path",
				[]byte("data"),
			}

			mockQuery = &mock.AppConn{
				QuerySyncFn: func(_ abciTypes.RequestQuery) (abciTypes.ResponseQuery, error) {
					return abciTypes.ResponseQuery{}, queryErr
				},
			}
		)

		h := NewHandler(mockQuery)

		res, err := h.QueryHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
		assert.Contains(t, err.Message, queryErr.Error())
	})

	t.Run("Valid query", func(t *testing.T) {
		t.Parallel()

		var (
			height           = int64(10)
			expectedResponse = abciTypes.ResponseQuery{
				Height: height,
			}

			params = []any{
				"some/path",       // path
				[]byte("payload"), // data
				height,            // height
				true,              // prove
			}

			expectedRequest = abciTypes.RequestQuery{
				Path:   "some/path",
				Data:   []byte("payload"),
				Height: 10,
				Prove:  true,
			}

			mockQuery = &mock.AppConn{
				QuerySyncFn: func(req abciTypes.RequestQuery) (abciTypes.ResponseQuery, error) {
					assert.Equal(t, expectedRequest, req)

					return expectedResponse, nil
				},
			}
		)

		h := NewHandler(mockQuery)

		res, err := h.QueryHandler(nil, params)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultABCIQuery)
		require.True(t, ok)

		assert.Equal(t, expectedResponse, result.Response)
	})

	t.Run("Valid query with defaults", func(t *testing.T) {
		t.Parallel()

		var (
			params = []any{
				// path="", height=0, prove=false defaults
				nil,
				[]byte("data-only"),
			}
			expectedRequest = abciTypes.RequestQuery{
				Path:   "",
				Data:   []byte("data-only"),
				Height: 0,
				Prove:  false,
			}
			expectedResponse = abciTypes.ResponseQuery{}

			mockQuery = &mock.AppConn{
				QuerySyncFn: func(req abciTypes.RequestQuery) (abciTypes.ResponseQuery, error) {
					assert.Equal(t, expectedRequest, req)

					return expectedResponse, nil
				},
			}
		)

		h := NewHandler(mockQuery)

		res, err := h.QueryHandler(nil, params)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultABCIQuery)
		require.True(t, ok)

		assert.Equal(t, expectedResponse, result.Response)
	})
}

func TestHandler_InfoHandler(t *testing.T) {
	t.Parallel()

	t.Run("Params not allowed", func(t *testing.T) {
		t.Parallel()

		var (
			params = []any{"unexpected"}

			mockQuery = &mock.AppConn{
				InfoSyncFn: func(_ abciTypes.RequestInfo) (abciTypes.ResponseInfo, error) {
					t.FailNow()

					return abciTypes.ResponseInfo{}, nil
				},
			}
		)

		h := NewHandler(mockQuery)

		res, err := h.InfoHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.InvalidParamsErrorCode, err.Code)
	})

	t.Run("Info error", func(t *testing.T) {
		t.Parallel()

		var (
			infoErr = errors.New("info failed")
			params  = []any(nil)

			mockQuery = &mock.AppConn{
				InfoSyncFn: func(req abciTypes.RequestInfo) (abciTypes.ResponseInfo, error) {
					// The request should always be empty
					assert.Equal(t, abciTypes.RequestInfo{}, req)

					return abciTypes.ResponseInfo{}, infoErr
				},
			}
		)

		h := NewHandler(mockQuery)

		res, err := h.InfoHandler(nil, params)
		require.Nil(t, res)
		require.NotNil(t, err)

		assert.Equal(t, spec.ServerErrorCode, err.Code)
		assert.Contains(t, err.Message, infoErr.Error())
	})

	t.Run("Valid info", func(t *testing.T) {
		t.Parallel()

		var (
			expectedResponse = abciTypes.ResponseInfo{
				ResponseBase: abciTypes.ResponseBase{
					Data: []byte("some-info"),
				},
				ABCIVersion: "v1.2.3",
			}

			mockQuery = &mock.AppConn{
				InfoSyncFn: func(req abciTypes.RequestInfo) (abciTypes.ResponseInfo, error) {
					// The request should always be empty
					assert.Equal(t, abciTypes.RequestInfo{}, req)

					return expectedResponse, nil
				},
			}
		)

		h := NewHandler(mockQuery)

		res, err := h.InfoHandler(nil, nil)
		require.Nil(t, err)
		require.NotNil(t, res)

		result, ok := res.(*ResultABCIInfo)
		require.True(t, ok)

		assert.Equal(t, expectedResponse, result.Response)
	})
}
