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
// ContinuousVestingAccount tests

func TestContinuousVestingAccount_GetVestedCoins(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, Coins{NewCoin("ugnot", 1000)}, pubKey, 0, 0)

	originalVesting := Coins{NewCoin("ugnot", 1000)}
	startTime := int64(100)
	endTime := int64(200)

	cva, err := NewContinuousVestingAccount(baseAcc, originalVesting, startTime, endTime)
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

func TestContinuousVestingAccount_GetVestingCoins(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, Coins{NewCoin("ugnot", 1000)}, pubKey, 0, 0)

	originalVesting := Coins{NewCoin("ugnot", 1000)}
	cva, err := NewContinuousVestingAccount(baseAcc, originalVesting, 100, 200)
	require.NoError(t, err)

	tests := []struct {
		name      string
		blockTime time.Time
		wantAmt   int64
	}{
		{"before start", time.Unix(50, 0), 1000},
		{"halfway", time.Unix(150, 0), 500},
		{"after end", time.Unix(300, 0), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vesting := cva.GetVestingCoins(tt.blockTime)
			gotAmt := vesting.AmountOf("ugnot")
			assert.Equal(t, tt.wantAmt, gotAmt, "vesting amount mismatch")
		})
	}
}

func TestContinuousVestingAccount_LockedCoins(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, Coins{NewCoin("ugnot", 1000)}, pubKey, 0, 0)

	originalVesting := Coins{NewCoin("ugnot", 1000)}
	cva, err := NewContinuousVestingAccount(baseAcc, originalVesting, 100, 200)
	require.NoError(t, err)

	// Before start: all coins are locked.
	locked := cva.LockedCoins(time.Unix(50, 0))
	assert.Equal(t, int64(1000), locked.AmountOf("ugnot"))

	// Halfway: half locked.
	locked = cva.LockedCoins(time.Unix(150, 0))
	assert.Equal(t, int64(500), locked.AmountOf("ugnot"))

	// After end: none locked.
	locked = cva.LockedCoins(time.Unix(300, 0))
	assert.True(t, locked.IsZero(), "expected no locked coins after vesting ends")
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

	originalVesting := Coins{
		NewCoin("uatom", 500),
		NewCoin("ugnot", 1000),
	}
	cva, err := NewContinuousVestingAccount(baseAcc, originalVesting, 100, 200)
	require.NoError(t, err)

	// Halfway: half of each vested.
	vested := cva.GetVestedCoins(time.Unix(150, 0))
	assert.Equal(t, int64(500), vested.AmountOf("ugnot"))
	assert.Equal(t, int64(250), vested.AmountOf("uatom"))

	vesting := cva.GetVestingCoins(time.Unix(150, 0))
	assert.Equal(t, int64(500), vesting.AmountOf("ugnot"))
	assert.Equal(t, int64(250), vesting.AmountOf("uatom"))
}

func TestContinuousVestingAccount_Validate(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, NewCoins(), pubKey, 0, 0)
	vesting := Coins{NewCoin("ugnot", 1000)}

	tests := []struct {
		name      string
		startTime int64
		endTime   int64
		wantErr   bool
	}{
		{"valid", 100, 200, false},
		{"start equals end", 100, 100, true},
		{"start after end", 200, 100, true},
		{"negative end time", 100, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewContinuousVestingAccount(baseAcc, vesting, tt.startTime, tt.endTime)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
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
			OriginalVesting: Coins{NewCoin("ugnot", 1000)},
			EndTime:         200,
		},
		StartTime: 100,
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

	dva, err := NewDelayedVestingAccount(baseAcc, Coins{NewCoin("ugnot", 1000)}, 200)
	require.NoError(t, err)

	tests := []struct {
		name      string
		blockTime time.Time
		wantAmt   int64
	}{
		{"before cliff", time.Unix(100, 0), 0},
		{"at cliff", time.Unix(200, 0), 1000},
		{"after cliff", time.Unix(300, 0), 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vested := dva.GetVestedCoins(tt.blockTime)
			gotAmt := vested.AmountOf("ugnot")
			assert.Equal(t, tt.wantAmt, gotAmt, "vested amount mismatch")
		})
	}
}

