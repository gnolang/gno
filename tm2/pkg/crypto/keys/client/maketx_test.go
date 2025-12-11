package client

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

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
