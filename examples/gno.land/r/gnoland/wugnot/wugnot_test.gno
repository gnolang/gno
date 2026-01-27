package wugnot

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
)

func TestDepositOriginGetTokens(cur realm, t *testing.T) {
	user := testutils.TestAddress("victim")
	proxy := chain.PackageAddress("gno.land/r/some/proxy")

	testing.IssueCoins(user, chain.Coins{{"ugnot", 5000}})
	testing.SetOriginCaller(user)
	testing.SetOriginSend(chain.Coins{{"ugnot", 1000}})

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/some/proxy"))
	Deposit(cross)

	uassert.Equal(t, int64(0), Token.BalanceOf(proxy))
	uassert.Equal(t, int64(1000), Token.BalanceOf(user))
}
