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

// A malformed address must report an error. Before the fix the error was written
// into res and then overwritten by the success path, so a bad address came back as
// an empty balance — indistinguishable from an address holding nothing.
func TestBalancesRejectsMalformedAddress(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	h := NewHandler(env.bankk)
	res := h.Query(env.ctx, abci.RequestQuery{
		Path: "bank/balances/not-a-bech32-address",
	})
	require.NotNil(t, res.Error, "a malformed address must not report an empty balance")
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
