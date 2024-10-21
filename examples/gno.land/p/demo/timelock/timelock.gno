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
//  adminRole := accesscontrol.NewRole("admin", std.Address("admin-address"))
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

package timelock

import (
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/accesscontrol"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
)

// Represents the status of a planned operation
type OperationState int

const (
	Unset OperationState = iota
	Pending
	Ready
	Done
)

func (os OperationState) StateToString() string {
	switch os {
	case Unset:
		return "Unset"
	case Pending:
		return "Pending"
	case Ready:
		return "Ready"
	case Done:
		return "Done"
	default:
		return "Unknown"
	}
}

// OperationStatus represents the status of an operation
type OperationStatus struct {
	sheduleTime int64
	isDone      bool
}

// TimeLockUtil stores the necessary parameters for the timelock operations
type TimeLockUtil struct {
	timestamps    *avl.Tree // id -> time.Time
	accessControl *accesscontrol.Role
	minDelay      uint64
}

// New instance of TimeLockUtil
func NewTimeLockUtil(timestamps *avl.Tree, accessControl *accesscontrol.Role, minDelay uint64) (*TimeLockUtil, error) {
	if timestamps == nil || accessControl == nil {
		return nil, ErrNilTimestampsOrAccessControl
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
		return ErrInsufficientDelay
	}

	if tl.timestamps.Has(id.Binary()) {
		return ErrOperationAlreadyScheduled
	}

	timestamp := time.Now().Unix() + int64(delay)
	status := OperationStatus{sheduleTime: timestamp, isDone: false}
	tl.timestamps.Set(id.Binary(), status)

	std.Emit(
		"TimeLockScheduled",
		"id", id.String(),
		"delay", strconv.FormatInt(int64(delay), 10),
	)

	return nil
}

// Remove operation
func (tl *TimeLockUtil) Remove(id seqid.ID) {
	tl.timestamps.Remove(id.Binary())

	std.Emit(
		"TimeLockRemoved",
		"id", id.String(),
	)
}

// Cancels a planned operation
func (tl *TimeLockUtil) Cancel(id seqid.ID) error {
	if !tl.IsPending(id) {
		return ErrOperationCancelNotPending
	}

	tl.timestamps.Remove(id.Binary())

	std.Emit(
		"TimeLockCancelled",
		"id", id.String(),
	)
	return nil
}

// Executes a pending operation
func (tl *TimeLockUtil) Execute(id seqid.ID) error {
	if !tl.IsPending(id) {
		return ErrOperationExecuteNotPending
	}

	status, err := tl.GetOperationStatus(id)
	if err != nil {
		return err
	}
	status.isDone = true
	tl.timestamps.Set(id.Binary(), status)

	std.Emit(
		"TimeLockExecuted",
		"id", id.String(),
	)

	return nil
}

// Update the minimum lead time for future operations
func (tl *TimeLockUtil) UpdateDelay(newDelay uint64) error {
	if std.PrevRealm().Addr() != tl.accessControl.Ownable.Owner() {
		return ErrUpadateDelay
	}

	std.Emit(
		"TimeLockMinDelayChanged",
		"oldDelay", strconv.FormatInt(int64(tl.minDelay), 10),
		"newDelay", strconv.FormatInt(int64(newDelay), 10),
	)

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
	status, err := tl.GetOperationStatus(id)
	if err != nil {
		return Unset, err
	}
	if status.isDone {
		return Done, nil
	}
	if status.sheduleTime == 0 {
		return Unset, nil
	}
	if status.sheduleTime > time.Now().Unix() {
		return Pending, nil
	}
	return Ready, nil
}

// Returns the status of an operation
func (tl *TimeLockUtil) GetOperationStatus(id seqid.ID) (OperationStatus, error) {
	value, ok := tl.timestamps.Get(id.Binary())

	if !ok {
		return OperationStatus{}, nil // Return an empty status if the operation is not found
	}
	if status, ok := value.(OperationStatus); ok {
		return status, nil
	} else {
		return OperationStatus{}, ErrUnexpectedType
	}
}
