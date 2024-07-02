package users

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestClaimDefaultName(t *testing.T) {
	profile := "hello"
	println(ClaimDefaultName(profile))
}
