package staker

import (
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/seqid"
)

// Events
const (
	EventStake         = "Stake"
	EventBeginUnstake  = "BeginUnstake"
	EventWithdraw      = "Withdraw"
	EventMemberAdded   = "MemberAdded"
	EventMemberRemoved = "MemberRemoved"
)

// Attribute key names
const (
	EventStakerKey              = "staker"
	EventDelegateeKey           = "delegatee"
	EventAmountKey              = "amount"
	EventUnlockAtKey            = "unlock_at"
	EventUnstakeIDKey           = "unstake_id"
	EventMemberKey              = "member"
	EventTotalWithdrawnKey      = "total_withdrawn"
	EventRemainingUnstakesKey   = "remaining_unstakes"
	EventWithdrawnUnstakeIDsKey = "withdrawn_unstake_ids"
	EventCallerKey              = "caller"
	EventTimestampKey           = "timestamp"
	EventCooldownPeriodKey      = "cooldown_period"
	EventPendingUnstakesKey     = "pending_unstakes"
	EventVlsBalanceKey          = "vls_balance"
	EventXvlsBalanceKey         = "xvls_balance"
	EventUnstakeInfoKey         = "unstake_info"
)

func emitStake(caller, staker, delegatee std.Address, amount int64) {
	std.Emit(
		EventStake,
		EventCallerKey, caller.String(),
		EventStakerKey, staker.String(),
		EventDelegateeKey, delegatee.String(),
		EventAmountKey, strconv.FormatInt(amount, 10),
		EventCooldownPeriodKey, strconv.FormatInt(UnstakeLockPeriod(), 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

func emitBeginUnstake(caller, staker, delegatee std.Address, amount int64, unlockAt int64, id seqid.ID) {
	std.Emit(
		EventBeginUnstake,
		EventCallerKey, caller.String(),
		EventStakerKey, staker.String(),
		EventDelegateeKey, delegatee.String(),
		EventAmountKey, strconv.FormatInt(amount, 10),
		EventUnlockAtKey, strconv.FormatInt(unlockAt, 10),
		EventUnstakeIDKey, id.String(),
		EventCooldownPeriodKey, strconv.FormatInt(UnstakeLockPeriod(), 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

func emitWithdraw(caller, staker std.Address, totalWithdrawn int64, remainingUnstakes int, withdrawnIDs []string) {
	std.Emit(
		EventWithdraw,
		EventCallerKey, caller.String(),
		EventStakerKey, staker.String(),
		EventTotalWithdrawnKey, strconv.FormatInt(totalWithdrawn, 10),
		EventRemainingUnstakesKey, strconv.Itoa(remainingUnstakes),
		EventWithdrawnUnstakeIDsKey, strings.Join(withdrawnIDs, ","),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

func emitMemberAdded(caller, member std.Address) {
	std.Emit(
		EventMemberAdded,
		EventCallerKey, caller.String(),
		EventMemberKey, member.String(),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

func emitMemberRemoved(caller, member std.Address) {
	std.Emit(
		EventMemberRemoved,
		EventCallerKey, caller.String(),
		EventMemberKey, member.String(),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}
