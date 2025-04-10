package auth

import (
	"bytes"
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

	h := NewHandler(AccountKeeper{}, GasPriceKeeper{})
	res := h.Process(sdk.NewContext(sdk.RunTxModeDeliver, nil, &bft.Header{ChainID: "test-chain"}, nil), tu.NewTestMsg())
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized auth message type"))
}

func TestQueryAccount(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.acck, env.gk)
	_, pubkey, addr := tu.KeyTestPubAddr()

	acc := env.acck.NewAccountWithAddress(env.ctx, addr)

	req := abci.RequestQuery{
		Path: fmt.Sprintf("auth/%s/%s", QueryAccount, addr),
		Data: []byte{},
	}
	res := h.Query(env.ctx, req)
	require.Nil(t, res.Error) // no account is set and no error returned anyway
	require.NotNil(t, res)

	// set account
	acc.SetAccountNumber(uint64(1))
	acc.SetSequence(uint64(20))
	acc.SetCoins(std.NewCoins(std.NewCoin("foo", 10)))
	acc.SetPubKey(pubkey)
	env.acck.SetAccount(env.ctx, acc)

	res = h.Query(env.ctx, req)
	require.Nil(t, res.Error)
	require.NotNil(t, res)
	bz, err := amino.MarshalJSONIndent(acc, "", "  ")
	require.NoError(t, err)

	require.True(t, bytes.Equal(res.Data, bz))
}

func TestQueryGasPrice(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.acck, env.gk)

	req := abci.RequestQuery{
		Path: fmt.Sprintf("auth/%s", QueryGasPrice),
		Data: []byte{},
	}

	res := h.Query(env.ctx, req)
	require.Nil(t, res.Error) // no gasprice is set and no error returned anyway
	require.NotNil(t, res)
	gp := std.GasPrice{
		Gas: 100,
		Price: std.Coin{
			Denom:  "token",
			Amount: 10,
		},
	}
	env.gk.SetGasPrice(env.ctx, gp)

	var gp2 std.GasPrice
	res = h.Query(env.ctx, req)
	require.Nil(t, res.Error)
	require.NotNil(t, res)
	require.NoError(t, amino.UnmarshalJSON(res.Data, &gp2))
	require.True(t, gp == gp2)
}

func TestQuerierRouteNotFound(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.acck, env.gk)
	req := abci.RequestQuery{
		Path: "auth/notexist",
		Data: []byte{},
	}
	res := h.Query(env.ctx, req)
	require.Error(t, res.Error)
}
