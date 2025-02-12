package protocol_fee

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"

	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/qux"

	_ "gno.land/r/onbloc/baz"
	_ "gno.land/r/onbloc/foo"
)

var (
	adminAddr  = consts.ADMIN
	adminRealm = std.NewUserRealm(adminAddr)
)

func TestDistributeProtocolFee(t *testing.T) {
	// admin > protocol_fee
	// send qux, bar for testing
	std.TestSetRealm(adminRealm)
	bar.Transfer(consts.PROTOCOL_FEE_ADDR, 1000)
	qux.Transfer(consts.PROTOCOL_FEE_ADDR, 1000)

	uassert.Equal(t, bar.BalanceOf(consts.PROTOCOL_FEE_ADDR), uint64(1000))
	uassert.Equal(t, bar.BalanceOf(consts.DEV_OPS), uint64(0))
	uassert.Equal(t, bar.BalanceOf(consts.GOV_STAKER_ADDR), uint64(0))

	uassert.Equal(t, qux.BalanceOf(consts.PROTOCOL_FEE_ADDR), uint64(1000))
	uassert.Equal(t, qux.BalanceOf(consts.DEV_OPS), uint64(0))
	uassert.Equal(t, qux.BalanceOf(consts.GOV_STAKER_ADDR), uint64(0))

	DistributeProtocolFee()

	uassert.Equal(t, bar.BalanceOf(consts.PROTOCOL_FEE_ADDR), uint64(0))
	uassert.Equal(t, bar.BalanceOf(consts.DEV_OPS), uint64(0))
	uassert.Equal(t, bar.BalanceOf(consts.GOV_STAKER_ADDR), uint64(1000))

	uassert.Equal(t, qux.BalanceOf(consts.PROTOCOL_FEE_ADDR), uint64(0))
	uassert.Equal(t, qux.BalanceOf(consts.DEV_OPS), uint64(0))
	uassert.Equal(t, qux.BalanceOf(consts.GOV_STAKER_ADDR), uint64(1000))
}

func TestSetDevOpsPctByAdminNoPermission(t *testing.T) {
	dummy := testutils.TestAddress("dummy")
	dummyRealm := std.NewUserRealm(dummy)
	std.TestSetRealm(dummyRealm)

	uassert.PanicsWithMessage(
		t, `caller(g1v36k6mteta047h6lta047h6lta047h6lz7gmv8) has no permission`, func() { SetDevOpsPctByAdmin(123) },
	)
}

func TestSetDevOpsPctByAdminInvalidFee(t *testing.T) {
	std.TestSetRealm(adminRealm)

	uassert.PanicsWithMessage(
		t,
		`[GNOSWAP-PROTOCOL_FEE-002] invalid percentage || pct(100001) should not be bigger than 10000`,
		func() {
			SetDevOpsPctByAdmin(100001)
		},
	)
}

func TestSetDevOpsPctByAdmin(t *testing.T) {
	std.TestSetRealm(adminRealm)

	uassert.Equal(t, GetDevOpsPct(), uint64(0))

	SetDevOpsPctByAdmin(123)

	uassert.Equal(t, GetDevOpsPct(), uint64(123))
}
