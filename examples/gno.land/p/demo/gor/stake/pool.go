package stake

import std "github.com/gnolang/gno/stdlibs/stdshim"

// for test burn all of them
func (s *Stake) BurnPool(amount uint64) {
	println("burn pool")
	caller := std.GetOrigCaller()
	pkgAddr := std.GetOrigPkgAddr()

	println("caller: ", caller)
	println("pkgAddr: ", pkgAddr)

	// amount, denom := s.GovToken.Balance(pkgAddr, "ugnot")

	println("amout: ", amount)
	println("denom: ", s.GovToken.GetDenom())
	if amount != 0 {
		// origPkgpath is set, coins can only be sent from this pkg, banker bind to pkg 1<=>1
		// banker := std.GetBanker(std.BankerTypeRealmSend)
		// coins := std.Coins{std.Coin{Denom: denom, Amount: int64(amount)}}
		// banker.SendCoins(pkgAddr, std.Address(""), coins)
		s.Transfer(pkgAddr, std.Address(""), amount)
	}
	println("pool burned")
}

func (s *Stake) Refund(amount uint64) {
	println("refund")
	caller := std.GetOrigCaller()
	pkgAddr := std.GetOrigPkgAddr()

	println("caller: ", caller)
	println("pkgAddr: ", pkgAddr)
	// amount, denom := s.GovToken.Balance(pkgAddr, "ugnot")
	println("amout: ", amount)
	println("denom: ", s.GovToken.GetDenom())

	s.Transfer(pkgAddr, caller, amount)
}
