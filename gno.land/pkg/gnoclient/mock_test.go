package gnoclient

import (
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Signer mock
type (
	mockSign     func(cfg SignCfg) (*std.Tx, error)
	mockInfo     func() keys.Info
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

func (m *mockSigner) Info() keys.Info {
	if m.info != nil {
		return m.info()
	}
	return nil
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
	mockBroadcastTxCommit    func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error)
	mockABCIQuery            func(path string, data []byte) (*ctypes.ResultABCIQuery, error)
	mockABCIInfo             func() (*ctypes.ResultABCIInfo, error)
	mockABCIQueryWithOptions func(path string, data []byte, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error)
	mockBroadcastTxAsync     func(tx types.Tx) (*ctypes.ResultBroadcastTx, error)
	mockBroadcastTxSync      func(tx types.Tx) (*ctypes.ResultBroadcastTx, error)
	mockGenesis              func() (*ctypes.ResultGenesis, error)
	mockBlockchainInfo       func(minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error)
	mockNetInfo              func() (*ctypes.ResultNetInfo, error)
	mockDumpConsensusState   func() (*ctypes.ResultDumpConsensusState, error)
	mockConsensusState       func() (*ctypes.ResultConsensusState, error)
	mockConsensusParams      func(height *int64) (*ctypes.ResultConsensusParams, error)
	mockHealth               func() (*ctypes.ResultHealth, error)
	mockBlock                func(height *int64) (*ctypes.ResultBlock, error)
	mockBlockResults         func(height *int64) (*ctypes.ResultBlockResults, error)
	mockCommit               func(height *int64) (*ctypes.ResultCommit, error)
	mockValidators           func(height *int64) (*ctypes.ResultValidators, error)
	mockStatus               func() (*ctypes.ResultStatus, error)
	mockUnconfirmedTxs       func(limit int) (*ctypes.ResultUnconfirmedTxs, error)
	mockNumUnconfirmedTxs    func() (*ctypes.ResultUnconfirmedTxs, error)
	mockTx                   func(hash []byte) (*ctypes.ResultTx, error)
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

func (m *mockRPCClient) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	if m.broadcastTxCommit != nil {
		return m.broadcastTxCommit(tx)
	}
	return nil, nil
}

func (m *mockRPCClient) ABCIQuery(path string, data []byte) (*ctypes.ResultABCIQuery, error) {
	if m.abciQuery != nil {
		return m.abciQuery(path, data)
	}
	return nil, nil
}

func (m *mockRPCClient) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	if m.abciInfo != nil {
		return m.ABCIInfo()
	}
	return nil, nil
}

func (m *mockRPCClient) ABCIQueryWithOptions(path string, data []byte, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	if m.abciQueryWithOptions != nil {
		return m.abciQueryWithOptions(path, data, opts)
	}
	return nil, nil
}

func (m *mockRPCClient) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	if m.broadcastTxAsync != nil {
		return m.broadcastTxAsync(tx)
	}
	return nil, nil
}

func (m *mockRPCClient) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	if m.broadcastTxSync != nil {
		return m.broadcastTxSync(tx)
	}
	return nil, nil
}

func (m *mockRPCClient) Genesis() (*ctypes.ResultGenesis, error) {
	if m.genesis != nil {
		return m.genesis()
	}
	return nil, nil
}

func (m *mockRPCClient) BlockchainInfo(minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error) {
	if m.blockchainInfo != nil {
		return m.blockchainInfo(minHeight, maxHeight)
	}
	return nil, nil
}

func (m *mockRPCClient) NetInfo() (*ctypes.ResultNetInfo, error) {
	if m.netInfo != nil {
		return m.netInfo()
	}
	return nil, nil
}

func (m *mockRPCClient) DumpConsensusState() (*ctypes.ResultDumpConsensusState, error) {
	if m.dumpConsensusState != nil {
		return m.dumpConsensusState()
	}
	return nil, nil
}

func (m *mockRPCClient) ConsensusState() (*ctypes.ResultConsensusState, error) {
	if m.consensusState != nil {
		return m.consensusState()
	}
	return nil, nil
}

func (m *mockRPCClient) ConsensusParams(height *int64) (*ctypes.ResultConsensusParams, error) {
	if m.consensusParams != nil {
		return m.consensusParams(height)
	}
	return nil, nil
}

func (m *mockRPCClient) Health() (*ctypes.ResultHealth, error) {
	if m.health != nil {
		return m.health()
	}
	return nil, nil
}

func (m *mockRPCClient) Block(height *int64) (*ctypes.ResultBlock, error) {
	if m.block != nil {
		return m.block(height)
	}
	return nil, nil
}

func (m *mockRPCClient) BlockResults(height *int64) (*ctypes.ResultBlockResults, error) {
	if m.blockResults != nil {
		return m.blockResults(height)
	}
	return nil, nil
}

func (m *mockRPCClient) Commit(height *int64) (*ctypes.ResultCommit, error) {
	if m.commit != nil {
		return m.commit(height)
	}
	return nil, nil
}

func (m *mockRPCClient) Validators(height *int64) (*ctypes.ResultValidators, error) {
	if m.validators != nil {
		return m.validators(height)
	}
	return nil, nil
}

func (m *mockRPCClient) Status() (*ctypes.ResultStatus, error) {
	if m.status != nil {
		return m.status()
	}
	return nil, nil
}

func (m *mockRPCClient) UnconfirmedTxs(limit int) (*ctypes.ResultUnconfirmedTxs, error) {
	if m.unconfirmedTxs != nil {
		return m.unconfirmedTxs(limit)
	}
	return nil, nil
}

func (m *mockRPCClient) NumUnconfirmedTxs() (*ctypes.ResultUnconfirmedTxs, error) {
	if m.numUnconfirmedTxs != nil {
		return m.numUnconfirmedTxs()
	}
	return nil, nil
}

func (m *mockRPCClient) Tx(hash []byte) (*ctypes.ResultTx, error) {
	if m.tx != nil {
		return m.tx(hash)
	}

	return nil, nil
}
