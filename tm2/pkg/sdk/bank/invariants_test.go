package bank

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// rawSet writes straight to the store, past every keeper guard. Most of the states
// below are unreachable through the public API by design — that is what the guards
// are for — so raw writes are the only way to test that they are reported.
func rawSet(t *testing.T, env testEnv, key, value []byte) {
	t.Helper()
	env.ctx.Store(env.key).Set(nil, key, value)
}

func fundHealthy(t *testing.T, env testEnv, addr crypto.Address) {
	t.Helper()
	require.NoError(t, env.bankk.SetCoins(env.ctx, addr, std.Coins{
		{Denom: testRealmDenom, Amount: 5},
		{Denom: testAccountDenom, Amount: 50},
	}))
}

// State the keeper produces must always be reported healthy. A check that fires on
// legitimate state is worse than no check: it would halt nodes on a working chain.
// Why the key/field agreement check in auth.AccountKeyspaceInvariant is worth its
// cost, measured rather than argued. SetAccount files an account under its own
// GetAddress(), while every reader looks it up by the address it already has, so a
// misfiled object silently redirects writes.
func TestAMisfiledAccountRedirectsCreditsAndBreaksSupply(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx
	a := crypto.AddressFromPreimage([]byte("redirect-a"))
	b := crypto.AddressFromPreimage([]byte("redirect-b"))

	require.NoError(t, env.bankk.SetCoins(ctx, b, std.Coins{{Denom: testAccountDenom, Amount: 1000}}))
	env.bankk.RecomputeSupply(ctx)

	swept := func() int64 {
		totals, err := computeSupply(ctx, env.bankk.ViewKeeper, 0)
		require.NoError(t, err)
		return totals[testAccountDenom]
	}
	require.Equal(t, int64(1000), swept())
	require.Equal(t, int64(1000), env.bankk.TotalSupply(ctx, testAccountDenom))

	// File B's account object under A's key. Reached past the keeper because
	// SetAccount would file it under B again, which is the whole point.
	rawSet(t, env, auth.AddressStoreKey(a), amino.MustMarshalAny(env.acck.GetAccount(ctx, b)))

	// The misfiling alone doubles the swept total: B's balance is now counted at both
	// keys, while the recorded supply is untouched.
	require.Equal(t, int64(2000), swept())
	require.Equal(t, int64(1000), env.bankk.TotalSupply(ctx, testAccountDenom))
	msg, broken := SupplyInvariant(env.bankk.ViewKeeper)(ctx)
	require.True(t, broken, "a doubled sweep must break the supply invariant")
	require.Contains(t, msg, "is held")

	// And an ordinary credit to A lands on B, because the object A's key holds says
	// its address is B.
	require.NoError(t, env.bankk.AddCoins(ctx, a, std.Coins{{Denom: testAccountDenom, Amount: 1}}))
	require.Equal(t, int64(1001), env.bankk.GetCoin(ctx, b, testAccountDenom),
		"the credit to A must be observed landing on B")
	require.Equal(t, int64(1000), env.bankk.GetCoin(ctx, a, testAccountDenom),
		"A gains nothing from being credited")
	require.Equal(t, int64(2001), swept())

	msg, broken = auth.AccountKeyspaceInvariant(env.acck)(ctx)
	require.True(t, broken)
	require.Contains(t, msg, "would be filed under")
}

