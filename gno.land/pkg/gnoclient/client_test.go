package gnoclient

import (
	"context"
	"errors"
	"testing"

	"github.com/gnolang/gno/gnovm/stdlibs/chain"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abciErrors "github.com/gnolang/gno/tm2/pkg/bft/abci/example/errors"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/blocks"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mempool"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gno.land/pkg/keyscli"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	abciTypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/abci"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

var testGasFee = ugnot.ValueString(10000)

func TestRender(t *testing.T) {
	t.Parallel()
	testRealmPath := "gno.land/r/tests/vm/deep/very/deep"
	expectedRender := []byte("it works!")

	client := Client{
		Signer: &mockSigner{
			sign: func(cfg SignCfg) (*std.Tx, error) {
				return &std.Tx{}, nil
			},
			info: func() (keys.Info, error) {
				return &mockKeysInfo{
					getAddress: func() crypto.Address {
						adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
						return adr
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			abciQuery: func(ctx context.Context, path string, data []byte) (*abciTypes.ResultABCIQuery, error) {
				res := &abciTypes.ResultABCIQuery{
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
						adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
						return adr
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTxCommit, error) {
				res := &mempool.ResultBroadcastTxCommit{
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
		GasFee:         testGasFee,
		AccountNumber:  1,
		SequenceNumber: 1,
		Memo:           "Test memo",
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg := []vm.MsgCall{
		{
			Caller:  caller.GetAddress(),
			PkgPath: "gno.land/r/tests/vm/deep/very/deep",
			Func:    "Render",
			Args:    []string{""},
			Send:    std.Coins{{Denom: ugnot.Denom, Amount: int64(100)}},
		},
	}

	res, err := client.Call(cfg, msg...)
	assert.NoError(t, err)
	require.NotNil(t, res)
	expected := "it works!"
	assert.Equal(t, string(res.DeliverTx.Data), expected)

	res, err = callSigningSeparately(t, client, cfg, msg...)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, string(res.DeliverTx.Data), expected)
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
						adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
						return adr
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTxCommit, error) {
				res := &mempool.ResultBroadcastTxCommit{
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
		GasFee:         testGasFee,
		AccountNumber:  1,
		SequenceNumber: 1,
		Memo:           "Test memo",
	}

	caller, err := client.Signer.Info()
	require.NoError(t, err)

	msg := []vm.MsgCall{
		{
			Caller:  caller.GetAddress(),
			PkgPath: "gno.land/r/tests/vm/deep/very/deep",
			Func:    "Render",
			Args:    []string{""},
			Send:    std.Coins{{Denom: ugnot.Denom, Amount: int64(100)}},
		},
		{
			Caller:  caller.GetAddress(),
			PkgPath: "gno.land/r/gnoland/wugnot",
			Func:    "Deposit",
			Args:    []string{""},
			Send:    std.Coins{{Denom: ugnot.Denom, Amount: int64(1000)}},
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

	res, err = callSigningSeparately(t, client, cfg, msg...)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestCallErrors(t *testing.T) {
	t.Parallel()

	// These tests don't actually sign
	mockAddress, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	testCases := []struct {
		name          string
		client        Client
		cfg           BaseTxCfg
		msgs          []vm.MsgCall
		expectedError string
	}{
		{
			name: "Invalid Signer",
			client: Client{
				Signer:    nil,
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgCall{
				{
					Caller:  mockAddress,
					PkgPath: "gno.land/r/random/path",
					Func:    "RandomName",
					Send:    nil,
					Args:    []string{},
				},
			},
			expectedError: ErrMissingSigner.Error(),
		},
		{
			name: "Invalid RPCClient",
			client: Client{
				&mockSigner{},
				nil,
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgCall{
				{
					Caller:  mockAddress,
					PkgPath: "gno.land/r/random/path",
					Func:    "RandomName",
					Send:    nil,
					Args:    []string{},
				},
			},
			expectedError: ErrMissingRPCClient.Error(),
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
					Caller:  mockAddress,
					PkgPath: "gno.land/r/random/path",
					Func:    "RandomName",
				},
			},
			expectedError: ErrInvalidGasFee.Error(),
		},
		{
			name: "Negative Gas Wanted",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      -1,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgCall{
				{
					Caller:  mockAddress,
					PkgPath: "gno.land/r/random/path",
					Func:    "RandomName",
					Send:    nil,
					Args:    []string{},
				},
			},
			expectedError: ErrInvalidGasWanted.Error(),
		},
		{
			name: "0 Gas Wanted",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      0,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgCall{
				{
					Caller:  mockAddress,
					PkgPath: "gno.land/r/random/path",
					Func:    "RandomName",
					Send:    nil,
					Args:    []string{},
				},
			},
			expectedError: ErrInvalidGasWanted.Error(),
		},
		{
			name: "Invalid PkgPath",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgCall{
				{
					Caller:  mockAddress,
					PkgPath: "",
					Func:    "RandomName",
					Send:    nil,
					Args:    []string{},
				},
			},
			expectedError: vm.InvalidPkgPathError{}.Error(),
		},
		{
			name: "Invalid FuncName",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgCall{
				{
					Caller:  mockAddress,
					PkgPath: "gno.land/r/random/path",
					Func:    "",
					Send:    nil,
					Args:    []string{},
				},
			},
			expectedError: vm.InvalidExprError{}.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.Call(tc.cfg, tc.msgs...)
			assert.Nil(t, res)
			assert.ErrorContains(t, err, tc.expectedError)
		})
	}
}

func TestClient_Send_Errors(t *testing.T) {
	t.Parallel()

	// These tests don't actually sign
	mockAddress, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	toAddress, _ := crypto.AddressFromBech32("g14a0y9a64dugh3l7hneshdxr4w0rfkkww9ls35p")
	testCases := []struct {
		name          string
		client        Client
		cfg           BaseTxCfg
		msgs          []bank.MsgSend
		expectedError string
	}{
		{
			name: "Invalid Signer",
			client: Client{
				Signer:    nil,
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []bank.MsgSend{
				{
					FromAddress: mockAddress,
					ToAddress:   toAddress,
					Amount:      std.Coins{{Denom: ugnot.Denom, Amount: int64(1)}},
				},
			},
			expectedError: ErrMissingSigner.Error(),
		},
		{
			name: "Invalid RPCClient",
			client: Client{
				&mockSigner{},
				nil,
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []bank.MsgSend{
				{
					FromAddress: mockAddress,
					ToAddress:   toAddress,
					Amount:      std.Coins{{Denom: ugnot.Denom, Amount: int64(1)}},
				},
			},
			expectedError: ErrMissingRPCClient.Error(),
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
					FromAddress: mockAddress,
					ToAddress:   toAddress,
					Amount:      std.Coins{{Denom: ugnot.Denom, Amount: int64(1)}},
				},
			},
			expectedError: ErrInvalidGasFee.Error(),
		},
		{
			name: "Negative Gas Wanted",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      -1,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []bank.MsgSend{
				{
					FromAddress: mockAddress,
					ToAddress:   toAddress,
					Amount:      std.Coins{{Denom: ugnot.Denom, Amount: int64(1)}},
				},
			},
			expectedError: ErrInvalidGasWanted.Error(),
		},
		{
			name: "0 Gas Wanted",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      0,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []bank.MsgSend{
				{
					FromAddress: mockAddress,
					ToAddress:   toAddress,
					Amount:      std.Coins{{Denom: ugnot.Denom, Amount: int64(1)}},
				},
			},
			expectedError: ErrInvalidGasWanted.Error(),
		},
		{
			name: "Invalid To Address",
			client: Client{
				Signer: &mockSigner{
					info: func() (keys.Info, error) {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
								return adr
							},
						}, nil
					},
				},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []bank.MsgSend{
				{
					FromAddress: mockAddress,
					ToAddress:   crypto.Address{},
					Amount:      std.Coins{{Denom: ugnot.Denom, Amount: int64(1)}},
				},
			},
			expectedError: std.InvalidAddressError{}.Error(),
		},
		{
			name: "Invalid Send Coins",
			client: Client{
				Signer: &mockSigner{
					info: func() (keys.Info, error) {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
								return adr
							},
						}, nil
					},
				},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []bank.MsgSend{
				{
					FromAddress: mockAddress,
					ToAddress:   toAddress,
					Amount:      std.Coins{{Denom: ugnot.Denom, Amount: int64(-1)}},
				},
			},
			expectedError: std.InvalidCoinsError{}.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.Send(tc.cfg, tc.msgs...)
			assert.Nil(t, res)
			assert.ErrorContains(t, err, tc.expectedError)
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
						adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
						return adr
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTxCommit, error) {
				res := &mempool.ResultBroadcastTxCommit{
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
		GasFee:         testGasFee,
		AccountNumber:  1,
		SequenceNumber: 1,
		Memo:           "Test memo",
	}

	fileBody := `package main
import (
	"std"
	"gno.land/p/nt/ufmt"
	"gno.land/r/tests/vm/deep/very/deep"
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
	expected := "hi gnoclient!\n"
	assert.Equal(t, expected, string(res.DeliverTx.Data))

	res, err = runSigningSeparately(t, client, cfg, msg)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
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
						adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
						return adr
					},
				}, nil
			},
		},
		RPCClient: &mockRPCClient{
			broadcastTxCommit: func(ctx context.Context, tx types.Tx) (*mempool.ResultBroadcastTxCommit, error) {
				res := &mempool.ResultBroadcastTxCommit{
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
		GasFee:         testGasFee,
		AccountNumber:  1,
		SequenceNumber: 1,
		Memo:           "Test memo",
	}

	fileBody := `package main
import (
	"std"
	"gno.land/p/nt/ufmt"
	"gno.land/r/tests/vm/deep/very/deep"
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
	expected := "hi gnoclient!\nhi gnoclient!\n"
	assert.Equal(t, expected, string(res.DeliverTx.Data))

	res, err = runSigningSeparately(t, client, cfg, msg1, msg2)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, expected, string(res.DeliverTx.Data))
}

func TestRunErrors(t *testing.T) {
	t.Parallel()

	// These tests don't actually sign
	mockAddress, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	testCases := []struct {
		name          string
		client        Client
		cfg           BaseTxCfg
		msgs          []vm.MsgRun
		expectedError string
	}{
		{
			name: "Invalid Signer",
			client: Client{
				Signer:    nil,
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgRun{
				{
					Caller: mockAddress,
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
			expectedError: ErrMissingSigner.Error(),
		},
		{
			name: "Invalid RPCClient",
			client: Client{
				&mockSigner{},
				nil,
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs:          []vm.MsgRun{},
			expectedError: ErrMissingRPCClient.Error(),
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
					Caller: mockAddress,
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
			expectedError: ErrInvalidGasFee.Error(),
		},
		{
			name: "Negative Gas Wanted",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      -1,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgRun{
				{
					Caller: mockAddress,
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
			expectedError: ErrInvalidGasWanted.Error(),
		},
		{
			name: "0 Gas Wanted",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      0,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgRun{
				{
					Caller: mockAddress,
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
			expectedError: ErrInvalidGasWanted.Error(),
		},
		{
			name: "Invalid Empty Package",
			client: Client{
				Signer: &mockSigner{
					info: func() (keys.Info, error) {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
								return adr
							},
						}, nil
					},
				},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgRun{
				{
					Caller:  mockAddress,
					Package: &std.MemPackage{Name: "", Path: " "},
					Send:    nil,
				},
			},
			expectedError: vm.InvalidPkgPathError{}.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.Run(tc.cfg, tc.msgs...)
			assert.Nil(t, res)
			assert.ErrorContains(t, err, tc.expectedError)
		})
	}
}

// AddPackage tests
func TestAddPackageErrors(t *testing.T) {
	t.Parallel()

	// These tests don't actually sign
	mockAddress, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

	testCases := []struct {
		name          string
		client        Client
		cfg           BaseTxCfg
		msgs          []vm.MsgAddPackage
		expectedError string
	}{
		{
			name: "Invalid Signer",
			client: Client{
				Signer:    nil,
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgAddPackage{
				{
					Creator: mockAddress,
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
					MaxDeposit: nil,
				},
			},
			expectedError: ErrMissingSigner.Error(),
		},
		{
			name: "Invalid RPCClient",
			client: Client{
				&mockSigner{},
				nil,
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs:          []vm.MsgAddPackage{},
			expectedError: ErrMissingRPCClient.Error(),
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
					Creator: mockAddress,
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
					MaxDeposit: nil,
				},
			},
			expectedError: ErrInvalidGasFee.Error(),
		},
		{
			name: "Negative Gas Wanted",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      -1,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgAddPackage{
				{
					Creator: mockAddress,
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
					MaxDeposit: nil,
				},
			},
			expectedError: ErrInvalidGasWanted.Error(),
		},
		{
			name: "0 Gas Wanted",
			client: Client{
				Signer:    &mockSigner{},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      0,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgAddPackage{
				{
					Creator: mockAddress,
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
					MaxDeposit: nil,
				},
			},
			expectedError: ErrInvalidGasWanted.Error(),
		},
		{
			name: "Invalid Empty Package",
			client: Client{
				Signer: &mockSigner{
					info: func() (keys.Info, error) {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
								return adr
							},
						}, nil
					},
				},
				RPCClient: &mockRPCClient{},
			},
			cfg: BaseTxCfg{
				GasWanted:      100000,
				GasFee:         testGasFee,
				AccountNumber:  1,
				SequenceNumber: 1,
				Memo:           "Test memo",
			},
			msgs: []vm.MsgAddPackage{
				{
					Creator:    mockAddress,
					Package:    &std.MemPackage{Name: "", Path: ""},
					MaxDeposit: nil,
				},
			},
			expectedError: vm.InvalidPkgPathError{}.Error(),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.AddPackage(tc.cfg, tc.msgs...)
			assert.Nil(t, res)
			assert.ErrorContains(t, err, tc.expectedError)
		})
	}
}

// Block tests
func TestBlock(t *testing.T) {
	t.Parallel()

	height := int64(5)
	client := &Client{
		Signer: &mockSigner{},
		RPCClient: &mockRPCClient{
			block: func(ctx context.Context, height *int64) (*blocks.ResultBlock, error) {
				return &blocks.ResultBlock{
					BlockMeta: &types.BlockMeta{
						BlockID: types.BlockID{},
						Header:  types.Header{},
					},
					Block: &types.Block{
						Header: types.Header{
							Height: *height,
						},
						Data:       types.Data{},
						LastCommit: nil,
					},
				}, nil
			},
		},
	}

	block, err := client.Block(height)
	require.NoError(t, err)
	assert.Equal(t, height, block.Block.GetHeight())
}

func TestBlockResults(t *testing.T) {
	t.Parallel()

	height := int64(5)
	client := &Client{
		Signer: &mockSigner{},
		RPCClient: &mockRPCClient{
			blockResults: func(ctx context.Context, height *int64) (*blocks.ResultBlockResults, error) {
				return &blocks.ResultBlockResults{
					Height:  *height,
					Results: nil,
				}, nil
			},
		},
	}

	blockResult, err := client.BlockResult(height)
	require.NoError(t, err)
	assert.Equal(t, height, blockResult.Height)
}

func TestLatestBlockHeight(t *testing.T) {
	t.Parallel()

	latestHeight := int64(5)

	client := &Client{
		Signer: &mockSigner{},
		RPCClient: &mockRPCClient{
			status: func(ctx context.Context, heightGte *int64) (*status.ResultStatus, error) {
				return &status.ResultStatus{
					SyncInfo: status.SyncInfo{
						LatestBlockHeight: latestHeight,
					},
				}, nil
			},
		},
	}

	head, err := client.LatestBlockHeight()
	require.NoError(t, err)
	assert.Equal(t, latestHeight, head)
}

func TestBlockErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		height        int64
		expectedError error
	}{
		{
			name: "Invalid RPCClient",
			client: Client{
				&mockSigner{},
				nil,
			},
			height:        1,
			expectedError: ErrMissingRPCClient,
		},
		{
			name: "Invalid height",
			client: Client{
				&mockSigner{},
				&mockRPCClient{},
			},
			height:        0,
			expectedError: ErrInvalidBlockHeight,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.Block(tc.height)
			assert.Nil(t, res)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestBlockResultErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		height        int64
		expectedError error
	}{
		{
			name: "Invalid RPCClient",
			client: Client{
				&mockSigner{},
				nil,
			},
			height:        1,
			expectedError: ErrMissingRPCClient,
		},
		{
			name: "Invalid height",
			client: Client{
				&mockSigner{},
				&mockRPCClient{},
			},
			height:        0,
			expectedError: ErrInvalidBlockHeight,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.BlockResult(tc.height)
			assert.Nil(t, res)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

func TestLatestBlockHeightErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		client        Client
		expectedError error
	}{
		{
			name: "Invalid RPCClient",
			client: Client{
				&mockSigner{},
				nil,
			},
			expectedError: ErrMissingRPCClient,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := tc.client.LatestBlockHeight()
			assert.Equal(t, int64(0), res)
			assert.ErrorIs(t, err, tc.expectedError)
		})
	}
}

// The same as client.Call, but test signing separately
func callSigningSeparately(t *testing.T, client Client, cfg BaseTxCfg, msgs ...vm.MsgCall) (*mempool.ResultBroadcastTxCommit, error) {
	t.Helper()
	tx, err := NewCallTx(cfg, msgs...)
	assert.NoError(t, err)
	require.NotNil(t, tx)
	signedTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)
	require.NotNil(t, signedTx)
	res, err := client.BroadcastTxCommit(signedTx)
	assert.NoError(t, err)
	require.NotNil(t, res)
	return res, nil
}

// The same as client.Run, but test signing separately
func runSigningSeparately(t *testing.T, client Client, cfg BaseTxCfg, msgs ...vm.MsgRun) (*mempool.ResultBroadcastTxCommit, error) {
	t.Helper()
	tx, err := NewRunTx(cfg, msgs...)
	assert.NoError(t, err)
	require.NotNil(t, tx)
	signedTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)
	require.NotNil(t, signedTx)
	res, err := client.BroadcastTxCommit(signedTx)
	assert.NoError(t, err)
	require.NotNil(t, res)
	return res, nil
}

// The same as client.Send, but test signing separately
func sendSigningSeparately(t *testing.T, client Client, cfg BaseTxCfg, msgs ...bank.MsgSend) (*mempool.ResultBroadcastTxCommit, error) {
	t.Helper()
	tx, err := NewSendTx(cfg, msgs...)
	assert.NoError(t, err)
	require.NotNil(t, tx)
	signedTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)
	require.NotNil(t, signedTx)
	res, err := client.BroadcastTxCommit(signedTx)
	assert.NoError(t, err)
	require.NotNil(t, res)
	return res, nil
}

// The same as client.AddPackage, but test signing separately
func addPackageSigningSeparately(t *testing.T, client Client, cfg BaseTxCfg, msgs ...vm.MsgAddPackage) (*mempool.ResultBroadcastTxCommit, error) {
	t.Helper()
	tx, err := NewAddPackageTx(cfg, msgs...)
	assert.NoError(t, err)
	require.NotNil(t, tx)
	signedTx, err := client.SignTx(*tx, cfg.AccountNumber, cfg.SequenceNumber)
	assert.NoError(t, err)
	require.NotNil(t, signedTx)
	res, err := client.BroadcastTxCommit(signedTx)
	assert.NoError(t, err)
	require.NotNil(t, res)
	return res, nil
}

func TestClient_EstimateGas(t *testing.T) {
	t.Parallel()

	t.Run("RPC client not set", func(t *testing.T) {
		t.Parallel()

		c := &Client{
			RPCClient: nil, // not set
		}

		estimate, err := c.EstimateGas(&std.Tx{})

		assert.Zero(t, estimate)
		assert.ErrorIs(t, err, ErrMissingRPCClient)
	})

	t.Run("unsuccessful query, rpc error", func(t *testing.T) {
		t.Parallel()

		var (
			rpcErr        = errors.New("rpc error")
			mockRPCClient = &mockRPCClient{
				abciQuery: func(ctx context.Context, path string, data []byte) (*abciTypes.ResultABCIQuery, error) {
					require.Equal(t, simulatePath, path)

					var tx std.Tx

					require.NoError(t, amino.Unmarshal(data, &tx))

					return nil, rpcErr
				},
			}
		)

		c := &Client{
			RPCClient: mockRPCClient,
		}

		estimate, err := c.EstimateGas(&std.Tx{})

		assert.Zero(t, estimate)
		assert.ErrorIs(t, err, rpcErr)
	})

	t.Run("unsuccessful query, process error", func(t *testing.T) {
		t.Parallel()

		var (
			response = &abciTypes.ResultABCIQuery{
				Response: abci.ResponseQuery{
					ResponseBase: abci.ResponseBase{
						Error: abciErrors.UnknownError{},
					},
				},
			}
			mockRPCClient = &mockRPCClient{
				abciQuery: func(ctx context.Context, path string, data []byte) (*abciTypes.ResultABCIQuery, error) {
					require.Equal(t, simulatePath, path)

					var tx std.Tx

					require.NoError(t, amino.Unmarshal(data, &tx))

					return response, nil
				},
			}
		)

		c := &Client{
			RPCClient: mockRPCClient,
		}

		estimate, err := c.EstimateGas(&std.Tx{})

		assert.Zero(t, estimate)
		assert.ErrorIs(t, err, abciErrors.UnknownError{})
	})

	t.Run("invalid response format", func(t *testing.T) {
		t.Parallel()

		var (
			response = &abciTypes.ResultABCIQuery{
				Response: abci.ResponseQuery{
					Value: []byte("totally valid amino"),
				},
			}
			mockRPCClient = &mockRPCClient{
				abciQuery: func(ctx context.Context, path string, data []byte) (*abciTypes.ResultABCIQuery, error) {
					require.Equal(t, simulatePath, path)

					var tx std.Tx

					require.NoError(t, amino.Unmarshal(data, &tx))

					return response, nil
				},
			}
		)

		c := &Client{
			RPCClient: mockRPCClient,
		}

		estimate, err := c.EstimateGas(&std.Tx{})

		assert.Zero(t, estimate)
		assert.ErrorContains(t, err, "unable to unmarshal simulation response")
	})

	t.Run("valid gas estimation", func(t *testing.T) {
		t.Parallel()

		var (
			gasUsed     = int64(100000)
			deliverResp = &abci.ResponseDeliverTx{
				GasUsed: gasUsed,
			}
		)

		// Encode the response
		encodedResp, err := amino.Marshal(deliverResp)
		require.NoError(t, err)

		var (
			response = &abciTypes.ResultABCIQuery{
				Response: abci.ResponseQuery{
					Value: encodedResp, // valid amino binary
				},
			}
			mockRPCClient = &mockRPCClient{
				abciQuery: func(ctx context.Context, path string, data []byte) (*abciTypes.ResultABCIQuery, error) {
					require.Equal(t, simulatePath, path)

					var tx std.Tx

					require.NoError(t, amino.Unmarshal(data, &tx))

					return response, nil
				},
			}
		)

		c := &Client{
			RPCClient: mockRPCClient,
		}

		estimate, err := c.EstimateGas(&std.Tx{})

		require.NoError(t, err)
		assert.Equal(t, gasUsed, estimate)
	})

	t.Run("valid simulate", func(t *testing.T) {
		t.Parallel()

		var (
			gasUsed     = int64(100000)
			deliverResp = &abci.ResponseDeliverTx{
				GasUsed: gasUsed,
				ResponseBase: abci.ResponseBase{
					Events: []abci.Event{
						&chain.StorageDepositEvent{
							BytesDelta: 10,
							FeeDelta:   std.Coin{Denom: ugnot.Denom, Amount: 1000},
						},
					},
				},
			}
		)

		// Encode the response
		encodedResp, err := amino.Marshal(deliverResp)
		require.NoError(t, err)

		var (
			response = &abciTypes.ResultABCIQuery{
				Response: abci.ResponseQuery{
					Value: encodedResp, // valid amino binary
				},
			}
			mockRPCClient = &mockRPCClient{
				abciQuery: func(ctx context.Context, path string, data []byte) (*abciTypes.ResultABCIQuery, error) {
					require.Equal(t, simulatePath, path)

					var tx std.Tx

					require.NoError(t, amino.Unmarshal(data, &tx))

					return response, nil
				},
			}
		)

		c := &Client{
			RPCClient: mockRPCClient,
		}

		deliverTx, err := c.Simulate(&std.Tx{})

		require.NoError(t, err)
		assert.Equal(t, gasUsed, deliverTx.GasUsed)

		bytesDelta, coinsDelta, hasStorageEvents := keyscli.GetStorageInfo(deliverTx.Events)
		assert.Equal(t, true, hasStorageEvents)
		assert.Equal(t, int64(10), bytesDelta)
		assert.Equal(t, "1000ugnot", coinsDelta.String())
	})
}
