package std

import (
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// VestingSchedule tests

func TestVestingSchedule_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		schedule VestingSchedule
		wantErr  bool
	}{
		{
			"valid",
			VestingSchedule{OriginalVesting: Coins{NewCoin("ugnot", 100)}, StartTime: 100, EndTime: 200},
			false,
		},
		{
			"negative end time",
			VestingSchedule{OriginalVesting: Coins{NewCoin("ugnot", 100)}, StartTime: 100, EndTime: -1},
			true,
		},
		{
			"start >= end",
			VestingSchedule{OriginalVesting: Coins{NewCoin("ugnot", 100)}, StartTime: 200, EndTime: 100},
			true,
		},
		{
			"zero vesting is valid",
			VestingSchedule{OriginalVesting: Coins{}, StartTime: 100, EndTime: 200},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.schedule.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// ContinuousVestingAccount tests

func TestContinuousVestingAccount_GetVestedCoins(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, Coins{NewCoin("ugnot", 1000)}, pubKey, 0, 0)

	schedule := VestingSchedule{
		OriginalVesting: Coins{NewCoin("ugnot", 1000)},
		StartTime:       100,
		EndTime:         200,
	}

	cva, err := NewContinuousVestingAccount(baseAcc, schedule)
	require.NoError(t, err)

	tests := []struct {
		name      string
		blockTime time.Time
		wantAmt   int64
	}{
		{"before start", time.Unix(50, 0), 0},
		{"at start", time.Unix(100, 0), 0},
		{"halfway", time.Unix(150, 0), 500},
		{"at end", time.Unix(200, 0), 1000},
		{"after end", time.Unix(300, 0), 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vested := cva.GetVestedCoins(tt.blockTime)
			gotAmt := vested.AmountOf("ugnot")
			assert.Equal(t, tt.wantAmt, gotAmt, "vested amount mismatch")
		})
	}
}

func TestContinuousVestingAccount_LockedCoins(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, Coins{NewCoin("ugnot", 1000)}, pubKey, 0, 0)

	cva, err := NewContinuousVestingAccount(baseAcc, VestingSchedule{
		OriginalVesting: Coins{NewCoin("ugnot", 1000)},
		StartTime:       100,
		EndTime:         200,
	})
	require.NoError(t, err)

	assert.Equal(t, int64(1000), cva.LockedCoins(time.Unix(50, 0)).AmountOf("ugnot"))
	assert.Equal(t, int64(500), cva.LockedCoins(time.Unix(150, 0)).AmountOf("ugnot"))
	assert.True(t, cva.LockedCoins(time.Unix(300, 0)).IsZero())
}

func TestContinuousVestingAccount_MultiDenom(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, Coins{
		NewCoin("uatom", 500),
		NewCoin("ugnot", 1000),
	}, pubKey, 0, 0)

	cva, err := NewContinuousVestingAccount(baseAcc, VestingSchedule{
		OriginalVesting: Coins{NewCoin("uatom", 500), NewCoin("ugnot", 1000)},
		StartTime:       100,
		EndTime:         200,
	})
	require.NoError(t, err)

	vested := cva.GetVestedCoins(time.Unix(150, 0))
	assert.Equal(t, int64(500), vested.AmountOf("ugnot"))
	assert.Equal(t, int64(250), vested.AmountOf("uatom"))

	vesting := cva.GetVestingCoins(time.Unix(150, 0))
	assert.Equal(t, int64(500), vesting.AmountOf("ugnot"))
	assert.Equal(t, int64(250), vesting.AmountOf("uatom"))
}

func TestContinuousVestingAccount_AminoRoundTrip(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()

	original := &ContinuousVestingAccount{
		BaseVestingAccount: BaseVestingAccount{
			BaseAccount: BaseAccount{
				Address:       addr,
				PubKey:        pubKey,
				Coins:         Coins{NewCoin("ugnot", 500)},
				AccountNumber: 42,
				Sequence:      7,
			},
			VestingSchedule: VestingSchedule{
				OriginalVesting: Coins{NewCoin("ugnot", 1000)},
				StartTime:       100,
				EndTime:         200,
			},
		},
	}

	bz, err := amino.MarshalAny(original)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	var got interface{}
	err = amino.UnmarshalAny(bz, &got)
	require.NoError(t, err)

	result, ok := got.(*ContinuousVestingAccount)
	require.True(t, ok, "expected *ContinuousVestingAccount, got %T", got)

	assert.Equal(t, original.Address, result.Address)
	assert.True(t, original.PubKey.Equals(result.PubKey))
	assert.True(t, original.Coins.IsEqual(result.Coins))
	assert.Equal(t, original.AccountNumber, result.AccountNumber)
	assert.Equal(t, original.Sequence, result.Sequence)
	assert.True(t, original.OriginalVesting.IsEqual(result.OriginalVesting))
	assert.Equal(t, original.StartTime, result.StartTime)
	assert.Equal(t, original.EndTime, result.EndTime)
}

