package gnoclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var (
	addr1 = crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	addr2 = crypto.MustAddressFromString("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
)

func TestRender(t *testing.T) {
	t.Parallel()
	testRealmPath := "gno.land/r/demo/deep/very/deep"

	expectedRender := []byte("hi gnoclient!\n")

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				return &std.Tx{}, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
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

// Call tests
func TestCallSingle(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				return &std.Tx{}, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
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

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg := []vm.MsgCall{
		{
			Caller:  caller.GetAddress(),
			PkgPath: "gno.land/r/demo/deep/very/deep",
			Func:    "Render",
			Args:    []string{""},
			Send:    std.Coins{{Denom: "ugnot", Amount: 100}},
		},
	}

	res, err := client.Call(cfg, msg...)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))

	res, err = callSigningSeparately(t, client, cfg, msg...)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestCallSingle_Sponsor(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				cfg.Tx.Signatures = make([]std.Signature, 2)
				return &cfg.Tx, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
						},
					},
				}
				return res, nil
			},
		},
	}

	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasWanted:      100000,
			GasFee:         "10000ugnot",
			AccountNumber:  1,
			SequenceNumber: 1,
			Memo:           "Test memo",
		},
		SponsorAddress: addr2,
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg := vm.MsgCall{
		Caller:  caller.GetAddress(),
		PkgPath: "gno.land/r/demo/deep/very/deep",
		Func:    "Render",
		Args:    []string{""},
		Send:    std.Coins{{Denom: "ugnot", Amount: 100}},
	}

	tx, err := client.NewSponsorTransaction(cfg, msg)
	assert.NoError(t, err)

	presignedTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*presignedTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestCallMultiple(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				return &std.Tx{}, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
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

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg := []vm.MsgCall{
		{
			Caller:  caller.GetAddress(),
			PkgPath: "gno.land/r/demo/deep/very/deep",
			Func:    "Render",
			Args:    []string{""},
			Send:    std.Coins{{Denom: "ugnot", Amount: 100}},
		},
		{
			Caller:  caller.GetAddress(),
			PkgPath: "gno.land/r/demo/wugnot",
			Func:    "Deposit",
			Args:    []string{""},
			Send:    std.Coins{{Denom: "ugnot", Amount: 100}},
		},
		{
			Caller:  caller.GetAddress(),
			PkgPath: "gno.land/r/demo/tamagotchi",
			Func:    "Feed",
			Args:    []string{""},
			Send:    nil,
		},
	}

	res, err := client.Call(cfg, msg...)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))

	res, err = callSigningSeparately(t, client, cfg, msg...)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestCallMultiple_Sponsor(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				cfg.Tx.Signatures = make([]std.Signature, 2)
				return &cfg.Tx, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
						},
					},
				}
				return res, nil
			},
		},
	}

	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasWanted:      100000,
			GasFee:         "10000ugnot",
			AccountNumber:  1,
			SequenceNumber: 1,
			Memo:           "Test memo",
		},
		SponsorAddress: addr2,
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg1 := vm.MsgCall{
		Caller:  caller.GetAddress(),
		PkgPath: "gno.land/r/demo/deep/very/deep",
		Func:    "Render",
		Args:    []string{""},
		Send:    std.Coins{{Denom: "ugnot", Amount: 100}},
	}

	msg2 := vm.MsgCall{
		Caller:  caller.GetAddress(),
		PkgPath: "gno.land/r/demo/wugnot",
		Func:    "Deposit",
		Args:    []string{""},
		Send:    std.Coins{{Denom: "ugnot", Amount: 100}},
	}

	msg3 := vm.MsgCall{
		Caller:  caller.GetAddress(),
		PkgPath: "gno.land/r/demo/tamagotchi",
		Func:    "Feed",
		Args:    []string{""},
		Send:    nil,
	}

	tx, err := client.NewSponsorTransaction(cfg, msg1, msg2, msg3)
	assert.NoError(t, err)

	presignedTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*presignedTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestCallErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		cfg           BaseTxCfg
		msgs          []vm.MsgCall
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
			msgs: []vm.MsgCall{
				{
					Caller:  addr1,
					PkgPath: "random/path",
					Func:    "RandomName",
					Send:    nil,
					Args:    []string{},
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
			msgs: []vm.MsgCall{
				{
					Caller:  addr1,
					PkgPath: "random/path",
					Func:    "RandomName",
					Send:    nil,
					Args:    []string{},
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
			msgs: []vm.MsgCall{
				{
					PkgPath: "random/path",
					Func:    "RandomName",
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
			msgs: []vm.MsgCall{
				{
					Caller:  addr1,
					PkgPath: "random/path",
					Func:    "RandomName",
					Send:    nil,
					Args:    []string{},
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
			msgs: []vm.MsgCall{
				{
					Caller:  addr1,
					PkgPath: "random/path",
					Func:    "RandomName",
					Send:    nil,
					Args:    []string{},
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
			msgs: []vm.MsgCall{
				{
					Caller:  addr1,
					PkgPath: "",
					Func:    "RandomName",
					Send:    nil,
					Args:    []string{},
				},
			},
			expectedError: vm.InvalidPkgPathError{},
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
			msgs: []vm.MsgCall{
				{
					Caller:  addr1,
					PkgPath: "gno.land/r/random/path",
					Func:    "",
					Send:    nil,
					Args:    []string{},
				},
			},
			expectedError: vm.InvalidExprError{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.Call(tc.cfg, tc.msgs...)
			assert.Nil(t, res)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

// Send tests
func TestSendSingle(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				return &std.Tx{}, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
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

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	receiver, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")

	msg := []bank.MsgSend{
		{
			FromAddress: caller.GetAddress(),
			ToAddress:   receiver,
			Amount:      std.Coins{{Denom: "ugnot", Amount: 100}},
		},
	}

	res, err := client.Send(cfg, msg...)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))

	res, err = sendSigningSeparately(t, client, cfg, msg...)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestSendSingle_Sponsor(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				cfg.Tx.Signatures = make([]std.Signature, 2)
				return &cfg.Tx, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
						},
					},
				}
				return res, nil
			},
		},
	}

	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasWanted:      100000,
			GasFee:         "10000ugnot",
			AccountNumber:  1,
			SequenceNumber: 1,
			Memo:           "Test memo",
		},
		SponsorAddress: addr2,
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	receiver, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")

	msg := bank.MsgSend{
		FromAddress: caller.GetAddress(),
		ToAddress:   receiver,
		Amount:      std.Coins{{Denom: "ugnot", Amount: 100}},
	}

	tx, err := client.NewSponsorTransaction(cfg, msg)
	assert.NoError(t, err)

	presignedTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*presignedTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestSendMultiple(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				return &std.Tx{}, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
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

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg1 := bank.MsgSend{
		FromAddress: caller.GetAddress(),
		ToAddress:   addr2,
		Amount:      std.Coins{{Denom: "ugnot", Amount: 100}},
	}

	msg2 := bank.MsgSend{
		FromAddress: caller.GetAddress(),
		ToAddress:   addr2,
		Amount:      std.Coins{{Denom: "ugnot", Amount: 200}},
	}

	res, err := client.Send(cfg, msg1, msg2)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))

	res, err = sendSigningSeparately(t, client, cfg, msg1, msg2)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestSendMultiple_Sponsor(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				cfg.Tx.Signatures = make([]std.Signature, 2)
				return &cfg.Tx, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
						},
					},
				}
				return res, nil
			},
		},
	}

	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasWanted:      100000,
			GasFee:         "10000ugnot",
			AccountNumber:  1,
			SequenceNumber: 1,
			Memo:           "Test memo",
		},
		SponsorAddress: addr2,
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	receiver, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")

	msg1 := bank.MsgSend{
		FromAddress: caller.GetAddress(),
		ToAddress:   receiver,
		Amount:      std.Coins{{Denom: "ugnot", Amount: 100}},
	}

	msg2 := bank.MsgSend{
		FromAddress: caller.GetAddress(),
		ToAddress:   receiver,
		Amount:      std.Coins{{Denom: "ugnot", Amount: 200}},
	}

	tx, err := client.NewSponsorTransaction(cfg, msg1, msg2)
	assert.NoError(t, err)

	presignedTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*presignedTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestSendErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		cfg           BaseTxCfg
		msgs          []bank.MsgSend
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
			msgs: []bank.MsgSend{
				{
					FromAddress: addr1,
					ToAddress:   addr2,
					Amount:      std.Coins{{Denom: "ugnot", Amount: 1}},
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
			msgs: []bank.MsgSend{
				{
					FromAddress: addr1,
					ToAddress:   addr2,
					Amount:      std.Coins{{Denom: "ugnot", Amount: 1}},
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
			msgs: []bank.MsgSend{
				{
					FromAddress: addr1,
					ToAddress:   addr2,
					Amount:      std.Coins{{Denom: "ugnot", Amount: 1}},
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
			msgs: []bank.MsgSend{
				{
					FromAddress: addr1,
					ToAddress:   addr2,
					Amount:      std.Coins{{Denom: "ugnot", Amount: 1}},
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
			msgs: []bank.MsgSend{
				{
					FromAddress: addr1,
					ToAddress:   addr2,
					Amount:      std.Coins{{Denom: "ugnot", Amount: 1}},
				},
			},
			expectedError: ErrInvalidGasWanted,
		},
		{
			name: "Invalid To Address",
			client: Client{
				Signer: &mockSigner{
					info: func() (keys.Info, error) {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								return addr1
							},
						}, nil
					},
				},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []bank.MsgSend{
				{
					FromAddress: addr1,
					ToAddress:   crypto.Address{},
					Amount:      std.Coins{{Denom: "ugnot", Amount: 1}},
				},
			},
			expectedError: std.InvalidAddressError{},
		},
		{
			name: "Invalid Send Coins",
			client: Client{
				Signer: &mockSigner{
					info: func() (keys.Info, error) {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								return addr1
							},
						}, nil
					},
				},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []bank.MsgSend{
				{
					FromAddress: addr1,
					ToAddress:   addr2,
					Amount:      std.Coins{{Denom: "ugnot", Amount: -1}},
				},
			},
			expectedError: std.InvalidCoinsError{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.Send(tc.cfg, tc.msgs...)
			assert.Nil(t, res)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

// Run tests
func TestRunSingle(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				return &std.Tx{}, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
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

	fileBody := `package main
import (
	"std"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/deep/very/deep"
)
func main() {
	println(ufmt.Sprintf("%s", deep.Render("gnoclient!")))
}`

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg := vm.MsgRun{
		Caller: caller.GetAddress(),
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody,
				},
			},
		},
		Send: nil,
	}

	res, err := client.Run(cfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))

	res, err = runSigningSeparately(t, client, cfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestRunSingle_Sponsor(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				cfg.Tx.Signatures = make([]std.Signature, 2)
				return &cfg.Tx, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte("hi gnoclient!\n"),
						},
					},
				}
				return res, nil
			},
		},
	}

	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasWanted:      100000,
			GasFee:         "10000ugnot",
			AccountNumber:  1,
			SequenceNumber: 1,
			Memo:           "Test memo",
		},
		SponsorAddress: addr2,
	}

	fileBody := `package main
import (
	"std"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/deep/very/deep"
)
func main() {
	println(ufmt.Sprintf("%s", deep.Render("gnoclient!")))
}`

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg := vm.MsgRun{
		Caller: caller.GetAddress(),
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody,
				},
			},
		},
		Send: nil,
	}

	tx, err := client.NewSponsorTransaction(cfg, msg)
	assert.NoError(t, err)

	presignedTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*presignedTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestRunMultiple(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				return &std.Tx{}, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
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

	fileBody := `package main
import (
	"std"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/deep/very/deep"
)
func main() {
	println(ufmt.Sprintf("%s", deep.Render("gnoclient!")))
}`

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg1 := vm.MsgRun{
		Caller: caller.GetAddress(),
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main1.gno",
					Body: fileBody,
				},
			},
		},
		Send: nil,
	}

	msg2 := vm.MsgRun{
		Caller: caller.GetAddress(),
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main2.gno",
					Body: fileBody,
				},
			},
		},
		Send: nil,
	}

	res, err := client.Run(cfg, msg1, msg2)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))

	res, err = runSigningSeparately(t, client, cfg, msg1, msg2)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestRunMultiple_Sponsor(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				cfg.Tx.Signatures = make([]std.Signature, 2)
				return &cfg.Tx, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
						},
					},
				}
				return res, nil
			},
		},
	}

	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasWanted:      100000,
			GasFee:         "10000ugnot",
			AccountNumber:  1,
			SequenceNumber: 1,
			Memo:           "Test memo",
		},
		SponsorAddress: addr2,
	}

	fileBody := `package main
import (
	"std"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/deep/very/deep"
)
func main() {
	println(ufmt.Sprintf("%s", deep.Render("gnoclient!")))
}`

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg1 := vm.MsgRun{
		Caller: caller.GetAddress(),
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main1.gno",
					Body: fileBody,
				},
			},
		},
		Send: nil,
	}

	msg2 := vm.MsgRun{
		Caller: caller.GetAddress(),
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main2.gno",
					Body: fileBody,
				},
			},
		},
		Send: nil,
	}

	tx, err := client.NewSponsorTransaction(cfg, msg1, msg2)
	assert.NoError(t, err)

	presignedTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*presignedTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestRunErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		cfg           BaseTxCfg
		msgs          []vm.MsgRun
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
			msgs: []vm.MsgRun{
				{
					Caller: addr1,
					Package: &std.MemPackage{
						Name: "",
						Path: "",
						Files: []*std.MemFile{
							{
								Name: "file1.gno",
								Body: "",
							},
						},
					},
					Send: nil,
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
			msgs:          []vm.MsgRun{},
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
			msgs: []vm.MsgRun{
				{
					Caller: addr1,
					Package: &std.MemPackage{
						Name: "",
						Path: "",
						Files: []*std.MemFile{
							{
								Name: "file1.gno",
								Body: "",
							},
						},
					},
					Send: nil,
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
			msgs: []vm.MsgRun{
				{
					Caller: addr1,
					Package: &std.MemPackage{
						Name: "",
						Path: "",
						Files: []*std.MemFile{
							{
								Name: "file1.gno",
								Body: "",
							},
						},
					},
					Send: nil,
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
			msgs: []vm.MsgRun{
				{
					Caller: addr1,
					Package: &std.MemPackage{
						Name: "",
						Path: "",
						Files: []*std.MemFile{
							{
								Name: "file1.gno",
								Body: "",
							},
						},
					},
					Send: nil,
				},
			},
			expectedError: ErrInvalidGasWanted,
		},
		{
			name: "Invalid Empty Package",
			client: Client{
				Signer: &mockSigner{
					info: func() (keys.Info, error) {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								return addr1
							},
						}, nil
					},
				},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgRun{
				{
					Caller: addr1,
					Package: &std.MemPackage{
						Name: "",
						Path: " ",
					},
					Send: nil,
				},
			},
			expectedError: vm.InvalidPkgPathError{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.Run(tc.cfg, tc.msgs...)
			assert.Nil(t, res)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

// AddPackage tests
func TestAddPackageSingle(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				return &std.Tx{}, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
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

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg := vm.MsgAddPackage{
		Creator: caller.GetAddress(),
		Package: &std.MemPackage{
			Name: "hello",
			Path: "gno.land/p/demo/hello",
			Files: []*std.MemFile{
				{
					Name: "file1.gno",
					Body: "",
				},
			},
		},
		Deposit: nil,
	}

	res, err := client.AddPackage(cfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))

	res, err = addPackageSigningSeparately(t, client, cfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestAddPackageSingle_Sponsor(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				cfg.Tx.Signatures = make([]std.Signature, 2)
				return &cfg.Tx, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
						},
					},
				}
				return res, nil
			},
		},
	}

	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasWanted:      100000,
			GasFee:         "10000ugnot",
			AccountNumber:  1,
			SequenceNumber: 1,
			Memo:           "Test memo",
		},
		SponsorAddress: addr2,
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg := vm.MsgAddPackage{
		Creator: caller.GetAddress(),
		Package: &std.MemPackage{
			Name: "hello",
			Path: "gno.land/p/demo/hello",
			Files: []*std.MemFile{
				{
					Name: "file1.gno",
					Body: "",
				},
			},
		},
		Deposit: nil,
	}

	tx, err := client.NewSponsorTransaction(cfg, msg)
	assert.NoError(t, err)

	sponsorTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*sponsorTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestAddPackageMultiple(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				return &std.Tx{}, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
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

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msgs := []vm.MsgAddPackage{
		{
			Creator: caller.GetAddress(),
			Package: &std.MemPackage{
				Name: "hello",
				Path: "gno.land/p/demo/hello",
				Files: []*std.MemFile{
					{
						Name: "file1.gno",
						Body: "",
					},
				},
			},
			Deposit: nil,
		},
		{
			Creator: caller.GetAddress(),
			Package: &std.MemPackage{
				Name: "goodbye",
				Path: "gno.land/p/demo/goodbye",
				Files: []*std.MemFile{
					{
						Name: "file1.gno",
						Body: "",
					},
				},
			},
			Deposit: nil,
		},
	}

	res, err := client.AddPackage(cfg, msgs...)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))

	res, err = addPackageSigningSeparately(t, client, cfg, msgs...)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestAddPackageMultiple_Sponsor(t *testing.T) {
	t.Parallel()

	expected := "hi gnoclient!\n"

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				cfg.Tx.Signatures = make([]std.Signature, 2)
				return &cfg.Tx, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						return addr1
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
				res := &ctypes.ResultBroadcastTxCommit{
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Data: []byte(expected),
						},
					},
				}
				return res, nil
			},
		},
	}

	cfg := SponsorTxCfg{
		BaseTxCfg: BaseTxCfg{
			GasWanted:      100000,
			GasFee:         "10000ugnot",
			AccountNumber:  1,
			SequenceNumber: 1,
			Memo:           "Test memo",
		},
		SponsorAddress: addr2,
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg1 := vm.MsgAddPackage{
		Creator: caller.GetAddress(),
		Package: &std.MemPackage{
			Name: "hello",
			Path: "gno.land/p/demo/hello",
			Files: []*std.MemFile{
				{
					Name: "file1.gno",
					Body: "",
				},
			},
		},
		Deposit: nil,
	}

	msg2 := vm.MsgAddPackage{
		Creator: caller.GetAddress(),
		Package: &std.MemPackage{
			Name: "goodbye",
			Path: "gno.land/p/demo/goodbye",
			Files: []*std.MemFile{
				{
					Name: "file1.gno",
					Body: "",
				},
			},
		},
		Deposit: nil,
	}

	tx, err := client.NewSponsorTransaction(cfg, msg1, msg2)
	assert.NoError(t, err)

	sponsorTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*sponsorTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestAddPackageErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		cfg           BaseTxCfg
		msgs          []vm.MsgAddPackage
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
			msgs: []vm.MsgAddPackage{
				{
					Creator: addr1,
					Package: &std.MemPackage{
						Name: "",
						Path: "",
						Files: []*std.MemFile{
							{
								Name: "file1.gno",
								Body: "",
							},
						},
					},
					Deposit: nil,
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
			msgs:          []vm.MsgAddPackage{},
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
			msgs: []vm.MsgAddPackage{
				{
					Creator: addr1,
					Package: &std.MemPackage{
						Name: "",
						Path: "",
						Files: []*std.MemFile{
							{
								Name: "file1.gno",
								Body: "",
							},
						},
					},
					Deposit: nil,
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
			msgs: []vm.MsgAddPackage{
				{
					Creator: addr1,
					Package: &std.MemPackage{
						Name: "",
						Path: "",
						Files: []*std.MemFile{
							{
								Name: "file1.gno",
								Body: "",
							},
						},
					},
					Deposit: nil,
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
			msgs: []vm.MsgAddPackage{
				{
					Creator: addr1,
					Package: &std.MemPackage{
						Name: "",
						Path: "",
						Files: []*std.MemFile{
							{
								Name: "file1.gno",
								Body: "",
							},
						},
					},
					Deposit: nil,
				},
			},
			expectedError: ErrInvalidGasWanted,
		},
		{
			name: "Invalid Empty Package",
			client: Client{
				Signer: &mockSigner{
					info: func() (keys.Info, error) {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								return addr1
							},
						}, nil
					},
				},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         "10000ugnot",
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgAddPackage{
				{
					Creator: addr1,
					Package: &std.MemPackage{
						Name: "",
						Path: "",
					},
					Deposit: nil,
				},
			},
			expectedError: vm.InvalidPkgPathError{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.AddPackage(tc.cfg, tc.msgs...)
			assert.Nil(t, res)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestNewSponsorTransaction(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		cfg           SponsorTxCfg
		msgs          []std.Msg
		expectedError error
	}{
		{
			name: "Invalid Client",
			client: Client{
				Signer:    nil, // invalid signer
				RPCClient: &mockRPCClient{},
			},
			cfg: SponsorTxCfg{
				BaseTxCfg: BaseTxCfg{
					GasWanted:      100000,
					GasFee:         "10000ugnot",
					AccountNumber:  1,
					SequenceNumber: 1,
					Memo:           "Test memo",
				},
				SponsorAddress: addr2,
			},
			expectedError: ErrMissingSigner,
		},
		{
			name: "Invalid SponsorTxCfg",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: SponsorTxCfg{
				BaseTxCfg: BaseTxCfg{
					GasWanted:      -1,
					GasFee:         "10000ugnot",
					AccountNumber:  1,
					SequenceNumber: 1,
					Memo:           "Test memo",
				},
				SponsorAddress: crypto.Address{}, // invalid sponsor address
			},
			expectedError: ErrInvalidSponsorAddress,
		},
		{
			name: "Empty message list",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: SponsorTxCfg{
				BaseTxCfg: BaseTxCfg{
					GasWanted:      100000,
					GasFee:         "10000ugnot",
					AccountNumber:  1,
					SequenceNumber: 1,
					Memo:           "Test memo",
				},
				SponsorAddress: addr2,
			},

			msgs: []std.Msg{}, // no messages provided

			expectedError: ErrNoMessages,
		},
		{
			name: "Signer not found",
			client: Client{
				Signer: &mockSigner{
					info: func() (keys.Info, error) {
						return nil, errors.New("failed to get signer info") // signer not found
					},
				},
				RPCClient: &mockRPCClient{},
			},
			cfg: SponsorTxCfg{
				BaseTxCfg: BaseTxCfg{
					GasWanted:      100000,
					GasFee:         "10000ugnot",
					AccountNumber:  1,
					SequenceNumber: 1,
					Memo:           "Test memo",
				},
				SponsorAddress: addr2,
			},

			msgs: []std.Msg{}, // no messages provided

			expectedError: ErrNoMessages,
		},
		{
			name: "All messages aren't the same type",
			client: Client{
				Signer: &mockSigner{
					info: func() (keys.Info, error) {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								return addr1
							},
						}, nil
					},
				},
				RPCClient: &mockRPCClient{},
			},
			cfg: SponsorTxCfg{
				BaseTxCfg: BaseTxCfg{
					GasWanted:      100000,
					GasFee:         "10000ugnot",
					AccountNumber:  1,
					SequenceNumber: 1,
					Memo:           "Test memo",
				},
				SponsorAddress: addr2,
			},

			// MixedMessage is invalid
			msgs: []std.Msg{
				vm.MsgCall{
					Caller:  addr1,
					PkgPath: "gno.land/r/demo/deep/very/deep",
					Func:    "Render",
					Args:    []string{""},
					Send:    std.Coins{{Denom: "ugnot", Amount: 100}},
				},
				bank.MsgSend{
					FromAddress: addr1,
					ToAddress:   addr2,
					Amount:      std.Coins{{Denom: "ugnot", Amount: 100}},
				},
			},
			expectedError: ErrMixedMessageTypes,
		},
		{
			name: "At least one invalid message",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: SponsorTxCfg{
				BaseTxCfg: BaseTxCfg{
					GasWanted:      100000,
					GasFee:         "10000ugnot",
					AccountNumber:  1,
					SequenceNumber: 1,
					Memo:           "Test memo",
				},
				SponsorAddress: addr2,
			},
			msgs: []std.Msg{
				// invalid message send
				bank.MsgSend{
					FromAddress: addr1,
					ToAddress:   crypto.Address{},
					Amount:      std.Coins{{Denom: "ugnot", Amount: 10000}},
				},
			},
			expectedError: std.InvalidAddressError{},
		},
		{
			name: "Failed to parse gas fee",
			client: Client{
				Signer: &mockSigner{
					info: func() (keys.Info, error) {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								return addr1
							},
						}, nil
					},
				},
				RPCClient: &mockRPCClient{},
			},
			cfg: SponsorTxCfg{
				BaseTxCfg: BaseTxCfg{
					GasWanted:      100000,
					GasFee:         "xxx", // invalid gas fee
					AccountNumber:  1,
					SequenceNumber: 1,
					Memo:           "Test memo",
				},
				SponsorAddress: addr2,
			},
			msgs: []std.Msg{
				vm.MsgCall{
					Caller:  addr1,
					PkgPath: "gno.land/r/demo/deep/very/deep",
					Func:    "Render",
					Args:    []string{""},
					Send:    std.Coins{{Denom: "ugnot", Amount: 100}},
				},
			},
			expectedError: errors.New("invalid coin expression: xxx"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.NewSponsorTransaction(tc.cfg, tc.msgs...)
			assert.Nil(t, res)
			assert.Equal(t, tc.expectedError.Error(), err.Error())
		})
	}
}

func TestSignTx(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		tx            std.Tx
		expectedError error
	}{
		{
			name: "Failed to sign transaction",
			client: Client{
				Signer: &mockSigner{
					info: func() (keys.Info, error) {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								return addr1
							},
						}, nil
					},
					sign: func(cfg SignCfg) (*std.Tx, error) {
						return nil, errors.New("failed to sign transaction")
					},
				},
				RPCClient: &mockRPCClient{
					abciQuery: func(path string, data []byte) (*ctypes.ResultABCIQuery, error) {
						acc := std.NewBaseAccount(addr1, nil, nil, 0, 0)
						accData, _ := amino.MarshalJSON(acc)

						return &ctypes.ResultABCIQuery{
							Response: abci.ResponseQuery{
								ResponseBase: abci.ResponseBase{
									Data: accData,
								},
							},
						}, nil
					},
				},
			},
			tx:            std.Tx{},
			expectedError: errors.New("failed to sign transaction"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.SignTx(tc.tx, 0, 0)
			assert.Nil(t, res)
			assert.Equal(t, tc.expectedError.Error(), err.Error())
		})
	}
}

