package core

import (
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
)

func TestABCIQuery(t *testing.T) {
	t.Parallel()

	origBlockStore, origProxyAppQuery := blockStore, proxyAppQuery
	defer func() {
		blockStore, proxyAppQuery = origBlockStore, origProxyAppQuery
	}()

	tests := []struct {
		name           string
		path           string
		data           []byte
		height         int64
		prove          bool
		mockHeight     int64
		mockQueryResp  *abci.ResponseQuery
		mockQueryErr   error
		expectedResult *ctypes.ResultABCIQuery
		expectedError  string
	}{
		{
			name:       "valid query",
			path:       "/a/b/c",
			data:       []byte("data"),
			height:     10,
			prove:      false,
			mockHeight: 20,
			mockQueryResp: &abci.ResponseQuery{
				Key:    []byte("key"),
				Value:  []byte("result"),
				Height: 10,
			},
			expectedResult: &ctypes.ResultABCIQuery{
				Response: abci.ResponseQuery{
					Key:    []byte("key"),
					Value:  []byte("result"),
					Height: 10,
				},
			},
		},
		{
			name:           "negative height",
			height:         -1,
			mockHeight:     20,
			expectedResult: nil,
			expectedError:  "height cannot be negative",
		},
		{
			name:           "future height",
			height:         30,
			mockHeight:     20,
			expectedResult: nil,
			expectedError:  "requested height 30 is in the future (latest height is 20)",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockBS := &mockBlockStore{
				heightFn: func() int64 { return tt.mockHeight },
			}
			blockStore = mockBS

			mockPAQ := &mockProxyAppQuery{
				queryFn: func(req abci.RequestQuery) (abci.ResponseQuery, error) {
					if tt.mockQueryResp != nil {
						return *tt.mockQueryResp, tt.mockQueryErr
					}
					return abci.ResponseQuery{}, tt.mockQueryErr
				},
			}
			proxyAppQuery = mockPAQ

			result, err := ABCIQuery(&rpctypes.Context{}, tt.path, tt.data, tt.height, tt.prove)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}
