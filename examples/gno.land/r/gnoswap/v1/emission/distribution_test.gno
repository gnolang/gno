package emission

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/gns"
)

func TestChangeDistributionPctByAdmin(t *testing.T) {
	resetObject(t)

	originCallback := callbackStakerEmissionChange
	callbackStakerEmissionChange = func(amount uint64) {}

	tests := []struct {
		name        string
		shouldPanic bool
		panicMsg    string
		setup       func()
		callerRealm std.Realm
		targets     []int
		pcts        []uint64
		verify      func()
	}{
		{
			name:        "panic if caller is not admin",
			shouldPanic: true,
			panicMsg:    `caller(g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm) has no permission`,
			targets:     []int{1, 2, 3, 4},
			pcts:        []uint64{1000, 2000, 3000, 4000},
		},
		{
			name:        "panic if target is invalid",
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-EMISSION-002] invalid emission target || invalid target(9)",
			callerRealm: adminRealm,
			targets:     []int{1, 2, 3, 9},
			pcts:        []uint64{1000, 2000, 3000, 4000},
		},
		{
			name:        "panic if sum of percentages is not 100%",
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-EMISSION-003] invalid emission percentage || sum of all pct should be 100% (10000 bps), got 10001",
			callerRealm: adminRealm,
			targets:     []int{1, 2, 3, 4},
			pcts:        []uint64{1000, 2000, 3000, 4001},
		},
		{
			name:        "success if admin",
			shouldPanic: false,
			callerRealm: adminRealm,
			targets:     []int{1, 2, 3, 4},
			pcts:        []uint64{1000, 2000, 3000, 4000},
			verify: func() {
				uassert.Equal(t, uint64(1000), GetDistributionBpsPct(int(1)))
				uassert.Equal(t, uint64(2000), GetDistributionBpsPct(int(2)))
				uassert.Equal(t, uint64(3000), GetDistributionBpsPct(int(3)))
				uassert.Equal(t, uint64(4000), GetDistributionBpsPct(int(4)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			if tt.callerRealm != (std.Realm{}) {
				std.TestSetRealm(tt.callerRealm)
			}

			target01, target02, target03, target04 := sliceToFourInt(t, tt.targets)
			pct01, pct02, pct03, pct04 := sliceToFourUint64(t, tt.pcts)

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					ChangeDistributionPctByAdmin(
						target01, pct01,
						target02, pct02,
						target03, pct03,
						target04, pct04,
					)
				})
			} else {
				uassert.NotPanics(t, func() {
					ChangeDistributionPctByAdmin(
						target01, pct01,
						target02, pct02,
						target03, pct03,
						target04, pct04,
					)
				})
				tt.verify()
			}
		})
	}
	callbackStakerEmissionChange = originCallback
}

func TestChangeDistributionPct(t *testing.T) {
	resetObject(t)

	originCallback := callbackStakerEmissionChange
	callbackStakerEmissionChange = func(amount uint64) {}

	tests := []struct {
		name        string
		shouldPanic bool
		panicMsg    string
		setup       func()
		callerRealm std.Realm
		targets     []int
		pcts        []uint64
		verify      func()
	}{
		{
			name:        "panic if caller is not governance",
			shouldPanic: true,
			panicMsg:    `caller(g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm) has no permission`,
			targets:     []int{1, 2, 3, 4},
			pcts:        []uint64{1000, 2000, 3000, 4000},
		},
		{
			name:        "panic if target is invalid",
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-EMISSION-002] invalid emission target || invalid target(9)",
			callerRealm: govRealm,
			targets:     []int{1, 2, 3, 9},
			pcts:        []uint64{1000, 2000, 3000, 4000},
		},
		{
			name:        "panic if sum of percentages is not 100%",
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-EMISSION-003] invalid emission percentage || sum of all pct should be 100% (10000 bps), got 10001",
			callerRealm: govRealm,
			targets:     []int{1, 2, 3, 4},
			pcts:        []uint64{1000, 2000, 3000, 4001},
		},
		{
			name:        "success if governance",
			shouldPanic: false,
			callerRealm: govRealm,
			targets:     []int{1, 2, 3, 4},
			pcts:        []uint64{1000, 2000, 3000, 4000},
			verify: func() {
				uassert.Equal(t, uint64(1000), GetDistributionBpsPct(1))
				uassert.Equal(t, uint64(2000), GetDistributionBpsPct(2))
				uassert.Equal(t, uint64(3000), GetDistributionBpsPct(3))
				uassert.Equal(t, uint64(4000), GetDistributionBpsPct(4))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			if tt.callerRealm != (std.Realm{}) {
				std.TestSetRealm(tt.callerRealm)
			}

			target01, target02, target03, target04 := sliceToFourInt(t, tt.targets)
			pct01, pct02, pct03, pct04 := sliceToFourUint64(t, tt.pcts)

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					ChangeDistributionPct(
						target01, pct01,
						target02, pct02,
						target03, pct03,
						target04, pct04,
					)
				})
			} else {
				uassert.NotPanics(t, func() {
					ChangeDistributionPct(
						target01, pct01,
						target02, pct02,
						target03, pct03,
						target04, pct04,
					)
				})
				tt.verify()
			}
		})
	}
	callbackStakerEmissionChange = originCallback
}

func TestChangeDistributionPcts(t *testing.T) {
	resetObject(t)

	originCallback := callbackStakerEmissionChange
	callbackStakerEmissionChange = func(amount uint64) {}

	changeDistributionPcts(
		1, 1000,
		2, 2000,
		3, 3000,
		4, 4000,
	)
	uassert.Equal(t, uint64(1000), GetDistributionBpsPct(1))
	uassert.Equal(t, uint64(2000), GetDistributionBpsPct(2))
	uassert.Equal(t, uint64(3000), GetDistributionBpsPct(3))
	uassert.Equal(t, uint64(4000), GetDistributionBpsPct(4))

	callbackStakerEmissionChange = originCallback
}

