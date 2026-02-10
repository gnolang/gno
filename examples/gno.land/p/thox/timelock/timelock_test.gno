package timelock

import (
	"chain/runtime"
	"testing"
	"time"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/seqid"
	"gno.land/p/nt/uassert"
	"gno.land/p/thox/accesscontrol"
)

func TestTimelock(t *testing.T) {
	// Initialization
	timestamps := avl.NewTree()
	minDelay := uint64(2) // 2 seconds to simplify testing
	accessControl, _ := accesscontrol.NewRole("admin", runtime.OriginCaller())
	timelockUtil, err := NewTimeLock(timestamps, accessControl, minDelay)

	// Generate a new ID from time.Now().UnixNano() with seconds added to guarantee uniqueness
	newID := func(offset int64) seqid.ID {
		return seqid.ID(time.Now().UnixNano() + offset)
	}

	uassert.NoError(t, err, "Failed to create TimeLock instance")

	// Test Schedule
	t.Run("Schedule", func(t *testing.T) {
		id := newID(0)
		delay := uint64(3) // 3 seconds

		err := timelockUtil.Schedule(id, delay)

		uassert.NoError(t, err, "Schedule failed")

		status, err := timelockUtil.GetOperationStatus(id)

		uassert.NoError(t, err, "failed to get operation status")
		uassert.NotEmpty(t, status.sheduleTime, "operation status not set or invalid")
	})

	// Test Cancel
	t.Run("Cancel", func(t *testing.T) {
		id := newID(1)

		// Plan a new operation to ensure it is unique
		err := timelockUtil.Schedule(id, uint64(3))
		uassert.NoError(t, err, "Failed to schedule operation for cancellation")

		err = timelockUtil.Cancel(id)
		uassert.NoError(t, err, "Cancel failed")

		status, err := timelockUtil.GetOperationStatus(id)
		uassert.NoError(t, err, "failed to get operation status")
		uassert.Empty(t, status.sheduleTime, "operation not cancelled")
	})

	// Test Execute
	t.Run("Execute", func(t *testing.T) {
		id := newID(2)
		delay := uint64(3) // 3 seconds
		futureTime := time.Now().Unix() + int64(delay)

		// Schedule the operation with a future timestamp
		err := timelockUtil.Schedule(id, delay)
		uassert.NoError(t, err, "Failed to schedule operation for execution")

		// Simulates the passage of time by setting the timestamp to a future time
		timestamps.Set(id.Binary(), OperationStatus{sheduleTime: futureTime, isDone: false})

		err = timelockUtil.Execute(id)
		uassert.NoError(t, err, "Execute failed")

		state, err := timelockUtil.GetOperationState(id)
		uassert.NoError(t, err, "failed to get operation state")
		uassert.Equal(t, Done.StateToString(), state.StateToString(), "operation not executed")
	})

	// Test UpdateDelay
	t.Run("UpdateDelay", func(t *testing.T) {
		newDelay := uint64(4) // 4 seconds

		err := timelockUtil.UpdateDelay(newDelay)
		uassert.NoError(t, err, "UpdateDelay failed")

		uassert.Equal(t, newDelay, timelockUtil.minDelay, "minDelay not updated")
	})
}
