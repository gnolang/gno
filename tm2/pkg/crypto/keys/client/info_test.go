package client

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/gnovm/stdlibs/chain"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/gas"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
)

func TestGetStorageInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		events            []abci.Event
		expectedBytes     int64
		expectedCoins     std.Coins
		expectedHasEvents bool
	}{
		{
			name:              "no storage events",
			events:            []abci.Event{},
			expectedBytes:     0,
			expectedCoins:     nil,
			expectedHasEvents: false,
		},
		{
			name: "single deposit event",
			events: []abci.Event{
				chain.StorageDepositEvent{
					BytesDelta: 100,
					FeeDelta:   std.Coin{Denom: "ugnot", Amount: 1000},
					PkgPath:    "gno.land/r/demo",
				},
			},
			expectedBytes:     100,
			expectedCoins:     std.Coins{std.Coin{Denom: "ugnot", Amount: 1000}},
			expectedHasEvents: true,
		},
		{
			name: "single unlock event without withholding",
			events: []abci.Event{
				chain.StorageUnlockEvent{
					BytesDelta:     -50,
					FeeRefund:      std.Coin{Denom: "ugnot", Amount: 500},
					PkgPath:        "gno.land/r/demo",
					RefundWithheld: false,
				},
			},
			expectedBytes:     -50,
			expectedCoins:     std.Coins{std.Coin{Denom: "ugnot", Amount: -500}},
			expectedHasEvents: true,
		},
		{
			name: "single unlock event with withholding",
			events: []abci.Event{
				chain.StorageUnlockEvent{
					BytesDelta:     -50,
					FeeRefund:      std.Coin{Denom: "ugnot", Amount: 500},
					PkgPath:        "gno.land/r/demo",
					RefundWithheld: true,
				},
			},
			expectedBytes:     -50,
			expectedCoins:     nil,
			expectedHasEvents: true,
		},
		{
			name: "multiple deposit events",
			events: []abci.Event{
				chain.StorageDepositEvent{
					BytesDelta: 100,
					FeeDelta:   std.Coin{Denom: "ugnot", Amount: 1000},
					PkgPath:    "gno.land/r/demo1",
				},
				chain.StorageDepositEvent{
					BytesDelta: 200,
					FeeDelta:   std.Coin{Denom: "ugnot", Amount: 2000},
					PkgPath:    "gno.land/r/demo2",
				},
			},
			expectedBytes:     300,
			expectedCoins:     std.Coins{std.Coin{Denom: "ugnot", Amount: 3000}},
			expectedHasEvents: true,
		},
		{
			name: "mixed deposit and unlock events",
			events: []abci.Event{
				chain.StorageDepositEvent{
					BytesDelta: 100,
					FeeDelta:   std.Coin{Denom: "ugnot", Amount: 1000},
					PkgPath:    "gno.land/r/demo1",
				},
				chain.StorageUnlockEvent{
					BytesDelta:     -30,
					FeeRefund:      std.Coin{Denom: "ugnot", Amount: 300},
					PkgPath:        "gno.land/r/demo2",
					RefundWithheld: false,
				},
			},
			expectedBytes:     70,
			expectedCoins:     std.Coins{std.Coin{Denom: "ugnot", Amount: 700}},
			expectedHasEvents: true,
		},
		{
			name: "mixed events with non-storage events",
			events: []abci.Event{
				chain.Event{
					Type: "custom",
					Attributes: []chain.EventAttribute{
						{Key: "key", Value: "value"},
					},
					PkgPath: "gno.land/r/demo",
				},
				chain.StorageDepositEvent{
					BytesDelta: 100,
					FeeDelta:   std.Coin{Denom: "ugnot", Amount: 1000},
					PkgPath:    "gno.land/r/demo",
				},
			},
			expectedBytes:     100,
			expectedCoins:     std.Coins{std.Coin{Denom: "ugnot", Amount: 1000}},
			expectedHasEvents: true,
		},
		{
			name:              "nil events",
			events:            nil,
			expectedBytes:     0,
			expectedCoins:     nil,
			expectedHasEvents: false,
		},
		{
			name: "zero bytes delta deposit",
			events: []abci.Event{
				chain.StorageDepositEvent{
					BytesDelta: 0,
					FeeDelta:   std.Coin{Denom: "ugnot", Amount: 0},
					PkgPath:    "gno.land/r/demo",
				},
			},
			expectedBytes:     0,
			expectedCoins:     nil, // zero-value coins are not added to the slice
			expectedHasEvents: true,
		},
		{
			name: "mixed withheld and non-withheld unlocks",
			events: []abci.Event{
				chain.StorageUnlockEvent{
					BytesDelta:     -30,
					FeeRefund:      std.Coin{Denom: "ugnot", Amount: 300},
					PkgPath:        "gno.land/r/demo1",
					RefundWithheld: false,
				},
				chain.StorageUnlockEvent{
					BytesDelta:     -20,
					FeeRefund:      std.Coin{Denom: "ugnot", Amount: 200},
					PkgPath:        "gno.land/r/demo2",
					RefundWithheld: true,
				},
			},
			expectedBytes:     -50,
			expectedCoins:     std.Coins{std.Coin{Denom: "ugnot", Amount: -300}},
			expectedHasEvents: true,
		},
		{
			name: "large byte deltas",
			events: []abci.Event{
				chain.StorageDepositEvent{
					BytesDelta: 1000000,
					FeeDelta:   std.Coin{Denom: "ugnot", Amount: 1000000000},
					PkgPath:    "gno.land/r/demo",
				},
			},
			expectedBytes:     1000000,
			expectedCoins:     std.Coins{std.Coin{Denom: "ugnot", Amount: 1000000000}},
			expectedHasEvents: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bytesDelta, coinsDelta, hasEvents := GetStorageInfo(tt.events)

			assert.Equal(t, tt.expectedBytes, bytesDelta, "bytes delta mismatch")
			assert.Equal(t, tt.expectedCoins, coinsDelta, "coins delta mismatch")
			assert.Equal(t, tt.expectedHasEvents, hasEvents, "hasEvents mismatch")
		})
	}
}

func TestPrintTxInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		tx           std.Tx
		res          *ctypes.ResultBroadcastTxCommit
		verbosity    int
		expectOutput []string // strings that should appear in output
	}{
		{
			name: "basic transaction with verbosity 0",
			tx: std.Tx{
				Fee: std.Fee{
					GasFee: std.Coin{Denom: "ugnot", Amount: 1000},
				},
			},
			res: &ctypes.ResultBroadcastTxCommit{
				DeliverTx: abci.ResponseDeliverTx{
					ResponseBase: abci.ResponseBase{
						Data: []byte("transaction successful"),
					},
					GasWanted: 100000,
					GasUsed: gas.GasDetail{
						Total: gas.Detail{
							OperationCount: 10,
							GasConsumed:    50000,
						},
					},
				},
				Height: 12345,
				Hash:   []byte("test_hash"),
			},
			verbosity: 0,
			expectOutput: []string{
				"transaction successful",
				"OK!",
				"GAS WANTED: 100000",
				"GAS USED:   50000",
				"HEIGHT:     12345",
			},
		},
		{
			name: "transaction with storage deposit",
			tx: std.Tx{
				Fee: std.Fee{
					GasFee: std.Coin{Denom: "ugnot", Amount: 1000},
				},
			},
			res: &ctypes.ResultBroadcastTxCommit{
				DeliverTx: abci.ResponseDeliverTx{
					ResponseBase: abci.ResponseBase{
						Data: []byte("success"),
						Events: []abci.Event{
							chain.StorageDepositEvent{
								BytesDelta: 100,
								FeeDelta:   std.Coin{Denom: "ugnot", Amount: 500},
								PkgPath:    "gno.land/r/demo",
							},
						},
					},
					GasWanted: 100000,
					GasUsed: gas.GasDetail{
						Total: gas.Detail{
							OperationCount: 10,
							GasConsumed:    50000,
						},
					},
				},
				Height: 12345,
				Hash:   []byte("test_hash"),
			},
			verbosity: 0,
			expectOutput: []string{
				"success",
				"OK!",
				"STORAGE DELTA:  100 bytes",
				"STORAGE FEE:",
				"500ugnot",
				"TOTAL TX COST:",
				"1500ugnot",
			},
		},
		{
			name: "transaction with storage refund",
			tx: std.Tx{
				Fee: std.Fee{
					GasFee: std.Coin{Denom: "ugnot", Amount: 1000},
				},
			},
			res: &ctypes.ResultBroadcastTxCommit{
				DeliverTx: abci.ResponseDeliverTx{
					ResponseBase: abci.ResponseBase{
						Data: []byte("success"),
						Events: []abci.Event{
							chain.StorageUnlockEvent{
								BytesDelta:     -50,
								FeeRefund:      std.Coin{Denom: "ugnot", Amount: 300},
								PkgPath:        "gno.land/r/demo",
								RefundWithheld: false,
							},
						},
					},
					GasWanted: 100000,
					GasUsed: gas.GasDetail{
						Total: gas.Detail{
							OperationCount: 10,
							GasConsumed:    50000,
						},
					},
				},
				Height: 12345,
				Hash:   []byte("test_hash"),
			},
			verbosity: 0,
			expectOutput: []string{
				"success",
				"OK!",
				"STORAGE DELTA:  -50 bytes",
				"STORAGE REFUND:",
				"300ugnot",
				"TOTAL TX COST:",
				"700ugnot",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create test IO
			io := commands.NewTestIO()
			var outBuf bytes.Buffer
			io.SetOut(commands.WriteNopCloser(&outBuf))

			// Call PrintTxInfo
			PrintTxInfo(tt.tx, tt.res, io, tt.verbosity)

			// Check output contains expected strings
			output := outBuf.String()
			for _, expected := range tt.expectOutput {
				assert.Contains(t, output, expected, "output should contain %q", expected)
			}
		})
	}
}

