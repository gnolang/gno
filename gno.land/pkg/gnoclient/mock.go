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

// Signer
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

type mockKeysInfo struct{}

func (m mockKeysInfo) GetAddress() crypto.Address {
	adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	return adr
}

func (m mockKeysInfo) GetType() keys.KeyType {
	return 0
}

func (m mockKeysInfo) GetName() string {
	return "mockKeyInfoName"
}

func (m mockKeysInfo) GetPubKey() crypto.PubKey {
	pubkey, _ := crypto.PubKeyFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	return pubkey
}

func (m mockKeysInfo) GetPath() (*hd.BIP44Params, error) {
	return nil, nil
}

// RPC Client
type (
	mockBroadcastTxCommit func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error)
	mockABCIQuery         func(path string, data []byte) (*ctypes.ResultABCIQuery, error)
)

type mockRPCClient struct {
	broadcastTxCommit mockBroadcastTxCommit
	abciQuery         mockABCIQuery
}

func (m mockRPCClient) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	if m.broadcastTxCommit != nil {
		return m.broadcastTxCommit(tx)
	}

	return nil, nil
}

func (m mockRPCClient) ABCIQuery(path string, data []byte) (*ctypes.ResultABCIQuery, error) {
	if m.abciQuery != nil {
		return m.abciQuery(path, data)
	}
	return nil, nil
}

// Unused RPC Client functions

func (m mockRPCClient) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	panic("implement me")
}

func (m mockRPCClient) ABCIQueryWithOptions(path string, data []byte, opts client.ABCIQueryOptions) (*ctypes.ResultABCIQuery, error) {
	panic("implement me")
}

func (m mockRPCClient) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	panic("implement me")
}

func (m mockRPCClient) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	panic("implement me")
}

func (m mockRPCClient) Genesis() (*ctypes.ResultGenesis, error) {
	panic("implement me")
}

func (m mockRPCClient) BlockchainInfo(minHeight, maxHeight int64) (*ctypes.ResultBlockchainInfo, error) {
	panic("implement me")
}

func (m mockRPCClient) NetInfo() (*ctypes.ResultNetInfo, error) {
	panic("implement me")
}

func (m mockRPCClient) DumpConsensusState() (*ctypes.ResultDumpConsensusState, error) {
	panic("implement me")
}

func (m mockRPCClient) ConsensusState() (*ctypes.ResultConsensusState, error) {
	panic("implement me")
}

func (m mockRPCClient) ConsensusParams(height *int64) (*ctypes.ResultConsensusParams, error) {
	panic("implement me")
}

func (m mockRPCClient) Health() (*ctypes.ResultHealth, error) {
	panic("implement me")
}

func (m mockRPCClient) Block(height *int64) (*ctypes.ResultBlock, error) {
	panic("implement me")
}

func (m mockRPCClient) BlockResults(height *int64) (*ctypes.ResultBlockResults, error) {
	panic("implement me")
}

func (m mockRPCClient) Commit(height *int64) (*ctypes.ResultCommit, error) {
	panic("implement me")
}

func (m mockRPCClient) Validators(height *int64) (*ctypes.ResultValidators, error) {
	panic("implement me")
}

func (m mockRPCClient) Status() (*ctypes.ResultStatus, error) {
	panic("implement me")
}

func (m mockRPCClient) UnconfirmedTxs(limit int) (*ctypes.ResultUnconfirmedTxs, error) {
	panic("implement me")
}

func (m mockRPCClient) NumUnconfirmedTxs() (*ctypes.ResultUnconfirmedTxs, error) {
	panic("implement me")
}
