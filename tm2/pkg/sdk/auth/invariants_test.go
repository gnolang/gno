package auth

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
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
				t.Helper()
				bz := amino.MustMarshalAny(std.NewBaseAccount(other, nil, nil, 0, 0))
				rawSet(t, env, AddressStoreKey(master), bz)
			},
		},
		{
			// Accepted by the store; amino returns a nil account with no error, so
			// "decoded without error" is not enough to call an account usable.
			"zero-length value", "does not decode",
			func(t *testing.T, env testEnv) {
				t.Helper()
				rawSet(t, env, AddressStoreKey(master), []byte{})
			},
		},
		{
			"undecodable value", "does not decode",
			func(t *testing.T, env testEnv) {
				t.Helper()
				rawSet(t, env, AddressStoreKey(master), []byte{0xff, 0xfe, 0xfd})
			},
		},
		{
			"key of an unrecognised shape", "unrecognised shape",
			func(t *testing.T, env testEnv) {
				t.Helper()
				rawSet(t, env, append(AddressStoreKey(master), 0xAA), []byte{1})
			},
		},
		{
			// Session length but no /s/ infix: a length-only classifier would hand
			// this to a session decoder.
			"session-length key without the infix", "unrecognised shape",
			func(t *testing.T, env testEnv) {
				t.Helper()
				key := append(AddressStoreKey(master), make([]byte, SessionStoreKeyLen-AccountStoreKeyLen)...)
				rawSet(t, env, key, []byte{1})
			},
		},
		{
			"duplicate account number", "used more than once",
			func(t *testing.T, env testEnv) {
				t.Helper()
				a := env.acck.NewAccountWithAddress(env.ctx, master)
				env.acck.SetAccount(env.ctx, a)
				dup := std.NewBaseAccount(other, nil, nil, a.GetAccountNumber(), 0)
				env.acck.SetAccount(env.ctx, dup)
			},
		},
		{
			"account number at or above the counter", "above the global counter",
			func(t *testing.T, env testEnv) {
				t.Helper()
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

// The account-number checks and the two decode reports below have no violating
// state reachable through the keeper, so each reaches past it. Without these the
// checks could be deleted with the rest of the suite still green.
func TestAccountKeyspaceInvariantReportsCounterAndDecodeFaults(t *testing.T) {
	t.Parallel()

	t.Run("counter unreadable skips the number checks", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		rawSet(t, env, []byte(GlobalAccountNumberKey), []byte{0xff, 0xfe, 0xfd})

		msg, broken := AccountKeyspaceInvariant(env.acck)(env.ctx)
		require.True(t, broken)
		require.Contains(t, msg, "global account number is unreadable")
	})

	t.Run("counter above the allocation bound skips uniqueness", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		// The bitset is sized from this number, so the check must refuse to allocate
		// rather than trust it. Reported, not silently skipped.
		rawSet(t, env, []byte(GlobalAccountNumberKey), amino.MustMarshal(uint64(maxUniquenessBits+1)))

		msg, broken := AccountKeyspaceInvariant(env.acck)(env.ctx)
		require.True(t, broken)
		require.Contains(t, msg, "uniqueness was NOT verified")
	})

	t.Run("session account that does not decode", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		master := crypto.AddressFromPreimage([]byte("decode-master"))
		env.acck.SetAccount(env.ctx, env.acck.NewAccountWithAddress(env.ctx, master))
		session := crypto.AddressFromPreimage([]byte("decode-session"))
		rawSet(t, env, SessionStoreKey(master, session), []byte{0xff, 0xfe, 0xfd})

		msg, broken := AccountKeyspaceInvariant(env.acck)(env.ctx)
		require.True(t, broken)
		require.Contains(t, msg, "does not decode")
	})

	t.Run("delegated account filed at a regular path", func(t *testing.T) {
		t.Parallel()
		// The mirror of "session path holds a non-delegated account" below, and the
		// state bank.SetCoins has to survive: SetAccount files by the account's own
		// address, so a session account reaches the regular keyspace through it.
		env := setupTestEnv()
		master := crypto.AddressFromPreimage([]byte("regular-path-master"))
		_, sessPub, _ := tu.KeyTestPubAddr()
		sess := env.acck.NewSessionAccount(env.ctx, master, sessPub)
		env.acck.SetAccount(env.ctx, sess)

		msg, broken := AccountKeyspaceInvariant(env.acck)(env.ctx)
		require.True(t, broken)
		require.Contains(t, msg, "filed at a regular path")
	})
}

