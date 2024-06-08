package timelock

import (
    "strconv"
    "std"
    "testing"
    "gno.land/p/demo/avl"
    "gno.land/p/demo/accesscontrol"
    "gno.land/p/demo/seqid"
    "time"
)

func TestTimelockUtil(t *testing.T) {
    // Initialization
    timestamps := avl.NewTree()
    minDelay := int64(2) // 2 secondes pour simplifier les tests
    accessControl := accesscontrol.NewRoleData("admin", std.GetOrigCaller())
    timelockUtil := NewTimelockUtil(timestamps, accessControl)

    // Generate a new ID from time.Now().UnixNano() with seconds added to guarantee uniqueness
    newID := func(offset int64) seqid.ID {
        return seqid.ID(time.Now().UnixNano() + offset)
    }

    // Test Schedule
    t.Run("Schedule", func(t *testing.T) {
        id := newID(0)
        delay := int64(3) // 3 seconds

        defer func() {
            if r := recover(); r != nil {
                t.Errorf("Schedule panicked: %v", r)
            }
        }()
        timelockUtil.Schedule(id, delay, minDelay)

        if !timestamps.Has(id.Binary()) {
            t.Errorf("Schedule failed: timestamp not set")
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
        timelockUtil.Schedule(id, int64(3), minDelay)
        timelockUtil.Cancel(id)

        if timestamps.Has(id.Binary()) {
            t.Errorf("Cancel failed: timestamp not removed")
        }
    })

    // Test Execute
    t.Run("Execute", func(t *testing.T) {
        id := newID(2)
        delay := int64(3) // 3 secondes
        futureTime := time.Now().Unix() + delay

        // Schedule the operation with a future timestamp
        timelockUtil.Schedule(id, delay, minDelay)

        // Simulates the passage of time by setting the timestamp directly in the past
        timestamps.Set(id.Binary(), futureTime-4)

        defer func() {
            if r := recover(); r != nil {
                t.Errorf("Execute panicked: %v", r)
            }
        }()
        timelockUtil.Execute(id)

        if state := timelockUtil.GetOperationState(id); state != Done {
            t.Errorf("Execute failed: state is %v, expected Done", state)
        }
    })

    // Test UpdateDelay
    t.Run("UpdateDelay", func(t *testing.T) {
        newDelay := int64(4) // 4 seconds

        defer func() {
            if r := recover(); r != nil {
                t.Errorf("UpdateDelay panicked: %v", r)
            }
        }()
        timelockUtil.UpdateDelay(&minDelay, newDelay)

        if minDelay != newDelay {
            t.Errorf("UpdateDelay failed: minDelay is %v, expected %v", minDelay, newDelay)
        }
    })
}
