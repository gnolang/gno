package timelock

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/accesscontrol"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
)

func TestTimelockUtil(t *testing.T) {
	// Initialization
	timestamps := avl.NewTree()
	minDelay := uint64(2) // 2 seconds to simplify testing
	accessControl := accesscontrol.NewRole("admin", std.GetOrigCaller())
	timelockUtil, err := NewTimeLockUtil(timestamps, accessControl, minDelay)

	// Generate a new ID from time.Now().UnixNano() with seconds added to guarantee uniqueness
	newID := func(offset int64) seqid.ID {
		return seqid.ID(time.Now().UnixNano() + offset)
	}

	// Test Schedule
	t.Run("Schedule", func(t *testing.T) {
		id := newID(0)
		delay := uint64(3) // 3 seconds

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Schedule panicked: %v", r)
			}
		}()
		timelockUtil.Schedule(id, delay)

		if status, err := timelockUtil.GetOperationStatus(id); err != nil || status.sheduleTime == 0 {
			t.Errorf("Schedule failed: operation status not set or invalid")
		}
	})

	// Test Cancel
	t.Run("Cancel", func(t *testing.T) {
		id := newID(1)

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Cancel panicked: %v", r)
			}
		}()
		// Plan a new operation to ensure it is unique
		timelockUtil.Schedule(id, uint64(3))
		timelockUtil.Cancel(id)

		if status, err := timelockUtil.GetOperationStatus(id); err == nil && status.sheduleTime != 0 {
			t.Errorf("Cancel failed: operation status not removed")
		}
	})

	// Test Execute
	t.Run("Execute", func(t *testing.T) {
		id := newID(2)
		delay := uint64(3) // 3 seconds
		futureTime := time.Now().Unix() + int64(delay)

		// Schedule the operation with a future timestamp
		timelockUtil.Schedule(id, delay)

		// Simulates the passage of time by setting the timestamp to a future time
		timestamps.Set(id.Binary(), OperationStatus{sheduleTime: futureTime, isDone: false})

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Execute panicked: %v", r)
			}
		}()
		timelockUtil.Execute(id)

		if state, err := timelockUtil.GetOperationState(id); state != Done {
			t.Errorf("Execute failed: state is %v, expected Done", state)
		}
	})

	// Test UpdateDelay
	t.Run("UpdateDelay", func(t *testing.T) {
		newDelay := uint64(4) // 4 seconds

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("UpdateDelay panicked: %v", r)
			}
		}()
		timelockUtil.UpdateDelay(newDelay)

		if timelockUtil.minDelay != newDelay {
			t.Errorf("UpdateDelay failed: minDelay is %v, expected %v", minDelay, newDelay)
		}
	})
}
