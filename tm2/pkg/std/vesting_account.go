package std

import (
	"fmt"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// VestingAccount defines an account type that vests coins via a vesting schedule.
type VestingAccount interface {
	Account

	// LockedCoins returns the set of coins that are not spendable.
	// Equivalent to GetVestingCoins(blockTime).
	LockedCoins(blockTime time.Time) Coins

	GetVestedCoins(blockTime time.Time) Coins
	GetVestingCoins(blockTime time.Time) Coins
	GetStartTime() int64
	GetEndTime() int64
	GetOriginalVesting() Coins
}

// SpendableCoins returns the total spendable coins for a vesting account.
// It is the total balance minus locked coins.
func SpendableCoins(va VestingAccount, blockTime time.Time) Coins {
	locked := va.LockedCoins(blockTime)
	balance := va.GetCoins()
	if locked.IsZero() {
		return balance
	}
	if balance.IsZero() {
		return Coins{}
	}
	result := make(Coins, 0, len(balance))
	for _, c := range balance {
		lockedAmt := locked.AmountOf(c.Denom)
		spendable := c.Amount - lockedAmt
		if spendable > 0 {
			result = append(result, Coin{c.Denom, spendable})
		}
	}
	return result
}

// -----------------------------------------------------------------------------
// BaseVestingAccount

// BaseVestingAccount provides common fields for vesting account types.
// It is embedded in concrete vesting account types.
type BaseVestingAccount struct {
	BaseAccount

	OriginalVesting Coins `json:"original_vesting" yaml:"original_vesting"`
	EndTime         int64 `json:"end_time" yaml:"end_time"`
}

// ProtoBaseVestingAccount returns a prototype for BaseVestingAccount.
func ProtoBaseVestingAccount() Account {
	return &BaseVestingAccount{}
}

// String implements fmt.Stringer.
func (bva BaseVestingAccount) String() string {
	var pubkey string

	if bva.PubKey != nil {
		pubkey = crypto.PubKeyToBech32(bva.PubKey)
	}

	return fmt.Sprintf(`VestingAccount:
  Address:          %s
  Pubkey:           %s
  Coins:            %s
  AccountNumber:    %d
  Sequence:         %d
  OriginalVesting:  %s
  EndTime:          %d`,
		bva.Address, pubkey, bva.Coins, bva.AccountNumber, bva.Sequence,
		bva.OriginalVesting, bva.EndTime,
	)
}

// GetOriginalVesting returns the original vesting amount.
func (bva BaseVestingAccount) GetOriginalVesting() Coins {
	return bva.OriginalVesting
}

// GetEndTime returns the vesting end time.
func (bva BaseVestingAccount) GetEndTime() int64 {
	return bva.EndTime
}

// Validate checks for errors on the account fields.
func (bva BaseVestingAccount) Validate() error {
	if bva.EndTime < 0 {
		return fmt.Errorf("end time cannot be negative: %d", bva.EndTime)
	}
	if !bva.OriginalVesting.IsValid() && !bva.OriginalVesting.IsZero() {
		return fmt.Errorf("invalid original vesting coins: %s", bva.OriginalVesting)
	}
	return nil
}

// -----------------------------------------------------------------------------
// ContinuousVestingAccount

// ContinuousVestingAccount implements a continuous (linear) vesting schedule.
// Coins vest linearly from StartTime to EndTime.
type ContinuousVestingAccount struct {
	BaseVestingAccount

	StartTime int64 `json:"start_time" yaml:"start_time"`
}

// NewContinuousVestingAccount creates a new ContinuousVestingAccount.
func NewContinuousVestingAccount(
	baseAcc *BaseAccount,
	originalVesting Coins,
	startTime, endTime int64,
) (*ContinuousVestingAccount, error) {
	bva := &BaseVestingAccount{
		BaseAccount:     *baseAcc,
		OriginalVesting: originalVesting,
		EndTime:         endTime,
	}

	cva := &ContinuousVestingAccount{
		BaseVestingAccount: *bva,
		StartTime:          startTime,
	}

	if err := cva.Validate(); err != nil {
		return nil, err
	}
	return cva, nil
}

// ProtoContinuousVestingAccount returns a prototype.
func ProtoContinuousVestingAccount() Account {
	return &ContinuousVestingAccount{}
}

// String implements fmt.Stringer.
func (cva ContinuousVestingAccount) String() string {
	var pubkey string

	if cva.PubKey != nil {
		pubkey = crypto.PubKeyToBech32(cva.PubKey)
	}

	return fmt.Sprintf(`ContinuousVestingAccount:
  Address:          %s
  Pubkey:           %s
  Coins:            %s
  AccountNumber:    %d
  Sequence:         %d
  OriginalVesting:  %s
  StartTime:        %d
  EndTime:          %d`,
		cva.Address, pubkey, cva.Coins, cva.AccountNumber, cva.Sequence,
		cva.OriginalVesting, cva.StartTime, cva.EndTime,
	)
}

// GetVestedCoins returns the total amount of vested coins at blockTime.
// If blockTime <= StartTime, no coins are vested.
// If blockTime >= EndTime, all original vesting coins are vested.
// Otherwise, coins vest linearly between StartTime and EndTime.
func (cva ContinuousVestingAccount) GetVestedCoins(blockTime time.Time) Coins {
	var vestedCoins Coins

	bt := blockTime.Unix()
	if bt <= cva.StartTime {
		return vestedCoins
	}
	if bt >= cva.EndTime {
		return cva.OriginalVesting
	}

	elapsed := bt - cva.StartTime
	totalDuration := cva.EndTime - cva.StartTime

	for _, ovc := range cva.OriginalVesting {
		// vestedAmt = originalAmt * elapsed / totalDuration
		product, ok := overflow.Mul(ovc.Amount, elapsed)
		if !ok {
			panic(fmt.Sprintf(
				"vesting calculation overflow: amount=%d * elapsed=%d",
				ovc.Amount, elapsed,
			))
		}
		vestedAmt := product / totalDuration
		if vestedAmt > 0 {
			vestedCoins = append(vestedCoins, Coin{ovc.Denom, vestedAmt})
		}
	}

	return vestedCoins
}

// GetVestingCoins returns the total amount of vesting coins at blockTime.
func (cva ContinuousVestingAccount) GetVestingCoins(blockTime time.Time) Coins {
	return cva.OriginalVesting.SubUnsafe(cva.GetVestedCoins(blockTime))
}

// LockedCoins returns the set of coins that are not spendable.
// Without delegation, locked coins equal vesting coins.
func (cva ContinuousVestingAccount) LockedCoins(blockTime time.Time) Coins {
	return cva.GetVestingCoins(blockTime)
}

// GetStartTime returns the vesting start time.
func (cva ContinuousVestingAccount) GetStartTime() int64 {
	return cva.StartTime
}

// Validate checks for errors on the account fields.
func (cva ContinuousVestingAccount) Validate() error {
	if cva.GetStartTime() >= cva.GetEndTime() {
		return fmt.Errorf(
			"vesting start-time (%d) must be before end-time (%d)",
			cva.StartTime, cva.EndTime,
		)
	}
	return cva.BaseVestingAccount.Validate()
}

// -----------------------------------------------------------------------------
// DelayedVestingAccount

// DelayedVestingAccount vests all coins at EndTime (cliff vesting).
// Before EndTime, no coins are vested. After EndTime, all coins are vested.
type DelayedVestingAccount struct {
	BaseVestingAccount
}

// NewDelayedVestingAccount creates a new DelayedVestingAccount.
func NewDelayedVestingAccount(
	baseAcc *BaseAccount,
	originalVesting Coins,
	endTime int64,
) (*DelayedVestingAccount, error) {
	bva := &BaseVestingAccount{
		BaseAccount:     *baseAcc,
		OriginalVesting: originalVesting,
		EndTime:         endTime,
	}

	dva := &DelayedVestingAccount{
		BaseVestingAccount: *bva,
	}

	if err := dva.Validate(); err != nil {
		return nil, err
	}
	return dva, nil
}

// ProtoDelayedVestingAccount returns a prototype.
func ProtoDelayedVestingAccount() Account {
	return &DelayedVestingAccount{}
}

// String implements fmt.Stringer.
func (dva DelayedVestingAccount) String() string {
	var pubkey string

	if dva.PubKey != nil {
		pubkey = crypto.PubKeyToBech32(dva.PubKey)
	}

	return fmt.Sprintf(`DelayedVestingAccount:
  Address:          %s
  Pubkey:           %s
  Coins:            %s
  AccountNumber:    %d
  Sequence:         %d
  OriginalVesting:  %s
  EndTime:          %d`,
		dva.Address, pubkey, dva.Coins, dva.AccountNumber, dva.Sequence,
		dva.OriginalVesting, dva.EndTime,
	)
}

// GetVestedCoins returns the total amount of vested coins at blockTime.
// All coins vest at EndTime (cliff).
func (dva DelayedVestingAccount) GetVestedCoins(blockTime time.Time) Coins {
	if blockTime.Unix() >= dva.EndTime {
		return dva.OriginalVesting
	}
	return nil
}

// GetVestingCoins returns the total amount of vesting coins at blockTime.
func (dva DelayedVestingAccount) GetVestingCoins(blockTime time.Time) Coins {
	return dva.OriginalVesting.SubUnsafe(dva.GetVestedCoins(blockTime))
}

// LockedCoins returns the set of coins that are not spendable.
// Without delegation, locked coins equal vesting coins.
func (dva DelayedVestingAccount) LockedCoins(blockTime time.Time) Coins {
	return dva.GetVestingCoins(blockTime)
}

// GetStartTime returns zero since delayed vesting has no start time.
func (dva DelayedVestingAccount) GetStartTime() int64 {
	return 0
}

// Validate checks for errors on the account fields.
func (dva DelayedVestingAccount) Validate() error {
	return dva.BaseVestingAccount.Validate()
}
