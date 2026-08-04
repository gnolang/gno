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

	// Seed through the keeper, not by writing the account object directly: "foo"
	// is a split-tier denom, so putting it in the account object builds a state
	// the keeper cannot produce, and the assertion would then hold even if tier
	// routing were entirely broken.
	require.NoError(t, env.bankk.SetCoins(env.ctx, addr, std.NewCoins(std.NewCoin("foo", 10))))
	res = h.Query(env.ctx, req)
	require.Nil(t, res.Error)
	require.NotNil(t, res)
	require.NoError(t, amino.UnmarshalJSON(res.Data, &coins))
	require.True(t, coins.AmountOf("foo") == 10)
}

// A malformed address must report an error and carry no data. The error alone is not
// enough to pin the fix: it was set before the fix too, and only the fall-through to
// the success path — which populated Data from a GetCoins on the zero address — was
// removed. Asserting Data is empty is what makes this test fail without the return.
func TestBalancesRejectsMalformedAddress(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.bankk)
	res := h.Query(env.ctx, abci.RequestQuery{
		Path: "bank/balances/not-a-bech32-address",
	})
	require.NotNil(t, res.Error, "a malformed address must report an error")
	require.Empty(t, res.Data,
		"a malformed address must not also carry a balance in Data")
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
