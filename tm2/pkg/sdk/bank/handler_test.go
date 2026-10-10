package bank

import (
	"fmt"
	"strings"
	"testing"
	"time"

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

	env := setupTestEnv()
	from := crypto.AddressFromPreimage([]byte("handler-send-from"))
	to := crypto.AddressFromPreimage([]byte("handler-send-to"))
	amount := std.NewCoins(std.NewCoin("ugnot", 5))
	require.NoError(t, env.bankk.SetCoins(env.ctx, from, amount))

	res := NewHandler(env.bankk).Process(env.ctx, NewMsgSend(from, to, amount))
	require.True(t, res.IsOK(), res.Log)
	require.Equal(t, []sdk.Event{TransferEvent{
		From: from.String(), To: to.String(), Coins: amount,
	}}, env.ctx.EventLogger().Events())
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

// envAt is a test environment whose block clock reads blockTime. Each caller
// builds its own: the handler writes nothing, but the store behind the keeper
// is not safe for concurrent use and these tests run in parallel.
func envAt(t *testing.T, blockTime int64) (testEnv, sdk.Context) {
	t.Helper()

	env := setupTestEnv()
	return env, env.ctx.WithBlockHeader(&bft.Header{
		Height:  1,
		ChainID: "test-chain-id",
		Time:    time.Unix(blockTime, 0),
	})
}

// spendableEnv is an account holding total, with schedule applied, queried at
// blockTime. total goes through SetCoins so split-tier denoms land in the tier
// the keeper would put them in, not all in the account object.
func spendableEnv(t *testing.T, total std.Coins, schedule std.VestingSchedule, blockTime int64) (
	bankHandler, sdk.Context, crypto.Address,
) {
	t.Helper()

	env, ctx := envAt(t, blockTime)

	_, _, addr := tu.KeyTestPubAddr()
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	acc.SetVesting(schedule)
	env.acck.SetAccount(ctx, acc)
	require.NoError(t, env.bankk.SetCoins(ctx, addr, total))

	return NewHandler(env.bankk), ctx, addr
}

func querySpendableAt(t *testing.T, h bankHandler, ctx sdk.Context, addr crypto.Address) AccountSpendable {
	t.Helper()

	res := h.Query(ctx, abci.RequestQuery{
		Path: fmt.Sprintf("bank/%s/%s", QuerySpendable, addr),
	})
	require.Nil(t, res.Error)

	var got AccountSpendable
	require.NoError(t, amino.UnmarshalJSON(res.Data, &got))
	return got
}

// The point of the endpoint: it evaluates the curve, rather than handing back
// the schedule for the caller to evaluate. Mid-schedule is the only case where
// a wrong implementation is visibly wrong -- at either end every implementation
// agrees.
func TestQuerySpendableMidSchedule(t *testing.T) {
	t.Parallel()

	const start, end = 1_000_000, 1_000_000 + 1000
	h, ctx, addr := spendableEnv(t,
		std.NewCoins(std.NewCoin(testAccountDenom, 10_000)),
		std.VestingSchedule{
			OriginalVesting: std.NewCoins(std.NewCoin(testAccountDenom, 8_000)),
			StartTime:       start,
			EndTime:         end,
		},
		start+250, // a quarter through
	)

	got := querySpendableAt(t, h, ctx, addr)

	// 8000 granted, a quarter vested = 2000; 6000 still locked; the 2000 that
	// was never part of the grant is spendable throughout.
	require.Equal(t, int64(10_000), got.Coins.AmountOf(testAccountDenom))
	require.Equal(t, int64(6_000), got.Locked.AmountOf(testAccountDenom))
	require.Equal(t, int64(4_000), got.Spendable.AmountOf(testAccountDenom))
	require.Equal(t, int64(start+250), got.BlockTime)
}

// THE REASON THIS LIVES IN bank. A schedule may name a denom that lives outside
// the account object -- gno.land accepts one from a genesis balances file, and
// SubtractCoins is deliberately tier-agnostic about enforcing it. Reading only
// the account tier would report the split-tier denom as locked and omit it from
// spendable, telling a caller nothing can move while a transfer would be allowed.
func TestQuerySpendableCoversSplitTierDenoms(t *testing.T) {
	t.Parallel()

	const start, end = 7_000_000, 7_000_000 + 1000
	const otherDenom = "ibc/atom" // not in accountTierTestDenoms

	h, ctx, addr := spendableEnv(t,
		std.NewCoins(
			std.NewCoin(testAccountDenom, 500),
			std.NewCoin(otherDenom, 1_000),
		),
		std.VestingSchedule{
			OriginalVesting: std.NewCoins(std.NewCoin(otherDenom, 800)),
			StartTime:       start,
			EndTime:         end,
		},
		start+250, // a quarter of the 800 vested = 200; 600 locked
	)

	got := querySpendableAt(t, h, ctx, addr)

	// Guard the guard: if the denom were account-tier this would prove nothing.
	require.NotContains(t, accountTierTestDenoms, otherDenom)

	require.Equal(t, int64(1_000), got.Coins.AmountOf(otherDenom),
		"the split-tier balance must be reported, not silently zero")
	require.Equal(t, int64(600), got.Locked.AmountOf(otherDenom))
	require.Equal(t, int64(400), got.Spendable.AmountOf(otherDenom),
		"a split-tier denom must appear in spendable, not be omitted as if fully locked")

	// The untouched gas denom rides along unaffected.
	require.Equal(t, int64(500), got.Spendable.AmountOf(testAccountDenom))
}

// The endpoint must agree with the transfer path, or it is worse than useless:
// a caller would size a transfer by it and have the transfer rejected. This
// pins them to the same function rather than to each other's arithmetic.
func TestQuerySpendableMatchesLockedCoins(t *testing.T) {
	t.Parallel()

	const start, end = 2_000_000, 2_000_000 + 3600
	total := std.NewCoins(std.NewCoin(testAccountDenom, 1_000_000))
	schedule := std.VestingSchedule{
		OriginalVesting: std.NewCoins(std.NewCoin(testAccountDenom, 900_000)),
		StartTime:       start,
		EndTime:         end,
	}

	for _, offset := range []int64{-1, 0, 1, 900, 1800, 3599, 3600, 7200} {
		h, ctx, addr := spendableEnv(t, total, schedule, start+offset)
		got := querySpendableAt(t, h, ctx, addr)

		want := schedule.LockedCoins(time.Unix(start+offset, 0))
		require.Equal(t, want.AmountOf(testAccountDenom), got.Locked.AmountOf(testAccountDenom),
			"offset %d: locked disagrees with std.VestingSchedule.LockedCoins", offset)
		require.Equal(t, total.AmountOf(testAccountDenom)-want.AmountOf(testAccountDenom),
			got.Spendable.AmountOf(testAccountDenom), "offset %d: spendable", offset)
	}
}

// A delayed schedule vests nothing until EndTime and everything after, so it
// separates the two curves: a continuous implementation would report a
// partially-vested amount at the same instant.
func TestQuerySpendableDelayedIsACliff(t *testing.T) {
	t.Parallel()

	const start, end = 3_000_000, 3_000_000 + 1000
	schedule := std.VestingSchedule{
		OriginalVesting: std.NewCoins(std.NewCoin(testAccountDenom, 500)),
		EndTime:         end,
		Type:            std.VestingDelayed,
	}

	for _, tc := range []struct {
		name              string
		at                int64
		locked, spendable int64
	}{
		{"just before the cliff", end - 1, 500, 100},
		{"at the cliff", end, 0, 600},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h, ctx, addr := spendableEnv(t,
				std.NewCoins(std.NewCoin(testAccountDenom, 600)), schedule, tc.at)
			got := querySpendableAt(t, h, ctx, addr)

			require.Equal(t, tc.locked, got.Locked.AmountOf(testAccountDenom))
			require.Equal(t, tc.spendable, got.Spendable.AmountOf(testAccountDenom))
		})
	}

	// Guard the guard, AT THE INSTANT THE SUBTESTS ASSERT: one tick before the
	// cliff a continuous schedule has all but released the grant while the
	// delayed one still locks every unit of it. Compared anywhere else this
	// would not protect the assertions above.
	continuous := schedule
	continuous.StartTime = start
	continuous.Type = std.VestingContinuous
	require.NotEqual(t,
		schedule.LockedCoins(time.Unix(end-1, 0)).AmountOf(testAccountDenom),
		continuous.LockedCoins(time.Unix(end-1, 0)).AmountOf(testAccountDenom),
		"the two curves agree at end-1, so the delayed assertions prove nothing")
}