// The invariants ADR fixes a rule: no invariant may be breakable by an unprivileged
// actor, because a periodic check that halts the local node is otherwise a remote
// DoS. It names the state that got a proposed check dropped — a session address that
// is also a regular account, which anyone induces by sending coins there, since
// AddCoins creates the recipient's account object.
//
// This pins that the shipped set tolerates it. The transfer is the reachable path:
// a bare AddCoins would create coins from nothing and break the supply invariant,
// but no external caller has one — it is absent from vm.BankKeeperI and no message
// handler exposes it.
func TestCreditingASessionAddressBreaksNoInvariant(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	ctx := env.ctx

	master := crypto.AddressFromPreimage([]byte("dos-master"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, master))
	_, sessPub, _ := tu.KeyTestPubAddr()
	sess := env.acck.NewSessionAccount(ctx, master, sessPub)
	env.acck.SetSessionAccount(ctx, master, sess)
	sessAddr := sess.GetAddress()

	payer := crypto.AddressFromPreimage([]byte("dos-payer"))
	require.NoError(t, env.bankk.MintCoins(ctx, payer, std.Coins{{Denom: testAccountDenom, Amount: 100}}))
	env.bankk.RecomputeSupply(ctx)

	require.NoError(t, env.bankk.SendCoins(ctx, payer, sessAddr,
		std.Coins{{Denom: testAccountDenom, Amount: 1}}))
	require.NotNil(t, env.acck.GetAccount(ctx, sessAddr),
		"precondition: the credit must have created a regular account at the session address")

	for name, inv := range map[string]sdk.Invariant{
		"auth": auth.AllInvariants(env.acck),
		"bank": AllInvariants(env.bankk.ViewKeeper),
	} {
		msg, broken := inv(ctx)
		require.False(t, broken, "%s invariants must tolerate a state anyone can induce:\n%s", name, msg)
	}
}

func TestInvariantsHealthyOnKeeperState(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	a := crypto.AddressFromPreimage([]byte("a"))
	b := crypto.AddressFromPreimage([]byte("b"))
	fundHealthy(t, env, a)
	// fundHealthy uses SetCoins, which is supply-blind by design; seed the counter
	// so the supply invariant has something consistent to check against.
	env.bankk.RecomputeSupply(env.ctx)
	require.NoError(t, env.bankk.MintCoins(env.ctx, b, std.Coins{{Denom: "atom", Amount: 7}}))
	require.NoError(t, env.bankk.SendCoins(env.ctx, a, b, std.Coins{{Denom: testAccountDenom, Amount: 1}}))

	for name, inv := range map[string]func() (string, bool){
		"balance-keys": func() (string, bool) { return BalanceKeysInvariant(env.bankk.ViewKeeper)(env.ctx) },
		"account-tier": func() (string, bool) { return AccountTierInvariant(env.bankk.ViewKeeper)(env.ctx) },
		"auth":         func() (string, bool) { return auth.AllInvariants(env.acck)(env.ctx) },
		"all":          func() (string, bool) { return AllInvariants(env.bankk.ViewKeeper)(env.ctx) },
	} {
		msg, broken := inv()
		require.False(t, broken, "%s reported a violation on healthy state:\n%s", name, msg)
	}
}

func TestBalanceKeysInvariantReportsCorruption(t *testing.T) {
	t.Parallel()

	enc := encodeBalance(5)
	addr := crypto.AddressFromPreimage([]byte("holder"))

	cases := []struct {
		name, want string
		poison     func(t *testing.T, env testEnv)
	}{
		{
			"value of the wrong width", "expected 8 bytes, got 3",
			func(t *testing.T, env testEnv) {
				t.Helper()
				rawSet(t, env, BalanceKey(addr, "atom"), []byte{1, 2, 3})
			},
		},
		{
			// A zero-length value is accepted by the store; only nil is refused.
			"zero-length value", "expected 8 bytes, got 0",
			func(t *testing.T, env testEnv) {
				t.Helper()
				rawSet(t, env, BalanceKey(addr, "atom"), []byte{})
			},
		},
		{
			"stored zero", "0 out of range",
			func(t *testing.T, env testEnv) {
				t.Helper()
				rawSet(t, env, BalanceKey(addr, "atom"), make([]byte, 8))
			},
		},
		{
			"amount over MaxInt64", "out of range",
			func(t *testing.T, env testEnv) {
				t.Helper()
				rawSet(t, env, BalanceKey(addr, "atom"), []byte{255, 255, 255, 255, 255, 255, 255, 255})
			},
		},
		{
			"key too short to hold an address", "malformed balance key",
			func(t *testing.T, env testEnv) {
				t.Helper()
				rawSet(t, env, append([]byte(BalancePrefix), addr[:9]...), enc)
			},
		},
		{
			"key with no denom", "malformed balance key",
			func(t *testing.T, env testEnv) {
				t.Helper()
				rawSet(t, env, BalancePrefixKey(addr), enc)
			},
		},
		{
			"invalid denom", "invalid denom",
			func(t *testing.T, env testEnv) {
				t.Helper()
				rawSet(t, env, BalanceKey(addr, "UPPER"), enc)
			},
		},
		{
			// The allowlist grew without migrating: the denom now belongs in the
			// account object but an old split key survives.
			"split key for an account-tier denom", "the allowlist grew",
			func(t *testing.T, env testEnv) {
				t.Helper()
				env.bankk.setSplitBalance(env.ctx, addr, testAccountDenom, 500)
			},
		},
		{
			// Realm-shaped but unissuable: base name over the 16-byte limit.
			"realm-shaped denom no realm could issue", "no realm could issue",
			func(t *testing.T, env testEnv) {
				t.Helper()
				rawSet(t, env, BalanceKey(addr, "/gno.land/r/x:"+strings.Repeat("z", 17)), enc)
			},
		},
		{
			"realm-shaped denom with no base name", "no realm could issue",
			func(t *testing.T, env testEnv) {
				t.Helper()
				rawSet(t, env, BalanceKey(addr, "/nocolon"), enc)
			},
		},
		{
			// Every keeper credit path calls ensureAccount, so only a raw write
			// can leave a balance with no account to sign for it.
			"balance with no account object", "no account object",
			func(t *testing.T, env testEnv) {
				t.Helper()
				rawSet(t, env, BalanceKey(crypto.AddressFromPreimage([]byte("ghost")), "atom"), enc)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			env := setupTestEnv()
			fundHealthy(t, env, addr)
			tc.poison(t, env)

			var msg string
			var broken bool
			require.NotPanics(t, func() {
				msg, broken = BalanceKeysInvariant(env.bankk.ViewKeeper)(env.ctx)
			}, "an invariant must report corruption, never panic on it")
			require.True(t, broken, "not reported")
			require.Contains(t, msg, tc.want)
		})
	}
}

