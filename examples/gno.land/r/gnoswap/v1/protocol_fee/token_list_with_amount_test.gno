package protocol_fee

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	_ "gno.land/r/onbloc/baz"
	_ "gno.land/r/onbloc/foo"
)

var (
	dummyRealm = std.NewCodeRealm("gno.land/r/dummy")

	adminAddr  = consts.ADMIN
	adminUser  = adminAddr
	adminRealm = std.NewUserRealm(adminAddr)
)

func TestTokenList(t *testing.T) {
	tokenListWithAmount = map[string]uint64{
		"gno.land/r/foo": 100,
		"gno.land/r/bar": 0,
		"gno.land/r/baz": 200,
	}

	uassert.Equal(t, len(TokenList(true)), 2)
	uassert.Equal(t, len(TokenList(false)), 3)
}

func TestAddToProtocolFee(t *testing.T) {
	tokenListWithAmount = map[string]uint64{}

	tests := []struct {
		name      string
		tokenPath string
		amount    uint64
		want      uint64
	}{
		{
			name:      "add foo to protocol fee",
			tokenPath: "gno.land/r/foo",
			amount:    100,
			want:      100,
		},
		{
			name:      "add baz to protocol fee",
			tokenPath: "gno.land/r/baz",
			amount:    50,
			want:      50,
		},
		{
			name:      "add more baz to protocol fee",
			tokenPath: "gno.land/r/baz",
			amount:    10,
			want:      60,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			std.TestSetRealm(std.NewCodeRealm(consts.POOL_PATH))
			AddToProtocolFee(test.tokenPath, test.amount)

			uassert.Equal(t, tokenListWithAmount[test.tokenPath], test.want)
		})
	}
}

func TestClearTokenListWithAmount(t *testing.T) {
	tokenListWithAmount = map[string]uint64{
		"gno.land/r/foo": 100,
		"gno.land/r/baz": 200,
	}

	tests := []struct {
		name        string
		prevRealm   std.Realm
		want        map[string]uint64
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:        "no permission to clear",
			prevRealm:   dummyRealm,
			shouldPanic: true,
			panicMsg:    "caller(g1lvx5ssxvuz5tttx6uza3myv8xy6w36a46fv7sy) has no permission",
		},
		{
			name:      "clear protocol fee",
			prevRealm: std.NewCodeRealm(consts.GOV_STAKER_PATH),
			want:      map[string]uint64{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			std.TestSetRealm(test.prevRealm)

			if test.shouldPanic {
				uassert.PanicsWithMessage(t, test.panicMsg, func() {
					ClearTokenListWithAmount()
				})
			} else {
				ClearTokenListWithAmount()
				uassert.Equal(t, len(tokenListWithAmount), len(test.want))
			}
		})
	}
}

