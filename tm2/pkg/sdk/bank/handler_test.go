package bank

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
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

func TestHandlerEmitsTransferEvents(t *testing.T) {
	t.Parallel()

	t.Run("send", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		from := crypto.AddressFromPreimage([]byte("handler-send-from"))
		to := crypto.AddressFromPreimage([]byte("handler-send-to"))
		amount := std.NewCoins(std.NewCoin("ugnot", 5))
		require.NoError(t, env.bankk.SetCoins(env.ctx, from, amount))

		res := NewHandler(env.bankk).Process(env.ctx, NewMsgSend(from, to, amount))
		require.True(t, res.IsOK(), res.Log)
		require.Equal(t, []sdk.Event{TransferEvent{
			From: from.String(), To: to.String(), Amount: amount,
		}}, env.ctx.EventLogger().Events())
	})

	t.Run("multisend", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		from := crypto.AddressFromPreimage([]byte("handler-multisend-from"))
		to := crypto.AddressFromPreimage([]byte("handler-multisend-to"))
		amount := std.NewCoins(std.NewCoin("ugnot", 5))
		require.NoError(t, env.bankk.SetCoins(env.ctx, from, amount))

		res := NewHandler(env.bankk).Process(env.ctx, NewMsgMultiSend(
			[]Input{NewInput(from, amount)}, []Output{NewOutput(to, amount)},
		))
		require.True(t, res.IsOK(), res.Log)
		require.Equal(t, []sdk.Event{TransferEvent{
			To: to.String(), Amount: amount,
		}}, env.ctx.EventLogger().Events())
	})
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

func TestQuerySupply(t *testing.T) {
	t.Parallel()

	addr := crypto.AddressFromPreimage([]byte("holder"))

	// Each subtest gets its own env: the burn case mutates supply, and sharing one
	// env would make the outcomes depend on execution order.
	minted := func(t *testing.T) func(string) abci.ResponseQuery {
		t.Helper()
		env := setupTestEnv()
		require.NoError(t, env.bankk.MintCoins(env.ctx, addr, std.Coins{
			{Denom: testRealmDenom, Amount: 42},
			{Denom: testAccountDenom, Amount: 1000},
		}))
		h := NewHandler(env.bankk)
		return func(path string) abci.ResponseQuery {
			return h.Query(env.ctx, abci.RequestQuery{Path: path})
		}
	}

	t.Run("a simple denom", func(t *testing.T) {
		t.Parallel()
		res := minted(t)("bank/supply/" + testAccountDenom)
		require.Nil(t, res.Error)
		// amino renders int64 as a quoted string
		require.Equal(t, `"1000"`, string(res.Data))
	})

	// The case a path split cannot handle: a realm denom contains slashes, which is
	// why the handler takes the path remainder rather than a component.
	t.Run("a realm denom with slashes", func(t *testing.T) {
		t.Parallel()
		require.Contains(t, testRealmDenom, "/", "the point of this case")
		res := minted(t)("bank/supply/" + testRealmDenom)
		require.Nil(t, res.Error)
		require.Equal(t, `"42"`, string(res.Data))
	})

	t.Run("a denom nobody holds is zero, not an error", func(t *testing.T) {
		t.Parallel()
		res := minted(t)("bank/supply/atom")
		require.Nil(t, res.Error)
		require.Equal(t, `"0"`, string(res.Data))
	})

	// The distinct message is the whole point of the empty check: ValidateDenom would
	// reject "" anyway, so asserting only that an error came back cannot tell whether
	// the caller is told what to do about it.
	t.Run("a missing denom says what to supply", func(t *testing.T) {
		t.Parallel()
		query := minted(t)
		for _, path := range []string{"bank/supply", "bank/supply/"} {
			res := query(path)
			require.NotNil(t, res.Error, "path %q must be rejected", path)
			require.Contains(t, res.Log, "requires a denom",
				"path %q must say what is missing, not just that it is invalid", path)
			require.Empty(t, res.Data)
		}
	})

	t.Run("a malformed denom is named, not silently zero", func(t *testing.T) {
		t.Parallel()
		query := minted(t)
		for _, denom := range []string{"UPPER", "a b", strings.Repeat("z", 275)} {
			res := query("bank/supply/" + denom)
			require.NotNil(t, res.Error, "denom %q must be rejected", denom)
			require.Empty(t, res.Data)
		}
	})

	t.Run("burning to zero reports zero", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		require.NoError(t, env.bankk.MintCoins(env.ctx, addr,
			std.Coins{{Denom: testRealmDenom, Amount: 42}}))
		require.NoError(t, env.bankk.BurnCoins(env.ctx, addr,
			std.Coins{{Denom: testRealmDenom, Amount: 42}}))
		res := NewHandler(env.bankk).Query(env.ctx,
			abci.RequestQuery{Path: "bank/supply/" + testRealmDenom})
		require.Nil(t, res.Error)
		require.Equal(t, `"0"`, string(res.Data))
	})
}

func TestPathRemainder(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ path, want string }{
		{"bank/supply/ugnot", "ugnot"},
		{"bank/supply//gno.land/r/x:tok", "/gno.land/r/x:tok"},
		{"bank/supply/", ""},
		{"bank/supply", ""},
		{"bank", ""},
		{"", ""},
	} {
		require.Equal(t, tc.want, pathRemainder(tc.path, 2), "path %q", tc.path)
	}
}
