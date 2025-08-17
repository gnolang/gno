package vls

import (
	"std"
	"strconv"
	"time"
)

const (
	EventMint              = "Mint"
	EventBurn              = "Burn"
	EventTransfer          = "Transfer"
	EventApproval          = "Approval"
	EventGovernanceUpdated = "GovernanceUpdated"
)

const (
	EventFromKey      = "from"
	EventToKey        = "to"
	EventAmountKey    = "amount"
	EventOwnerKey     = "owner"
	EventSpenderKey   = "spender"
	EventValueKey     = "value"
	EventOldPkgPathKey = "old_pkg_path"
	EventNewPkgPathKey = "new_pkg_path"
	EventCallerKey    = "caller"
	EventTimestampKey = "timestamp"
	EventTotalSupplyKey = "total_supply"
	EventBalanceKey   = "balance"
)

// emitMint emits an event when VLS tokens are minted
func emitMint(caller, to std.Address, amount int64) {
	std.Emit(
		EventMint,
		EventCallerKey, caller.String(),
		EventToKey, to.String(),
		EventAmountKey, strconv.FormatInt(amount, 10),
		EventTotalSupplyKey, strconv.FormatInt(TotalSupply(), 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// emitBurn emits an event when VLS tokens are burned
func emitBurn(caller, from std.Address, amount int64) {
	std.Emit(
		EventBurn,
		EventCallerKey, caller.String(),
		EventFromKey, from.String(),
		EventAmountKey, strconv.FormatInt(amount, 10),
		EventTotalSupplyKey, strconv.FormatInt(TotalSupply(), 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// emitTransfer emits an event when VLS tokens are transferred
func emitTransfer(caller, from, to std.Address, amount int64) {
	std.Emit(
		EventTransfer,
		EventCallerKey, caller.String(),
		EventFromKey, from.String(),
		EventToKey, to.String(),
		EventAmountKey, strconv.FormatInt(amount, 10),
		EventBalanceKey, strconv.FormatInt(BalanceOf(from), 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// emitApproval emits an event when VLS token allowance is approved
func emitApproval(owner, spender std.Address, value int64) {
	std.Emit(
		EventApproval,
		EventOwnerKey, owner.String(),
		EventSpenderKey, spender.String(),
		EventValueKey, strconv.FormatInt(value, 10),
		EventBalanceKey, strconv.FormatInt(BalanceOf(owner), 10),
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}

// emitGovernanceUpdated emits an event when the governance contract is updated
func emitGovernanceUpdated(caller std.Address, oldPkgPath, newPkgPath string) {
	std.Emit(
		EventGovernanceUpdated,
		EventCallerKey, caller.String(),
		EventOldPkgPathKey, oldPkgPath,
		EventNewPkgPathKey, newPkgPath,
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}
