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

func TestRender(t *testing.T) {
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

// Call tests
func TestCallSingle(t *testing.T) {
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

func TestCallMultiple(t *testing.T) {
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

func TestClient_Send_Errors(t *testing.T) {
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
					info: func() keys.Info {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
								return adr
							},
						}
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
					info: func() keys.Info {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
								return adr
							},
						}
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
			expectedError: ErrInvalidSendAmount,
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

func TestRunMultiple(t *testing.T) {
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
					info: func() keys.Info {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
								return adr
							},
						}
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
					info: func() keys.Info {
						return &mockKeysInfo{
							getAddress: func() crypto.Address {
								adr, _ := crypto.AddressFromBech32("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
								return adr
							},
						}
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

// Block tests
func TestBlock(t *testing.T) {
	t.Parallel()

	height := int64(5)
	client := &Client{
		Signer: &mockSigner{},
		RPCClient: &mockRPCClient{
			block: func(height *int64) (*ctypes.ResultBlock, error) {
				return &ctypes.ResultBlock{
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
			blockResults: func(height *int64) (*ctypes.ResultBlockResults, error) {
				return &ctypes.ResultBlockResults{
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
			status: func() (*ctypes.ResultStatus, error) {
				return &ctypes.ResultStatus{
					SyncInfo: ctypes.SyncInfo{
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
