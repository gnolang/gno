package uuid

import (
	"testing"
	"time"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/uuid"
)

func TestID(t *testing.T) {
	u := uuid.NewUUID()	

	id1 := u.ID()
	id2 := u.ID()
	uuid1 := uuid.IDToUUID(id1)
	uuid2 := uuid.IDToUUID(id2)

	if uuid1 == uuid2 {
		t.Errorf("The IDs generated are identical: %d", uuid1)
	}

	if uuid2 <= uuid1 {
		t.Errorf("The second ID should be higher than the first; uuid1: %d	, uuid2: %d", uuid1, uuid2)
	}

	t.Logf("uuid1: %s , uuid2: %s", uuid1, uuid2)


	const numIDs = 11
	idMap := make(map[uint64]bool)

	for i := 0; i < numIDs; i++ {
		id := u.ID()
		if _, exists := idMap[id]; exists {
			t.Errorf("Duplicate ID found: %d", id)
		}
		idMap[id] = true
		uuidStr := uuid.IDToUUID(id)
		t.Logf("Generated UUID %d: %s", i, uuidStr)
	}

	t.Logf("setmachineID: %d", u.MachineID)
	t.Logf("Counter: %d", u.counter)

}

func TestCounter(t *testing.T) {
	u := uuid.NewUUID()
	initialCounter := u.counter

	
	for i := 0; i < 10; i++ {
		u.ID()
	}

	if u.counter != initialCounter+10 {
		t.Errorf("Counter did not increment correctly: expected %d, got %d", initialCounter+10, u.counter)
	} else {
		t.Logf("Counter incremented correctly: expected %d, got %d", initialCounter+10, u.counter)
	}

	
	u.counter = int64(uuid.MaxTimestamp) + 1
	u.ID()
	if u.counter != 0 {
		t.Errorf("Counter did not reset correctly: expected %d, got %d", 0, u.counter)
	} else {
		t.Logf("Counter reset correctly: expected %d, got %d", 0, u.counter)
	}
}