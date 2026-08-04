package auth

import (
	"github.com/gnolang/gno/tm2/pkg/amino"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// rawSet writes straight to the account store, past every keeper guard.
func rawSet(t *testing.T, env testEnv, key, value []byte) {
	t.Helper()
	env.ctx.Store(env.acck.key).Set(nil, key, value)
}

func TestAccountKeyspaceInvariantHealthy(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	for _, seed := range []string{"a", "b", "c"} {
		acc := env.acck.NewAccountWithAddress(env.ctx, crypto.AddressFromPreimage([]byte(seed)))
		env.acck.SetAccount(env.ctx, acc)
	}
	msg, broken := AccountKeyspaceInvariant(env.acck)(env.ctx)
	require.False(t, broken, "healthy keyspace reported broken:\n%s", msg)
}

func TestAccountKeyspaceInvariantReportsCorruption(t *testing.T) {
	t.Parallel()

	master := crypto.AddressFromPreimage([]byte("master"))
	other := crypto.AddressFromPreimage([]byte("other"))

	cases := []struct {
		name, want string
		poison     func(t *testing.T, env testEnv)
	}{
		{
			// The check both reviewers rated most valuable: SetAccount files an
			// account under its own Address field, so a mismatch redirects writes.
			"account filed under the wrong key", "would be filed under",
			func(t *testing.T, env testEnv) {
				bz := amino.MustMarshalAny(std.NewBaseAccount(other, nil, nil, 0, 0))
				rawSet(t, env, AddressStoreKey(master), bz)
			},
		},
		{
			// Accepted by the store; amino returns a nil account with no error, so
			// "decoded without error" is not enough to call an account usable.
			"zero-length value", "does not decode",
			func(t *testing.T, env testEnv) { rawSet(t, env, AddressStoreKey(master), []byte{}) },
		},
		{
			"undecodable value", "does not decode",
			func(t *testing.T, env testEnv) {
				rawSet(t, env, AddressStoreKey(master), []byte{0xff, 0xfe, 0xfd})
			},
		},
		{
			"key of an unrecognised shape", "unrecognised shape",
			func(t *testing.T, env testEnv) {
				rawSet(t, env, append(AddressStoreKey(master), 0xAA), []byte{1})
			},
		},
		{
			// Session length but no /s/ infix: a length-only classifier would hand
			// this to a session decoder.
			"session-length key without the infix", "unrecognised shape",
			func(t *testing.T, env testEnv) {
				key := append(AddressStoreKey(master), make([]byte, SessionStoreKeyLen-AccountStoreKeyLen)...)
				rawSet(t, env, key, []byte{1})
			},
		},
		{
			"duplicate account number", "used more than once",
			func(t *testing.T, env testEnv) {
				a := env.acck.NewAccountWithAddress(env.ctx, master)
				env.acck.SetAccount(env.ctx, a)
				dup := std.NewBaseAccount(other, nil, nil, a.GetAccountNumber(), 0)
				env.acck.SetAccount(env.ctx, dup)
			},
		},
		{
			"account number at or above the counter", "above the global counter",
			func(t *testing.T, env testEnv) {
				env.acck.SetAccount(env.ctx, std.NewBaseAccount(master, nil, nil, 1<<20, 0))
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			env := setupTestEnv()
			tc.poison(t, env)

			var msg string
			var broken bool
			require.NotPanics(t, func() {
				msg, broken = AccountKeyspaceInvariant(env.acck)(env.ctx)
			}, "an invariant must report corruption, never panic on it")
			require.True(t, broken, "not reported")
			require.Contains(t, msg, tc.want)
		})
	}
}

// The keeper's own accessors panic on state the invariant reports. That asymmetry is
// the reason the invariants read raw rather than going through the keeper.
func TestInvariantReportsWhatIterateAccountsPanicsOn(t *testing.T) {
	t.Parallel()

	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte("poisoned"))
	rawSet(t, env, AddressStoreKey(addr), []byte{0xff, 0xfe, 0xfd})

	require.Panics(t, func() {
		env.acck.IterateAccounts(env.ctx, func(std.Account) bool { return false })
	}, "precondition: the keeper accessor panics on this state")

	require.NotPanics(t, func() {
		_, broken := AccountKeyspaceInvariant(env.acck)(env.ctx)
		require.True(t, broken)
	})
}
