package pkgerr

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestErr(t *testing.T) {
	uassert.ErrorContains(t, errorOut(), currRealmLink)
	uassert.ErrorContains(t, errorOut(), "err test")
	uassert.ErrorContains(t, errorOut(), "err test")
}
