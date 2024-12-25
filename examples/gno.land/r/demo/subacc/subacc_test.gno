package subacc_test

import (
	"path"
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/demo/authbanker"
)

var (
	alice = testutils.TestAddress("alice")
)

func TestSubacc(t *testing.T) {
	accountId := path.Join("/", std.DerivePkgAddr("gno.land/r/demo/subacc").String(), alice.String(), "foo")

	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigSend(std.Coins{{"ugnot", 42}}, nil)
	authbanker.FundVault(accountId)

	println(authbanker.GetCoins(accountId))

	// XXX: ./examples/gno.land/r/demo/subacc: test pkg: panic: unexpected unreal object
}