func TestPrintGasDetail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		detail       gas.GasDetail
		verbosity    int
		expectOutput []string
		notExpected  []string
	}{
		{
			name: "verbosity 1 - shows only totals",
			detail: gas.GasDetail{
				Total: gas.Detail{
					OperationCount: 100,
					GasConsumed:    50000,
				},
				Operations: [gas.OperationListMaxSize]gas.Detail{
					gas.OpCPUAdd: {
						OperationCount: 50,
						GasConsumed:    25000,
					},
					gas.OpMemoryAllocPerByte: {
						OperationCount: 30,
						GasConsumed:    15000,
					},
					gas.OpStoreReadFlat: {
						OperationCount: 20,
						GasConsumed:    10000,
					},
				},
			},
			verbosity: 1,
			expectOutput: []string{
				"GAS USED",
				"Operation: 100",
				"Gas: 50000",
			},
			notExpected: []string{
				"CPUAdd",
				"MemoryAllocPerByte",
			},
		},
		{
			name: "verbosity 2 - shows categories and operations",
			detail: gas.GasDetail{
				Total: gas.Detail{
					OperationCount: 100,
					GasConsumed:    50000,
				},
				Operations: [gas.OperationListMaxSize]gas.Detail{
					gas.OpCPUAdd: {
						OperationCount: 50,
						GasConsumed:    25000,
					},
					gas.OpMemoryAllocPerByte: {
						OperationCount: 30,
						GasConsumed:    15000,
					},
					gas.OpStoreReadFlat: {
						OperationCount: 20,
						GasConsumed:    10000,
					},
				},
			},
			verbosity: 2,
			expectOutput: []string{
				"GAS USED",
				"Operation: 100",
				"Gas: 50000",
			},
		},
		{
			name: "verbosity 3 - shows all including zero operations",
			detail: gas.GasDetail{
				Total: gas.Detail{
					OperationCount: 50,
					GasConsumed:    25000,
				},
				Operations: [gas.OperationListMaxSize]gas.Detail{
					gas.OpCPUAdd: {
						OperationCount: 50,
						GasConsumed:    25000,
					},
					// Other operations have zero counts
				},
			},
			verbosity: 3,
			expectOutput: []string{
				"GAS USED",
				"Operation: 50",
				"Gas: 25000",
				"Operation: 0",
				"Gas: 0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create test IO
			io := commands.NewTestIO()
			var outBuf bytes.Buffer
			io.SetOut(commands.WriteNopCloser(&outBuf))

			// Call printGasDetail
			printGasDetail(tt.detail, io, tt.verbosity)

			// Check output contains expected strings
			output := outBuf.String()
			for _, expected := range tt.expectOutput {
				assert.Contains(t, output, expected, "output should contain %q", expected)
			}

			// Check output does not contain unexpected strings
			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, output, notExpected, "output should not contain %q", notExpected)
			}
		})
	}
}
