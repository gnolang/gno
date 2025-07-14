package common

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"

	"gno.land/r/gnoswap/v1/access"
)

func TestSetLimitCaller(t *testing.T) {
	t.Run("initial check", func(t *testing.T) {
		uassert.True(t, GetLimitCaller())
	})

	t.Run("with admin privilege, success", func(t *testing.T) {
		func(cur realm) {
			adminAddr, _ := access.GetAddress(access.ROLE_ADMIN)
			testing.SetRealm(std.NewUserRealm(adminAddr))

			SetLimitCaller(cross, false)
			uassert.False(t, GetLimitCaller())
		}(cross)
	})
}
