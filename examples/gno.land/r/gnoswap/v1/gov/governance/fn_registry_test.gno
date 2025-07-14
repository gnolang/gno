package governance

import (
	"errors"
	"testing"

	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"

	en "gno.land/r/gnoswap/v1/emission"
	pl "gno.land/r/gnoswap/v1/pool"
	pf "gno.land/r/gnoswap/v1/protocol_fee"
	rr "gno.land/r/gnoswap/v1/router"
	sr "gno.land/r/gnoswap/v1/staker"
)

func TestCreateParameterHandlers(t *testing.T) {
	tests := []struct {
		name             string
		path             string
		function         string
		params           []string
		validate         func(t *testing.T)
		expectedHasError bool
		expectedHasPanic bool
		expectedHasAbort bool
		expectedErrorMsg string
		expectedPanicMsg string
		expectedAbortMsg string
	}{
		{
			name:     "create parameter handlers is success by emission change distribution pct",
			path:     consts.EMISSION_PATH,
			function: "ChangeDistributionPct",
			params:   []string{"1", "7000", "2", "1500", "3", "1000", "4", "500"},
			validate: func(t *testing.T) {
				uassert.Equal(t, en.GetDistributionBpsPct(1), int64(7000))
				uassert.Equal(t, en.GetDistributionBpsPct(2), int64(1500))
				uassert.Equal(t, en.GetDistributionBpsPct(3), int64(1000))
				uassert.Equal(t, en.GetDistributionBpsPct(4), int64(500))
			},
			expectedHasPanic: false,
			expectedHasAbort: false,
			expectedHasError: false,
		},
		{
			name:     "create parameter handlers is success by router set swap fee",
			path:     consts.ROUTER_PATH,
			function: "SetSwapFee",
			params:   []string{"1000"},
			validate: func(t *testing.T) {
				uassert.Equal(t, rr.GetSwapFee(), uint64(1000))
			},
			expectedHasPanic: false,
			expectedHasAbort: false,
			expectedHasError: false,
		},
		{
			name:     "create parameter handlers is success by staker set deposit gns amount",
			path:     consts.STAKER_PATH,
			function: "SetDepositGnsAmount",
			params:   []string{"1000"},
			validate: func(t *testing.T) {
				uassert.Equal(t, sr.GetDepositGnsAmount(), int64(1000))
			},
			expectedHasPanic: false,
			expectedHasAbort: false,
			expectedHasError: false,
		},
		{
			name:     "create parameter handlers is success by pool set pool creation fee",
			path:     consts.POOL_PATH,
			function: "SetPoolCreationFee",
			params:   []string{"500"},
			validate: func(t *testing.T) {
				uassert.Equal(t, pl.GetPoolCreationFee(), int64(500))
			},
			expectedHasPanic: false,
			expectedHasAbort: false,
			expectedHasError: false,
		},
		{
			name:     "create parameter handlers is success by pool set withdrawal fee",
			path:     consts.POOL_PATH,
			function: "SetWithdrawalFee",
			params:   []string{"600"},
			validate: func(t *testing.T) {
				uassert.Equal(t, pl.GetWithdrawalFee(), uint64(600))
			},
		},
		{
			name:     "create parameter handlers is success by protocol fee set dev ops pct",
			path:     consts.PROTOCOL_FEE_PATH,
			function: "SetDevOpsPct",
			params:   []string{"900"},
			validate: func(t *testing.T) {
				uassert.Equal(t, pf.GetDevOpsPct(), uint64(900))
			},
			expectedHasPanic: false,
			expectedHasAbort: false,
			expectedHasError: false,
		},
		{
			name:     "create parameter handlers is success by router set swap fee",
			path:     consts.ROUTER_PATH,
			function: "SetSwapFee",
			params:   []string{"400"},
			validate: func(t *testing.T) {
				uassert.Equal(t, rr.GetSwapFee(), uint64(400))
			},
		},
		{
			name:     "create parameter handlers is success by staker set deposit gns amount",
			path:     consts.STAKER_PATH,
			function: "SetDepositGnsAmount",
			params:   []string{"400"},
			validate: func(t *testing.T) {
				uassert.Equal(t, sr.GetDepositGnsAmount(), int64(400))
			},
			expectedHasPanic: false,
			expectedHasAbort: false,
			expectedHasError: false,
		},
		{
			name:     "create parameter handlers is success by staker set un staking fee",
			path:     consts.STAKER_PATH,
			function: "SetUnStakingFee",
			params:   []string{"100"},
			validate: func(t *testing.T) {
				uassert.Equal(t, sr.GetUnstakingFee(), uint64(100))
			},
			expectedHasPanic: false,
			expectedHasAbort: false,
			expectedHasError: false,
		},
		{
			name:     "create parameter handlers is success by protocol fee set dev ops pct",
			path:     consts.PROTOCOL_FEE_PATH,
			function: "SetDevOpsPct",
			params:   []string{"900"},
			validate: func(t *testing.T) {
				uassert.Equal(t, pf.GetDevOpsPct(), uint64(900))
			},
		},
		{
			name:     "create parameter handlers is success by router set swap fee",
			path:     consts.ROUTER_PATH,
			function: "SetSwapFee",
			params:   []string{"400"},
			validate: func(t *testing.T) {
				uassert.Equal(t, rr.GetSwapFee(), uint64(400))
			},
			expectedHasPanic: false,
			expectedHasAbort: false,
			expectedHasError: false,
		},
		{
			name:             "create parameter handlers is failed by router set swap fee with invalid value",
			path:             consts.ROUTER_PATH,
			function:         "SetSwapFee",
			params:           []string{"-100"},
			validate:         func(t *testing.T) {},
			expectedHasPanic: false,
			expectedHasAbort: true,
			expectedHasError: false,
			expectedAbortMsg: "invalid uint64 value: -100",
		},
		{
			name:             "create parameter handlers is failed by staker set un staking fee with invalid value",
			path:             consts.STAKER_PATH,
			function:         "SetUnStakingFee",
			params:           []string{"10001"},
			validate:         func(t *testing.T) {},
			expectedHasPanic: false,
			expectedHasAbort: true,
			expectedHasError: false,
			expectedAbortMsg: "[GNOSWAP-STAKER-008] invalid unstaking fee || fee(10001) must be in range 0 ~ 10000",
		},
		{
			name:             "create parameter handlers is failed by emission change distribution pct with invalid value",
			path:             consts.EMISSION_PATH,
			function:         "ChangeDistributionPct",
			params:           []string{"1", "10001", "2", "3000"},
			validate:         func(t *testing.T) {},
			expectedHasPanic: false,
			expectedHasAbort: false,
			expectedHasError: true,
			expectedErrorMsg: "invalid parameters for %!s((unhandled)). expected 8 but got 4",
		},
		{
			name:             "create parameter handlers is failed by access create role with invalid parameter",
			path:             consts.ACCESS_PATH,
			function:         "CreateRole",
			params:           []string{"test_role", "invalid_address"},
			validate:         func(t *testing.T) {},
			expectedHasPanic: false,
			expectedHasAbort: true,
			expectedHasError: false,
			expectedAbortMsg: "invalid address: invalid_address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			registry := createParameterHandlers()
			handler, _ := registry.Handler(tt.path, tt.function)

			var err error

			// when
			if tt.expectedHasPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicMsg, func() {
					handler.Execute(tt.params)
				})
			} else if tt.expectedHasAbort {
				uassert.AbortsWithMessage(t, tt.expectedAbortMsg, func() {
					handler.Execute(tt.params)
				})
			} else {
				err = handler.Execute(tt.params)
			}

			if tt.expectedHasError {
				uassert.ErrorIs(t, err, errors.New(tt.expectedErrorMsg))
			} else {
				uassert.NoError(t, err)
			}

			// then
			tt.validate(t)
		})
	}
}
