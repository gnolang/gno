package auth

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	tu "github.com/gnolang/gno/tm2/pkg/sdk/testutils"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// Revoking every session must leave none behind. A session carries a spend
// limit, so one that survives keeps the authority to spend against it.
//
// Worth its own test because RemoveAllSessions gathers the keys under a live
// iterator and deletes them afterwards: a walk that stopped early, or keys that
// aliased one another, would delete some and look like it deleted all. Eight
// sessions, so an off-by-one or a single-delete both show up.
func TestRemoveAllSessionsRemovesEveryOne(t *testing.T) {
	env, _, _, master := setupSessionEnv(t)
	ctx := env.ctx

	const n = 8
	addrs := make([]crypto.Address, 0, n)
	for i := range n {
		_, spub, saddr := tu.KeyTestPubAddr()
		sa := env.acck.NewSessionAccount(ctx, master, spub)
		da := sa.(std.DelegatedAccount)
		da.SetExpiresAt(0)
		da.SetSpendLimit(std.Coins{std.NewCoin("ugnot", int64(100+i))})
		env.acck.SetSessionAccount(ctx, master, sa)
		addrs = append(addrs, saddr)
	}

	before := 0
	env.acck.IterateSessions(ctx, master, func(std.Account) bool { before++; return false })
	require.Equal(t, n, before, "setup must create %d sessions", n)

	env.acck.RemoveAllSessions(ctx, master)

	after := 0
	env.acck.IterateSessions(ctx, master, func(std.Account) bool { after++; return false })
	require.Zero(t, after, "revoke-all must leave no session behind")

	// And none is individually reachable either.
	for i, a := range addrs {
		require.Nil(t, env.acck.GetSessionAccount(ctx, master, a),
			"session %d survived revoke-all and keeps its spend authority", i)
	}
}
