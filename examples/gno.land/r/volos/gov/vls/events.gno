package vls

import (
	"std"
	"strconv"
	"time"
)

// Events
const (
	EventMint              = "Mint"
	EventBurn              = "Burn"
	EventTransfer          = "Transfer"
	EventApproval          = "Approval"
	EventGovernanceUpdated = "GovernanceUpdated"
)

// Attribute key names
const (
	EventFromKey        = "from"
	EventToKey          = "to"
	EventAmountKey      = "amount"
	EventOwnerKey       = "owner"
	EventSpenderKey     = "spender"
	EventValueKey       = "value"
	EventOldPkgPathKey  = "old_pkg_path"
	EventNewPkgPathKey  = "new_pkg_path"
	EventCallerKey      = "caller"
	EventTimestampKey   = "timestamp"
	EventTotalSupplyKey = "total_supply"
	EventBalanceKey     = "balance"
)

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

func emitGovernanceUpdated(caller std.Address, oldPkgPath, newPkgPath string) {
	std.Emit(
		EventGovernanceUpdated,
		EventCallerKey, caller.String(),
		EventOldPkgPathKey, oldPkgPath,
		EventNewPkgPathKey, newPkgPath,
		EventTimestampKey, strconv.FormatInt(time.Now().Unix(), 10),
	)
}
