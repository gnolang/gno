package inner

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

var (
	oc std.Address = testutils.TestAddress("origin caller ")
	or string      = "gno.land/r/demo/outer_realm"
)

func TestOuterCallInner(t *testing.T) {
	std.TestSetOrigCaller(oc)
	shouldEQ(t, std.GetOrigCaller(), oc)

	std.TestSetPrevRealm(or)
	pr := std.PrevRealm()
	shouldEQ(t, pr.Addr(), std.DerivePkgAddr(or))
	shouldEQ(t, pr.PkgPath(), or)
	shouldEQ(t, std.GetOrigCaller(), oc)
}

func shouldEQ(t *testing.T, got, want interface{}) {
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
