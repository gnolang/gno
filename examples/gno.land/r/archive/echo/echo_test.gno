package echo

import (
	"testing"

	"gno.land/p/nt/urequire/v0"
)

func Test(t *testing.T) {
	urequire.Equal(t, "aa", Render("aa"))
	urequire.Equal(t, "", Render(""))
}
