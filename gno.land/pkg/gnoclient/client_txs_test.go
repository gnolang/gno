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
	expectedRender := []byte("it works!")

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

func TestCallSingle_Sponsor(t *testing.T) {
	t.Parallel()

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
							Data: []byte("it works!"),
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

	msg := MsgCall{
		PkgPath:  "gno.land/r/demo/deep/very/deep",
		FuncName: "Render",
		Args:     []string{""},
		Send:     "100ugnot",
	}

	tx, err := client.NewSponsorTransaction(cfg, msg)
	assert.NoError(t, err)

	presignedTx, err := client.SignTransaction(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*presignedTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), "it works!")
}

func TestCallMultiple(t *testing.T) {
	t.Parallel()

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

func TestCallMultiple_Sponsor(t *testing.T) {
	t.Parallel()

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
							Data: []byte("it works!"),
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

	msg1 := MsgCall{
		PkgPath:  "gno.land/r/demo/deep/very/deep",
		FuncName: "Render",
		Args:     []string{""},
		Send:     "100ugnot",
	}

	msg2 := MsgCall{
		PkgPath:  "gno.land/r/demo/wugnot",
		FuncName: "Deposit",
		Args:     []string{""},
		Send:     "1000ugnot",
	}

	msg3 := MsgCall{
		PkgPath:  "gno.land/r/demo/tamagotchi",
		FuncName: "Feed",
		Args:     []string{""},
		Send:     "",
	}

	tx, err := client.NewSponsorTransaction(cfg, msg1, msg2, msg3)
	assert.NoError(t, err)

	presignedTx, err := client.SignTransaction(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*presignedTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), "it works!")
}

func TestCallErrors(t *testing.T) {
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

	receiver, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")

	msg := []MsgSend{
		{
			ToAddress: receiver,
			Send:      "100ugnot",
		},
	}

	res, err := client.Send(cfg, msg...)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), "it works!")
}

func TestSendSingle_Sponsor(t *testing.T) {
	t.Parallel()

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
							Data: []byte("it works!"),
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

	receiver, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")

	msg := MsgSend{
		ToAddress: receiver,
		Send:      "100ugnot",
	}

	tx, err := client.NewSponsorTransaction(cfg, msg)
	assert.NoError(t, err)

	presignedTx, err := client.SignTransaction(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*presignedTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), "it works!")
}

func TestSendMultiple(t *testing.T) {
	t.Parallel()

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

	receiver, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")

	msg1 := MsgSend{
		ToAddress: receiver,
		Send:      "100ugnot",
	}

	msg2 := MsgSend{
		ToAddress: receiver,
		Send:      "200ugnot",
	}

	res, err := client.Send(cfg, msg1, msg2)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), "it works!")
}

func TestSendMultiple_Sponsor(t *testing.T) {
	t.Parallel()

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
							Data: []byte("it works!"),
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

	receiver, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")

	msg1 := MsgSend{
		ToAddress: receiver,
		Send:      "100ugnot",
	}

	msg2 := MsgSend{
		ToAddress: receiver,
		Send:      "200ugnot",
	}

	tx, err := client.NewSponsorTransaction(cfg, msg1, msg2)
	assert.NoError(t, err)

	presignedTx, err := client.SignTransaction(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*presignedTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), "it works!")
}

func TestSendErrors(t *testing.T) {
	t.Parallel()

	toAddress, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")
	testCases := []struct {
		name          string
		client        Client
		cfg           BaseTxCfg
		msgs          []MsgSend
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
			msgs: []MsgSend{
				{
					ToAddress: toAddress,
					Send:      "1ugnot",
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
			msgs: []MsgSend{
				{
					ToAddress: toAddress,
					Send:      "1ugnot",
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
			msgs: []MsgSend{
				{
					ToAddress: toAddress,
					Send:      "1ugnot",
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
			msgs: []MsgSend{
				{
					ToAddress: toAddress,
					Send:      "1ugnot",
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
			msgs: []MsgSend{
				{
					ToAddress: toAddress,
					Send:      "1ugnot",
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
			msgs: []MsgSend{
				{
					ToAddress: crypto.Address{},
					Send:      "1ugnot",
				},
			},
			expectedError: ErrInvalidToAddress,
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
			msgs: []MsgSend{
				{
					ToAddress: toAddress,
					Send:      "-1ugnot",
				},
			},
			expectedError: ErrInvalidAmount,
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
							Data: []byte("hi gnoclient!\n"),
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

	msg := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody,
				},
			},
		},
		Send: "",
	}

	res, err := client.Run(cfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "hi gnoclient!\n", string(res.DeliverTx.Data))
}

func TestRunSingle_Sponsor(t *testing.T) {
	t.Parallel()

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

	msg := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main.gno",
					Body: fileBody,
				},
			},
		},
		Send: "",
	}

	tx, err := client.NewSponsorTransaction(cfg, msg)
	assert.NoError(t, err)

	presignedTx, err := client.SignTransaction(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*presignedTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	require.NotNil(t, res)
	assert.Equal(t, "hi gnoclient!\n", string(res.DeliverTx.Data))
}

func TestRunMultiple(t *testing.T) {
	t.Parallel()

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
							Data: []byte("hi gnoclient!\nhi gnoclient!\n"),
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

	msg1 := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main1.gno",
					Body: fileBody,
				},
			},
		},
		Send: "",
	}

	msg2 := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main2.gno",
					Body: fileBody,
				},
			},
		},
		Send: "",
	}

	res, err := client.Run(cfg, msg1, msg2)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "hi gnoclient!\nhi gnoclient!\n", string(res.DeliverTx.Data))
}

