package protocol_fee

import (
	"std"
	"strconv"
	"testing"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/json"
	"gno.land/p/demo/uassert"
)

func TestApiGetAccuTransferToGovStaker(t *testing.T) {
	addAccuToGovStaker("gno.land/r/onbloc/bar", 1000)
	addAccuToGovStaker("gno.land/r/onbloc/qux", 2000)

	result := ApiGetAccuTransferToGovStaker()
	uassert.Equal(t, `{"height":"123","now":"1234567890","gno.land/r/onbloc/bar":"1000","gno.land/r/onbloc/qux":"2000"}`, result)

	std.TestSetRealm(std.NewCodeRealm("gno.land/r/gnoswap/v1/gov/staker"))
	ClearAccuTransferToGovStaker()
	result = ApiGetAccuTransferToGovStaker()
	uassert.Equal(t, `{"height":"123","now":"1234567890"}`, result)
}

func TestApiGetHistoryTransferToGovStaker(t *testing.T) {
	std.TestSkipHeights(1)
	addAccuToGovStaker("gno.land/r/onbloc/bar", 1000)
	addAccuToGovStaker("gno.land/r/onbloc/qux", 2000)
	updateDistributedToGovStakerHistory("gno.land/r/onbloc/bar", 1000)
	updateDistributedToGovStakerHistory("gno.land/r/onbloc/qux", 2000)
	resultHistory := ApiGetHistoryTransferToGovStaker()
	uassert.Equal(t, `{"height":"124","now":"1234567892","gno.land/r/onbloc/bar":"1000","gno.land/r/onbloc/qux":"2000"}`, resultHistory)
	resultAccu := ApiGetAccuTransferToGovStaker()
	uassert.Equal(t, `{"height":"124","now":"1234567892","gno.land/r/onbloc/bar":"1000","gno.land/r/onbloc/qux":"2000"}`, resultAccu)

	std.TestSetRealm(std.NewCodeRealm("gno.land/r/gnoswap/v1/gov/staker"))
	ClearAccuTransferToGovStaker()
	resultHistory = ApiGetHistoryTransferToGovStaker()
	uassert.Equal(t, `{"height":"124","now":"1234567892","gno.land/r/onbloc/bar":"1000","gno.land/r/onbloc/qux":"2000"}`, resultHistory)
	resultAccu = ApiGetAccuTransferToGovStaker()
	uassert.Equal(t, `{"height":"124","now":"1234567892"}`, resultAccu)
}

func TestApiGetAccuTransferToDevOps(t *testing.T) {
	std.TestSkipHeights(1)
	addAccuToDevOps("gno.land/r/onbloc/bar", 1000)
	addAccuToDevOps("gno.land/r/onbloc/qux", 2000)

	result := ApiGetAccuTransferToDevOps()
	uassert.Equal(t, `{"height":"124","now":"1234567892","gno.land/r/onbloc/bar":"1000","gno.land/r/onbloc/qux":"2000"}`, result)
}

func TestApiGetHistoryTransferToDevOps(t *testing.T) {
	std.TestSkipHeights(1)
	addAccuToDevOps("gno.land/r/onbloc/bar", 1000)
	addAccuToDevOps("gno.land/r/onbloc/qux", 2000)
	updateDistributedToDevOpsHistory("gno.land/r/onbloc/bar", 1000)
	updateDistributedToDevOpsHistory("gno.land/r/onbloc/qux", 2000)
	resultHistory := ApiGetHistoryTransferToDevOps()
	uassert.Equal(t, `{"height":"124","now":"1234567892","gno.land/r/onbloc/bar":"1000","gno.land/r/onbloc/qux":"2000"}`, resultHistory)
	resultAccu := ApiGetAccuTransferToDevOps()
	uassert.Equal(t, `{"height":"124","now":"1234567892","gno.land/r/onbloc/bar":"2000","gno.land/r/onbloc/qux":"4000"}`, resultAccu)
}
