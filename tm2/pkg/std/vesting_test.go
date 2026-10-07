package std

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ugnot(n int64) Coins { return Coins{{"ugnot", n}} }

func at(unix int64) time.Time { return time.Unix(unix, 0) }

// linear vests 1000ugnot from t=100 to t=200.
func linear() VestingSchedule {
	return VestingSchedule{OriginalVesting: ugnot(1000), StartTime: 100, EndTime: 200}
}

// cliff vests 1000ugnot all at once at t=200.
func cliff() VestingSchedule {
	return VestingSchedule{OriginalVesting: ugnot(1000), EndTime: 200, Type: VestingDelayed}
}

func TestVestingSchedule_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		vs      VestingSchedule
		wantErr string
	}{
		{"the zero schedule is valid", VestingSchedule{}, ""},
		{"linear", linear(), ""},
		{"cliff", cliff(), ""},
		{
			// The zero schedule is what every account without vesting carries, so
			// the other fields must not be examined when there is no amount.
			"no amount, nonsense times", VestingSchedule{StartTime: 500, EndTime: 1}, "",
		},
		{
			"linear starting after it ends", VestingSchedule{
				OriginalVesting: ugnot(1), StartTime: 300, EndTime: 100,
			}, "must be before end-time",
		},
		{
			"linear starting exactly when it ends", VestingSchedule{
				OriginalVesting: ugnot(1), StartTime: 100, EndTime: 100,
			}, "must be before end-time",
		},
		{
			// A cliff has no start, so one that was supplied is not checked.
			"cliff ignores its start time", VestingSchedule{
				OriginalVesting: ugnot(1), StartTime: 300, EndTime: 100, Type: VestingDelayed,
			}, "",
		},
		{
			"end time of zero", VestingSchedule{
				OriginalVesting: ugnot(1), EndTime: 0, Type: VestingDelayed,
			}, "end time must be positive",
		},
		{
			"negative end time", VestingSchedule{
				OriginalVesting: ugnot(1), EndTime: -5, Type: VestingDelayed,
			}, "end time must be positive",
		},
		{
			"unsorted coins", VestingSchedule{
				OriginalVesting: Coins{{"zzz", 1}, {"aaa", 1}}, StartTime: 1, EndTime: 2,
			}, "invalid original vesting coins",
		},
		{
			"negative amount", VestingSchedule{
				OriginalVesting: Coins{{"ugnot", -5}}, StartTime: 1, EndTime: 2,
			}, "invalid original vesting coins",
		},
		{
			// A negative start wraps EndTime-StartTime and blockTime-StartTime,
			// and the vested amount then computes from the wrapped values.
			"linear starting before the epoch", VestingSchedule{
				OriginalVesting: ugnot(1), StartTime: -1, EndTime: 100,
			}, "start-time cannot be negative",
		},
		{
			"linear spanning the whole int64 range", VestingSchedule{
				OriginalVesting: ugnot(1), StartTime: math.MinInt64, EndTime: math.MaxInt64,
			}, "start-time cannot be negative",
		},
		{
			// A cliff never reads its start, so it is not constrained.
			"cliff with a negative start", VestingSchedule{
				OriginalVesting: ugnot(1), StartTime: -1, EndTime: 100, Type: VestingDelayed,
			}, "",
		},
		{
			"unknown type", VestingSchedule{
				OriginalVesting: ugnot(1), StartTime: 1, EndTime: 2, Type: "quarterly",
			}, "unknown vesting type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.vs.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			// std.Err* renders its detail only under %+v; Error() is the bare
			// type name, which is the same for every case here.
			assert.Contains(t, fmt.Sprintf("%+v", err), tt.wantErr)
		})
	}
}

