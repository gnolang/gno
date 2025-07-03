package init

import (
	"gno.land/p/demo/uassert"
	"testing"
)

var (
	h = s
)

func TestInit(t *testing.T) {
	uassert.Empty(t, h)
}
