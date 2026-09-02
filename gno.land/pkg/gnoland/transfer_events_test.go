package gnoland

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestDeliverTxIncludesMsgRunSendEvent(t *testing.T) {
	t.Parallel()

	key := getDummyKey(t)
	addr := key.PubKey().Address()
	send := std.NewCoins(std.NewCoin("ugnot", 42))
	_, deliver := inertChain(t, vm.DefaultGenesisState(), []crypto.PrivKey{key})

	res := deliver(t, []std.Msg{vm.NewMsgRun(addr, send, []*std.MemFile{{
		Name: "main.gno",
		Body: "package main; func main() {}",
	}})}, key)

	require.True(t, res.IsOK(), res.Log)
	require.Equal(t, []abci.Event{bank.TransferEvent{
		From: addr.String(), To: addr.String(), Amount: send,
	}}, res.Events)
}
