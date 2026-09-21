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

// Local is a Client implementation that directly executes the rpc functions
// on a given node's Environment, without going through any network
// connection.
//
// The Environment is owned by the Node and accessible via Node.RPCEnvironment()
// once the node has been started. A Local client bound to one Environment
// does not share state with other nodes in the same process.
//
// This implementation is useful for:
//
//   - Running tests against a node in-process without the overhead
//     of going through an http server
//   - Communication between an ABCI app and Tendermint core when they
//     are compiled in process
//
// For real clients, you probably want to use the [HTTP] client. For more
// powerful control during testing, you probably want the "client/mock" package.
type Local struct {
	Logger *slog.Logger
	env    *core.Environment
	ctx    *rpctypes.Context
}

// NewLocal configures a client that calls the given rpc/core Environment
// directly, without requiring a network connection. The Environment must
// be non-nil; pass the result of node.RPCEnvironment() after the node has
// been started.
func NewLocal(env *core.Environment) *Local {
	return &Local{
		Logger: log.NewNoopLogger(),
		env:    env,
		ctx:    &rpctypes.Context{},
	}
}

var _ Client = (*Local)(nil)

// SetLogger allows to set a logger on the client.
func (c *Local) SetLogger(l *slog.Logger) {
	c.Logger = l
}

func (c *Local) Status(_ context.Context, heightGte *int64) (*ctypes.ResultStatus, error) {
	return c.env.Status(c.ctx, heightGte)
}

func (c *Local) ABCIInfo(_ context.Context) (*ctypes.ResultABCIInfo, error) {
	return c.env.ABCIInfo(c.ctx)
}

func (c *Local) ABCIQuery(ctx context.Context, path string, data []byte) (*ctypes.ResultABCIQuery, error) {
	return c.ABCIQueryWithOptions(ctx, path, data, DefaultABCIQueryOptions)
}

func (c *Local) ABCIQueryWithOptions(_ context.Context, path string, data []byte, opts ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	return c.env.ABCIQuery(c.ctx, path, data, opts.Height, opts.Prove)
}

func (c *Local) BroadcastTxCommit(_ context.Context, tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	return c.env.BroadcastTxCommit(c.ctx, tx)
}

func (c *Local) BroadcastTxAsync(_ context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.env.BroadcastTxAsync(c.ctx, tx)
}

func (c *Local) BroadcastTxSync(_ context.Context, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.env.BroadcastTxSync(c.ctx, tx)
}

func (c *Local) UnconfirmedTxs(_ context.Context, limit int) (*ctypes.ResultUnconfirmedTxs, error) {
	return c.env.UnconfirmedTxs(c.ctx, limit)
}

func (c *Local) NumUnconfirmedTxs(_ context.Context) (*ctypes.ResultUnconfirmedTxs, error) {
	return c.env.NumUnconfirmedTxs(c.ctx)
}

func (c *Local) NetInfo(_ context.Context) (*ctypes.ResultNetInfo, error) {
	return c.env.NetInfo(c.ctx)
}

func (c *Local) DumpConsensusState(_ context.Context) (*ctypes.ResultDumpConsensusState, error) {
	return c.env.DumpConsensusState(c.ctx)
}

func (c *Local) ConsensusState(_ context.Context) (*ctypes.ResultConsensusState, error) {
	return c.env.ConsensusState(c.ctx)
}

func (c *Local) ConsensusParams(_ context.Context, height *int64) (*ctypes.ResultConsensusParams, error) {
	return c.env.ConsensusParams(c.ctx, height)
}

func (c *Local) Health(_ context.Context) (*ctypes.ResultHealth, error) {
	return c.env.Health(c.ctx)
}

func (c *Local) BlockchainInfo(_ context.Context, minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error) {
	return c.env.BlockchainInfo(c.ctx, minHeight, maxHeight)
}

func (c *Local) Genesis(_ context.Context) (*ctypes.ResultGenesis, error) {
	return c.env.Genesis(c.ctx)
}

func (c *Local) Block(_ context.Context, height *int64) (*ctypes.ResultBlock, error) {
	return c.env.Block(c.ctx, height)
}

func (c *Local) BlockResults(_ context.Context, height *int64) (*ctypes.ResultBlockResults, error) {
	return c.env.BlockResults(c.ctx, height)
}

func (c *Local) Commit(_ context.Context, height *int64) (*ctypes.ResultCommit, error) {
	return c.env.Commit(c.ctx, height)
}

func (c *Local) Validators(_ context.Context, height *int64) (*ctypes.ResultValidators, error) {
	return c.env.Validators(c.ctx, height)
}

func (c *Local) Tx(_ context.Context, hash []byte) (*ctypes.ResultTx, error) {
	return c.env.Tx(c.ctx, hash)
}
