package gnoclient

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/abci"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/blocks"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/consensus"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/health"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mempool"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/net"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/status"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/tx"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Signer mock
type (
	mockSign     func(cfg SignCfg) (*std.Tx, error)
	mockInfo     func() (keys.Info, error)
	mockValidate func() error
)

type mockSigner struct {
	sign     mockSign
	info     mockInfo
	validate mockValidate
}

func (m *mockSigner) Sign(cfg SignCfg) (*std.Tx, error) {
	if m.sign != nil {
		return m.sign(cfg)
	}
	return nil, nil
}

func (m *mockSigner) Info() (keys.Info, error) {
	if m.info != nil {
		return m.info()
	}
	return nil, nil
}

func (m *mockSigner) Validate() error {
	if m.validate != nil {
		return m.validate()
	}
	return nil
}

// Keys Info mock
type (
	mockGetAddress func() crypto.Address
	mockGetType    func() keys.KeyType
	mockGetName    func() string
	mockGetPubKey  func() crypto.PubKey
	mockGetPath    func() (*hd.BIP44Params, error)
)

type mockKeysInfo struct {
	getAddress mockGetAddress
	getType    mockGetType
	getName    mockGetName
	getPubKey  mockGetPubKey
	getPath    mockGetPath
}

func (m *mockKeysInfo) GetAddress() crypto.Address {
	if m.getAddress != nil {
		return m.getAddress()
	}
	return crypto.Address{}
}

func (m *mockKeysInfo) GetType() keys.KeyType {
	if m.getType != nil {
		return m.getType()
	}
	return 0
}

func (m *mockKeysInfo) GetName() string {
	if m.getName != nil {
		return m.getName()
	}
	return ""
}

func (m *mockKeysInfo) GetPubKey() crypto.PubKey {
	if m.getPubKey != nil {
		return m.getPubKey()
	}
	return nil
}

func (m *mockKeysInfo) GetPath() (*hd.BIP44Params, error) {
	if m.getPath != nil {
		return m.getPath()
	}
	return nil, nil
}

// RPC Client mock
type (
	mockBroadcastTxCommit    func(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTxCommit, error)
	mockABCIQuery            func(ctx context.Context, path string, data []byte) (*abci.ResultABCIQuery, error)
	mockABCIInfo             func(ctx context.Context) (*abci.ResultABCIInfo, error)
	mockABCIQueryWithOptions func(ctx context.Context, path string, data []byte, opts client.ABCIQueryOptions) (*abci.ResultABCIQuery, error)
	mockBroadcastTxAsync     func(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTx, error)
	mockBroadcastTxSync      func(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTx, error)
	mockGenesis              func(ctx context.Context) (*net.ResultGenesis, error)
	mockBlockchainInfo       func(ctx context.Context, minHeight, maxHeight int64) (*blocks.ResultBlockchainInfo, error)
	mockNetInfo              func(ctx context.Context) (*net.ResultNetInfo, error)
	mockDumpConsensusState   func(ctx context.Context) (*consensus.ResultDumpConsensusState, error)
	mockConsensusState       func(ctx context.Context) (*consensus.ResultConsensusState, error)
	mockConsensusParams      func(ctx context.Context, height *int64) (*consensus.ResultConsensusParams, error)
	mockHealth               func(ctx context.Context) (*health.ResultHealth, error)
	mockBlock                func(ctx context.Context, height *int64) (*blocks.ResultBlock, error)
	mockBlockResults         func(ctx context.Context, height *int64) (*blocks.ResultBlockResults, error)
	mockCommit               func(ctx context.Context, height *int64) (*blocks.ResultCommit, error)
	mockValidators           func(ctx context.Context, height *int64) (*consensus.ResultValidators, error)
	mockStatus               func(ctx context.Context, heightGte *int64) (*status.ResultStatus, error)
	mockUnconfirmedTxs       func(ctx context.Context, limit int) (*mempool.ResultUnconfirmedTxs, error)
	mockNumUnconfirmedTxs    func(ctx context.Context) (*mempool.ResultUnconfirmedTxs, error)
	mockTx                   func(ctx context.Context, hash []byte) (*tx.ResultTx, error)
)

type mockRPCClient struct {
	broadcastTxCommit    mockBroadcastTxCommit
	abciQuery            mockABCIQuery
	abciInfo             mockABCIInfo
	abciQueryWithOptions mockABCIQueryWithOptions
	broadcastTxAsync     mockBroadcastTxAsync
	broadcastTxSync      mockBroadcastTxSync
	genesis              mockGenesis
	blockchainInfo       mockBlockchainInfo
	netInfo              mockNetInfo
	dumpConsensusState   mockDumpConsensusState
	consensusState       mockConsensusState
	consensusParams      mockConsensusParams
	health               mockHealth
	block                mockBlock
	blockResults         mockBlockResults
	commit               mockCommit
	validators           mockValidators
	status               mockStatus
	unconfirmedTxs       mockUnconfirmedTxs
	numUnconfirmedTxs    mockNumUnconfirmedTxs
	tx                   mockTx
}