func TestExecuteSponsorTransaction(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		tx            std.Tx
		expectedError error
	}{
		{
			name: "Invalid Client",
			client: Client{
				Signer:    nil,
				RPCClient: &mockRPCClient{},
			},
			tx:            std.Tx{},
			expectedError: ErrMissingSigner,
		},
		{
			name: "Invalid transaction",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			tx: std.Tx{
				Fee: std.NewFee(1000, std.NewCoin("ugnot", 10)),
				Msgs: []std.Msg{
					vm.MsgCall{
						Caller: addr1,
					},
				},
				Signatures: []std.Signature{}, // no signatures provided
			},
			expectedError: errors.New("no signatures error"),
		},
		{
			name: "tx is not a sponsor transaction",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			tx: std.Tx{
				Fee: std.NewFee(1000, std.NewCoin("ugnot", 10)),
				Msgs: []std.Msg{ // missing noop msg
					bank.MsgSend{
						FromAddress: addr1,
						ToAddress:   addr2,
						Amount:      std.NewCoins(std.NewCoin("gnot", 1000)),
					},
				},
				Signatures: []std.Signature{
					{
						PubKey:    nil,
						Signature: nil,
					},
				},
			},
			expectedError: ErrInvalidSponsorTx,
		},
		{
			name: "signAndBroadcastTxCommit error",
			client: Client{
				Signer: &mockSigner{
					info: func() (keys.Info, error) {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								return addr1
							},
						}, nil
					},
					sign: func(cfg SignCfg) (*std.Tx, error) {
						return nil, errors.New("failed to sign tx") // failed to sign tx
					},
				},
				RPCClient: &mockRPCClient{
					abciQuery: func(path string, data []byte) (*ctypes.ResultABCIQuery, error) {
						acc := std.NewBaseAccount(addr1, std.NewCoins(), nil, 0, 0)
						accData, _ := amino.MarshalJSON(acc)

						return &ctypes.ResultABCIQuery{
							Response: abci.ResponseQuery{
								ResponseBase: abci.ResponseBase{
									Data: accData,
								},
							},
						}, nil
					},
				},
			},
			tx: std.Tx{
				Fee: std.NewFee(1000, std.NewCoin("ugnot", 10)),
				Msgs: []std.Msg{
					vm.MsgNoop{
						Caller: addr2,
					},
					bank.MsgSend{
						FromAddress: addr1,
						ToAddress:   addr2,
						Amount:      std.NewCoins(std.NewCoin("gnot", 1000)),
					},
				},
				Signatures: []std.Signature{
					{
						PubKey:    nil,
						Signature: nil,
					},
					{
						PubKey:    nil,
						Signature: nil,
					},
				},
			},
			expectedError: errors.New("failed to sign tx"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.ExecuteSponsorTransaction(tc.tx, 0, 0)
			assert.Nil(t, res)
			assert.Equal(t, tc.expectedError.Error(), err.Error())
		})
	}
}

