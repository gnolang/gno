package core

import (
	"fmt"
	"sync"
	"testing"

	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnvironment_IsolatedBetweenInstances verifies that two Environment
// instances with different mock BlockStores do not share state. This is the
// property that the old package-global design made impossible: whoever
// called SetBlockStore last owned every handler.
func TestEnvironment_IsolatedBetweenInstances(t *testing.T) {
	t.Parallel()

	const (
		heightA int64 = 100
		heightB int64 = 200
	)

	envA := &Environment{
		BlockStore: &mockBlockStore{
			heightFn: func() int64 { return heightA },
			loadBlockMetaFn: func(h int64) *types.BlockMeta {
				return &types.BlockMeta{Header: types.Header{Height: h}}
			},
			loadBlockFn: func(h int64) *types.Block {
				return &types.Block{Header: types.Header{Height: h}}
			},
		},
	}
	envB := &Environment{
		BlockStore: &mockBlockStore{
			heightFn: func() int64 { return heightB },
			loadBlockMetaFn: func(h int64) *types.BlockMeta {
				return &types.BlockMeta{Header: types.Header{Height: h}}
			},
			loadBlockFn: func(h int64) *types.Block {
				return &types.Block{Header: types.Header{Height: h}}
			},
		},
	}

	// Hammer both environments in parallel. If handlers were still reading
	// from a package-level global, all calls would route through whichever
	// env was constructed last and the heights would collide.
	const iters = 200
	errsA := make(chan error, iters)
	errsB := make(chan error, iters)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer close(errsA)
		for range iters {
			res, err := envA.Block(&rpctypes.Context{}, nil)
			if err != nil {
				errsA <- err
				return
			}
			if res.Block.Header.Height != heightA {
				errsA <- assertHeightMismatch(heightA, res.Block.Header.Height)
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		defer close(errsB)
		for range iters {
			res, err := envB.Block(&rpctypes.Context{}, nil)
			if err != nil {
				errsB <- err
				return
			}
			if res.Block.Header.Height != heightB {
				errsB <- assertHeightMismatch(heightB, res.Block.Header.Height)
				return
			}
		}
	}()
	wg.Wait()

	for err := range errsA {
		require.NoError(t, err, "envA handler read wrong state")
	}
	for err := range errsB {
		require.NoError(t, err, "envB handler read wrong state")
	}
}

// TestEnvironment_StartStopIdempotent verifies that Start is idempotent before
// Stop, and that Stop is idempotent (callable multiple times safely), including
// on an Environment without an EventSwitch (which skips txDispatcher creation).
func TestEnvironment_StartStopIdempotent(t *testing.T) {
	t.Parallel()

	env := &Environment{} // no EventSwitch; dispatcher stays nil
	require.NoError(t, env.Start())
	require.NoError(t, env.Start()) // second Start is a no-op
	assert.Nil(t, env.txDispatcher)

	require.NoError(t, env.Stop())
	require.NoError(t, env.Stop()) // second Stop is a no-op

	// Start after Stop must panic.
	assert.Panics(t, func() { _ = env.Start() })
}

// TestEnvironment_BroadcastTxCommitRejectsUnstarted verifies that
// BroadcastTxCommit returns a clear error when called on an Environment
// that has no dispatcher (not started or no EventSwitch), rather than
// panicking on a nil deref.
func TestEnvironment_BroadcastTxCommitRejectsUnstarted(t *testing.T) {
	t.Parallel()

	env := &Environment{}
	_, err := env.BroadcastTxCommit(&rpctypes.Context{}, types.Tx("x"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Environment not started")
}

// assertHeightMismatch produces a readable error for the parallelism test.
type heightMismatchError struct{ want, got int64 }

func (e heightMismatchError) Error() string {
	return fmt.Sprintf("height mismatch: want %d, got %d", e.want, e.got)
}

func assertHeightMismatch(want, got int64) error {
	return heightMismatchError{want: want, got: got}
}
