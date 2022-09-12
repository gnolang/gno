package vault

import (
	"std"
	"testing"
	"time"

	"gno.land/p/grc/grc20"
)

func TestVault(t *testing.T) {
	_ = std.Address
	_ = time.Now
	_ = grc20.NewToken
	/*
		fooToken := grc20.NewToken("The Foo Token", "FOO", 4)
		recoverAddr := std.Address("...")
		vault := New(fooToken, 1000, recoverAddr, time.Minute)

		t.Log("test", fooToken.GetName(), barToken.GetName())
	*/
}
