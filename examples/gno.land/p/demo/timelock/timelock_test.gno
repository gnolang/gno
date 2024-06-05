package timelock

import (
    "std"
    "gno.land/p/demo/avl"
    "gno.land/p/demo/accesscontrol"
    "gno.land/p/timelock"
    "testing"
    "time"
)

func TestTimelockUtil(t *testing.T) {
    // Initialisation
    timestamps := avl.NewTree()
    minDelay := int64(3600) // 1 heure en secondes
    accessControl := accesscontrol.NewRoleData(std.GetOrigCaller())
    timelockUtil := timelock.NewTimelockUtil()

    // Test Schedule
    t.Run("Schedule", func(t *testing.T) {
        id := std.Hash("operation1")
        delay := int64(7200) // 2 heures

        defer func() {
            if r := recover(); r != nil {
                t.Errorf("Schedule panicked: %v", r)
            }
        }()
        timelockUtil.Schedule(timestamps, id, delay, minDelay, accessControl)

        if !timestamps.Has(id.String()) {
            t.Errorf("Schedule failed: timestamp not set")
        }
    })

    // Test Cancel
    t.Run("Cancel", func(t *testing.T) {
        id := std.Hash("operation1")

        defer func() {
            if r := recover(); r != nil {
                t.Errorf("Cancel panicked: %v", r)
            }
        }()
        timelockUtil.Cancel(timestamps, id, accessControl)

        if timestamps.Has(id.String()) {
            t.Errorf("Cancel failed: timestamp not removed")
        }
    })

    // Test Execute
    t.Run("Execute", func(t *testing.T) {
        id := std.Hash("operation2")
        delay := int64(1) // 1 second
        timelockUtil.Schedule(timestamps, id, delay, minDelay, accessControl)
        time.Sleep(2 * time.Second) // Attendre que l'opération soit prête

        defer func() {
            if r := recover(); r != nil {
                t.Errorf("Execute panicked: %v", r)
            }
        }()
        timelockUtil.Execute(timestamps, id, accessControl)

        if state := timelockUtil.getOperationState(timestamps, id); state != timelock.Done {
            t.Errorf("Execute failed: state is %v, expected Done", state)
        }
    })

    // Test UpdateDelay
    t.Run("UpdateDelay", func(t *testing.T) {
        newDelay := int64(1800) // 30 minutes

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

    // Test HashOperation
    t.Run("HashOperation", func(t *testing.T) {
        target := std.Address("target")
        value := int64(100)
        data := []byte("data")
        predecessor := std.Hash("predecessor")
        salt := std.Hash("salt")
        expectedHash := std.Hash(std.HashValue(target.String() + std.ItoA(value) + string(data) + predecessor.String() + salt.String()))

        hash := timelockUtil.HashOperation(target, value, data, predecessor, salt)
        if hash != expectedHash {
            t.Errorf("HashOperation failed: hash is %v, expected %v", hash, expectedHash)
        }
    })

    // Test HashOperationBatch
    t.Run("HashOperationBatch", func(t *testing.T) {
        targets := []std.Address{std.Address("target1"), std.Address("target2")}
        values := []int64{100, 200}
        payloads := [][]byte{[]byte("data1"), []byte("data2")}
        predecessor := std.Hash("predecessor")
        salt := std.Hash("salt")
        hashInput := targets[0].String() + std.ItoA(values[0]) + string(payloads[0]) +
            targets[1].String() + std.ItoA(values[1]) + string(payloads[1]) + predecessor.String() + salt.String()
        expectedHash := std.Hash(std.HashValue(hashInput))

        hash := timelockUtil.HashOperationBatch(targets, values, payloads, predecessor, salt)
        if hash != expectedHash {
            t.Errorf("HashOperationBatch failed: hash is %v, expected %v", hash, expectedHash)
        }
    })
}
