package common

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"
)

func TestSetLimitCaller(t *testing.T) {
	t.Run("initial check", func(t *testing.T) {
		uassert.True(t, GetLimitCaller())
	})

	t.Run("with non-admin privilege, panics", func(t *testing.T) {
		uassert.PanicsWithMessage(t,
			`[GNOSWAP-COMMON-001] caller has no permission || only Admin can set halt, called from g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm`,
			func() { SetLimitCaller(false) },
		)
	})

	t.Run("with admin privilege, success", func(t *testing.T) {
		std.TestSetRealm(std.NewUserRealm(consts.ADMIN))
		SetLimitCaller(false)
		uassert.False(t, GetLimitCaller())
	})
}
