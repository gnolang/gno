package governance

import (
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
	registry := createParameterHandlers()

	if registry == nil {
		t.Fatal("registry is nil")
	}

	tests := []struct {
		name     string
		path     string
		function string
		params   []string
		wantErr  bool
		validate func(t *testing.T)
	}{
		{
			name:     "Emission_ChangeDistributionPct",
			path:     consts.EMISSION_PATH,
			function: "ChangeDistributionPct",
			params:   []string{"1", "7000", "2", "1500", "3", "1000", "4", "500"},
			wantErr:  false,
			validate: func(t *testing.T) {
				uassert.Equal(t, en.GetDistributionBpsPct(1), uint64(7000))
				uassert.Equal(t, en.GetDistributionBpsPct(2), uint64(1500))
				uassert.Equal(t, en.GetDistributionBpsPct(3), uint64(1000))
				uassert.Equal(t, en.GetDistributionBpsPct(4), uint64(500))
			},
		},
		{
			name:     "Router SetSwapFee: Pass",
			path:     consts.ROUTER_PATH,
			function: "SetSwapFee",
			params:   []string{"1000"},
			wantErr:  false,
			validate: func(t *testing.T) {
				uassert.Equal(t, rr.GetSwapFee(), uint64(1000))
			},
		},
		{
			name:     "set deposit gns amount",
			path:     consts.STAKER_PATH,
			function: "SetDepositGnsAmount",
			params:   []string{"1000"},
			wantErr:  false,
			validate: func(t *testing.T) {
				uassert.Equal(t, sr.GetDepositGnsAmount(), uint64(1000))
			},
		},
		{
			name:     "set pool creation fee",
			path:     consts.POOL_PATH,
			function: "SetPoolCreationFee",
			params:   []string{"500"},
			wantErr:  false,
			validate: func(t *testing.T) {
				uassert.Equal(t, pl.GetPoolCreationFee(), uint64(500))
			},
		},
		{
			name:     "SetWithdrawalFee",
			path:     consts.POOL_PATH,
			function: "SetWithdrawalFee",
			params:   []string{"600"},
			wantErr:  false,
			validate: func(t *testing.T) {
				uassert.Equal(t, pl.GetWithdrawalFee(), uint64(600))
			},
		},
		{
			name:     "ProtocolFee_SetDevOpsPct",
			path:     consts.PROTOCOL_FEE_PATH,
			function: "SetDevOpsPct",
			params:   []string{"900"},
			wantErr:  false,
			validate: func(t *testing.T) {
				uassert.Equal(t, pf.GetDevOpsPct(), uint64(900))
			},
		},
		{
			name:     "Router_SetSwapFee",
			path:     consts.ROUTER_PATH,
			function: "SetSwapFee",
			params:   []string{"400"},
			wantErr:  false,
			validate: func(t *testing.T) {
				uassert.Equal(t, rr.GetSwapFee(), uint64(400))
			},
		},
		{
			name:     "Staker_SetDepositGnsAmount",
			path:     consts.STAKER_PATH,
			function: "SetDepositGnsAmount",
			params:   []string{"400"},
			wantErr:  false,
			validate: func(t *testing.T) {
				uassert.Equal(t, sr.GetDepositGnsAmount(), uint64(400))
			},
		},
		{
			name:     "Staker_SetUnStakingFee",
			path:     consts.STAKER_PATH,
			function: "SetUnStakingFee",
			params:   []string{"100"},
			wantErr:  false,
			validate: func(t *testing.T) {
				uassert.Equal(t, sr.GetUnstakingFee(), uint64(100))
			},
		},
		{
			name:     "ProtocolFee_SetDevOpsPct",
			path:     consts.PROTOCOL_FEE_PATH,
			function: "SetDevOpsPct",
			params:   []string{"900"},
			wantErr:  false,
			validate: func(t *testing.T) {
				uassert.Equal(t, pf.GetDevOpsPct(), uint64(900))
			},
		},
		{
			name:     "Router_SetSwapFee",
			path:     consts.ROUTER_PATH,
			function: "SetSwapFee",
			params:   []string{"400"},
			wantErr:  false,
			validate: func(t *testing.T) {
				uassert.Equal(t, rr.GetSwapFee(), uint64(400))
			},
		},
		{
			name:     "Staker_SetDepositGnsAmount",
			path:     consts.STAKER_PATH,
			function: "SetDepositGnsAmount",
			params:   []string{"400"},
			wantErr:  false,
			validate: func(t *testing.T) {
				uassert.Equal(t, sr.GetDepositGnsAmount(), uint64(400))
			},
		},
		{
			name:     "Emission_ChangeDistributionPct_InvalidTotal_Exceed100",
			path:     consts.EMISSION_PATH,
			function: "ChangeDistributionPct",
			params:   []string{"1", "8000", "2", "3000"},
			wantErr:  true,
			validate: func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						return
					}
				}()
				// existing state must be preserved in the event of a panic
				uassert.Equal(t, en.GetDistributionBpsPct(1), uint64(7000))
			},
		},
		{
			name:     "Router_SetSwapFee_InvalidValue",
			path:     consts.ROUTER_PATH,
			function: "SetSwapFee",
			params:   []string{"-100"},
			wantErr:  true,
			validate: func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						return
					}
				}()
				uassert.Equal(t, rr.GetSwapFee(), uint64(400))
			},
		},
		{
			name:     "Staker_SetUnStakingFee_TooHigh",
			path:     consts.STAKER_PATH,
			function: "SetUnStakingFee",
			params:   []string{"10001"},
			wantErr:  true,
			validate: func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						return
					}
				}()
				uassert.Equal(t, sr.GetUnstakingFee(), uint64(100))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					return
				} else if r != nil {
					t.Fatal("unexpected panic")
				}
			}()
			handler, _ := registry.Get(tt.path, tt.function)
			err := handler(tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantErr {
				tt.validate(t)
			}
		})
	}
}
