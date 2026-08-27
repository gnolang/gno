package std

import (
	"fmt"
	"math/big"
	"time"
)

// VestingSchedule locks part of an account's balance until it vests.
//
// It is a field on BaseAccount rather than a separate account type, so every
// account has one and almost every account's is the zero value, which locks
// nothing. Enforcement therefore never depends on which concrete account type
// is stored -- reading the field always gives the right answer, where a type
// assertion silently gave none for a type that had been missed.
//
// Times are unix seconds. Nothing compares them against the chain's own clock,
// so a schedule that already ended when the chain starts is valid and vests
// everything immediately. That has to stay allowed: replaying an old genesis
// onto a fork is legitimate and its schedules are legitimately in the past.
type VestingSchedule struct {
	OriginalVesting Coins `json:"original_vesting,omitempty" yaml:"original_vesting,omitempty"`
	// StartTime is when linear vesting begins. Delayed schedules have no start
	// time and ignore it.
	StartTime int64               `json:"start_time,omitempty" yaml:"start_time,omitempty"`
	EndTime   int64               `json:"end_time,omitempty" yaml:"end_time,omitempty"`
	Type      VestingScheduleType `json:"type,omitempty" yaml:"type,omitempty"`
}

// VestingScheduleType selects the curve. It is read every time coins are
// counted, so it cannot drift out of agreement with how an account actually
// vests.
type VestingScheduleType string

const (
	VestingContinuous VestingScheduleType = ""        // default -- linear vesting
	VestingDelayed    VestingScheduleType = "delayed" // cliff vesting
)

// Validate checks the schedule fields.
func (vs VestingSchedule) Validate() error {
	if vs.IsZero() {
		return nil // no schedule; the other fields do not matter
	}
	if vs.EndTime <= 0 {
		return ErrInvalidVestingSchedule(fmt.Sprintf("end time must be positive: %d", vs.EndTime))
	}
	// Only linear vesting has a start. Requiring one for a cliff would make
	// genesis supply a number that nothing reads.
	if vs.Type != VestingDelayed {
		// A start before the epoch is meaningless for a chain, and rejecting it
		// is what keeps the arithmetic below safe: with 0 <= StartTime < EndTime,
		// neither EndTime-StartTime nor blockTime-StartTime can overflow int64.
		// A start of MinInt64 wraps both, and the vested amount then computes
		// from the wrapped values and can exceed the whole grant.
		if vs.StartTime < 0 {
			return ErrInvalidVestingSchedule(fmt.Sprintf(
				"vesting start-time cannot be negative: %d", vs.StartTime))
		}
		if vs.StartTime >= vs.EndTime {
			return ErrInvalidVestingSchedule(fmt.Sprintf(
				"vesting start-time (%d) must be before end-time (%d)",
				vs.StartTime, vs.EndTime,
			))
		}
	}
	if !vs.OriginalVesting.IsValid() {
		return ErrInvalidVestingSchedule(fmt.Sprintf("invalid original vesting coins: %s", vs.OriginalVesting))
	}
	if vs.Type != VestingContinuous && vs.Type != VestingDelayed {
		return ErrInvalidVestingSchedule(fmt.Sprintf("unknown vesting type: %q", vs.Type))
	}
	return nil
}

// IsZero returns true if the schedule has no vesting amount.
func (vs VestingSchedule) IsZero() bool {
	return vs.OriginalVesting.IsZero()
}

// VestedCoins returns the coins that have vested by blockTime.
func (vs VestingSchedule) VestedCoins(blockTime time.Time) Coins {
	if vs.IsZero() {
		return nil
	}
	now := blockTime.Unix()
	if now >= vs.EndTime {
		return vs.OriginalVesting
	}
	if vs.Type == VestingDelayed {
		return nil // a cliff vests nothing until EndTime
	}
	if now <= vs.StartTime {
		return nil
	}

	elapsed := now - vs.StartTime
	total := vs.EndTime - vs.StartTime

	var vested Coins
	for _, ovc := range vs.OriginalVesting {
		// big.Int because amount*elapsed overflows int64 for large grants. The
		// quotient always fits: elapsed < total, so it is below ovc.Amount.
		product := new(big.Int).Mul(big.NewInt(ovc.Amount), big.NewInt(elapsed))
		amount := new(big.Int).Div(product, big.NewInt(total)).Int64()
		if amount > 0 {
			vested = append(vested, Coin{ovc.Denom, amount})
		}
	}
	return vested
}

// LockedCoins returns the coins that have not vested by blockTime, which are
// the ones that cannot be transferred out.
//
// "Not spendable" means not transferable. Locked coins can still leave the
// account as gas fees and storage deposits, which debit through the
// unrestricted path and never consult a schedule. So a schedule caps what a
// holder can move to another address; it does not reserve a balance.
func (vs VestingSchedule) LockedCoins(blockTime time.Time) Coins {
	if vs.IsZero() {
		return nil
	}
	return vs.OriginalVesting.SubUnsafe(vs.VestedCoins(blockTime))
}

// String implements fmt.Stringer.
func (vs VestingSchedule) String() string {
	if vs.IsZero() {
		return "none"
	}
	kind := "continuous"
	if vs.Type == VestingDelayed {
		return fmt.Sprintf("%s delayed until %d", vs.OriginalVesting, vs.EndTime)
	}
	return fmt.Sprintf("%s %s from %d to %d", vs.OriginalVesting, kind, vs.StartTime, vs.EndTime)
}
