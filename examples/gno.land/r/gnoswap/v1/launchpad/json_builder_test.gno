package launchpad

import (
	"gno.land/p/demo/json"
	u256 "gno.land/p/gnoswap/uint256"
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestTimeInfoBuilder(t *testing.T) {
	info := TimeInfo{
		height: 100,
		time:   1234567890,
	}

	b := json.Builder()
	TimeInfoBuilder(b, "test", info)

	result := b.Node()
	expected := `{"testHeight":"100","testTime":"1234567890"}`

	res, err := json.Marshal(result)
	uassert.NoError(t, err)
	uassert.Equal(t, string(res), expected)
}

func TestProjectStatsBuilder(t *testing.T) {
	stats := ProjectStats{
		totalDeposit:      1000,
		actualDeposit:     800,
		totalParticipant:  50,
		actualParticipant: 40,
		totalCollected:    200,
	}

	b := json.Builder()
	ProjectStatsBuilder(b, stats)

	result := b.Node()
	expected := `{"totalDeposit":"1000","actualDeposit":"800","totalParticipant":"50","actualParticipant":"40","totalCollected":"200"}`

	res, err := json.Marshal(result)
	uassert.NoError(t, err)
	uassert.Equal(t, string(res), expected)
}

func TestTierBuilder(t *testing.T) {
	amountPerBlock := u256.MustFromDecimal("1000000")
	tier := Tier{
		collectWaitDuration:    100,
		tierAmount:             5000,
		tierAmountPerBlockX128: amountPerBlock,
		started:                TimeInfo{height: 100, time: 1234567890},
		ended:                  TimeInfo{height: 200, time: 1234567899},
		totalDepositAmount:     1000,
		actualDepositAmount:    800,
		totalParticipant:       50,
		actualParticipant:      40,
		userCollectedAmount:    300,
		calculatedAmount:       400,
	}

	b := json.Builder()
	TierBuilder(b, "test", tier)

	result := b.Node()
	expected := `{"testCollectWaitDuration":"100","testTierAmount":"5000","testTierAmountPerBlockX128":"1000000","testStartedHeight":"100","testStartedTime":"1234567890","testEndedHeight":"200","testEndedTime":"1234567899","testTotalDepositAmount":"1000","testActualDepositAmount":"800","testTotalParticipant":"50","testActualParticipant":"40","testUserCollectedAmount":"300","testCalculatedAmount":"400"}`

	res, err := json.Marshal(result)
	uassert.NoError(t, err)
	uassert.Equal(t, string(res), expected)
}

func TestProjectBuilder(t *testing.T) {
	project := Project{
		name:          "Test Project",
		tokenPath:     "gno.land/r/test",
		depositAmount: 1000,
		recipient:     std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
		conditions: map[string]Condition{
			"token1": {tokenPath: "path1", minAmount: 100},
		},
		tiersRatios: map[uint64]uint64{
			30:  20,
			90:  30,
			180: 50,
		},
		created: TimeInfo{height: 100, time: 1234567890},
		started: TimeInfo{height: 110, time: 1234567899},
		ended:   TimeInfo{height: 200, time: 1234568000},
		stats: ProjectStats{
			totalDeposit:      1000,
			actualDeposit:     800,
			totalParticipant:  50,
			actualParticipant: 40,
			totalCollected:    200,
		},
		refund: RefundInfo{
			amount: 100,
			height: 150,
			time:   1234567950,
		},
	}

	b := json.Builder()
	ProjectBuilder(b, project)

	data := b.Node()

	result, err := json.Marshal(data)
	uassert.NoError(t, err)

	expected := `{"name":"Test Project","tokenPath":"gno.land/r/test","depositAmount":"1000","recipient":"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5","conditionsToken":"token1","conditionsAmount":"100","tier30Ratio":"20","tier90Ratio":"30","tier180Ratio":"50","createdHeight":"100","createdTime":"1234567890","startedHeight":"110","startedTime":"1234567899","endedHeight":"200","endedTime":"1234568000","totalDeposit":"1000","actualDeposit":"800","totalParticipant":"50","actualParticipant":"40","totalCollected":"200","refundedAmount":"100","refundedHeight":"150","refundedTime":"1234567950"}`
	uassert.Equal(t, string(result), expected)
}
