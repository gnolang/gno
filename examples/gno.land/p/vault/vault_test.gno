package vault

import (
	"std"
	"testing"
	"time"

	"gno.land/p/grc/grc20"
	"gno.land/p/testutils"
	"gno.land/p/ufmt"
)

func TestVault(t *testing.T) {
	alice := testutils.TestAddress("alice")
	std.TestSetOrigCaller(alice)

	origCaller := std.GetOrigCaller()
	fooAdminToken := grc20.NewAdminToken("The Foo Token", "FOO", 4)
	if err := fooAdminToken.Mint(origCaller, 1000); err != nil {
		t.Error(err)
	}
	fooToken := fooAdminToken.GRC20()
	recoverAddr := std.Address("...")
	lockDuration := uint(5)
	v := New(fooToken, 1000, recoverAddr, lockDuration)

	debugVault(t, v, origCaller)
}

func debugVault(t *testing.T, v Vault, caller std.Address) string {
	t.Helper()

	impl, ok := v.(*impl)
	if !ok {
		t.Fatal("should not happen")
	}

	pkgaddr := std.GetOrigPkgAddr()
	balanceOfCaller, _ := impl.token.BalanceOf(caller)
	balanceOfOwner, _ := impl.token.BalanceOf(impl.owner)
	balanceOfPkg, _ := impl.token.BalanceOf(pkgaddr)

	line := ufmt.Sprintf(
		"redeemMinHeight=%d unvaultedAmount=%d\n"+
			"  owner=%s  balanceOf(owner)=%d\n"+
			"  caller=%s balanceOf(caller)=%d\n"+
			"  pkg=%s    balanceOf(pkg)=%d",
		impl.redeemMinHeight, impl.unvaultedAmount,
		impl.owner.String(), balanceOfOwner,
		caller.String(), balanceOfCaller,
		pkgaddr.String(), balanceOfPkg,
	)
	t.Log(line)
}
