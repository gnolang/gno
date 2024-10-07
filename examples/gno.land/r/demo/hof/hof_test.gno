package hof

import (
	"gno.land/p/demo/uassert"
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestRegister(t *testing.T) {
	userRealm := std.NewUserRealm(testutils.TestAddress("leon"))
	std.TestSetRealm(userRealm)
	uassert.PanicsWithMessage(t, ErrNonCodeCall.Error(), Register)

	rlm1 := std.NewCodeRealm("gno.land/r/leon/home")
	std.TestSetRealm(rlm1)

	Register()
	Register()
	//uassert.NotPanics(t, Register)
	//
	//Register()
	//uassert.PanicsWithMessage(t, ErrAlreadyExists.Error(), Register)
}
