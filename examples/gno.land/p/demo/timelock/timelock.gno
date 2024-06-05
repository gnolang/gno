package timelock

import (
	"std"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/accesscontrol"
	"time"
)

// Marks operations completed
const DONE_TIMESTAMP = 1

// Represents the status of a planned operation
type OperationState int

const (
	Unset OperationState = iota
	Waiting
	Ready
	Done
)

// Stock all methods for the Timelock contract
type TimelockUtil struct{}

// New instance of TimelockUtil
func NewTimelockUtil() *TimelockUtil {
	return &TimelockUtil{}
}

// Planifie une operation a execuetr apres un delai minimum, verifie que le delai est superieur au delai minimum
func (tl *TimelockUtil) Schedule(timestamps *avl.Tree, id std.Hash, delay int64, minDelay int64, accessControl *accesscontrol.RoelData) {
	if deay < minDelay {
		panic("timelockutil: schedule: insufficient delay")
	}
	if timestamps.Has(id.String()) {
		panic("timelockutil: schedule: operation already scheduled")
	}
	timestamps.Set(id.String(), time.Now().Unix() + delay)
	std.Emit("CallScheduled", "id", id.String(), "delay", delay)
}

// Annule une operation planifiee et verifie que l'operation est en attente
func (tl *TimelockUtil) Cancel(timestamps *avl.Tree, id std.Hash, accessControl *accesscontrol.RoleData) {
    if !tl.isOperationPending(timestamps, id) {
        panic("timelock: cancel: operation not pending")
    }
    timestamps.Remove(id.String())
    std.Emit("Cancelled", "id", id.String())
}


// Execute une operation prete, verifie que loperation est prete
func (tl *TimelockUtil) Execute(timestamps *avl.Tree, id std.Hash, accessControl *accesscontrol.RoleData) {
    if !tl.isOperationReady(timestamps, id) {
        panic("timelock: execute:  operation not ready")
    }
    timestamps.Set(id.String(), DONE_TIMESTAMP)
    std.Emit("CallExecuted", "id", id.String())
}

// Met a jour le delai minimum pour les opertaions futures, verifie que lappliquant est ladmin
func (tl *TimelockUtil) UpdateDelay(minDelay *int64, newDelay int64, accessControl *accesscontrol.RoleData) {
    if std.GetOrigCaller() != accessControl.GetRoleAdmin() {
        panic("timelock: updatedelay: only admin can update delay")
    }
    std.Emit("MinDelayChange", "oldDelay", *minDelay, "newDelay", newDelay)
    *minDelay = newDelay
}

// Retourne le hash d'une operation unique
func (tl *TimelockUtil) HashOperation(target std.Address, value int64, data []byte, predecessor std.Hash, salt std.Hash) std.Hash {
    return std.Hash(std.HashValue(target.String() + std.ItoA(value) + string(data) + predecessor.String() + salt.String()))
}

// Retourne le hash d'un lot d'operations
func (tl *TimelockUtil) HashOperationBatch(targets []std.Address, values []int64, payloads [][]byte, predecessor std.Hash, salt std.Hash) std.Hash {
    hashInput := ""
    for i, target := range targets {
        hashInput += target.String() + std.ItoA(values[i]) + string(payloads[i])
    }
    return std.Hash(std.HashValue(hashInput + predecessor.String() + salt.String()))
}

// Verifie si une operation est en attente
func (tl *TimelockUtil) isOperationPending(timestamps *avl.Tree, id std.Hash) bool {
    state := tl.getOperationState(timestamps, id)
    return state == Waiting || state == Ready
}

// Verifie si une operation est prete
func (tl *TimelockUtil) isOperationReady(timestamps *avl.Tree, id std.Hash) bool {
    return tl.getOperationState(timestamps, id) == Ready
}

// Retourne l'etat d'une operation
func (tl *TimelockUtil) getOperationState(timestamps *avl.Tree, id std.Hash) OperationState {
    timestamp := tl.getTimestamp(timestamps, id)
    if timestamp == 0 {
        return Unset
    } else if timestamp == DONE_TIMESTAMP {
        return Done
    } else if timestamp > time.Now().Unix() {
        return Waiting
    } else {
        return Ready
    }
}

// Retourne le timestamp d'une operation
func (tl *TimelockUtil) getTimestamp(timestamps *avl.Tree, id std.Hash) int64 {
    value := timestamps.Get(id.String())
    if value == nil {
        return 0
    }
    return value.(int64)
}

