package client

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestHandleDeliverResultCallsOnFailure(t *testing.T) {
	called := false
	cfg := &BaseCfg{BaseOptions: BaseOptions{OnTxFailure: func(tx std.Tx, res *ctypes.ResultBroadcastTxCommit) {
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

	simBz, simGasWanted, err := buildSimulationTxBytes(&tx, bz, 25)
	require.NoError(t, err)
	require.Equal(t, int64(25), simGasWanted)

	var simTx std.Tx
	require.NoError(t, amino.Unmarshal(simBz, &simTx))
	require.Equal(t, int64(25), simTx.Fee.GasWanted)
}

func TestBuildSimulationTxBytesUsesFallbackWhenConsensusMaxGasUndefined(t *testing.T) {
	tx := std.Tx{Fee: std.Fee{GasWanted: 10}}
	bz, err := amino.Marshal(&tx)
	require.NoError(t, err)

	simBz, simGasWanted, err := buildSimulationTxBytes(&tx, bz, -1)
	require.NoError(t, err)
	require.Equal(t, simulationMaxGasFallback, simGasWanted)

	var simTx std.Tx
	require.NoError(t, amino.Unmarshal(simBz, &simTx))
	require.Equal(t, simulationMaxGasFallback, simTx.Fee.GasWanted)
}

func TestBuildSimulationTxBytesKeepsHigherOriginalGasWanted(t *testing.T) {
	tx := std.Tx{Fee: std.Fee{GasWanted: 100}}
	bz, err := amino.Marshal(&tx)
	require.NoError(t, err)

	simBz, simGasWanted, err := buildSimulationTxBytes(&tx, bz, 25)
	require.NoError(t, err)
	require.Equal(t, int64(100), simGasWanted)
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