func TestRunMultiple_Sponsor(t *testing.T) {
	t.Parallel()

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
							Data: []byte("hi gnoclient!\nhi gnoclient!\n"),
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

	msg1 := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main1.gno",
					Body: fileBody,
				},
			},
		},
		Send: "",
	}

	msg2 := MsgRun{
		Package: &std.MemPackage{
			Files: []*std.MemFile{
				{
					Name: "main2.gno",
					Body: fileBody,
				},
			},
		},
		Send: "",
	}

	tx, err := client.NewSponsorTransaction(cfg, msg1, msg2)
	assert.NoError(t, err)

	presignedTx, err := client.SignTransaction(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*presignedTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	require.NotNil(t, res)
	assert.Equal(t, "hi gnoclient!\nhi gnoclient!\n", string(res.DeliverTx.Data))
}

func TestRunErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		cfg           BaseTxCfg
		msgs          []MsgRun
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
			msgs: []MsgRun{
				{
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
					Send: "",
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
			msgs:          []MsgRun{},
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
			msgs: []MsgRun{
				{
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
					Send: "",
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
			msgs: []MsgRun{
				{
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
					Send: "",
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
			msgs: []MsgRun{
				{
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
					Send: "",
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
			msgs: []MsgRun{
				{
					Package: nil,
					Send:    "",
				},
			},
			expectedError: ErrEmptyPackage,
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
							Data: []byte("hi gnoclient!\n"),
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

	msg := MsgAddPackage{
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
		Deposit: "",
	}

	res, err := client.AddPackage(cfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "hi gnoclient!\n", string(res.DeliverTx.Data))
}

func TestAddPackageSingle_Sponsor(t *testing.T) {
	t.Parallel()

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

	msg := MsgAddPackage{
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
		Deposit: "",
	}

	tx, err := client.NewSponsorTransaction(cfg, msg)
	assert.NoError(t, err)

	sponsorTx, err := client.SignTransaction(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*sponsorTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	require.NotNil(t, res)
	assert.Equal(t, "hi gnoclient!\n", string(res.DeliverTx.Data))
}

func TestAddPackageMultiple(t *testing.T) {
	t.Parallel()

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
							Data: []byte("hi gnoclient!\n"),
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

	msgs := []MsgAddPackage{
		{
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
			Deposit: "",
		},
		{
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
			Deposit: "",
		},
	}

	res, err := client.AddPackage(cfg, msgs...)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "hi gnoclient!\n", string(res.DeliverTx.Data))
}

func TestAddPackageMultiple_Sponsor(t *testing.T) {
	t.Parallel()

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

	msg1 := MsgAddPackage{
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
		Deposit: "",
	}

	msg2 := MsgAddPackage{
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
		Deposit: "",
	}

	tx, err := client.NewSponsorTransaction(cfg, msg1, msg2)
	assert.NoError(t, err)

	sponsorTx, err := client.SignTransaction(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	res, err := client.ExecuteSponsorTransaction(*sponsorTx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)

	require.NotNil(t, res)
	assert.Equal(t, "hi gnoclient!\n", string(res.DeliverTx.Data))
}

func TestAddPackageErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		cfg           BaseTxCfg
		msgs          []MsgAddPackage
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
			msgs: []MsgAddPackage{
				{
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
					Deposit: "",
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
			msgs:          []MsgAddPackage{},
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
			msgs: []MsgAddPackage{
				{
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
					Deposit: "",
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
			msgs: []MsgAddPackage{
				{
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
					Deposit: "",
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
			msgs: []MsgAddPackage{
				{
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
					Deposit: "",
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
			msgs: []MsgAddPackage{
				{
					Package: nil,
					Deposit: "",
				},
			},
			expectedError: ErrEmptyPackage,
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
		msgs          []Msg
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

			msgs: []Msg{}, // no messages provided

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

			msgs: []Msg{}, // no messages provided

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
			msgs: []Msg{
				MsgCall{
					PkgPath:  "gno.land/r/demo/deep/very/deep",
					FuncName: "Render",
					Args:     []string{""},
					Send:     "100ugnot",
				},
				MsgSend{
					ToAddress: addr1,
					Send:      "100ugnot",
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
			msgs: []Msg{
				// invalid message send
				MsgSend{
					ToAddress: crypto.Address{},
					Send:      "10000ugnot",
				},
			},
			expectedError: ErrInvalidToAddress,
		},
		{
			name: "Failed to parse coin from message",
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
			msgs: []Msg{
				MsgCall{
					PkgPath:  "gno.land/r/demo/deep/very/deep",
					FuncName: "Render",
					Args:     []string{""},
					Send:     "xxx", // invalid coin
				},
			},
			expectedError: ErrInvalidAmount,
		},
		{
			name: "Invalid message type",
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
			msgs: []Msg{
				mockMsg{}, // invalid msg type
			},
			expectedError: ErrInvalidMsgType,
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
			msgs: []Msg{
				MsgCall{
					PkgPath:  "gno.land/r/demo/deep/very/deep",
					FuncName: "Render",
					Args:     []string{""},
					Send:     "100ugnot",
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
			assert.Equal(t, err.Error(), tc.expectedError.Error())
		})
	}
}

func TestSignTransaction(t *testing.T) {
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

			res, err := tc.client.SignTransaction(tc.tx, 0, 0)
			assert.Nil(t, res)
			assert.Equal(t, err.Error(), tc.expectedError.Error())
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
			assert.Equal(t, err.Error(), tc.expectedError.Error())
		})
	}
}
