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
    // Initialisation
    timestamps := avl.NewTree()
    minDelay := int64(2) // 2 secondes pour simplifier les tests
    accessControl := accesscontrol.NewRoleData(std.GetOrigCaller())
    timelockUtil := NewTimelockUtil()

    // Génération d'un nouvel ID à partir de time.Now().UnixNano() avec ajout de secondes pour garantir l'unicité
    newID := func(offset int64) seqid.ID {
        return seqid.ID(time.Now().UnixNano() + offset)
    }

    // Test Schedule
    t.Run("Schedule", func(t *testing.T) {
        id := newID(0) // Génération d'un nouvel ID unique
        delay := int64(3) // 3 secondes

        defer func() {
            if r := recover(); r != nil {
                t.Errorf("Schedule panicked: %v", r)
            }
        }()
        timelockUtil.Schedule(timestamps, id, delay, minDelay, accessControl)

        if !timestamps.Has(id.Binary()) {
            t.Errorf("Schedule failed: timestamp not set")
        }
    })

    // Test Cancel
    t.Run("Cancel", func(t *testing.T) {
        id := newID(1) // Génération d'un nouvel ID unique avec un offset

        defer func() {
            if r := recover(); r != nil {
                t.Errorf("Cancel panicked: %v", r)
            }
        }()
        // Planifie une nouvelle opération pour garantir qu'elle est unique
        timelockUtil.Schedule(timestamps, id, int64(3), minDelay, accessControl)
        timelockUtil.Cancel(timestamps, id, accessControl)

        if timestamps.Has(id.Binary()) {
            t.Errorf("Cancel failed: timestamp not removed")
        }
    })

    // Test Execute
    t.Run("Execute", func(t *testing.T) {
        id := newID(2) // Génération d'un nouvel ID unique avec un offset
        delay := int64(3) // 3 secondes
        futureTime := time.Now().Unix() + delay

        // Schedule the operation with future timestamp
        timelockUtil.Schedule(timestamps, id, delay, minDelay, accessControl)

        // Simulate time passing by directly setting the timestamp to the past
        timestamps.Set(id.Binary(), futureTime-4)

        defer func() {
            if r := recover(); r != nil {
                t.Errorf("Execute panicked: %v", r)
            }
        }()
        timelockUtil.Execute(timestamps, id, accessControl)

        if state := timelockUtil.GetOperationState(timestamps, id); state != Done {
            t.Errorf("Execute failed: state is %v, expected Done", state)
        }
    })

    // Test UpdateDelay
    t.Run("UpdateDelay", func(t *testing.T) {
        newDelay := int64(4) // 4 secondes

        defer func() {
            if r := recover(); r != nil {
                t.Errorf("UpdateDelay panicked: %v", r)
            }
        }()
        timelockUtil.UpdateDelay(&minDelay, newDelay, accessControl)

        if minDelay != newDelay {
            t.Errorf("UpdateDelay failed: minDelay is %v, expected %v", minDelay, newDelay)
        }
    })
}
