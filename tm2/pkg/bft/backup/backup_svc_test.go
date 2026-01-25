package backup

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/backup/backuppb"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
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
			client, cleanup := newTestClient(t, store)
			defer cleanup()

			stream, err := client.StreamBlocks(
				context.Background(),
				&backuppb.StreamBlocksRequest{
					StartHeight: tc.start,
					EndHeight:   tc.end,
				},
			)
			if tc.errContains == "" {
				require.NoError(t, err)
			} else if err != nil {
				require.ErrorContains(t, err, tc.errContains)
				return
			}

			data := []*types.Block(nil)
			var streamErr error
			for {
				msg, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					streamErr = err
					break
				}
				block := &types.Block{}
				require.NoError(t, amino.Unmarshal(msg.Data, block))
				data = append(data, block)
			}
			if tc.errContains == "" {
				require.NoError(t, streamErr)
			} else {
				require.ErrorContains(t, streamErr, tc.errContains)
			}
			require.Equal(t, tc.expectedResult, data)
		})
	}
}

func TestNewServer(t *testing.T) {
	require.NotPanics(t, func() {
		srv := grpc.NewServer()
		backuppb.RegisterBackupServiceServer(srv, NewBackupServiceHandler(&mockBlockStore{}))
		require.NotNil(t, srv)
	})
}

func newTestClient(t *testing.T, store blockStore) (backuppb.BackupServiceClient, func()) {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	backuppb.RegisterBackupServiceServer(server, NewBackupServiceHandler(store))

	go func() {
		_ = server.Serve(listener)
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
	conn, err := grpc.NewClient(
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	cleanup := func() {
		_ = conn.Close()
		server.Stop()
		_ = listener.Close()
	}

	return backuppb.NewBackupServiceClient(conn), cleanup
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
