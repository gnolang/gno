package init

import (
	"testing"

	"gno.land/p/demo/uassert"
)

var (
	h = s
)

func TestInit(t *testing.T) {
	uassert.Empty(t, h)
}
