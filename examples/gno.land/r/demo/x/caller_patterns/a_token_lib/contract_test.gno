package token_lib

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func Test(t testing.T) {
	std.TestSetOrigCaller(testutils.TestAddress("foobar"))
	// GetCaller()
	GetOrigCaller()
}