func (m *mockRPCClient) BroadcastTxCommit(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTxCommit, error) {
	if m.broadcastTxCommit != nil {
		return m.broadcastTxCommit(ctx, tx)
	}
	return nil, nil
}

func (m *mockRPCClient) ABCIQuery(ctx context.Context, path string, data []byte) (*abci.ResultABCIQuery, error) {
	if m.abciQuery != nil {
		return m.abciQuery(ctx, path, data)
	}
	return nil, nil
}

func (m *mockRPCClient) ABCIInfo(ctx context.Context) (*abci.ResultABCIInfo, error) {
	if m.abciInfo != nil {
		return m.abciInfo(ctx)
	}
	return nil, nil
}

func (m *mockRPCClient) ABCIQueryWithOptions(ctx context.Context, path string, data []byte, opts client.ABCIQueryOptions) (*abci.ResultABCIQuery, error) {
	if m.abciQueryWithOptions != nil {
		return m.abciQueryWithOptions(ctx, path, data, opts)
	}
	return nil, nil
}

func (m *mockRPCClient) BroadcastTxAsync(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTx, error) {
	if m.broadcastTxAsync != nil {
		return m.broadcastTxAsync(ctx, tx)
	}
	return nil, nil
}

func (m *mockRPCClient) BroadcastTxSync(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTx, error) {
	if m.broadcastTxSync != nil {
		return m.broadcastTxSync(ctx, tx)
	}
	return nil, nil
}

func (m *mockRPCClient) Genesis(ctx context.Context) (*net.ResultGenesis, error) {
	if m.genesis != nil {
		return m.genesis(ctx)
	}
	return nil, nil
}

func (m *mockRPCClient) BlockchainInfo(ctx context.Context, minHeight, maxHeight int64) (*blocks.ResultBlockchainInfo, error) {
	if m.blockchainInfo != nil {
		return m.blockchainInfo(ctx, minHeight, maxHeight)
	}
	return nil, nil
}

func (m *mockRPCClient) NetInfo(ctx context.Context) (*net.ResultNetInfo, error) {
	if m.netInfo != nil {
		return m.netInfo(ctx)
	}
	return nil, nil
}

func (m *mockRPCClient) DumpConsensusState(ctx context.Context) (*consensus.ResultDumpConsensusState, error) {
	if m.dumpConsensusState != nil {
		return m.dumpConsensusState(ctx)
	}
	return nil, nil
}

func (m *mockRPCClient) ConsensusState(ctx context.Context) (*consensus.ResultConsensusState, error) {
	if m.consensusState != nil {
		return m.consensusState(ctx)
	}
	return nil, nil
}

func (m *mockRPCClient) ConsensusParams(ctx context.Context, height *int64) (*consensus.ResultConsensusParams, error) {
	if m.consensusParams != nil {
		return m.consensusParams(ctx, height)
	}
	return nil, nil
}

func (m *mockRPCClient) Health(ctx context.Context) (*health.ResultHealth, error) {
	if m.health != nil {
		return m.health(ctx)
	}
	return nil, nil
}

func (m *mockRPCClient) Block(ctx context.Context, height *int64) (*blocks.ResultBlock, error) {
	if m.block != nil {
		return m.block(ctx, height)
	}
	return nil, nil
}

func (m *mockRPCClient) BlockResults(ctx context.Context, height *int64) (*blocks.ResultBlockResults, error) {
	if m.blockResults != nil {
		return m.blockResults(ctx, height)
	}
	return nil, nil
}

func (m *mockRPCClient) Commit(ctx context.Context, height *int64) (*blocks.ResultCommit, error) {
	if m.commit != nil {
		return m.commit(ctx, height)
	}
	return nil, nil
}

func (m *mockRPCClient) Validators(ctx context.Context, height *int64) (*consensus.ResultValidators, error) {
	if m.validators != nil {
		return m.validators(ctx, height)
	}
	return nil, nil
}

func (m *mockRPCClient) Status(ctx context.Context, heightGte *int64) (*status.ResultStatus, error) {
	if m.status != nil {
		return m.status(ctx, heightGte)
	}
	return nil, nil
}

func (m *mockRPCClient) UnconfirmedTxs(ctx context.Context, limit int) (*mempool.ResultUnconfirmedTxs, error) {
	if m.unconfirmedTxs != nil {
		return m.unconfirmedTxs(ctx, limit)
	}
	return nil, nil
}

func (m *mockRPCClient) NumUnconfirmedTxs(ctx context.Context) (*mempool.ResultUnconfirmedTxs, error) {
	if m.numUnconfirmedTxs != nil {
		return m.numUnconfirmedTxs(ctx)
	}
	return nil, nil
}

func (m *mockRPCClient) Tx(ctx context.Context, hash []byte) (*tx.ResultTx, error) {
	if m.tx != nil {
		return m.tx(ctx, hash)
	}

	return nil, nil
}
