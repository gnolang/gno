package gnoclient

import (
	"testing"

	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockResults(t *testing.T) {
	t.Parallel()

	height := int64(5)
	client := &Client{
		Signer: &mockSigner{},
		RPCClient: &mockRPCClient{
			blockResults: func(height *int64) (*ctypes.ResultBlockResults, error) {
				return &ctypes.ResultBlockResults{
					Height:  *height,
					Results: nil,
				}, nil
			},
		},
	}

	blockResult, err := client.BlockResult(height)
	require.NoError(t, err)
	assert.Equal(t, height, blockResult.Height)
}

func TestLatestBlockHeight(t *testing.T) {
	t.Parallel()

	latestHeight := int64(5)

	client := &Client{
		Signer: &mockSigner{},
		RPCClient: &mockRPCClient{
			status: func() (*ctypes.ResultStatus, error) {
				return &ctypes.ResultStatus{
					SyncInfo: ctypes.SyncInfo{
						LatestBlockHeight: latestHeight,
					},
				}, nil
			},
		},
	}

	head, err := client.LatestBlockHeight()
	require.NoError(t, err)
	assert.Equal(t, latestHeight, head)
}

func TestBlockErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		height        int64
		expectedError error
	}{
		{
			name: "Invalid RPCClient",
			client: Client{
				&mockSigner{},
				nil,
			},
			height:        1,
			expectedError: ErrMissingRPCClient,
		},
		{
			name: "Invalid height",
			client: Client{
				&mockSigner{},
				&mockRPCClient{},
			},
			height:        0,
			expectedError: ErrInvalidBlockHeight,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.Block(tc.height)
			assert.Nil(t, res)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestBlockResultErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		height        int64
		expectedError error
	}{
		{
			name: "Invalid RPCClient",
			client: Client{
				&mockSigner{},
				nil,
			},
			height:        1,
			expectedError: ErrMissingRPCClient,
		},
		{
			name: "Invalid height",
			client: Client{
				&mockSigner{},
				&mockRPCClient{},
			},
			height:        0,
			expectedError: ErrInvalidBlockHeight,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.BlockResult(tc.height)
			assert.Nil(t, res)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestLatestBlockHeightErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		expectedError error
	}{
		{
			name: "Invalid RPCClient",
			client: Client{
				&mockSigner{},
				nil,
			},
			expectedError: ErrMissingRPCClient,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.LatestBlockHeight()
			assert.Equal(t, int64(0), res)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}
