package staker

import (
	"std"
	"strconv"
	"time"
)

const (
	EventStake         = "Stake"
	EventBeginUnstake  = "BeginUnstake"
	EventWithdraw      = "Withdraw"
	EventMemberAdded   = "MemberAdded"
	EventMemberRemoved = "MemberRemoved"
)

const (
	EventStakerKey     = "staker"
	EventDelegateeKey  = "delegatee"
	EventAmountKey     = "amount"
	EventUnlockAtKey   = "unlock_at"
	EventMemberKey     = "member"
	EventTotalWithdrawnKey = "total_withdrawn"
	EventRemainingUnstakesKey = "remaining_unstakes"
	EventCallerKey     = "caller"
	EventTimestampKey  = "timestamp"
	EventCooldownPeriodKey = "cooldown_period"
	EventPendingUnstakesKey = "pending_unstakes"
	EventVlsBalanceKey = "vls_balance"
	EventXvlsBalanceKey = "xvls_balance"
	EventUnstakeInfoKey = "unstake_info"
)

// emitStake emits an event when VLS tokens are staked
func emitStake(caller, staker, delegatee std.Address, amount int64) {
	std.Emit(
		EventStake,
		EventCallerKey, caller.String(),
		EventStakerKey, staker.String(),
		EventDelegateeKey, delegatee.String(),
		EventAmountKey, strconv.FormatInt(amount, 10),
		EventCooldownPeriodKey, strconv.FormatInt(UnstakeLockPeriod, 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// emitBeginUnstake emits an event when unstaking begins
func emitBeginUnstake(caller, staker, delegatee std.Address, amount int64, unlockAt int64) {
	std.Emit(
		EventBeginUnstake,
		EventCallerKey, caller.String(),
		EventStakerKey, staker.String(),
		EventDelegateeKey, delegatee.String(),
		EventAmountKey, strconv.FormatInt(amount, 10),
		EventUnlockAtKey, strconv.FormatInt(unlockAt, 10),
		EventCooldownPeriodKey, strconv.FormatInt(UnstakeLockPeriod, 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// emitWithdraw emits an event when unstaked VLS tokens are withdrawn
func emitWithdraw(caller, staker std.Address, totalWithdrawn int64, remainingUnstakes int) {
	std.Emit(
		EventWithdraw,
		EventCallerKey, caller.String(),
		EventStakerKey, staker.String(),
		EventTotalWithdrawnKey, strconv.FormatInt(totalWithdrawn, 10),
		EventRemainingUnstakesKey, strconv.Itoa(remainingUnstakes),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// emitMemberAdded emits an event when a member is added to governance
func emitMemberAdded(caller, member std.Address) {
	std.Emit(
		EventMemberAdded,
		EventCallerKey, caller.String(),
		EventMemberKey, member.String(),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// emitMemberRemoved emits an event when a member is removed from governance
func emitMemberRemoved(caller, member std.Address) {
	std.Emit(
		EventMemberRemoved,
		EventCallerKey, caller.String(),
		EventMemberKey, member.String(),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
} 
