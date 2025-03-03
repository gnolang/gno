package pkgerr

import (
	"errors"
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/leon/pkgerr"
)

func TestErr(t *testing.T) {
	uassert.ErrorContains(t, errorOut(), currRealmLink)
	uassert.ErrorContains(t, errorOut(), "err test")
	uassert.ErrorContains(t, errorOut(), "err test")
}
