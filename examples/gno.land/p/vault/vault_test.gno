package vault

import (
	"std"
	"testing"

	"gno.land/p/grc/grc20"
)

func TestVault(t *testing.T) {
	fooToken := grc20.NewToken("Foo", "FOO", 4)
	vault := New(fooToken)

	t.Log("test", fooToken.GetName(), barToken.GetName())
}
