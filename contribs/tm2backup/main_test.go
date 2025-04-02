package main

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/backup"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/require"
)

func TestBackup(t *testing.T) {
	store := &mockBlockStore{height: 5, blocks: map[int64]*types.Block{
		1: {Header: types.Header{Height: 1}},
		2: {Header: types.Header{Height: 2}},
		3: {Header: types.Header{Height: 3}},
		4: {Header: types.Header{Height: 4}},
		5: {Header: types.Header{Height: 5}},
	}}

	mux := backup.NewMux(store)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	io := commands.NewTestIO()
	io.SetOut(os.Stdout)
	io.SetErr(os.Stderr)

	err := newRootCmd(io).ParseAndRun(context.Background(), []string{
		"--remote", srv.URL,
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
