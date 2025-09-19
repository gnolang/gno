// Package timelock provides a library for scheduling, cancelling, and
// executing time-locked operations in Gno. It ensures that
// operations are only carried out after a specified delay and offers
// mechanisms for managing and verifying the status of these operations.
// This package leverages an AVL tree for efficient management of timestamps
// and integrates role-based access control for administrative tasks.
//
// # Usage:
//
//	import "gno.land/p/demo/timelock"
//	import "gno.land/p/demo/accesscontrol"
//
//	Initialize timelock utility with an AVL tree and access control.
//	timestamps := avl.NewTree()
//	adminRole := accesscontrol.NewRole("admin", address("admin-address"))
//	timeLockUtil := timelock.NewTimeLockUtil(timestamps, adminRole, 30)
//
//	Schedule an operation with a delay of 60 seconds.
//	id := seqid.ID()
//	timeLockUtil.Schedule(id, 60)
//
//	Check if an operation is pending.
//	isPending := timeLockUtil.IsPending(id)
//
//	Execute the operation when it is pending.
//	if timeLockUtil.IsPending(id) {
//	    timeLockUtil.Execute(id)
//	}
//
//	Update the minimum delay for future operations.
//	timeLockUtil.UpdateDelay(45)
package timelock
