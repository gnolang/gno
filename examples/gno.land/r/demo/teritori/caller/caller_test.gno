package caller

import (
	"testing"

	"gno.land/r/demo/teritori/callee"
)

func TestInitPkgPath(t *testing.T) {
	caller := callee.GetCaller()
	expected := "gno.land/r/demo/teritori/caller"
	if caller != expected {
		t.Errorf("caller should be '%s', but got '%s'", expected, caller)
	}
}

func TestInitAddr(t *testing.T) {
	caller := callee.GetCallerAddress()
	expected := "g1r2h3mkm8ky54w60vphvk8q7uk36uyum9gsqyj3"
	if caller != expected {
		t.Errorf("caller address should be '%s', but got '%s'", expected, caller)
	}
}
