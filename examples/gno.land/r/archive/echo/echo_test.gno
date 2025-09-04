package echo

import (
	"testing"

	"gno.land/p/nt/urequire"
)

func Test(t *testing.T) {
	urequire.Equal(t, "aa", Render("aa"))
	urequire.Equal(t, "", Render(""))
}
