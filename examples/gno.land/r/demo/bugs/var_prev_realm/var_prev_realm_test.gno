package var_prev_realm

import (
	"std"
	"testing"
)

func TestVarPrevRealm(t *testing.T) {
	if initRealm.Addr() != varRealm.Addr() {
		t.Errorf("initRealm != varRealm")
	}

	correctRealmAddr := std.CurrentRealm().Addr()
	if correctRealmAddr != varRealm.Addr() {
		t.Errorf("correctRealmAddr != varRealm.Addr()")
	}

	if correctRealmAddr != initRealm.Addr() {
		t.Errorf("correctRealmAddr != initRealm.Addr()")
	}
}
