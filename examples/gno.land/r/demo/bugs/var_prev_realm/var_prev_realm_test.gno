package var_prev_realm

import (
	"std"
	"testing"
)

func TestVarPrevRealm(t *testing.T) {
	if initRealm.Addr() != varRealm.Addr() {
		t.Errorf("expected init realm  '%s' to be equal to var block realm '%s'", initRealm.Addr(), varRealm.Addr())
	}

	correctRealmAddr := std.DerivePkgAddr("gno.land/r/demo/bugs/var_prev_realm")

	if correctRealmAddr != varRealm.Addr() {
		t.Errorf("var block realm address incorrect, expected '%s', got '%s'", correctRealmAddr, varRealm.Addr())
	}

	if correctRealmAddr != initRealm.Addr() {
		t.Errorf("init realm address incorrect, expected '%s', got '%s'", correctRealmAddr, initRealm.Addr())
	}
}
