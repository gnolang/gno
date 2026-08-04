package client

import (
	"bytes"
	"errors"
	"flag"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
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

func TestMakeTxCfgGasProfileAcceptsDocumentedInvocation(t *testing.T) {
	t.Parallel()

	// Parse through RegisterFlags rather than a struct literal: -broadcast
	// defaults to true, so a struct-literal config represents a state the CLI
	// never produces. A guard rejecting "-gasprofile with -broadcast" would
	// look fine against a literal while rejecting every documented command.
	cfg := &MakeTxCfg{RootCfg: &BaseCfg{}}
	fs := flag.NewFlagSet("maketx", flag.ContinueOnError)
	cfg.RegisterFlags(fs)
	require.NoError(t, fs.Parse([]string{"-gasprofile", "gas.pprof"}))

	require.True(t, cfg.Broadcast, "-broadcast defaults to true")
	require.Equal(t, "gas.pprof", cfg.GasProfile)
	require.True(t, cfg.ShouldSign(), "-gasprofile must sign the tx")
	require.NoError(t, cfg.Validate(), "the documented -gasprofile invocation must be accepted")
}
