// PKGPATH: gno.land/r/demo/subrealm
package subrealm

import (
	"chain"
	"chain/banker"
)

// Sub-address treasury flow: the host issues coins to a sub-address
// (RealmIssue on the primary cur), then spends from it with a
// RealmSend banker constructed from the sub-token — no banker changes
// needed. RealmIssue and OriginSend banker types are forbidden for
// sub-tokens.
func tryNewBanker(_ int, bt banker.BankerType, rlm realm) {
	defer func() {
		println("recovered:", recover())
	}()
	banker.NewBanker(bt, rlm)
}

func main(cur realm) {
	sub := cur.Sub("dao/7")
	dest := address("g1destdestdestdestdestdestdestdestdest0")

	// Fund the sub-address using the host's primary issue authority.
	ib := banker.NewBanker(banker.BankerTypeRealmIssue, cur)
	denom := chain.CoinDenom(cur.PkgPath(), "gold")
	ib.IssueCoin(sub.Address(), denom, 1000)

	// Spend from the sub-address via a sub-token RealmSend banker.
	sb := banker.NewBanker(banker.BankerTypeRealmSend, sub)
	sb.SendCoins(sub.Address(), dest, chain.NewCoins(chain.NewCoin(denom, 400)))
	println("sub balance:", banker.NewReadonlyBanker().GetCoins(sub.Address()).String())
	println("dest balance:", banker.NewReadonlyBanker().GetCoins(dest).String())

	// Forbidden banker types for sub-tokens.
	tryNewBanker(0, banker.BankerTypeRealmIssue, sub)
	tryNewBanker(0, banker.BankerTypeOriginSend, sub)
}

// Output:
// sub balance: 600/gno.land/r/demo/subrealm:gold
// dest balance: 400/gno.land/r/demo/subrealm:gold
// recovered: BankerTypeRealmIssue not supported for sub-realm tokens
// recovered: BankerTypeOriginSend not supported for sub-realm tokens