func TestCalculateAmount(t *testing.T) {
	tests := []struct {
		name     string
		pct      uint64
		expected uint64
	}{
		{name: "5% of 1_000", pct: 500, expected: 50},
		{name: "10% of 1_000", pct: 1000, expected: 100},
		{name: "55% of 1_000", pct: 5500, expected: 550},
		{name: "100% of 1_000", pct: 10000, expected: 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uassert.Equal(t, tt.expected, calculateAmount(uint64(1000), tt.pct))
		})
	}
}

func TestTransferToTarget(t *testing.T) {
	resetObject(t)

	tests := []struct {
		name        string
		shouldPanic bool
		panicMsg    string
		setup       func()
		target      int
		amount      uint64
		verify      func()
	}{
		{
			name:        "invalid target",
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-EMISSION-002] invalid emission target || invalid target(9)",
			target:      9,
			amount:      100,
		},
		{
			name:        "not enough balance for emission",
			shouldPanic: true,
			panicMsg:    "insufficient balance",
			target:      LIQUIDITY_STAKER,
			amount:      1,
		},
		{
			name:   "transfer to LIQUIDITY_STAKER",
			target: LIQUIDITY_STAKER,
			setup: func() {
				std.TestSetRealm(adminRealm)
				gns.Transfer(consts.EMISSION_ADDR, 100000) // give enough balance for emission
			},
			amount: 100,
			verify: func() {
				uassert.Equal(t, uint64(100), distributedToStaker)
				uassert.Equal(t, uint64(100), accuDistributedToStaker)
			},
		},
		{
			name:   "transfer to DEVOPS",
			target: DEVOPS,
			amount: 200,
			verify: func() {
				uassert.Equal(t, uint64(200), distributedToDevOps)
				uassert.Equal(t, uint64(200), accuDistributedToDevOps)
			},
		},
		{
			name:   "transfer to COMMUNITY_POOL",
			target: COMMUNITY_POOL,
			amount: 300,
			verify: func() {
				uassert.Equal(t, uint64(300), distributedToCommunityPool)
				uassert.Equal(t, uint64(300), accuDistributedToCommunityPool)
			},
		},
		{
			name:   "transfer to GOV_STAKER",
			target: GOV_STAKER,
			amount: 400,
			verify: func() {
				uassert.Equal(t, uint64(400), distributedToGovStaker)
				uassert.Equal(t, uint64(400), accuDistributedToGovStaker)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					transferToTarget(tt.target, tt.amount)
				})
			} else {
				uassert.NotPanics(t, func() {
					transferToTarget(tt.target, tt.amount)
				})
				tt.verify()
			}
		})
	}
}

func TestDistributeToTarget(t *testing.T) {
	resetObject(t)

	tests := []struct {
		name         string
		shouldPanic  bool
		panicMsg     string
		setup        func()
		amount       uint64
		expectedLeft uint64
	}{
		{
			name:        "invalid target",
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-EMISSION-002] invalid emission target || invalid target(a)",
			setup: func() {
				distributionBpsPct.Set("a", uint64(1000))
			},
			amount: 100,
		},
		{
			name: "distributed all amount",
			setup: func() {
				distributionBpsPct.Remove("a")
			},
			amount:       1000,
			expectedLeft: 0,
		},
		{
			name:         "distributed partial amount",
			amount:       1001,
			expectedLeft: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					distributeToTarget(tt.amount)
				})
			} else {
				distributed := distributeToTarget(tt.amount)
				left := tt.amount - distributed
				uassert.Equal(t, tt.expectedLeft, left)
			}
		})
	}
}

func TestClearDistributedToStaker(t *testing.T) {
	distributedToStaker = 100

	tests := []struct {
		name        string
		expected    uint64
		callerRealm std.Realm
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:        "can not clear is caller is not staker",
			shouldPanic: true,
			panicMsg:    `caller(g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm) has no permission`,
		},
		{
			name:        "can clear if caller is staker",
			callerRealm: stakerRealm,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.callerRealm != (std.Realm{}) {
				std.TestSetRealm(tt.callerRealm)
			}

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					ClearDistributedToStaker()
				})
			} else {
				ClearDistributedToStaker()
				uassert.Equal(t, uint64(0), distributedToStaker)
			}
		})
	}

}

func TestClearClearDistributedToGovStaker(t *testing.T) {
	distributedToGovStaker = 100

	tests := []struct {
		name        string
		expected    uint64
		callerRealm std.Realm
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:        "can not clear is caller is not gov/staker",
			shouldPanic: true,
			panicMsg:    `caller(g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm) has no permission`,
		},
		{
			name:        "can clear if caller is gov/taker",
			callerRealm: govStakerRealm,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.callerRealm != (std.Realm{}) {
				std.TestSetRealm(tt.callerRealm)
			}

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					ClearDistributedToGovStaker()
				})
			} else {
				ClearDistributedToGovStaker()
				uassert.Equal(t, uint64(0), distributedToGovStaker)
			}
		})
	}

}

func sliceToFourInt(t *testing.T, slice []int) (int, int, int, int) {
	t.Helper()

	return slice[0], slice[1], slice[2], slice[3]
}

func sliceToFourUint64(t *testing.T, slice []uint64) (uint64, uint64, uint64, uint64) {
	t.Helper()

	return slice[0], slice[1], slice[2], slice[3]
}
