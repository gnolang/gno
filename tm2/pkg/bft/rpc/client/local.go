package client

import (
	"context"
	"log/slog"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/log"
)

// Local is a Client implementation that directly executes the rpc
// functions on a given node, without going through any network connection.
//
// As this connects directly to a Node instance, a Local client only works
// after the Node has been started. Note that the way this works is (alas)
// through the use of singletons in rpc/core. As a consequence, you may only
// have one active node at a time, and Local can only connect to that specific
// node. Keep this in mind for parallel tests, or attempting to simulate a
// network.
//
// This implementation is useful for:
//
//   - Running tests against a node in-process without the overhead
//     of going through an http server
//   - Communication between an ABCI app and Tendermint core when they
//     are compiled in process.
//
// For real clients, you probably want to use the [HTTP] client.  For more
// powerful control during testing, you probably want the "client/mock" package.
type Local struct {
	Logger *slog.Logger
	ctx    *rpctypes.Context
}

// NewLocal configures a client that calls the Node directly through rpc/core,
// without requiring a network connection. See [Local].
func NewLocal() *Local {
	return &Local{
		Logger: log.NewNoopLogger(),
		ctx:    &rpctypes.Context{},
	}
}

var _ Client = (*Local)(nil)

// SetLogger allows to set a logger on the client.
func (c *Local) SetLogger(l *slog.Logger) {
	c.Logger = l
}

func (c *Local) Status(_ context.Context, heightGte *int64) (*ctypes.ResultStatus, error) {
	return core.Status(c.ctx, heightGte)
}

func (c *Local) ABCIInfo(_ context.Context) (*ctypes.ResultABCIInfo, error) {
	return core.ABCIInfo(c.ctx)
}

func (c *Local) ABCIQuery(ctx context.Context, path string, data []byte) (*ctypes.ResultABCIQuery, error) {
	return c.ABCIQueryWithOptions(ctx, path, data, DefaultABCIQueryOptions)
}

func (c *Local) ABCIQueryWithOptions(_ context.Context, path string, data []byte, opts ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	return core.ABCIQuery(c.ctx, path, data, opts.Height, opts.Prove)
}

func (c *Local) BroadcastTxCommit(_ context.Context, tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	return core.BroadcastTxCommit(c.ctx, tx)
}

func (c *Local) BroadcastTxAsync(_ context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return core.BroadcastTxAsync(c.ctx, tx)
}

func (c *Local) BroadcastTxSync(_ context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return core.BroadcastTxSync(c.ctx, tx)
}

func (c *Local) UnconfirmedTxs(_ context.Context, limit int) (*ctypes.ResultUnconfirmedTxs, error) {
	return core.UnconfirmedTxs(c.ctx, limit)
}

func (c *Local) NumUnconfirmedTxs(_ context.Context) (*ctypes.ResultUnconfirmedTxs, error) {
	return core.NumUnconfirmedTxs(c.ctx)
}

func (c *Local) NetInfo(_ context.Context) (*ctypes.ResultNetInfo, error) {
	return core.NetInfo(c.ctx)
}

func (c *Local) DumpConsensusState(_ context.Context) (*ctypes.ResultDumpConsensusState, error) {
	return core.DumpConsensusState(c.ctx)
}

func (c *Local) ConsensusState(_ context.Context) (*ctypes.ResultConsensusState, error) {
	return core.ConsensusState(c.ctx)
}

func (c *Local) ConsensusParams(_ context.Context, height *int64) (*ctypes.ResultConsensusParams, error) {
	return core.ConsensusParams(c.ctx, height)
}

func (c *Local) Health(_ context.Context) (*ctypes.ResultHealth, error) {
	return core.Health(c.ctx)
}

func (c *Local) BlockchainInfo(_ context.Context, minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error) {
	return core.BlockchainInfo(c.ctx, minHeight, maxHeight)
}

func (c *Local) Genesis(_ context.Context) (*ctypes.ResultGenesis, error) {
	return core.Genesis(c.ctx)
}

func (c *Local) Block(_ context.Context, height *int64) (*ctypes.ResultBlock, error) {
	return core.Block(c.ctx, height)
}

func (c *Local) BlockResults(_ context.Context, height *int64) (*ctypes.ResultBlockResults, error) {
	return core.BlockResults(c.ctx, height)
}

func (c *Local) Commit(_ context.Context, height *int64) (*ctypes.ResultCommit, error) {
	return core.Commit(c.ctx, height)
}

func (c *Local) Validators(_ context.Context, height *int64) (*ctypes.ResultValidators, error) {
	return core.Validators(c.ctx, height)
}

func (c *Local) Tx(_ context.Context, hash []byte) (*ctypes.ResultTx, error) {
	return core.Tx(c.ctx, hash)
}
