package commondao

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// TestGenesisCouncilOrder pins the council of the DAO genesis.gno creates.
// renderCouncil walks the addrset, so the realm's own page lists the two
// founding members in bech32 order, not in the order genesis.gno passes
// them. No filetest renders DAO 1, so nothing else in the suite notices
// when either address moves.
func TestGenesisCouncilOrder(t *testing.T) {
	dao := mustGetDAO(1)
	council := dao.Council()
	urequire.Equal(t, 2, council.Size())

	var got []string
	council.IterateByOffset(0, council.Size(), func(addr address) bool {
		got = append(got, addr.String())
		return false
	})

	urequire.Equal(t, 2, len(got))
	uassert.Equal(t, "g1manfred47kzduec920z88wfr64ylksmdcedlf5", got[0])
	uassert.Equal(t, "g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx", got[1])
}
