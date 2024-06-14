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
//  timeLockUtil := timelock.NewTimeLockUtil(timestamps, adminRole, 30)
//
//  Schedule an operation with a delay of 60 seconds.
//  id := seqid.ID()
//  timeLockUtil.Schedule(id, 60)
//
//  Check if an operation is pending.
//  isPending := timeLockUtil.IsPending(id)
//
//  Execute the operation when it is pending.
//  if timeLockUtil.IsPending(id) {
//      timeLockUtil.Execute(id)
//  }
//
//  Update the minimum delay for future operations.
//  timeLockUtil.UpdateDelay(45)

package timelock // import "gno.land/p/demo/timelock"

import (
	"std"
	"strconv"
	"time"
	"errors"

	"gno.land/p/demo/ufmt"

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
	Pending
	Ready
	Done
)

// TimeLockUtil stores the necessary parameters for the timelock operations
type TimeLockUtil struct {
	timestamps    *avl.Tree
	accessControl *accesscontrol.RoleData
	minDelay      uint64
}

// New instance of TimeLockUtil
func NewTimeLockUtil(timestamps *avl.Tree, accessControl *accesscontrol.RoleData, minDelay uint64) (*TimeLockUtil, error) {
	if timestamps == nil || accessControl == nil {
		return nil, errors.New("timestamps and accesscontrol values must be different from nil")
	}
	return &TimeLockUtil{
		timestamps:    timestamps,
		accessControl: accessControl,
		minDelay:      minDelay,
	}, nil
}

// Schedules an operation to be carried out after a minimum delay
func (tl *TimeLockUtil) Schedule(id seqid.ID, delay uint64) error {
	if delay < tl.minDelay {
		return errors.New("timelockutil: Schedule: insufficient delay")
	}
	if tl.timestamps.Has(id.Binary()) {
		return errors.New("timelockutil: Schedule: operation already scheduled")
	}
	tl.timestamps.Set(id.Binary(), uint64(time.Now().Unix())+delay)
	std.Emit("CallScheduled", "id", id.String(), "delay", strconv.FormatInt(int64(delay), 10))
	return nil
}

// Remove operation
func (tl *TimeLockUtil) Remove(id seqid.ID) {
	tl.timestamps.Remove(id.Binary())
	std.Emit("Removed", "id", id.String())
}

// Cancels a planned operation
func (tl *TimeLockUtil) Cancel(id seqid.ID) error {
	if !tl.IsPending(id) {
		errors.New("timelock: Cancel: operation not pending")
	}
	tl.timestamps.Remove(id.Binary())
	std.Emit("Cancelled", "id", id.String())
	return nil
}

// Executes a pending operation
func (tl *TimeLockUtil) Execute(id seqid.ID) error {
	if !tl.IsPending(id) {
		errors.New("timelock: Execute: operation not pending")
	}
	tl.timestamps.Set(id.Binary(), DoneTimestamp)
	std.Emit("CallExecuted", "id", id.String())
	return nil
}

// Update the minimum lead time for future operations
func (tl *TimeLockUtil) UpdateDelay(newDelay uint64) error {
	if std.GetOrigCaller() != tl.accessControl.AdminRole {
		return errors.New("timelock: UpdateDelay: only admin can update delay")
	}
	std.Emit("MinDelayChange", "oldDelay", strconv.FormatInt(int64(tl.minDelay), 10), "newDelay", strconv.FormatInt(int64(newDelay), 10))
	tl.minDelay = newDelay
	return nil
}

// Checks if an operation is pending
func (tl *TimeLockUtil) IsPending(id seqid.ID) bool {
	state, err := tl.GetOperationState(id)
	if err != nil {
		// Handle the error appropriately; for now, we assume the operation is not pending if there's an error
		ufmt.Errorf("Error retrieving operation state: %v", err)
		return false
	}
	return state == Pending
}

// Checks if an operation is ready
func (tl *TimeLockUtil) IsReady(id seqid.ID) bool {
	state, err := tl.GetOperationState(id)
	if err != nil {
		// Handle the error appropriately; for now, we assume the operation is not pending if there's an error
		ufmt.Errorf("Error retrieving operation state: %v", err)
		return false
	}
	return state == Ready
}

// Returns the status of an operation
func (tl *TimeLockUtil) GetOperationState(id seqid.ID) (OperationState, error) {
	timestamp, err := tl.GetTimestamp(id)
	if err != nil {
		return Unset, err
	}
	if timestamp == 0 {
		return Unset, nil
	} else if timestamp == DoneTimestamp {
		return Done, nil
	} else if timestamp > time.Now().Unix() {
		return Pending, nil
	} else {
		return Ready, nil
	}
}

// Returns the timestamp of an operation
func (tl *TimeLockUtil) GetTimestamp(id seqid.ID) (int64, error) {
	value, ok := tl.timestamps.Get(id.Binary())
	if !ok {
		return 0, nil // Returning 0 as the timestamp and nil for no error
	}
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	default:
		return 0, errors.New("timelockutil: GetTimestamp: unexpected type")
	}
}