func TestVestingSchedule_LinearVesting(t *testing.T) {
	t.Parallel()

	vs := linear()
	for _, tc := range []struct {
		at             int64
		vested, locked int64
	}{
		{50, 0, 1000},   // before it starts
		{100, 0, 1000},  // exactly at the start
		{101, 10, 990},  // one second in
		{150, 500, 500}, // halfway
		{199, 990, 10},  // one second from the end
		{200, 1000, 0},  // exactly at the end
		{250, 1000, 0},  // after
	} {
		assert.Equal(t, tc.vested, vs.VestedCoins(at(tc.at)).AmountOf("ugnot"),
			"vested at t=%d", tc.at)
		assert.Equal(t, tc.locked, vs.LockedCoins(at(tc.at)).AmountOf("ugnot"),
			"locked at t=%d", tc.at)
	}
}

func TestVestingSchedule_CliffVesting(t *testing.T) {
	t.Parallel()

	vs := cliff()
	for _, tc := range []struct {
		at             int64
		vested, locked int64
	}{
		{50, 0, 1000},
		{150, 0, 1000}, // halfway through, a cliff has still vested nothing
		{199, 0, 1000},
		{200, 1000, 0}, // and then all at once
		{250, 1000, 0},
	} {
		assert.Equal(t, tc.vested, vs.VestedCoins(at(tc.at)).AmountOf("ugnot"),
			"vested at t=%d", tc.at)
		assert.Equal(t, tc.locked, vs.LockedCoins(at(tc.at)).AmountOf("ugnot"),
			"locked at t=%d", tc.at)
	}
}

// A cliff schedule that carries a start time must ignore it rather than vest
// linearly from it.
func TestVestingSchedule_CliffIgnoresStartTime(t *testing.T) {
	t.Parallel()

	vs := cliff()
	vs.StartTime = 100
	assert.Zero(t, vs.VestedCoins(at(150)).AmountOf("ugnot"),
		"a cliff must vest nothing halfway to its end")
}

// The zero schedule is what almost every account carries.
func TestVestingSchedule_ZeroLocksNothing(t *testing.T) {
	t.Parallel()

	var vs VestingSchedule
	assert.True(t, vs.IsZero())
	assert.Nil(t, vs.LockedCoins(at(0)))
	assert.Nil(t, vs.LockedCoins(at(math.MaxInt32)))
	assert.Nil(t, vs.VestedCoins(at(150)))
}

func TestVestingSchedule_MultiDenom(t *testing.T) {
	t.Parallel()

	vs := VestingSchedule{
		OriginalVesting: Coins{{"atom", 500}, {"ugnot", 1000}},
		StartTime:       100,
		EndTime:         200,
	}
	vested := vs.VestedCoins(at(150))
	assert.Equal(t, int64(250), vested.AmountOf("atom"))
	assert.Equal(t, int64(500), vested.AmountOf("ugnot"))

	locked := vs.LockedCoins(at(150))
	assert.Equal(t, int64(250), locked.AmountOf("atom"))
	assert.Equal(t, int64(500), locked.AmountOf("ugnot"))
}

// Vested must never go backwards, never exceed the original, and must always
// account for exactly the original amount together with what is locked.
func TestVestingSchedule_MonotonicAndConserved(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		vs   VestingSchedule
	}{
		{"tiny amount over a long schedule", VestingSchedule{
			OriginalVesting: ugnot(7), StartTime: 0, EndTime: 1000,
		}},
		{"amount smaller than the duration", VestingSchedule{
			OriginalVesting: ugnot(3), StartTime: 0, EndTime: 5000,
		}},
		{"prime duration", VestingSchedule{
			OriginalVesting: ugnot(1000), StartTime: 13, EndTime: 997,
		}},
		{"maximum amount", VestingSchedule{
			OriginalVesting: ugnot(math.MaxInt64), StartTime: 0, EndTime: 1000,
		}},
		{"the largest span Validate accepts", VestingSchedule{
			OriginalVesting: ugnot(1000), StartTime: 0, EndTime: math.MaxInt64,
		}},
		{"cliff", cliff()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			original := tc.vs.OriginalVesting.AmountOf("ugnot")
			var prev int64
			for now := tc.vs.StartTime - 2; now <= tc.vs.EndTime+2; now++ {
				vested := tc.vs.VestedCoins(at(now)).AmountOf("ugnot")
				locked := tc.vs.LockedCoins(at(now)).AmountOf("ugnot")

				require.GreaterOrEqual(t, vested, prev, "vested went backwards at t=%d", now)
				require.LessOrEqual(t, vested, original, "over-vested at t=%d", now)
				require.GreaterOrEqual(t, locked, int64(0), "locked went negative at t=%d", now)
				require.Equal(t, original, vested+locked,
					"vested+locked must equal the original at t=%d", now)
				prev = vested
			}
		})
	}
}