// -----------------------------------------------------------------------------
// DelayedVestingAccount tests

func TestDelayedVestingAccount_GetVestedCoins(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, Coins{NewCoin("ugnot", 1000)}, pubKey, 0, 0)

	dva, err := NewDelayedVestingAccount(baseAcc, VestingSchedule{
		OriginalVesting: Coins{NewCoin("ugnot", 1000)},
		EndTime:         200,
	})
	require.NoError(t, err)

	assert.Equal(t, int64(0), dva.GetVestedCoins(time.Unix(100, 0)).AmountOf("ugnot"))
	assert.Equal(t, int64(1000), dva.GetVestedCoins(time.Unix(200, 0)).AmountOf("ugnot"))
	assert.Equal(t, int64(1000), dva.GetVestedCoins(time.Unix(300, 0)).AmountOf("ugnot"))
}

func TestDelayedVestingAccount_LockedCoins(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, Coins{NewCoin("ugnot", 1000)}, pubKey, 0, 0)

	dva, err := NewDelayedVestingAccount(baseAcc, VestingSchedule{
		OriginalVesting: Coins{NewCoin("ugnot", 1000)},
		EndTime:         200,
	})
	require.NoError(t, err)

	assert.Equal(t, int64(1000), dva.LockedCoins(time.Unix(100, 0)).AmountOf("ugnot"))
	assert.True(t, dva.LockedCoins(time.Unix(300, 0)).IsZero())
}

func TestDelayedVestingAccount_AminoRoundTrip(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()

	original := &DelayedVestingAccount{
		BaseVestingAccount: BaseVestingAccount{
			BaseAccount: BaseAccount{
				Address:       addr,
				PubKey:        pubKey,
				Coins:         Coins{NewCoin("ugnot", 500)},
				AccountNumber: 42,
				Sequence:      7,
			},
			VestingSchedule: VestingSchedule{
				OriginalVesting: Coins{NewCoin("ugnot", 1000)},
				EndTime:         200,
			},
		},
	}

	bz, err := amino.MarshalAny(original)
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	var got interface{}
	err = amino.UnmarshalAny(bz, &got)
	require.NoError(t, err)

	result, ok := got.(*DelayedVestingAccount)
	require.True(t, ok, "expected *DelayedVestingAccount, got %T", got)

	assert.Equal(t, original.Address, result.Address)
	assert.True(t, original.PubKey.Equals(result.PubKey))
	assert.True(t, original.Coins.IsEqual(result.Coins))
	assert.Equal(t, original.AccountNumber, result.AccountNumber)
	assert.Equal(t, original.Sequence, result.Sequence)
	assert.True(t, original.OriginalVesting.IsEqual(result.OriginalVesting))
	assert.Equal(t, original.EndTime, result.EndTime)
}

// -----------------------------------------------------------------------------
// SpendableCoins tests

func TestSpendableCoins(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, Coins{NewCoin("ugnot", 1000)}, pubKey, 0, 0)

	cva, err := NewContinuousVestingAccount(baseAcc, VestingSchedule{
		OriginalVesting: Coins{NewCoin("ugnot", 500)},
		StartTime:       100,
		EndTime:         200,
	})
	require.NoError(t, err)

	spendable := SpendableCoins(cva, time.Unix(50, 0))
	assert.Equal(t, int64(500), spendable.AmountOf("ugnot"))

	spendable = SpendableCoins(cva, time.Unix(300, 0))
	assert.Equal(t, int64(1000), spendable.AmountOf("ugnot"))

	spendable = SpendableCoins(cva, time.Unix(150, 0))
	assert.Equal(t, int64(750), spendable.AmountOf("ugnot"))
}

