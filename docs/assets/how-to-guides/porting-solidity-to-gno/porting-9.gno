// Withdraw Function Test
func TestWithdraw(t *testing.T) {
	// If there's no participants for return
	shouldEqual(t, pendingReturns.Size(), 0)

	// If there's participants for return (data generation
	returnAddr := bidder01.String()
	returnAmount := int64(3)
	pendingReturns.Set(returnAddr, returnAmount)
	shouldEqual(t, pendingReturns.Size(), 1)
	shouldEqual(t, pendingReturns.Has(returnAddr), true)

	banker := std.GetBanker(std.BankerTypeRealmSend)
	pkgAddr := std.GetOrigPkgAddr()
	banker.SendCoins(pkgAddr, std.Address(returnAddr), std.Coins{{"ugnot", returnAmount}})
	shouldEqual(t, banker.GetCoins(std.Address(returnAddr)).String(), "3ugnot")
}