func TestTransferProtocolFee(t *testing.T) {
	// change devOpsPct to 49% (51% is for gov/staker)
	devOpsPct = 4900

	// send(and add) to protocol fee
	transferToProtocolFee(t, "gno.land/r/onbloc/foo", 100)
	transferToProtocolFee(t, "gno.land/r/onbloc/baz", 201)

	// call TransferProtocolFee
	std.TestSetRealm(std.NewCodeRealm(consts.GOV_STAKER_PATH))

	devOpsOldFoo := common.BalanceOf("gno.land/r/onbloc/foo", consts.DEV_OPS)
	devOpsOldBaz := common.BalanceOf("gno.land/r/onbloc/baz", consts.DEV_OPS)
	govStkaerOldFoo := common.BalanceOf("gno.land/r/onbloc/foo", consts.GOV_STAKER_ADDR)
	govStkaerOldBaz := common.BalanceOf("gno.land/r/onbloc/baz", consts.GOV_STAKER_ADDR)
	uassert.Equal(t, devOpsOldFoo, uint64(0))
	uassert.Equal(t, devOpsOldBaz, uint64(0))
	uassert.Equal(t, govStkaerOldFoo, uint64(0))
	uassert.Equal(t, govStkaerOldBaz, uint64(0))

	sentToGovStaker := TransferProtocolFee()
	// foo 100
	// -> devOps 49% => 49
	// -> gov/staker 51% => 51

	// baz 201
	// -> devOps 49% => 98
	// -> gov/staker 51% => 103

	// emitted event
	// EVENTS: [{"type":"TransferProtocolFee","attrs":[{"key":"prevAddr","value":"g17e3ykyqk9jmqe2y9wxe9zhep3p7cw56davjqwa"},{"key":"prevRealm","value":"gno.land/r/gnoswap/v1/gov/staker"},{"key":"toDevOps","value":"gno.land/r/onbloc/foo*FEE*49,gno.land/r/onbloc/baz*FEE*98"},{"key":"toGovStaker","value":"gno.land/r/onbloc/foo*FEE*51,gno.land/r/onbloc/baz*FEE*103"}],"pkg_path":"gno.land/r/gnoswap/v1/protocol_fee","func":"TransferProtocolFee"}]

	devOpsNewFoo := common.BalanceOf("gno.land/r/onbloc/foo", consts.DEV_OPS)
	devOpsNewBaz := common.BalanceOf("gno.land/r/onbloc/baz", consts.DEV_OPS)
	uassert.Equal(t, devOpsNewFoo, uint64(49))
	uassert.Equal(t, devOpsNewBaz, uint64(98))

	govStkaerNewFoo := common.BalanceOf("gno.land/r/onbloc/foo", consts.GOV_STAKER_ADDR)
	govStkaerNewBaz := common.BalanceOf("gno.land/r/onbloc/baz", consts.GOV_STAKER_ADDR)
	uassert.Equal(t, govStkaerNewFoo, uint64(51))
	uassert.Equal(t, govStkaerNewBaz, uint64(103))

	uassert.Equal(t, len(sentToGovStaker), 2)
	uassert.Equal(t, sentToGovStaker["gno.land/r/onbloc/foo"], uint64(51))
	uassert.Equal(t, sentToGovStaker["gno.land/r/onbloc/baz"], uint64(103))
}

func TestAssertOnlyPoolRouterStaker(t *testing.T) {
	tests := []struct {
		name        string
		prevRealm   std.Realm
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:      "caller is pool contract",
			prevRealm: std.NewCodeRealm(consts.POOL_PATH),
		},
		{
			name:      "caller is router contract",
			prevRealm: std.NewCodeRealm(consts.ROUTER_PATH),
		},
		{
			name:      "caller is staker contract",
			prevRealm: std.NewCodeRealm(consts.STAKER_PATH),
		},
		{
			name:        "caller is stranger",
			prevRealm:   dummyRealm,
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-PROTOCOL_FEE-001] caller has no permission",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			std.TestSetRealm(test.prevRealm)

			if test.shouldPanic {
				uassert.PanicsWithMessage(t, test.panicMsg, func() {
					assertOnlyPoolRouterStaker()
				})
			} else {
				uassert.NotPanics(t, func() {
					assertOnlyPoolRouterStaker()
				})
			}
		})
	}
}

func TestAssertOnlyGovStaker(t *testing.T) {
	tests := []struct {
		name        string
		prevRealm   std.Realm
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:      "caller is gov/staker contract",
			prevRealm: std.NewCodeRealm(consts.GOV_STAKER_PATH),
		},
		{
			name:        "caller is stranger",
			prevRealm:   dummyRealm,
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-PROTOCOL_FEE-001] caller has no permission",
		},
	}
}

// helper
func transferToProtocolFee(t *testing.T, tokenPath string, amount uint64) {
	t.Helper()

	std.TestSetRealm(adminRealm)

	tokenTeller := common.GetTokenTeller(tokenPath)
	tokenTeller.Transfer(consts.PROTOCOL_FEE_ADDR, amount)

	tokenListWithAmount[tokenPath] += amount
}