func TestSpendableCoins_NoVesting(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, Coins{NewCoin("ugnot", 1000)}, pubKey, 0, 0)

	cva, err := NewContinuousVestingAccount(baseAcc, VestingSchedule{
		OriginalVesting: Coins{},
		StartTime:       100,
		EndTime:         200,
	})
	require.NoError(t, err)

	spendable := SpendableCoins(cva, time.Unix(50, 0))
	assert.True(t, spendable.IsEqual(Coins{NewCoin("ugnot", 1000)}))
}

func TestSpendableCoins_ZeroBalance(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, NewCoins(), pubKey, 0, 0)

	cva, err := NewContinuousVestingAccount(baseAcc, VestingSchedule{
		OriginalVesting: Coins{},
		StartTime:       100,
		EndTime:         200,
	})
	require.NoError(t, err)

	spendable := SpendableCoins(cva, time.Unix(50, 0))
	assert.True(t, spendable.IsZero())
}

func TestContinuousVestingAccount_PartialVesting(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()

	// 1000 total, only 300 vesting. 700 should be immediately spendable.
	baseAcc := NewBaseAccount(addr, Coins{NewCoin("ugnot", 1000)}, pubKey, 0, 0)

	cva, err := NewContinuousVestingAccount(baseAcc, VestingSchedule{
		OriginalVesting: Coins{NewCoin("ugnot", 300)},
		StartTime:       100,
		EndTime:         200,
	})
	require.NoError(t, err)

	assert.Equal(t, int64(1000), cva.GetCoins().AmountOf("ugnot"), "total balance")
	assert.Equal(t, int64(300), cva.GetOriginalVesting().AmountOf("ugnot"), "vesting amount")

	// Before start: all 300 locked, 700 spendable.
	assert.Equal(t, int64(300), cva.LockedCoins(time.Unix(50, 0)).AmountOf("ugnot"))
	assert.Equal(t, int64(700), SpendableCoins(cva, time.Unix(50, 0)).AmountOf("ugnot"))

	// Halfway: 150 vested, 150 locked, 850 spendable.
	assert.Equal(t, int64(150), cva.LockedCoins(time.Unix(150, 0)).AmountOf("ugnot"))
	assert.Equal(t, int64(850), SpendableCoins(cva, time.Unix(150, 0)).AmountOf("ugnot"))

	// After end: 0 locked, all 1000 spendable.
	assert.True(t, cva.LockedCoins(time.Unix(300, 0)).IsZero())
	assert.Equal(t, int64(1000), SpendableCoins(cva, time.Unix(300, 0)).AmountOf("ugnot"))
}

func TestContinuousVestingAccount_PartialVestingMultiDenom(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()

	baseAcc := NewBaseAccount(addr, Coins{
		NewCoin("uatom", 500),
		NewCoin("ugnot", 1000),
	}, pubKey, 0, 0)

	cva, err := NewContinuousVestingAccount(baseAcc, VestingSchedule{
		OriginalVesting: Coins{NewCoin("uatom", 200), NewCoin("ugnot", 300)},
		StartTime:       100,
		EndTime:         200,
	})
	require.NoError(t, err)

	// Before start: ugnot locked=300 spendable=700, uatom locked=200 spendable=300.
	spendable := SpendableCoins(cva, time.Unix(50, 0))
	assert.Equal(t, int64(700), spendable.AmountOf("ugnot"))
	assert.Equal(t, int64(300), spendable.AmountOf("uatom"))

	// After end: everything spendable.
	spendable = SpendableCoins(cva, time.Unix(300, 0))
	assert.Equal(t, int64(1000), spendable.AmountOf("ugnot"))
	assert.Equal(t, int64(500), spendable.AmountOf("uatom"))
}

func TestContinuousVestingAccount_RejectsVestingExceedingBalance(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()

	// 500 total, 1000 vesting — should be rejected.
	baseAcc := NewBaseAccount(addr, Coins{NewCoin("ugnot", 500)}, pubKey, 0, 0)

	_, err := NewContinuousVestingAccount(baseAcc, VestingSchedule{
		OriginalVesting: Coins{NewCoin("ugnot", 1000)},
		StartTime:       100,
		EndTime:         200,
	})
	assert.Error(t, err)
}

// -----------------------------------------------------------------------------
// Interface compliance

func TestContinuousVestingAccount_SatisfiesVestingAccount(t *testing.T) {
	var va VestingAccount = &ContinuousVestingAccount{}
	assert.NotNil(t, va)
}

func TestDelayedVestingAccount_SatisfiesVestingAccount(t *testing.T) {
	var va VestingAccount = &DelayedVestingAccount{}
	assert.NotNil(t, va)
}