func TestAccountTierInvariantReportsStrandedDenoms(t *testing.T) {
	t.Parallel()

	addr := crypto.AddressFromPreimage([]byte("holder"))

	t.Run("denom stranded by a shrunk allowlist", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		fundHealthy(t, env, addr)
		// No raw writes: exactly the state a binary whose allowlist shrank would
		// produce — the ADR's documented "upgrade without replay". The gas denom is
		// already in the account object; under the new allowlist a further credit
		// routes to a split key, so the same denom ends up in both homes.
		shrunkBank := NewBankKeeper(env.acck, env.prmk.ForModule(ModuleName), env.key, nil)
		require.NoError(t, shrunkBank.AddCoins(env.ctx, addr,
			std.Coins{{Denom: testAccountDenom, Amount: 1}}))

		msg, broken := AccountTierInvariant(shrunkBank.ViewKeeper)(env.ctx)
		require.True(t, broken)
		require.Contains(t, msg, "not in the account-tier allowlist")
		require.Contains(t, msg, "double-homed", "the split key for the same denom must be noted")
	})

	t.Run("realm-issuable denom in the account object", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		acc := env.acck.NewAccountWithAddress(env.ctx, addr)
		require.NoError(t, acc.SetCoins(std.Coins{{Denom: testRealmDenom, Amount: 3}}))
		env.acck.SetAccount(env.ctx, acc)
		msg, broken := AccountTierInvariant(env.bankk.ViewKeeper)(env.ctx)
		require.True(t, broken)
		require.Contains(t, msg, "realm-issuable and must never be in the account tier")
	})

	t.Run("invalid vesting schedule", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		// Bypass the constructors, which reject this. Read from the concrete type:
		// the interface's GetStartTime is hardcoded to 0 for delayed accounts.
		env.acck.SetAccount(env.ctx, &std.DelayedVestingAccount{
			BaseVestingAccount: std.BaseVestingAccount{
				BaseAccount: *std.NewBaseAccount(addr, nil, nil, 0, 0),
				VestingSchedule: std.VestingSchedule{
					OriginalVesting: std.Coins{{Denom: testAccountDenom, Amount: 1}},
					StartTime:       300,
					EndTime:         100,
					Type:            std.VestingDelayed,
				},
			},
		})
		msg, broken := AccountTierInvariant(env.bankk.ViewKeeper)(env.ctx)
		require.True(t, broken)
		require.Contains(t, msg, "invalid vesting schedule")
	})
}
