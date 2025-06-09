package timelock

import "errors"

var (
	ErrNilTimestampsOrAccessControl = errors.New("timelock: timestamps and accesscontrol values must be different from nil")
	ErrInsufficientDelay            = errors.New("timelockutil: Schedule: insufficient delay")
	ErrOperationAlreadyScheduled    = errors.New("timelockutil: Schedule: operation already scheduled")
	ErrOperationNotPending          = errors.New("timelock: operation not pending")
	ErrUnexpectedType               = errors.New("timelockutil: GetTimestamp: unexpected type")
	ErrUpadateDelay                 = errors.New("timelock: UpdateDelay: only admin can update delay")
	ErrOperationCancelNotPending    = errors.New("timelock: Cancel: operation not pending")
	ErrOperationExecuteNotPending   = errors.New("timelock: Execute: operation not pending")
)
