package main

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/backup"
	"github.com/gnolang/gno/tm2/pkg/bft/backup/backuppb"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestBackup(t *testing.T) {
	store := &mockBlockStore{height: 5, blocks: map[int64]*types.Block{
		1: {Header: types.Header{Height: 1}},
		2: {Header: types.Header{Height: 2}},
		3: {Header: types.Header{Height: 3}},
		4: {Header: types.Header{Height: 4}},
		5: {Header: types.Header{Height: 5}},
	}}

	server := grpc.NewServer()
	backupService := backup.NewBackupServiceHandler(store)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	backuppb.RegisterBackupServiceServer(server, backupService)
	t.Cleanup(func() {
		server.Stop()
		_ = lis.Close()
	})

	go func() {
		if err := server.Serve(lis); err != nil {
			t.Error(err)
		}
	}()

	io := commands.NewTestIO()
	io.SetOut(os.Stdout)
	io.SetErr(os.Stderr)
	outDir := t.TempDir()

	err = newRootCmd(io).ParseAndRun(context.Background(), []string{
		"--remote", lis.Addr().String(),
		"--o", outDir,
	})
	require.NoError(t, err)
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