// The same as client.Call, but test signing separately
func callSigningSeparately(t *testing.T, client Client, cfg BaseTxCfg, msgs ...vm.MsgCall) (*ctypes.ResultBroadcastTxCommit, error) {
	t.Helper()
	tx, err := NewCallTx(cfg, msgs...)
	assert.NoError(t, err)
	require.NotNil(t, tx)
	signedTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)
	require.NotNil(t, signedTx)
	res, err := client.BroadcastTx(signedTx)
	assert.NoError(t, err)
	require.NotNil(t, res)
	return res, nil
}

// The same as client.Run, but test signing separately
func runSigningSeparately(t *testing.T, client Client, cfg BaseTxCfg, msgs ...vm.MsgRun) (*ctypes.ResultBroadcastTxCommit, error) {
	t.Helper()
	tx, err := NewRunTx(cfg, msgs...)
	assert.NoError(t, err)
	require.NotNil(t, tx)
	signedTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)
	require.NotNil(t, signedTx)
	res, err := client.BroadcastTx(signedTx)
	assert.NoError(t, err)
	require.NotNil(t, res)
	return res, nil
}

// The same as client.Send, but test signing separately
func sendSigningSeparately(t *testing.T, client Client, cfg BaseTxCfg, msgs ...bank.MsgSend) (*ctypes.ResultBroadcastTxCommit, error) {
	t.Helper()
	tx, err := NewSendTx(cfg, msgs...)
	assert.NoError(t, err)
	require.NotNil(t, tx)
	signedTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)
	require.NotNil(t, signedTx)
	res, err := client.BroadcastTx(signedTx)
	assert.NoError(t, err)
	require.NotNil(t, res)
	return res, nil
}

// The same as client.AddPackage, but test signing separately
func addPackageSigningSeparately(t *testing.T, client Client, cfg BaseTxCfg, msgs ...vm.MsgAddPackage) (*ctypes.ResultBroadcastTxCommit, error) {
	t.Helper()
	tx, err := NewAddPackageTx(cfg, msgs...)
	assert.NoError(t, err)
	require.NotNil(t, tx)
	signedTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)
	require.NotNil(t, signedTx)
	res, err := client.BroadcastTx(signedTx)
	assert.NoError(t, err)
	require.NotNil(t, res)
	return res, nil
}
