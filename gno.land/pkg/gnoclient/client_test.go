package gnoclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestClient_Render(t *testing.T) {
	t.Parallel()
	testRealmPath := "gno.land/r/demo/deep/very/deep"
	expectedRender := []byte("it works!")

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				return &std.Tx{}, nil
			},
			info: func() keys.Info {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
						return adr
					},
				}
			},
		},
		RPCClient: &mockRPCClient{
			abciQuery: func(path string, data []byte) (*ctypes.ResultABCIQuery, error) {
				res := &ctypes.ResultABCIQuery{
					Response: abci.ResponseQuery{
						ResponseBase: abci.ResponseBase{
							Data: expectedRender,
						},
					},
				}
				return res, nil
			},
		},
	}

	res, data, err := client.Render(testRealmPath, "")
	assert.NoError(t, err)
	assert.NotEmpty(t, data.Response.Data)
	assert.NotEmpty(t, res)
	assert.Equal(t, data.Response.Data, expectedRender)
}

func TestClient_CallSingle(t *testing.T) {
	t.Parallel()

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				return &std.Tx{}, nil
			},
			info: func() keys.Info {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
						return adr
					},
				}
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte("it works!"),
						},
					},
				}
				return res, nil
			},
		},
	}

	cfg := BaseTxCfg{
		GasWanted:      100000,
		GasFee:         "10000ugnot",
		AccountNumber:  1,
		SequenceNumber: 1,
		Memo:           "Test memo",
	}

	msg := []MsgCall{
		{
			PkgPath:  "gno.land/r/demo/deep/very/deep",
			FuncName: "Render",
			Args:     []string{""},
			Send:     "100ugnot",
		},
	}

	res, err := client.Call(cfg, msg...)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), "it works!")
}

func TestClient_CallMultiple(t *testing.T) {
	t.Parallel()

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				return &std.Tx{}, nil
			},
			info: func() keys.Info {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
						return adr
					},
				}
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					CheckTx: abci.ResponseCheckTx{
						ResponseBase: abci.ResponseBase{
							Error:  nil,
							Data:   nil,
							Events: nil,
							Log:    "",
							Info:   "",
						},
					},
				}

				return res, nil
			},
		},
	}

	cfg := BaseTxCfg{
		GasWanted:      100000,
		GasFee:         "10000ugnot",
		AccountNumber:  1,
		SequenceNumber: 1,
		Memo:           "Test memo",
	}

	msg := []MsgCall{
		{
			PkgPath:  "gno.land/r/demo/deep/very/deep",
			FuncName: "Render",
			Args:     []string{""},
			Send:     "100ugnot",
		},
		{
			PkgPath:  "gno.land/r/demo/wugnot",
			FuncName: "Deposit",
			Args:     []string{""},
			Send:     "1000ugnot",
		},
		{
			PkgPath:  "gno.land/r/demo/tamagotchi",
			FuncName: "Feed",
			Args:     []string{""},
			Send:     "",
		},
	}

	res, err := client.Call(cfg, msg...)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestClient_Call_Errors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		cfg           BaseTxCfg
		msgs          []MsgCall
		expectedError error
	}{
		{
			name: "Invalid Signer",
			client: Client{
				Signer:    nil,
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []MsgCall{
				{
					PkgPath:  "random/path",
					FuncName: "RandomName",
					Send:     "",
					Args:     []string{},
				},
			},
			expectedError: ErrMissingSigner,
		},
		{
			name: "Invalid RPCClient",
			client: Client{
				&mockSigner{},
				nil,
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []MsgCall{
				{
					PkgPath:  "random/path",
					FuncName: "RandomName",
					Send:     "",
					Args:     []string{},
				},
			},
			expectedError: ErrMissingRPCClient,
		},
		{
			name: "Invalid Gas Fee",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         "",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []MsgCall{
				{
					PkgPath:  "random/path",
					FuncName: "RandomName",
				},
			},
			expectedError: ErrInvalidGasFee,
		},
		{
			name: "Negative Gas Wanted",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      -1,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []MsgCall{
				{
					PkgPath:  "random/path",
					FuncName: "RandomName",
					Send:     "",
					Args:     []string{},
				},
			},
			expectedError: ErrInvalidGasWanted,
		},
		{
			name: "0 Gas Wanted",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      0,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []MsgCall{
				{
					PkgPath:  "random/path",
					FuncName: "RandomName",
					Send:     "",
					Args:     []string{},
				},
			},
			expectedError: ErrInvalidGasWanted,
		},
		{
			name: "Invalid PkgPath",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []MsgCall{
				{
					PkgPath:  "",
					FuncName: "RandomName",
					Send:     "",
					Args:     []string{},
				},
			},
			expectedError: ErrEmptyPkgPath,
		},
		{
			name: "Invalid FuncName",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []MsgCall{
				{
					PkgPath:  "random/path",
					FuncName: "",
					Send:     "",
					Args:     []string{},
				},
			},
			expectedError: ErrEmptyFuncName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.Call(tc.cfg, tc.msgs...)
			assert.Nil(t, res)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}
