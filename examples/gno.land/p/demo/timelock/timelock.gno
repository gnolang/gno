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

// TimelockUtil stores the necessary parameters for the timelock operations
type TimelockUtil struct {
    timestamps    *avl.Tree
    accessControl *accesscontrol.RoleData
}

// New instance of TimelockUtil
func NewTimelockUtil(timestamps *avl.Tree, accessControl *accesscontrol.RoleData) *TimelockUtil {
    return &TimelockUtil{
        timestamps:    timestamps,
        accessControl: accessControl,
    }
}

// Schedules an operation to be carried out after a minimum delay
func (tl *TimelockUtil) Schedule(id seqid.ID, delay int64, minDelay int64) {
    if delay < minDelay {
        panic("timelockutil: Schedule: insufficient delay")
    }
    if tl.timestamps.Has(id.Binary()) {
        panic("timelockutil: Schedule: operation already scheduled")
    }
    tl.timestamps.Set(id.Binary(), time.Now().Unix()+delay)
    std.Emit("CallScheduled", "id", id.String(), "delay", strconv.FormatInt(delay, 10))
}

// Cancels a planned operation
func (tl *TimelockUtil) Cancel(id seqid.ID) {
    if !tl.IsOperationPending(id) {
        panic("timelock: Cancel: operation not pending")
    }
    tl.timestamps.Remove(id.Binary())
    std.Emit("Cancelled", "id", id.String())
}

// Executes a ready operation
func (tl *TimelockUtil) Execute(id seqid.ID) {
    if !tl.IsOperationReady(id) {
        panic("timelock: Execute: operation not ready")
    }
    tl.timestamps.Set(id.Binary(), DONE_TIMESTAMP)
    std.Emit("CallExecuted", "id", id.String())
}

// Update the minimum lead time for future operations
func (tl *TimelockUtil) UpdateDelay(minDelay *int64, newDelay int64) {
    if std.GetOrigCaller() != tl.accessControl.GetRoleAdmin() {
        panic("timelock: UpdateDelay: only admin can update delay")
    }
    std.Emit("MinDelayChange", "oldDelay", strconv.FormatInt(*minDelay, 10), "newDelay", strconv.FormatInt(newDelay, 10))
    *minDelay = newDelay
}

// Checks if an operation is pending
func (tl *TimelockUtil) IsOperationPending(id seqid.ID) bool {
    state := tl.GetOperationState(id)
    return state == Waiting || state == Ready
}

// Checks if an operation is ready
func (tl *TimelockUtil) IsOperationReady(id seqid.ID) bool {
    return tl.GetOperationState(id) == Ready
}

// Returns the status of an operation
func (tl *TimelockUtil) GetOperationState(id seqid.ID) OperationState {
    timestamp := tl.GetTimestamp(id)
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

// Returns the timestamp of an operation
func (tl *TimelockUtil) GetTimestamp(id seqid.ID) int64 {
    value, ok := tl.timestamps.Get(id.Binary())
    if !ok {
        return 0
    }
    switch v := value.(type) {
    case int:
        return int64(v)
    case int64:
        return v
    default:
        panic("timelockutil: GetTimestamp: unexpected type")
    }
}