// A schedule can lock more than the account still holds; see querySpendable for
// why, and why an unclamped result reaches the client rather than erroring.
func TestQuerySpendableClampsWhenLockedExceedsBalance(t *testing.T) {
	t.Parallel()

	const start, end = 4_000_000, 4_000_000 + 1000
	h, ctx, addr := spendableEnv(t,
		std.NewCoins(std.NewCoin(testAccountDenom, 10)), // far less than the grant
		std.VestingSchedule{
			OriginalVesting: std.NewCoins(std.NewCoin(testAccountDenom, 1_000)),
			StartTime:       start,
			EndTime:         end,
		},
		start, // nothing vested yet: 1000 locked against a balance of 10
	)

	got := querySpendableAt(t, h, ctx, addr)

	require.Equal(t, int64(1_000), got.Locked.AmountOf(testAccountDenom))
	require.Equal(t, int64(0), got.Spendable.AmountOf(testAccountDenom))
	require.Empty(t, got.Spendable, "Coins carries no zero entries")
}

// An account with no schedule locks nothing. This is almost every account, so
// the endpoint has to be correct and cheap for it.
func TestQuerySpendableWithoutVesting(t *testing.T) {
	t.Parallel()

	h, ctx, addr := spendableEnv(t,
		std.NewCoins(std.NewCoin(testAccountDenom, 42)), std.VestingSchedule{}, 5_000_000)

	got := querySpendableAt(t, h, ctx, addr)

	require.Empty(t, got.Locked)
	require.Equal(t, int64(42), got.Spendable.AmountOf(testAccountDenom))
}