// The session checks each get their own test, and each asserts its own message.
// They all live in one sweep, so require.True(t, broken) alone would pass whenever any
// sibling fired — which is how a check gets deleted without its test noticing.
//
// All three states are reachable through the public API: SetSessionAccount writes to
// SessionStoreKey(master, acc.GetAddress()) and validates neither the master's
// existence, nor that the account agrees about its master, nor that it is delegated
// at all.
func TestAccountKeyspaceInvariantReportsSessionFaults(t *testing.T) {
	t.Parallel()

	t.Run("master has no account object", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		_, sessPub, _ := tu.KeyTestPubAddr()
		ghost := crypto.AddressFromPreimage([]byte("ghost-master"))
		// Deliberately no SetAccount for ghost.
		sess := env.acck.NewSessionAccount(env.ctx, ghost, sessPub)
		env.acck.SetSessionAccount(env.ctx, ghost, sess)

		msg, broken := AccountKeyspaceInvariant(env.acck)(env.ctx)
		require.True(t, broken)
		require.Contains(t, msg, "which has no account object")
	})

	t.Run("session claims a different master than it is filed under", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		_, sessPub, _ := tu.KeyTestPubAddr()
		filedUnder := crypto.AddressFromPreimage([]byte("m1"))
		claims := crypto.AddressFromPreimage([]byte("m2"))
		for _, m := range []crypto.Address{filedUnder, claims} {
			env.acck.SetAccount(env.ctx, env.acck.NewAccountWithAddress(env.ctx, m))
		}

		// NewSessionAccount sets the master *field*; SetSessionAccount decides the
		// master in the *key*. Neither consults the other, so building the session for
		// one master and filing it under another needs no struct literal and no raw
		// write — the mismatch is reachable through the public API alone. (The setter
		// cannot be used for this: SetMasterAddress is write-once.)
		sess := env.acck.NewSessionAccount(env.ctx, claims, sessPub)
		require.Equal(t, claims, sess.(std.DelegatedAccount).GetMasterAddress())
		env.acck.SetSessionAccount(env.ctx, filedUnder, sess)

		msg, broken := AccountKeyspaceInvariant(env.acck)(env.ctx)
		require.True(t, broken)
		require.Contains(t, msg, "but claims master")
	})

	t.Run("session path holds a non-delegated account", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		master := crypto.AddressFromPreimage([]byte("master"))
		env.acck.SetAccount(env.ctx, env.acck.NewAccountWithAddress(env.ctx, master))
		sessAddr := crypto.AddressFromPreimage([]byte("session"))

		// SetSessionAccount takes any std.Account, so a plain one lands at a session
		// path, where every reader expects to find a master to attribute it to.
		env.acck.SetSessionAccount(env.ctx, master,
			std.NewBaseAccount(sessAddr, nil, nil, env.acck.GetNextAccountNumber(env.ctx), 0))

		msg, broken := AccountKeyspaceInvariant(env.acck)(env.ctx)
		require.True(t, broken)
		require.Contains(t, msg, "does not hold a delegated account")
	})

	// And a well-formed session must be reported healthy, or the checks above are
	// firing on legitimate state rather than on the fault they name.
	t.Run("a well-formed session is healthy", func(t *testing.T) {
		t.Parallel()
		env := setupTestEnv()
		_, sessPub, _ := tu.KeyTestPubAddr()
		master := crypto.AddressFromPreimage([]byte("master"))
		env.acck.SetAccount(env.ctx, env.acck.NewAccountWithAddress(env.ctx, master))
		sess := env.acck.NewSessionAccount(env.ctx, master, sessPub)
		env.acck.SetSessionAccount(env.ctx, master, sess)

		msg, broken := AccountKeyspaceInvariant(env.acck)(env.ctx)
		require.False(t, broken, "a well-formed session must not be reported:\n%s", msg)
	})
}
