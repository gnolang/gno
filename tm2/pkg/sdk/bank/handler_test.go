package bank

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestInvalidMsg(t *testing.T) {
	t.Parallel()

	h := NewHandler(BankKeeper{})
	res := h.Process(sdk.NewContext(sdk.RunTxModeDeliver, nil, &bft.Header{ChainID: "test-chain"}, nil), tu.NewTestMsg())
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized bank message type"))
}

func TestBalances(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.bankk)
	_, _, addr := tu.KeyTestPubAddr()

	req := abci.RequestQuery{
		Path: fmt.Sprintf("bank/%s/%s", QueryBalance, addr.String()),
		Data: []byte{},
	}

	res := h.Query(env.ctx, req)
	require.Nil(t, res.Error) // the account does not exist, no error returned anyway
	require.NotNil(t, res)

	var coins std.Coins
	require.NoError(t, amino.UnmarshalJSON(res.Data, &coins))
	require.True(t, coins.IsZero())

	acc := env.acck.NewAccountWithAddress(env.ctx, addr)
	acc.SetCoins(std.NewCoins(std.NewCoin("foo", 10)))
	env.acck.SetAccount(env.ctx, acc)
	res = h.Query(env.ctx, req)
	require.Nil(t, res.Error)
	require.NotNil(t, res)
	require.NoError(t, amino.UnmarshalJSON(res.Data, &coins))
	require.True(t, coins.AmountOf("foo") == 10)
}

func TestQuerierRouteNotFound(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.bankk)
	req := abci.RequestQuery{
		Path: "bank/notfound",
		Data: []byte{},
	}
	res := h.Query(env.ctx, req)
	require.Error(t, res.Error)
}