// An address that has never transacted has no account at all. Answering with
// zeros rather than an error spares every caller a special case, and is true.
func TestQuerySpendableUnknownAccount(t *testing.T) {
	t.Parallel()

	env, ctx := envAt(t, 6_000_000)
	h := NewHandler(env.bankk)
	_, _, addr := tu.KeyTestPubAddr()

	got := querySpendableAt(t, h, ctx, addr)

	require.Empty(t, got.Coins)
	require.Empty(t, got.Locked)
	require.Empty(t, got.Spendable)
	require.Equal(t, int64(6_000_000), got.BlockTime)
}

// A malformed address is an error, not an empty answer that reads as "this
// address holds nothing". Data must be empty too, for the reason
// TestBalancesRejectsMalformedAddress spells out: the error alone would still be
// set if execution fell through to the success path.
func TestQuerySpendableRejectsBadAddress(t *testing.T) {
	t.Parallel()

	env, ctx := envAt(t, 6_000_000)
	h := NewHandler(env.bankk)

	res := h.Query(ctx, abci.RequestQuery{
		Path: fmt.Sprintf("bank/%s/not-a-bech32-address", QuerySpendable),
	})
	require.NotNil(t, res.Error, "a malformed address must report an error")
	require.Empty(t, res.Data,
		"a malformed address must not also carry a spendable set in Data")
}
