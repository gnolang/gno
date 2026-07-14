package client

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

func TestRequiredGasFee(t *testing.T) {
	tests := []struct {
		name      string
		gasWanted int64
		gasPrice  std.GasPrice
		want      int64
	}{
		{"small gas with large denominator", 1, std.GasPrice{Gas: 1_000_000, Price: std.Coin{Amount: 1000}}, 1},
		{"exact division", 2000, std.GasPrice{Gas: 1_000_000, Price: std.Coin{Amount: 1000}}, 2},
		{"remainder division", 2001, std.GasPrice{Gas: 1_000_000, Price: std.Coin{Amount: 1000}}, 3},
		{"large multiply", 3_000_000_000, std.GasPrice{Gas: 1_000_000, Price: std.Coin{Amount: 1_000_000}}, 3_000_000_000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, requiredGasFee(tt.gasWanted, tt.gasPrice))
		})
	}
}

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
