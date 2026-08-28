package client

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

func TestHandleDeliverResultCallsOnFailure(t *testing.T) {
	called := false
	cfg := &BaseCfg{BaseOptions: BaseOptions{OnTxFailure: func(commands.IO, std.Tx, *ctypes.ResultBroadcastTxCommit) {
		called = true
	}}}

	tx := std.Tx{}
	bres := &ctypes.ResultBroadcastTxCommit{
		DeliverTx: abci.ResponseDeliverTx{
			ResponseBase: abci.ResponseBase{Error: abci.StringError("fail")},
			GasWanted:    10,
			GasUsed:      20,
		},
	}

	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(&bytes.Buffer{}))
	err := handleDeliverResult(cfg, tx, bres, io)

	require.True(t, called, "OnTxFailure should be invoked")
	require.Error(t, err)
}

func TestHandleDeliverResultPrintsDefaultWhenNoCallback(t *testing.T) {
	cfg := &BaseCfg{BaseOptions: BaseOptions{}}
	tx := std.Tx{}
	bres := &ctypes.ResultBroadcastTxCommit{
		DeliverTx: abci.ResponseDeliverTx{
			ResponseBase: abci.ResponseBase{Error: abci.StringError("fail"), Info: "hint"},
			GasWanted:    7,
			GasUsed:      9,
		},
	}

	buf := &bytes.Buffer{}
	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(buf))

	err := handleDeliverResult(cfg, tx, bres, io)
	require.Error(t, err)

	output := buf.String()
	require.Contains(t, output, "GAS WANTED: 7")
	require.Contains(t, output, "GAS USED:   9")
	require.Contains(t, output, "INFO:")
	require.Contains(t, output, "hint")
}

func TestBuildSimulationTxBytesUsesConsensusMaxGas(t *testing.T) {
	tx := std.Tx{Fee: std.Fee{GasWanted: 10}}
	bz, err := amino.Marshal(&tx)
	require.NoError(t, err)

	simBz, rewritten, err := buildSimulationTxBytes(&tx, bz, 25)
	require.NoError(t, err)
	require.True(t, rewritten)

	var simTx std.Tx
	require.NoError(t, amino.Unmarshal(simBz, &simTx))
	require.Equal(t, int64(25), simTx.Fee.GasWanted)
}

func TestBuildSimulationTxBytesUsesFallbackWhenConsensusMaxGasUndefined(t *testing.T) {
	tx := std.Tx{Fee: std.Fee{GasWanted: 10}}
	bz, err := amino.Marshal(&tx)
	require.NoError(t, err)

	simBz, rewritten, err := buildSimulationTxBytes(&tx, bz, -1)
	require.NoError(t, err)
	require.True(t, rewritten)

	var simTx std.Tx
	require.NoError(t, amino.Unmarshal(simBz, &simTx))
	require.Equal(t, simulationMaxGasFallback, simTx.Fee.GasWanted)
}

func TestBuildSimulationTxBytesKeepsOriginalWhenMaxGasUnknown(t *testing.T) {
	tx := std.Tx{Fee: std.Fee{GasWanted: 10}}
	bz, err := amino.Marshal(&tx)
	require.NoError(t, err)

	simBz, rewritten, err := buildSimulationTxBytes(&tx, bz, 0)
	require.NoError(t, err)
	require.False(t, rewritten)
	require.Equal(t, bz, simBz)
}

func TestBuildSimulationTxBytesKeepsHigherOriginalGasWanted(t *testing.T) {
	tx := std.Tx{Fee: std.Fee{GasWanted: 100}}
	bz, err := amino.Marshal(&tx)
	require.NoError(t, err)

	simBz, rewritten, err := buildSimulationTxBytes(&tx, bz, 25)
	require.NoError(t, err)
	require.False(t, rewritten)
	require.Equal(t, bz, simBz)
}

func TestAppendSuggestedGasWanted(t *testing.T) {
	bres := &ctypes.ResultBroadcastTxCommit{
		DeliverTx: abci.ResponseDeliverTx{
			GasUsed: 100,
		},
	}

	appendSuggestedGasWanted(bres)
	require.Equal(t, "suggested gas-wanted (gas used + 5%): 105", bres.DeliverTx.Info)
}

func TestAppendSuggestedGasWantedAppendsExistingInfo(t *testing.T) {
	bres := &ctypes.ResultBroadcastTxCommit{
		DeliverTx: abci.ResponseDeliverTx{
			ResponseBase: abci.ResponseBase{Info: "estimated gas usage: 100"},
			GasUsed:      100,
		},
	}

	appendSuggestedGasWanted(bres)
	require.Equal(t, "estimated gas usage: 100, suggested gas-wanted (gas used + 5%): 105", bres.DeliverTx.Info)
}

func TestResolveMaxGasWarnsOnError(t *testing.T) {
	ch := make(chan consensusMaxGasResult, 1)
	ch <- consensusMaxGasResult{err: errors.New("connection refused")}

	errBuf := &bytes.Buffer{}
	io := commands.NewTestIO()
	io.SetErr(commands.WriteNopCloser(errBuf))

	maxGas := resolveMaxGas(ch, io)

	require.Equal(t, int64(0), maxGas)
	require.Contains(t, errBuf.String(), "warning")
	require.Contains(t, errBuf.String(), "connection refused")
}