// The account is the thing the bank asks, so the lock has to be readable
// through it and not only off the schedule.
func TestBaseAccount_LockedCoins(t *testing.T) {
	t.Parallel()

	addr := crypto.AddressFromPreimage([]byte("vester"))
	acc := NewBaseAccount(addr, ugnot(1000), nil, 0, 0)

	// No schedule: nothing locked, whatever the time.
	assert.True(t, acc.LockedCoins(at(150)).IsZero())
	assert.True(t, acc.GetVesting().IsZero())

	acc.SetVesting(linear())
	assert.Equal(t, int64(1000), acc.LockedCoins(at(50)).AmountOf("ugnot"))
	assert.Equal(t, int64(500), acc.LockedCoins(at(150)).AmountOf("ugnot"))
	assert.True(t, acc.LockedCoins(at(250)).IsZero())
	assert.Equal(t, linear(), acc.GetVesting())
}

// Every account type answers, so the bank never has to ask which one it holds.
func TestEveryAccountTypeReportsLockedCoins(t *testing.T) {
	t.Parallel()

	for _, acc := range []Account{
		&BaseAccount{},
		&BaseSessionAccount{},
	} {
		assert.True(t, acc.LockedCoins(at(150)).IsZero(), "%T", acc)
		assert.True(t, acc.GetVesting().IsZero(), "%T", acc)
	}
}

// A schedule survives being stored and read back, and an account without one
// costs no bytes to store.
func TestBaseAccount_VestingAminoRoundTrip(t *testing.T) {
	t.Parallel()

	addr := crypto.AddressFromPreimage([]byte("vester"))

	plain := NewBaseAccount(addr, ugnot(1000), nil, 7, 3)
	vesting := NewBaseAccount(addr, ugnot(1000), nil, 7, 3)
	vesting.SetVesting(linear())

	// The zero schedule costs nothing: explicitly setting one must produce the
	// same bytes as never touching the field. That is what keeps an account
	// without vesting encoding exactly as it did before the field existed, and
	// so keeps the field off the app hash.
	zeroed := NewBaseAccount(addr, ugnot(1000), nil, 7, 3)
	zeroed.SetVesting(VestingSchedule{})
	assert.Equal(t, amino.MustMarshal(plain), amino.MustMarshal(zeroed),
		"a zero schedule must add no bytes")
	assert.Greater(t, len(amino.MustMarshal(vesting)), len(amino.MustMarshal(plain)),
		"a real schedule must add some, or the comparison above proves nothing")

	var back BaseAccount
	require.NoError(t, amino.Unmarshal(amino.MustMarshal(vesting), &back))
	assert.Equal(t, linear(), back.GetVesting())
	assert.Equal(t, int64(500), back.LockedCoins(at(150)).AmountOf("ugnot"))

	var backPlain BaseAccount
	require.NoError(t, amino.Unmarshal(amino.MustMarshal(plain), &backPlain))
	assert.True(t, backPlain.GetVesting().IsZero(), "a stored account without a schedule reads back without one")
}

