package timelock

import (
	"chain"
	"chain/runtime"
	"strconv"
	"time"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/thox/accesscontrol"
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

// TimeLock stores the necessary parameters for the timelock operations
type TimeLock struct {
	timestamps    *avl.Tree // id -> time.Time
	accessControl *accesscontrol.Role
	minDelay      uint64
}

// New instance of TimeLock
func NewTimeLock(timestamps *avl.Tree, accessControl *accesscontrol.Role, minDelay uint64) (*TimeLock, error) {
	if timestamps == nil || accessControl == nil {
		return nil, ErrNilTimestampsOrAccessControl
	}

	return &TimeLock{
		timestamps:    timestamps,
		accessControl: accessControl,
		minDelay:      minDelay,
	}, nil
}

// Schedules an operation to be carried out after a minimum delay
func (tl *TimeLock) Schedule(id seqid.ID, delay uint64) error {
	if delay < tl.minDelay {
		return ErrInsufficientDelay
	}

	if tl.timestamps.Has(id.Binary()) {
		return ErrOperationAlreadyScheduled
	}

	timestamp := time.Now().Unix() + int64(delay)
	status := OperationStatus{sheduleTime: timestamp, isDone: false}
	tl.timestamps.Set(id.Binary(), status)

	chain.Emit(
		"TimeLockScheduled",
		"id", id.String(),
		"delay", strconv.FormatInt(int64(delay), 10),
	)

	return nil
}

// Remove operation
func (tl *TimeLock) Remove(id seqid.ID) {
	tl.timestamps.Remove(id.Binary())

	chain.Emit(
		"TimeLockRemoved",
		"id", id.String(),
	)
}

// Cancels a planned operation
func (tl *TimeLock) Cancel(id seqid.ID) error {
	if !tl.IsPending(id) {
		return ErrOperationCancelNotPending
	}

	tl.timestamps.Remove(id.Binary())

	chain.Emit(
		"TimeLockCancelled",
		"id", id.String(),
	)
	return nil
}

// Executes a pending operation
func (tl *TimeLock) Execute(id seqid.ID) error {
	if !tl.IsPending(id) {
		return ErrOperationExecuteNotPending
	}

	status, err := tl.GetOperationStatus(id)
	if err != nil {
		return err
	}
	status.isDone = true
	tl.timestamps.Set(id.Binary(), status)

	chain.Emit(
		"TimeLockExecuted",
		"id", id.String(),
	)

	return nil
}

// Update the minimum lead time for future operations
func (tl *TimeLock) UpdateDelay(newDelay uint64) error {
	if runtime.PreviousRealm().Address() != tl.accessControl.Ownable.Owner() {
		return ErrUpadateDelay
	}

	chain.Emit(
		"TimeLockMinDelayChanged",
		"oldDelay", strconv.FormatInt(int64(tl.minDelay), 10),
		"newDelay", strconv.FormatInt(int64(newDelay), 10),
	)

	tl.minDelay = newDelay

	return nil
}

// Checks if an operation is pending
func (tl *TimeLock) IsPending(id seqid.ID) bool {
	state, err := tl.GetOperationState(id)
	if err != nil {
		// Handle the error appropriately; for now, we assume the operation is not pending if there's an error
		ufmt.Errorf("Error retrieving operation state: %v", err)
		return false
	}

	return state == Pending
}

// Checks if an operation is ready
func (tl *TimeLock) IsReady(id seqid.ID) bool {
	state, err := tl.GetOperationState(id)
	if err != nil {
		// Handle the error appropriately; for now, we assume the operation is not pending if there's an error
		ufmt.Errorf("Error retrieving operation state: %v", err)
		return false
	}

	return state == Ready
}

// Returns the status of an operation
func (tl *TimeLock) GetOperationState(id seqid.ID) (OperationState, error) {
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
func (tl *TimeLock) GetOperationStatus(id seqid.ID) (OperationStatus, error) {
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