func TestResolveMaxGasReturnsValueOnSuccess(t *testing.T) {
	ch := make(chan consensusMaxGasResult, 1)
	ch <- consensusMaxGasResult{maxGas: 1_000_000}

	maxGas := resolveMaxGas(ch, commands.NewTestIO())

	require.Equal(t, int64(1_000_000), maxGas)
}

func TestOutOfGasLogTxGasWanted(t *testing.T) {
	log := store.OutOfGasLog(120, 100, 200, "simulation", true)
	require.Equal(t, "gas used (120) exceeds tx's gas wanted (100) during operation: simulation; simulate with consensus maximum (200) to get real transaction usage", log)
}

func TestOutOfGasLogMaxBlockGas(t *testing.T) {
	log := store.OutOfGasLog(120, 100, 100, "simulation", true)
	require.Equal(t, "gas used (120) exceeds max block gas (100) during operation: simulation", log)
}

func TestOutOfGasLogMaxBlockGasWhenWantedHigher(t *testing.T) {
	log := store.OutOfGasLog(120, 150, 100, "simulation", true)
	require.Equal(t, "gas used (120) exceeds max block gas (100) during operation: simulation", log)
}

func TestOutOfGasLogNoConsensusMaxGas(t *testing.T) {
	log := store.OutOfGasLog(120, 100, -1, "simulation", true)
	require.Equal(t, "gas used (120) exceeds tx's gas wanted (100) during operation: simulation", log)
}

func TestOutOfGasLogNoSimulateHintWhenDisabled(t *testing.T) {
	log := store.OutOfGasLog(120, 100, 200, "simulation", false)
	require.Equal(t, "gas used (120) exceeds tx's gas wanted (100) during operation: simulation", log)
}

func TestTxWithGasWanted(t *testing.T) {
	t.Parallel()

	newTx := func(gasWanted int64) *std.Tx {
		return &std.Tx{
			Fee:        std.NewFee(gasWanted, std.MustParseCoin("1ugnot")),
			Signatures: []std.Signature{{Signature: []byte("original")}},
		}
	}

	t.Run("raises GasWanted to maxGas", func(t *testing.T) {
		t.Parallel()
		simTx, rewritten := txWithGasWanted(newTx(10), 25)
		require.True(t, rewritten)
		require.Equal(t, int64(25), simTx.Fee.GasWanted)
	})

	t.Run("no rewrite when already at or above maxGas", func(t *testing.T) {
		t.Parallel()
		_, rewritten := txWithGasWanted(newTx(25), 25)
		require.False(t, rewritten)
	})

	t.Run("maxGas -1 falls back to the max", func(t *testing.T) {
		t.Parallel()
		simTx, rewritten := txWithGasWanted(newTx(10), -1)
		require.True(t, rewritten)
		require.Equal(t, simulationMaxGasFallback, simTx.Fee.GasWanted)
	})

	t.Run("maxGas 0 means unknown, so no rewrite", func(t *testing.T) {
		t.Parallel()
		_, rewritten := txWithGasWanted(newTx(10), 0)
		require.False(t, rewritten)
	})

	t.Run("does not mutate the original", func(t *testing.T) {
		t.Parallel()
		tx := newTx(10)
		simTx, _ := txWithGasWanted(tx, 25)
		require.Equal(t, int64(10), tx.Fee.GasWanted)
		// The copy inherits a signature that no longer matches it; callers
		// that simulate against a chain which verifies signatures must
		// clear and re-sign, which is what maketx does.
		require.Equal(t, tx.Signatures, simTx.Signatures)
	})
}

// TestTxNeedsSimulationSignature pins which transactions get a second signature
// when simulated.
//
// Signing the rewritten simulation transaction is necessary for messages the
// chain checks, and harmful for the rest: it produces a second broadcastable
// transaction and, on a hardware wallet, a second prompt showing a different
// GasWanted. Before this, every maketx paid that cost.
func TestTxNeedsSimulationSignature(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		route string
		typ   string
		want  bool
	}{
		{"add_package carries source", "vm", "add_package", true},
		{"run carries source", "vm", "run", true},
		{"enable_package compiles stored source", "vm", "enable_package", true},
		{"disable_package is gated the same way", "vm", "disable_package", true},
		{"an ordinary call is not", "vm", "call", false},
		{"a bank send is not", "bank", "send", false},
		{"a vm route alone is not enough", "vm", "something_new", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tx := std.Tx{Msgs: []std.Msg{stubMsg{route: tt.route, typ: tt.typ}}}
			assert.Equal(t, tt.want, txNeedsSimulationSignature(tx))
		})
	}

	t.Run("found behind another message", func(t *testing.T) {
		t.Parallel()
		tx := std.Tx{Msgs: []std.Msg{
			stubMsg{route: "bank", typ: "send"},
			stubMsg{route: "vm", typ: "add_package"},
		}}
		assert.True(t, txNeedsSimulationSignature(tx),
			"the whole transaction is signed once, so any code-bearing message counts")
	})

	t.Run("no messages", func(t *testing.T) {
		t.Parallel()
		assert.False(t, txNeedsSimulationSignature(std.Tx{}))
	})
}

// stubMsg is the smallest thing that satisfies std.Msg for this table.
type stubMsg struct {
	route string
	typ   string
}

func (m stubMsg) Route() string                { return m.route }
func (m stubMsg) Type() string                 { return m.typ }
func (m stubMsg) ValidateBasic() error         { return nil }
func (m stubMsg) GetSignBytes() []byte         { return nil }
func (m stubMsg) GetSigners() []crypto.Address { return nil }