func TestDelayedVestingAccount_LockedCoins(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, Coins{NewCoin("ugnot", 1000)}, pubKey, 0, 0)

	dva, err := NewDelayedVestingAccount(baseAcc, Coins{NewCoin("ugnot", 1000)}, 200)
	require.NoError(t, err)

	// Before cliff: all locked.
	locked := dva.LockedCoins(time.Unix(100, 0))
	assert.Equal(t, int64(1000), locked.AmountOf("ugnot"))

	// After cliff: none locked.
	locked = dva.LockedCoins(time.Unix(300, 0))
	assert.True(t, locked.IsZero(), "expected no locked coins after cliff")
}

func TestDelayedVestingAccount_GetStartTime(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, NewCoins(), pubKey, 0, 0)

	dva, err := NewDelayedVestingAccount(baseAcc, Coins{NewCoin("ugnot", 1000)}, 200)
	require.NoError(t, err)

	assert.Equal(t, int64(0), dva.GetStartTime(), "delayed vesting should have zero start time")
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
			OriginalVesting: Coins{NewCoin("ugnot", 1000)},
			EndTime:         200,
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

	cva, err := NewContinuousVestingAccount(baseAcc, Coins{NewCoin("ugnot", 500)}, 100, 200)
	require.NoError(t, err)

	// Before start: all vesting coins locked, spendable = 1000 - 500 = 500.
	spendable := SpendableCoins(cva, time.Unix(50, 0))
	assert.Equal(t, int64(500), spendable.AmountOf("ugnot"))

	// After end: no locked coins, all 1000 spendable.
	spendable = SpendableCoins(cva, time.Unix(300, 0))
	assert.Equal(t, int64(1000), spendable.AmountOf("ugnot"))

	// Halfway: 250 locked, spendable = 750.
	spendable = SpendableCoins(cva, time.Unix(150, 0))
	assert.Equal(t, int64(750), spendable.AmountOf("ugnot"))
}

func TestSpendableCoins_NoVesting(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, Coins{NewCoin("ugnot", 1000)}, pubKey, 0, 0)

	cva, err := NewContinuousVestingAccount(baseAcc, Coins{}, 100, 200)
	require.NoError(t, err)

	spendable := SpendableCoins(cva, time.Unix(50, 0))
	assert.True(t, spendable.IsEqual(Coins{NewCoin("ugnot", 1000)}),
		"all coins should be spendable when no vesting schedule")
}

func TestSpendableCoins_ZeroBalance(t *testing.T) {
	t.Parallel()

	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := pubKey.Address()
	baseAcc := NewBaseAccount(addr, NewCoins(), pubKey, 0, 0)

	cva, err := NewContinuousVestingAccount(baseAcc, Coins{NewCoin("ugnot", 500)}, 100, 200)
	require.NoError(t, err)

	spendable := SpendableCoins(cva, time.Unix(50, 0))
	assert.True(t, spendable.IsZero(), "zero balance should yield zero spendable")
}

// -----------------------------------------------------------------------------
// BaseVestingAccount validation tests

func TestBaseVestingAccount_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		endTime         int64
		originalVesting Coins
		wantErr         bool
	}{
		{"valid", 200, Coins{NewCoin("ugnot", 100)}, false},
		{"negative end time", -1, Coins{NewCoin("ugnot", 100)}, true},
		{"zero vesting is valid", 200, Coins{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bva := &BaseVestingAccount{
				OriginalVesting: tt.originalVesting,
				EndTime:         tt.endTime,
			}
			err := bva.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Interface compliance tests

func TestContinuousVestingAccount_SatisfiesVestingAccount(t *testing.T) {
	t.Parallel()

	var va VestingAccount = &ContinuousVestingAccount{}
	assert.NotNil(t, va)
}

func TestDelayedVestingAccount_SatisfiesVestingAccount(t *testing.T) {
	t.Parallel()

	var va VestingAccount = &DelayedVestingAccount{}
	assert.NotNil(t, va)
}

func TestContinuousVestingAccount_SatisfiesAccount(t *testing.T) {
	t.Parallel()

	var acc Account = &ContinuousVestingAccount{}
	assert.NotNil(t, acc)
}

func TestDelayedVestingAccount_SatisfiesAccount(t *testing.T) {
	t.Parallel()

	var acc Account = &DelayedVestingAccount{}
	assert.NotNil(t, acc)
}
