package factory

import (
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

func TestRenderEscapesTokenName(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("mallory")))

	CreateNFT(cross(cur), "x\n\n# Fake heading")

	got := Render("")
	uassert.False(t, strings.Contains(got, "\n# Fake heading"), "token name must not open a heading")
}
