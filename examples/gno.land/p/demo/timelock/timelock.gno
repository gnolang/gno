package timelock

import (
    "strconv"
    "std"
    "gno.land/p/demo/avl"
    "gno.land/p/demo/accesscontrol"
    "gno.land/p/demo/seqid"
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

// Planifie une opération à exécuter après un délai minimum
func (tl *TimelockUtil) Schedule(timestamps *avl.Tree, id seqid.ID, delay int64, minDelay int64, accessControl *accesscontrol.RoleData) {
    if delay < minDelay {
        panic("timelockutil: schedule: insufficient delay")
    }
    if timestamps.Has(id.Binary()) {
        panic("timelockutil: schedule: operation already scheduled")
    }
    timestamps.Set(id.Binary(), time.Now().Unix()+delay)
    std.Emit("CallScheduled", "id", id.String(), "delay", strconv.FormatInt(delay, 10))
}

// Annule une opération planifiée
func (tl *TimelockUtil) Cancel(timestamps *avl.Tree, id seqid.ID, accessControl *accesscontrol.RoleData) {
    if !tl.isOperationPending(timestamps, id) {
        panic("timelock: cancel: operation not pending")
    }
    timestamps.Remove(id.Binary())
    std.Emit("Cancelled", "id", id.String())
}

// Exécute une opération prête
func (tl *TimelockUtil) Execute(timestamps *avl.Tree, id seqid.ID, accessControl *accesscontrol.RoleData) {
    if !tl.isOperationReady(timestamps, id) {
        panic("timelock: execute: operation not ready")
    }
    timestamps.Set(id.Binary(), DONE_TIMESTAMP)
    std.Emit("CallExecuted", "id", id.String())
}

// Met à jour le délai minimum pour les opérations futures
func (tl *TimelockUtil) UpdateDelay(minDelay *int64, newDelay int64, accessControl *accesscontrol.RoleData) {
    if std.GetOrigCaller() != accessControl.GetRoleAdmin() {
        panic("timelock: updatedelay: only admin can update delay")
    }
    std.Emit("MinDelayChange", "oldDelay", strconv.FormatInt(*minDelay, 10), "newDelay", strconv.FormatInt(newDelay, 10))
    *minDelay = newDelay
}

// Vérifie si une opération est en attente
func (tl *TimelockUtil) isOperationPending(timestamps *avl.Tree, id seqid.ID) bool {
    state := tl.GetOperationState(timestamps, id)
    return state == Waiting || state == Ready
}

// Vérifie si une opération est prête
func (tl *TimelockUtil) isOperationReady(timestamps *avl.Tree, id seqid.ID) bool {
    return tl.GetOperationState(timestamps, id) == Ready
}

// Retourne l'état d'une opération
func (tl *TimelockUtil) GetOperationState(timestamps *avl.Tree, id seqid.ID) OperationState {
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

// Retourne le timestamp d'une opération
func (tl *TimelockUtil) getTimestamp(timestamps *avl.Tree, id seqid.ID) int64 {
    value, ok := timestamps.Get(id.Binary())
    if !ok {
        return 0
    }
    switch v := value.(type) {
    case int:
        return int64(v)
    case int64:
        return v
    default:
        panic("timelockutil: getTimestamp: unexpected type")
    }
}