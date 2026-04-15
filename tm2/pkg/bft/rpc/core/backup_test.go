package core

import (
	"context"
	"testing"

	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBackupBlocks_ClientDisconnect verifies that the loop stops early when
// the client's context is cancelled mid-stream (e.g. the WebSocket drops).
func TestBackupBlocks_ClientDisconnect(t *testing.T) {
	t.Parallel()

	const totalBlocks = 5
	const disconnectAfter = 2 // cancel context after writing this many blocks

	SetLogger(log.NewNoopLogger())

	ctx, cancel := context.WithCancel(context.Background())

	writeCount := 0
	conn := &callbackWSConn{
		ctx: ctx,
		writeRPCFn: func(rpctypes.RPCResponses) {
			writeCount++
			if writeCount >= disconnectAfter {
				cancel() // simulate the connection dropping
			}
		},
		tryWriteRPCFn:   func(rpctypes.RPCResponses) bool { return true },
		getRemoteAddrFn: func() string { return "test-addr" },
	}

	SetBlockStore(&mockBlockStore{
		heightFn: func() int64 { return totalBlocks },
		loadBlockFn: func(h int64) *types.Block {
			return &types.Block{Header: types.Header{Height: h}}
		},
	})

	rpcCtx := &rpctypes.Context{
		JSONReq: &rpctypes.RPCRequest{ID: rpctypes.JSONRPCStringID("test")},
		WSConn:  conn,
	}

	result, err := BackupBlocks(rpcCtx, 1, totalBlocks)

	// The function must return an error when the context is cancelled.
	require.Error(t, err)
	assert.Nil(t, result)

	// Exactly disconnectAfter blocks must have been streamed before the loop
	// detected the cancellation. If this is 0, it means the context check fired
	// before any block was sent (the `Done() != nil` bug).
	assert.Equal(t, disconnectAfter, writeCount,
		"expected exactly %d block(s) written before disconnect, got %d",
		disconnectAfter, writeCount)
}

// TestBackupBlocks_FullRangeSuccess verifies that all blocks are streamed and
// Done is set when no disconnect occurs.
func TestBackupBlocks_FullRangeSuccess(t *testing.T) {
	t.Parallel()

	const totalBlocks = 5

	SetLogger(log.NewNoopLogger())

	writeCount := 0
	conn := &callbackWSConn{
		ctx: t.Context(),
		writeRPCFn: func(rpctypes.RPCResponses) {
			writeCount++
		},
		tryWriteRPCFn:   func(rpctypes.RPCResponses) bool { return true },
		getRemoteAddrFn: func() string { return "test-addr" },
	}

	SetBlockStore(&mockBlockStore{
		heightFn: func() int64 { return totalBlocks },
		loadBlockFn: func(h int64) *types.Block {
			return &types.Block{Header: types.Header{Height: h}}
		},
	})

	rpcCtx := &rpctypes.Context{
		JSONReq: &rpctypes.RPCRequest{ID: rpctypes.JSONRPCStringID("test")},
		WSConn:  conn,
	}

	result, err := BackupBlocks(rpcCtx, 1, totalBlocks)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Done)
	assert.Equal(t, totalBlocks, writeCount)
}

// callbackWSConn is a flexible WSRPCConnection whose methods are provided as
// callbacks, making it easy to inject behaviour per test.
type callbackWSConn struct {
	ctx             context.Context
	getRemoteAddrFn func() string
	writeRPCFn      func(rpctypes.RPCResponses)
	tryWriteRPCFn   func(rpctypes.RPCResponses) bool
}

func (c *callbackWSConn) GetRemoteAddr() string                     { return c.getRemoteAddrFn() }
func (c *callbackWSConn) WriteRPCResponses(r rpctypes.RPCResponses) { c.writeRPCFn(r) }
func (c *callbackWSConn) TryWriteRPCResponses(r rpctypes.RPCResponses) bool {
	return c.tryWriteRPCFn(r)
}
func (c *callbackWSConn) Context() context.Context { return c.ctx }
