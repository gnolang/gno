package execctx

import (
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

func TestCallCredits(t *testing.T) {
	ugnot := func(n int64) std.Coins { return std.Coins{{Denom: "ugnot", Amount: n}} }

	c := NewCallCredits("gno.land/r/entry", ugnot(50))
	// The envelope is read once, by the entry realm entered from the user.
	require.Equal(t, ugnot(50), c.Take("", "gno.land/r/entry"))
	require.Empty(t, c.Take("", "gno.land/r/entry"), "second read must see zero")
	require.Empty(t, c.Unclaimed(), "an unread envelope is not an unclaimed forward")

	// Forwards are keyed by payer as well as payee.
	c.Credit("gno.land/r/router", "gno.land/r/vault", ugnot(7))
	c.Credit("gno.land/r/router", "gno.land/r/vault", ugnot(3))
	c.Credit("gno.land/r/other", "gno.land/r/vault", ugnot(1))
	require.Empty(t, c.Take("gno.land/r/third", "gno.land/r/vault"), "wrong payer reads zero")
	require.Equal(t, ugnot(10), c.Take("gno.land/r/router", "gno.land/r/vault"), "same pair sums")
	require.Equal(t, "1ugnot forwarded by gno.land/r/other to gno.land/r/vault: never read by the payee", c.Unclaimed())

	// Zero forwards record nothing; a nil ledger reads as empty.
	c.Credit("a", "b", nil)
	require.Equal(t, 1, strings.Count(c.Unclaimed(), "forwarded"))
	var nilLedger *CallCredits
	require.Empty(t, nilLedger.Take("a", "b"))
	require.Empty(t, nilLedger.Unclaimed())

	// Re-seeding replaces the envelope and keeps forwards.
	c.SeedEnvelope("gno.land/r/entry2", ugnot(5))
	c.SeedEnvelope("gno.land/r/entry3", ugnot(6))
	require.Empty(t, c.Take("", "gno.land/r/entry2"))
	require.Equal(t, ugnot(6), c.Take("", "gno.land/r/entry3"))
	require.Equal(t, 1, strings.Count(c.Unclaimed(), "forwarded"))
}
