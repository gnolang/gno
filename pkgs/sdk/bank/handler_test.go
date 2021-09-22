package bank

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/pkgs/amino"
	abci "github.com/gnolang/gno/pkgs/bft/abci/types"
	bft "github.com/gnolang/gno/pkgs/bft/types"
	"github.com/gnolang/gno/pkgs/sdk"
	tu "github.com/gnolang/gno/pkgs/sdk/testutils"
	"github.com/gnolang/gno/pkgs/std"
)

func TestInvalidMsg(t *testing.T) {
	h := NewHandler(BankKeeper{})
	res := h.Process(sdk.NewContext(nil, &bft.Header{}, false, nil), tu.NewTestMsg())
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized bank message type"))
}

func TestBalances(t *testing.T) {
	env := setupTestEnv()
	h := NewHandler(env.bank)
	req := abci.RequestQuery{
		Path: fmt.Sprintf("bank/%s", QueryBalance),
		Data: []byte{},
	}

	res := h.Query(env.ctx, req)
	require.NotNil(t, res.Error)

	_, _, addr := tu.KeyTestPubAddr()
	req.Data = amino.MustMarshalJSON(NewQueryBalanceParams(addr))
	res = h.Query(env.ctx, req)
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
	env := setupTestEnv()
	h := NewHandler(env.bank)
	req := abci.RequestQuery{
		Path: "bank/notfound",
		Data: []byte{},
	}
	res := h.Query(env.ctx, req)
	require.Error(t, res.Error)
}
