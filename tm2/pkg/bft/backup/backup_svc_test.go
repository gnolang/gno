package backup

import (
	"context"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/backup/backuppb"
	"github.com/gnolang/gno/tm2/pkg/bft/backup/backuppb/backuppbconnect"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/stretchr/testify/require"
)

func TestStreamBlocks(t *testing.T) {
	tcs := []struct {
		name           string
		initStore      func() *mockBlockStore
		start          int64
		end            int64
		expectedResult []*types.Block
		errContains    string
	}{
		{
			name: "one block store",
			initStore: func() *mockBlockStore {
				return &mockBlockStore{height: 1, blocks: map[int64]*types.Block{
					1: {Header: types.Header{Height: 1}},
				}}
			},
			expectedResult: []*types.Block{
				{Header: types.Header{Height: 1}},
			},
		},
		{
			name:  "range",
			start: 2,
			end:   4,
			initStore: func() *mockBlockStore {
				return &mockBlockStore{height: 5, blocks: map[int64]*types.Block{
					1: {Header: types.Header{Height: 1}},
					2: {Header: types.Header{Height: 2}},
					3: {Header: types.Header{Height: 3}},
					4: {Header: types.Header{Height: 4}},
					5: {Header: types.Header{Height: 5}},
				}}
			},
			expectedResult: []*types.Block{
				{Header: types.Header{Height: 2}},
				{Header: types.Header{Height: 3}},
				{Header: types.Header{Height: 4}},
			},
		},
		{
			name: "range no start",
			end:  4,
			initStore: func() *mockBlockStore {
				return &mockBlockStore{height: 5, blocks: map[int64]*types.Block{
					1: {Header: types.Header{Height: 1}},
					2: {Header: types.Header{Height: 2}},
					3: {Header: types.Header{Height: 3}},
					4: {Header: types.Header{Height: 4}},
					5: {Header: types.Header{Height: 5}},
				}}
			},
			expectedResult: []*types.Block{
				{Header: types.Header{Height: 1}},
				{Header: types.Header{Height: 2}},
				{Header: types.Header{Height: 3}},
				{Header: types.Header{Height: 4}},
			},
		},
		{
			name:  "range no end",
			start: 2,
			initStore: func() *mockBlockStore {
				return &mockBlockStore{height: 5, blocks: map[int64]*types.Block{
					1: {Header: types.Header{Height: 1}},
					2: {Header: types.Header{Height: 2}},
					3: {Header: types.Header{Height: 3}},
					4: {Header: types.Header{Height: 4}},
					5: {Header: types.Header{Height: 5}},
				}}
			},
			expectedResult: []*types.Block{
				{Header: types.Header{Height: 2}},
				{Header: types.Header{Height: 3}},
				{Header: types.Header{Height: 4}},
				{Header: types.Header{Height: 5}},
			},
		},
		{
			name: "range no params",
			initStore: func() *mockBlockStore {
				return &mockBlockStore{height: 5, blocks: map[int64]*types.Block{
					1: {Header: types.Header{Height: 1}},
					2: {Header: types.Header{Height: 2}},
					3: {Header: types.Header{Height: 3}},
					4: {Header: types.Header{Height: 4}},
					5: {Header: types.Header{Height: 5}},
				}}
			},
			expectedResult: []*types.Block{
				{Header: types.Header{Height: 1}},
				{Header: types.Header{Height: 2}},
				{Header: types.Header{Height: 3}},
				{Header: types.Header{Height: 4}},
				{Header: types.Header{Height: 5}},
			},
		},
		{
			name: "err nil block",
			initStore: func() *mockBlockStore {
				return &mockBlockStore{height: 3, blocks: map[int64]*types.Block{
					1: {Header: types.Header{Height: 1}},
					2: nil,
					3: {Header: types.Header{Height: 3}},
				}}
			},
			errContains: "block store returned nil block for height 2",
			expectedResult: []*types.Block{
				{Header: types.Header{Height: 1}},
			},
		},
		{
			name: "err empty store",
			initStore: func() *mockBlockStore {
				return &mockBlockStore{}
			},
			errContains: "block store returned invalid max height (0)",
		},
		{
			name:  "err reverse range",
			start: 2,
			end:   1,
			initStore: func() *mockBlockStore {
				return &mockBlockStore{height: 1, blocks: map[int64]*types.Block{1: {}, 2: {}}}
			},
			errContains: "end height must be >= than start height",
		},
		{
			name:  "err invalid start",
			start: -42,
			initStore: func() *mockBlockStore {
				return &mockBlockStore{height: 1, blocks: map[int64]*types.Block{1: {Header: types.Header{Height: 1}}}}
			},
			errContains: "start height must be >= 1, got -42",
		},
		{
			name: "err invalid end",
			end:  42,
			initStore: func() *mockBlockStore {
				return &mockBlockStore{height: 1, blocks: map[int64]*types.Block{1: {}}}
			},
			errContains: "end height must be <= 1",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			store := tc.initStore()
			mux := NewMux(store)
			srv := httptest.NewServer(mux)
			defer srv.Close()
			httpClient := srv.Client()
			client := backuppbconnect.NewBackupServiceClient(httpClient, srv.URL)

			stream, err := client.StreamBlocks(context.Background(), &connect.Request[backuppb.StreamBlocksRequest]{Msg: &backuppb.StreamBlocksRequest{
				StartHeight: tc.start,
				EndHeight:   tc.end,
			}})
			require.NoError(t, err)
			defer func() {
				require.NoError(t, stream.Close())
			}()

			data := []*types.Block(nil)
			for {
				if !stream.Receive() {
					err := stream.Err()
					if tc.errContains == "" {
						require.NoError(t, err)
					} else {
						require.ErrorContains(t, err, tc.errContains)
					}
					break
				}
				msg := stream.Msg()
				block := &types.Block{}
				require.NoError(t, amino.Unmarshal(msg.Data, block))
				data = append(data, block)
			}
			require.Equal(t, tc.expectedResult, data)
		})
	}
}

func TestNewServer(t *testing.T) {
	require.NotPanics(t, func() {
		serv := NewServer(DefaultConfig(), &mockBlockStore{})
		require.NotNil(t, serv)
	})
}

type mockBlockStore struct {
	height int64
	blocks map[int64]*types.Block
}

// Height implements blockStore.
func (m *mockBlockStore) Height() int64 {
	return m.height
}

// LoadBlock implements blockStore.
func (m *mockBlockStore) LoadBlock(height int64) *types.Block {
	return m.blocks[height]
}
