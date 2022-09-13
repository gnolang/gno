package vault

import (
	"std"
	"testing"
	"time"

	"gno.land/p/grc/grc20"
)

func TestVault(t *testing.T) {
	fooAdminToken := grc20.NewAdminToken("The Foo Token", "FOO", 4)
	fooToken := fooAdminToken.GRC20()
	recoverAddr := std.Address("...")
	vault := New(fooToken, 1000, recoverAddr, time.Minute)

	t.Log("test", fooToken.GetName())
}
