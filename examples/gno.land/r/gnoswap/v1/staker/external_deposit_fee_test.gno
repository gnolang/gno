package staker

import (
	"std"
	"testing"

	"gno.land/p/gnoswap/consts"
)

func TestGetDepositGnsAmount(t *testing.T) {
	expected := uint64(1_000_000_000)
	actual := GetDepositGnsAmount()

	if actual != expected {
		t.Errorf("GetDepositGnsAmount() = %d; want %d", actual, expected)
	}
}

func TestSetDepositGnsAmountByAdmin(t *testing.T) {
	tests := []struct {
		name        string
		caller      std.Address
		newAmount   uint64
		shouldPanic bool
	}{
		{
			name:        "Success - Admin sets deposit amount",
			caller:      consts.ADMIN,
			newAmount:   2_000_000_000,
			shouldPanic: false,
		},
		{
			name:        "Failure - Non-admin tries to set deposit amount",
			caller:      std.Address("user1"),
			newAmount:   2_000_000_000,
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std.TestSetOrigCaller(tt.caller)

			defer func() {
				r := recover()
				if r != nil {
					if !tt.shouldPanic {
						t.Errorf("Unexpected panic: %v", r)
					}
				} else if tt.shouldPanic {
					t.Errorf("Expected panic but did not occur")
				}
			}()

			SetDepositGnsAmountByAdmin(tt.newAmount)

			if !tt.shouldPanic {
				actual := GetDepositGnsAmount()
				if actual != tt.newAmount {
					t.Errorf("SetDepositGnsAmountByAdmin() = %d; want %d", actual, tt.newAmount)
				}
			}
		})
	}
}

func TestSetDepositGnsAmount(t *testing.T) {
	tests := []struct {
		name        string
		caller      std.Address
		newAmount   uint64
		shouldPanic bool
	}{
		{
			name:        "Success - Governance sets deposit amount",
			caller:      consts.GOV_GOVERNANCE_ADDR,
			newAmount:   3_000_000_000,
			shouldPanic: false,
		},
		{
			name:        "Failure - Non-governance tries to set deposit amount",
			caller:      std.Address("user2"),
			newAmount:   3_000_000_000,
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std.TestSetOrigCaller(tt.caller)

			defer func() {
				r := recover()
				if r != nil {
					if !tt.shouldPanic {
						t.Errorf("Unexpected panic: %v", r)
					}
				} else if tt.shouldPanic {
					t.Errorf("Expected panic but did not occur")
				}
			}()

			SetDepositGnsAmount(tt.newAmount)

			if !tt.shouldPanic {
				actual := GetDepositGnsAmount()
				if actual != tt.newAmount {
					t.Errorf("SetDepositGnsAmount() = %d; want %d", actual, tt.newAmount)
				}
			}
		})
	}
}
