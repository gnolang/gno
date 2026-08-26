package std

import (
	"fmt"
	"math/big"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// -----------------------------------------------------------------------------
// VestingSchedule

// VestingSchedule defines the parameters of a vesting schedule.
//
// Times are unix seconds. Nothing compares them against the chain's own clock,
// so a schedule that already ended when the chain starts is valid and vests
// everything immediately. That has to stay allowed: replaying an old genesis
// onto a fork is legitimate and its schedules are legitimately in the past.
type VestingSchedule struct {
	OriginalVesting Coins `json:"original_vesting" yaml:"original_vesting"`
	// StartTime is when linear vesting begins. Delayed schedules ignore it and
	// vest everything at EndTime, but Validate still requires it to be earlier
	// than EndTime for both.
	StartTime int64 `json:"start_time,omitempty" yaml:"start_time,omitempty"`
	EndTime   int64 `json:"end_time" yaml:"end_time"`
	// Type picks which account type genesis builds. Nothing reads it again
	// afterwards -- the stored Go type is what decides how coins vest -- so a
	// Type that disagrees with that type changes no behaviour and only misleads
	// whoever reads the state.
	Type VestingScheduleType `json:"type,omitempty" yaml:"type,omitempty"` // empty or "continuous" = linear; "delayed" = cliff
}

// VestingScheduleType discriminates between linear (continuous) and cliff (delayed) vesting.
type VestingScheduleType string

const (
	VestingContinuous VestingScheduleType = ""        // default — linear vesting
	VestingDelayed    VestingScheduleType = "delayed" // cliff vesting
)

// Validate checks the schedule fields.
func (vs VestingSchedule) Validate() error {
	if vs.EndTime < 0 {
		return ErrInvalidVestingSchedule(fmt.Sprintf("end time cannot be negative: %d", vs.EndTime))
	}
	if vs.StartTime >= vs.EndTime {
		return ErrInvalidVestingSchedule(fmt.Sprintf(
			"vesting start-time (%d) must be before end-time (%d)",
			vs.StartTime, vs.EndTime,
		))
	}
	if !vs.OriginalVesting.IsValid() && !vs.OriginalVesting.IsZero() {
		return ErrInvalidVestingSchedule(fmt.Sprintf("invalid original vesting coins: %s", vs.OriginalVesting))
	}
	return nil
}

// IsZero returns true if the schedule has no vesting amount.
func (vs VestingSchedule) IsZero() bool {
	return vs.OriginalVesting.IsZero()
}

// -----------------------------------------------------------------------------
// VestingAccount interface

// VestingAccount defines an account type that vests coins via a vesting schedule.
//
// This interface is the whole of the enforcement contract: the bank decides
// whether to apply a lock by type-asserting a stored account to it, so an
// account that does not implement it has no lock, whatever fields it carries.
type VestingAccount interface {
	Account

	// LockedCoins returns the set of coins that are not spendable at blockTime.
	//
	// "Not spendable" means not transferable. Locked coins can still leave the
	// account as gas fees and storage deposits, which debit through the
	// unrestricted path and never consult a schedule. So a schedule caps what a
	// holder can move to another address; it does not reserve a balance.
	LockedCoins(blockTime time.Time) Coins

	GetVestedCoins(blockTime time.Time) Coins
	GetVestingCoins(blockTime time.Time) Coins
	GetStartTime() int64
	GetEndTime() int64
	GetOriginalVesting() Coins
}

// The two account types a schedule can be enforced on. BaseVestingAccount is
// deliberately absent; see its own comment.
var (
	_ VestingAccount = (*ContinuousVestingAccount)(nil)
	_ VestingAccount = (*DelayedVestingAccount)(nil)
)

// SpendableCoins returns the account's unlocked coins at blockTime, computed as
// its own Coins minus LockedCoins.
//
// Scope: this sees only the coins held in the account object, i.e. genesis
// denominations. Realm-issued balances live outside it (see
// tm2/pkg/sdk/bank/balance.go), so this under-reports for an account holding
// any. It has no production caller — the bank enforces vesting per denomination,
// reading each from whichever tier holds it — and is kept for tests and external
// callers that only deal in gas denominations.
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

// BaseVestingAccount carries the fields the two vesting account types share. It
// is embedded by both and must never be stored as an account on its own.
//
// It has no vesting maths of its own, so it does not implement VestingAccount,
// so the bank's type assertion misses it and its schedule locks nothing: a bare
// one holding a fully-locked schedule can send its whole balance. Nothing
// constructs one today -- only the two constructors below build it, always
// embedded -- but it is amino-registered, so it can be decoded from state, and
// the bank's account-tier invariant reports one if it ever appears.
//
// Giving it the three methods would be worse than leaving it out: there is no
// correct answer for a type that does not know whether it vests linearly or at
// a cliff, and supplying one would turn a reported fault into a silent guess.
type BaseVestingAccount struct {
	BaseAccount
	VestingSchedule
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
  StartTime:        %d
  EndTime:          %d`,
		bva.Address, pubkey, bva.Coins, bva.AccountNumber, bva.Sequence,
		bva.OriginalVesting, bva.StartTime, bva.EndTime,
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

// GetStartTime returns the vesting start time.
func (bva BaseVestingAccount) GetStartTime() int64 {
	return bva.StartTime
}

// -----------------------------------------------------------------------------
// ContinuousVestingAccount

// ContinuousVestingAccount implements a continuous (linear) vesting schedule.
// Coins vest linearly from StartTime to EndTime.
type ContinuousVestingAccount struct {
	BaseVestingAccount
}

// NewContinuousVestingAccount creates a new ContinuousVestingAccount.
func NewContinuousVestingAccount(
	baseAcc *BaseAccount,
	schedule VestingSchedule,
) (*ContinuousVestingAccount, error) {
	if !baseAcc.Coins.IsAllGTE(schedule.OriginalVesting) {
		return nil, fmt.Errorf(
			"original vesting (%s) exceeds account balance (%s)",
			schedule.OriginalVesting, baseAcc.Coins,
		)
	}

	cva := &ContinuousVestingAccount{
		BaseVestingAccount: BaseVestingAccount{
			BaseAccount:     *baseAcc,
			VestingSchedule: schedule,
		},
	}

	if err := schedule.Validate(); err != nil {
		return nil, err
	}
	return cva, nil
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
		amount := big.NewInt(ovc.Amount)
		product := new(big.Int).Mul(amount, big.NewInt(elapsed))
		vestedAmt := new(big.Int).Div(product, big.NewInt(totalDuration)).Int64()
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
func (cva ContinuousVestingAccount) LockedCoins(blockTime time.Time) Coins {
	return cva.GetVestingCoins(blockTime)
}

// -----------------------------------------------------------------------------
// DelayedVestingAccount

// DelayedVestingAccount vests all coins at EndTime (cliff vesting).
type DelayedVestingAccount struct {
	BaseVestingAccount
}

// NewDelayedVestingAccount creates a new DelayedVestingAccount.
func NewDelayedVestingAccount(
	baseAcc *BaseAccount,
	schedule VestingSchedule,
) (*DelayedVestingAccount, error) {
	if !baseAcc.Coins.IsAllGTE(schedule.OriginalVesting) {
		return nil, fmt.Errorf(
			"original vesting (%s) exceeds account balance (%s)",
			schedule.OriginalVesting, baseAcc.Coins,
		)
	}

	dva := &DelayedVestingAccount{
		BaseVestingAccount: BaseVestingAccount{
			BaseAccount:     *baseAcc,
			VestingSchedule: schedule,
		},
	}

	if err := schedule.Validate(); err != nil {
		return nil, err
	}
	return dva, nil
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
func (dva DelayedVestingAccount) LockedCoins(blockTime time.Time) Coins {
	return dva.GetVestingCoins(blockTime)
}

// GetStartTime returns zero: delayed vesting has no start time, it vests at the
// cliff. This shadows the embedded schedule's StartTime, which genesis still
// has to supply and Validate still checks, so the two disagree whenever one was
// configured. Read the field, not this, when you want what was configured.
func (dva DelayedVestingAccount) GetStartTime() int64 {
	return 0
}