// An account with no schedule must encode exactly as it would have before the
// field was added, in both binary and JSON. That is what keeps the field off
// the app hash and out of query output.
func TestZeroScheduleCostsNothingToStore(t *testing.T) {
	t.Parallel()

	acc := NewBaseAccount(crypto.AddressFromPreimage([]byte("plain")), ugnot(1000), nil, 7, 3)

	assert.NotContains(t, string(amino.MustMarshalJSON(acc)), "vesting",
		"a zero schedule must not appear in JSON")

	acc.SetVesting(linear())
	assert.Contains(t, string(amino.MustMarshalJSON(acc)), "vesting",
		"and a real one must")
}

func TestVestingSchedule_String(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "none", VestingSchedule{}.String())
	assert.Contains(t, linear().String(), "continuous")
	assert.Contains(t, cliff().String(), "delayed")
}

// A schedule has to be visible in the account's own String output, which is what
// logs and CLI output use. An account without one must print exactly as it did
// before the field existed.
func TestBaseAccount_StringShowsAScheduleOnlyWhenThereIsOne(t *testing.T) {
	t.Parallel()

	acc := NewBaseAccount(crypto.AddressFromPreimage([]byte("vester")), ugnot(1000), nil, 7, 3)
	assert.NotContains(t, acc.String(), "Vesting",
		"an account with no schedule must not gain a line")

	acc.SetVesting(linear())
	assert.Contains(t, acc.String(), "Vesting")
	assert.Contains(t, acc.String(), "1000ugnot", "the amount must be shown")
}

// Whatever Validate accepts must conserve at any block time: the vested amount
// stays within the grant, and vested plus locked is exactly the grant.
//
// This sweeps the extremes of the int64 range rather than plausible dates,
// because the failure it guards is arithmetic, not calendar. A start before the
// epoch wraps both EndTime-StartTime and blockTime-StartTime, and the vested
// amount is then computed from the wrapped values: a grant of 1000 vested 2000,
// which drives locked negative and makes the spendable figure exceed the balance.
func TestVestingSchedule_AcceptedSchedulesAlwaysConserve(t *testing.T) {
	t.Parallel()

	const orig = 1000
	starts := []int64{math.MinInt64, math.MinInt64 + 1, -(1 << 62), -1, 0, 1, 1 << 20, 1 << 62, math.MaxInt64 - 1}
	ends := []int64{math.MinInt64, -1, 0, 1, 100, 1 << 20, 1 << 62, math.MaxInt64 - 1, math.MaxInt64}
	nows := []int64{
		math.MinInt64, math.MinInt64 + 1, -(1 << 62), -1, 0, 1, 99, 100,
		1 << 20, 1893456000, 1 << 62, math.MaxInt64 - 1, math.MaxInt64,
	}

	accepted := 0
	for _, start := range starts {
		for _, end := range ends {
			for _, kind := range []VestingScheduleType{VestingContinuous, VestingDelayed} {
				vs := VestingSchedule{
					OriginalVesting: ugnot(orig),
					StartTime:       start,
					EndTime:         end,
					Type:            kind,
				}
				if vs.Validate() != nil {
					continue
				}
				accepted++
				for _, now := range nows {
					vested := vs.VestedCoins(at(now)).AmountOf("ugnot")
					locked := vs.LockedCoins(at(now)).AmountOf("ugnot")

					require.GreaterOrEqual(t, vested, int64(0),
						"start=%d end=%d type=%q now=%d", start, end, kind, now)
					require.LessOrEqual(t, vested, int64(orig),
						"start=%d end=%d type=%q now=%d", start, end, kind, now)
					require.GreaterOrEqual(t, locked, int64(0),
						"start=%d end=%d type=%q now=%d", start, end, kind, now)
					require.Equal(t, int64(orig), vested+locked,
						"start=%d end=%d type=%q now=%d", start, end, kind, now)
				}
			}
		}
	}
	// Guards the sweep itself: a Validate that rejected everything would pass
	// every assertion above without checking anything.
	require.Greater(t, accepted, 50, "the sweep must actually accept schedules to check")
}
