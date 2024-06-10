// Package timelock provides a library for scheduling, cancelling, and
// executing time-locked operations in Gno. It ensures that
// operations are only carried out after a specified delay and offers
// mechanisms for managing and verifying the status of these operations.
// This package leverages an AVL tree for efficient management of timestamps
// and integrates role-based access control for administrative tasks.
//
// Example Usage:
//
//  import "gno.land/p/demo/timelock"
//  import "gno.land/p/demo/accesscontrol"
//
//  Initialize timelock utility with an AVL tree and access control.
//  timestamps := avl.NewTree()
//  adminRole := accesscontrol.NewRoleData("admin", std.Address("admin-address"))
//  timelockUtil := timelock.NewTimelockUtil(timestamps, adminRole)
//
//  Schedule an operation with a delay of 60 seconds and a minimum delay of 30 seconds.
//  id := seqid.ID()
//  timelockUtil.Schedule(id, 60, 30)
//
//  Check if an operation is pending.
//  isPending := timelockUtil.IsOperationPending(id)
//
//  Execute the operation when it is ready.
//  if timelockUtil.IsOperationReady(id) {
//      timelockUtil.Execute(id)
//  }
//
//  Update the minimum delay for future operations.
//  var minDelay int64 = 30
//  timelockUtil.UpdateDelay(&minDelay, 45)

package timelock // import "gno.land/p/demo/timelock"

import (
	"std"
	"strconv"
	"time"
	"ufmt"

	"gno.land/p/demo/accesscontrol"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
)

// Marks operations completed
const DoneTimestamp = 1

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
func MustNewTimelockUtil(timestamps *avl.Tree, accessControl *accesscontrol.RoleData) (*TimelockUtil, error) {
	if timestamps == nil || accessControl == nil {
		return nil, ufmt.Errorf("timestamps and accesscontrol must be differtent to nil")
	}
	return &TimelockUtil{
		timestamps:    timestamps,
		accessControl: accessControl,
	}, nil
}

// Schedules an operation to be carried out after a minimum delay
func (tl *TimelockUtil) Schedule(id seqid.ID, delay uint64, minDelay uint64) error {
	if delay < minDelay {
		return ufmt.Errorf("timelockutil: Schedule: insufficient delay")
	}
	if tl.timestamps.Has(id.Binary()) {
		return ufmt.Errorf("timelockutil: Schedule: operation already scheduled")
	}
	tl.timestamps.Set(id.Binary(), time.Now().Unix()+delay)
	std.Emit("CallScheduled", "id", id.String(), "delay", strconv.FormatInt(delay, 10))
	return nil
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
	tl.timestamps.Set(id.Binary(), DoneTimestamp)
	std.Emit("CallExecuted", "id", id.String())
}

// Update the minimum lead time for future operations
func (tl *TimelockUtil) UpdateDelay(minDelay *uint64, newDelay uint64) {
	if std.GetOrigCaller() != tl.accessControl.AdminRole {
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
	} else if timestamp == DoneTimestamp {
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
